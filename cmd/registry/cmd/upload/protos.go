// Copyright 2020 Google LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upload

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"gopkg.in/yaml.v3"
)

// The pattern of an API version directory.
var versionDirectory = regexp.MustCompile("v.*[1-9]+.*")

// The API Service Configuration contains important API properties.
type ServiceConfig struct {
	Type          string `yaml:"type"`
	Name          string `yaml:"name"`
	Title         string `yaml:"title"`
	Documentation struct {
		Summary string `yaml:"summary"`
	} `yaml:"documentation"`
}

func protosCommand() *cobra.Command {
	var baseURI string
	var root string
	var jobs int
	cmd := &cobra.Command{
		Use:   "protos DIRECTORY",
		Short: "Upload Protocol Buffer descriptions from a directory of specs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			parent, err := getParent(cmd)
			if err != nil {
				return fmt.Errorf("failed to identify parent project (%s)", err)
			}
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				return err
			}
			if err := visitor.VerifyLocation(ctx, client, parent); err != nil {
				return fmt.Errorf("parent does not exist (%s)", err)
			}
			// create a queue for upload tasks and wait for the workers to finish after filling it.
			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()

			for _, arg := range args {
				path, err := filepath.Abs(arg)
				if err != nil {
					return fmt.Errorf("invalid path: %s", err)
				}

				if root == "" {
					root = path
				}
				root, err := filepath.Abs(root)
				if err != nil {
					return fmt.Errorf("invalid path: %s", err)
				}

				if err := scanDirectoryForProtos(client, parent, baseURI, path, root, taskQueue); err != nil {
					log.FromContext(ctx).WithError(err).Debug("Failed to walk directory")
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "project ID to use for each upload (deprecated)")
	cmd.Flags().StringVar(&parent, "parent", "", "parent for the upload (projects/PROJECT/locations/LOCATION)")
	cmd.Flags().StringVar(&root, "protoc-root", "", "root directory to use for proto compilation, defaults to PATH")
	cmd.Flags().StringVar(&baseURI, "base-uri", "", "prefix to use for the source_uri field of each proto upload")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
}

func scanDirectoryForProtos(client connection.RegistryClient, parent, baseURI, start, root string, taskQueue chan<- tasks.Task) error {
	return filepath.Walk(start, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip everything that's not a YAML file in a versioned directory.
		container := path.Dir(filepath)
		if info.IsDir() || !strings.HasSuffix(filepath, ".yaml") || !versionDirectory.MatchString(path.Base(container)) {
			return nil
		}

		bytes, err := os.ReadFile(filepath)
		if err != nil {
			return err
		}

		sc := &ServiceConfig{}
		if err := yaml.Unmarshal(bytes, sc); err != nil {
			return err
		}

		// Skip invalid API service configurations.
		if sc.Type != "google.api.Service" || sc.Title == "" || sc.Name == "" {
			return nil
		}

		taskQueue <- &uploadProtoTask{
			client:         client,
			baseURI:        baseURI,
			parent:         parent,
			apiID:          strings.TrimSuffix(sc.Name, ".googleapis.com"),
			apiTitle:       sc.Title,
			apiDescription: strings.ReplaceAll(sc.Documentation.Summary, "\n", " "),
			path:           container,
			directory:      root,
		}

		// Skip the directory after we find an API service configuration.
		return fs.SkipDir
	})
}

type uploadProtoTask struct {
	client         connection.RegistryClient
	baseURI        string
	parent         string
	path           string
	directory      string
	apiID          string
	apiTitle       string
	apiDescription string
	versionID      string // computed at runtime
	specID         string // computed at runtime
	contents       []byte // computed at runtime
}

func (task *uploadProtoTask) String() string {
	return "upload proto " + task.path
}

func (task *uploadProtoTask) Run(ctx context.Context) error {
	// Populate API path fields using the file's path.
	task.populateFields()
	log.Infof(ctx, "Uploading apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)

	// Zip up the protos first; if that fails, skip the API.
	var err error
	if task.contents, err = task.zipContents(); err != nil {
		return err
	}
	// If the API does not exist, create it.
	if err := task.createAPI(ctx); err != nil {
		return err
	}
	// If the API version does not exist, create it.
	if err := task.createVersion(ctx); err != nil {
		return err
	}
	// Create or update the spec as needed.
	if err := task.createOrUpdateSpec(ctx); err != nil {
		return err
	}
	return nil
}

func (task *uploadProtoTask) populateFields() {
	parts := strings.Split(task.apiPath(), "/")

	versionPart := parts[len(parts)-1]
	task.versionID = sanitize(versionPart)

	specPart := strings.TrimSuffix(task.fileName(), ".zip")
	task.specID = sanitize(specPart)
}

func (task *uploadProtoTask) createAPI(ctx context.Context) error {
	// Create an API if needed (or update an existing one)
	response, err := task.client.UpdateApi(ctx, &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:        task.apiName(),
			DisplayName: task.apiTitle,
			Description: task.apiDescription,
		},
		AllowMissing: true,
	})
	if err == nil {
		log.Debugf(ctx, "Updated %s", response.Name)
	} else if status.Code(err) == codes.AlreadyExists {
		log.Debugf(ctx, "Found %s", task.apiName())
	} else {
		log.FromContext(ctx).WithError(err).Debugf("Failed to create API %s", task.apiName())
		// Returning this error ends all tasks, which seems appropriate to
		// handle situations where all might fail due to a common problem
		// (a missing project or incorrect project-id).
		return fmt.Errorf("failed to create %s, %s", task.apiName(), err)
	}

	return nil
}

func (task *uploadProtoTask) createVersion(ctx context.Context) error {
	// Create an API version if needed (or update an existing one)
	response, err := task.client.UpdateApiVersion(ctx, &rpc.UpdateApiVersionRequest{
		ApiVersion: &rpc.ApiVersion{
			Name: task.versionName(),
		},
		AllowMissing: true,
	})
	if err == nil {
		log.Debugf(ctx, "Updated %s", response.Name)
	} else {
		log.FromContext(ctx).WithError(err).Debugf("Failed to create version %s", task.versionName())
	}

	return nil
}

func (task *uploadProtoTask) createOrUpdateSpec(ctx context.Context) error {
	// Use the spec size and hash to avoid unnecessary uploads.
	spec, err := task.client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: task.specName(),
	})

	if err == nil && int(spec.GetSizeBytes()) == len(task.contents) && spec.GetHash() == hashForBytes(task.contents) {
		log.Debugf(ctx, "Matched already uploaded spec %s", task.specName())
		return nil
	}

	request := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:     task.specName(),
			MimeType: mime.ProtobufMimeType("+zip"),
			Filename: task.fileName(),
			Contents: task.contents,
		},
		AllowMissing: true,
	}
	if task.baseURI != "" {
		request.ApiSpec.SourceUri = fmt.Sprintf("%s/%s", task.baseURI, task.apiPath())
	}

	response, err := task.client.UpdateApiSpec(ctx, request)
	if err != nil {
		log.FromContext(ctx).WithError(err).Errorf("Error %s [contents-length: %d]", task.specName(), len(task.contents))
	} else {
		log.Debugf(ctx, "Updated %s", response.Name)
	}

	return nil
}

func (task *uploadProtoTask) apiName() string {
	return fmt.Sprintf("%s/apis/%s", task.parent, task.apiID)
}

func (task *uploadProtoTask) versionName() string {
	return fmt.Sprintf("%s/versions/%s", task.apiName(), task.versionID)
}

func (task *uploadProtoTask) specName() string {
	return fmt.Sprintf("%s/specs/%s", task.versionName(), task.specID)
}

func (task *uploadProtoTask) apiPath() string {
	prefix := task.directory + "/"
	return strings.TrimPrefix(task.path, prefix)
}

func (task *uploadProtoTask) fileName() string {
	parts := strings.Split(task.apiPath(), "/")
	return strings.Join(parts, "-") + ".zip"
}

func (task *uploadProtoTask) zipContents() ([]byte, error) {
	prefix := task.directory + "/"

	// Get the proto files in the main directory.
	protos, err := localProtos(task.path, prefix)
	if err != nil {
		return nil, err
	}

	// Compile the listed protos to get their dependencies.
	if len(protos) > 0 {
		protos, err = referencedProtos(protos, task.directory)
		if err != nil {
			return nil, err
		}
	}

	// Get the metadata files in the main directory.
	metadata, err := localMetadata(task.path, prefix)
	if err != nil {
		return nil, err
	}

	// Zip the listed files.
	contents, err := compress.ZipArchiveOfFiles(append(metadata, protos...), prefix)
	if err != nil {
		return nil, err
	}

	return contents.Bytes(), nil
}

// Collect the names of all metadata files in a source directory, stripping the prefix.
func localMetadata(source, prefix string) ([]string, error) {
	return localFiles(source, prefix, func(name string) bool {
		return strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".yaml")
	})
}

// Collect the names of all proto files in a source directory, stripping the prefix.
func localProtos(source, prefix string) ([]string, error) {
	return localFiles(source, prefix, func(name string) bool {
		return strings.HasSuffix(name, ".proto")
	})
}

// Collect the names of all matching files in a source directory, stripping the prefix.
func localFiles(source, prefix string, match func(string) bool) ([]string, error) {
	filenames := make([]string, 0)
	err := filepath.WalkDir(source, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if entry.IsDir() {
			return nil // Do nothing for the directory, but still walk its contents.
		} else if match(p) {
			filenames = append(filenames, strings.TrimPrefix(p, prefix))
		}
		return nil
	})
	return filenames, err
}

// Get all the protos that are referenced in the compilation of a list of protos.
func referencedProtos(protos []string, root string) ([]string, error) {
	tempDir, err := os.MkdirTemp("", "proto-import-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)
	args := []string{"-o", tempDir + "/proto.pb", "--include_imports"}
	args = append(args, protos...)
	cmd := exec.Command("protoc", args...)
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to compile protos with protoc: %s", err)
	}
	return protosFromFileDescriptorSet(tempDir + "/proto.pb")
}

// Get all the protos listed as dependencies in a file descriptor set.
func protosFromFileDescriptorSet(filename string) ([]string, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	fds := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(bytes, fds)
	if err != nil {
		return nil, err
	}
	filenameset := make(map[string]bool)
	for _, file := range fds.File {
		filename = *file.Name
		if !strings.HasPrefix(filename, "google/protobuf/") {
			filenameset[filename] = true
		}
	}
	filenames := make([]string, 0)
	for k := range filenameset {
		filenames = append(filenames, k)
	}
	sort.Strings(filenames)
	return filenames, nil
}
