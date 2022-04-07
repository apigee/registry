// Copyright 2020 Google LLC. All Rights Reserved.
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

package bulk

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	cmd := &cobra.Command{
		Use:   "protos",
		Short: "Bulk-upload Protocol Buffer descriptions from a directory of specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			projectID, err := cmd.Flags().GetString("project-id")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get project-id from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			// create a queue for upload tasks and wait for the workers to finish after filling it.
			jobs, err := cmd.Flags().GetInt("jobs")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get jobs from flags")
			}
			taskQueue, wait := core.WorkerPool(ctx, jobs)
			defer wait()

			for _, arg := range args {
				path, err := filepath.Abs(arg)
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Invalid path")
				}

				if err := scanDirectoryForProtos(ctx, client, projectID, baseURI, path, taskQueue); err != nil {
					log.FromContext(ctx).WithError(err).Debug("Failed to walk directory")
				}
			}
		},
	}

	cmd.Flags().StringVar(&baseURI, "base-uri", "", "Prefix to use for the source_uri field of each proto upload")
	return cmd
}

func scanDirectoryForProtos(ctx context.Context, client connection.Client, projectID, baseURI, root string, taskQueue chan<- core.Task) error {
	return filepath.Walk(root, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip everything that's not a YAML file in a versioned directory.
		parent := path.Dir(filepath)
		if info.IsDir() || !strings.HasSuffix(filepath, ".yaml") || !versionDirectory.MatchString(path.Base(parent)) {
			return nil
		}

		bytes, err := ioutil.ReadFile(filepath)
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
			projectID:      projectID,
			apiID:          strings.TrimSuffix(sc.Name, ".googleapis.com"),
			apiTitle:       sc.Title,
			apiDescription: strings.ReplaceAll(sc.Documentation.Summary, "\n", " "),
			path:           parent,
			directory:      root,
		}

		// Skip the directory after we find an API service configuration.
		return fs.SkipDir
	})
}

type uploadProtoTask struct {
	client         connection.Client
	baseURI        string
	projectID      string
	path           string
	directory      string
	apiID          string
	apiTitle       string
	apiDescription string
	versionID      string // computed at runtime
	specID         string // computed at runtime
}

func (task *uploadProtoTask) String() string {
	return "upload proto " + task.path
}

func (task *uploadProtoTask) Run(ctx context.Context) error {
	// Populate API path fields using the file's path.
	task.populateFields()
	log.Infof(ctx, "Uploading apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)

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

	specPart := task.fileName()
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
		return fmt.Errorf("Failed to create %s, %s", task.apiName(), err)
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
	contents, err := task.zipContents()
	if err != nil {
		return err
	}

	// Use the spec size and hash to avoid unnecessary uploads.
	spec, err := task.client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: task.specName(),
	})

	if err == nil && int(spec.GetSizeBytes()) == len(contents) && spec.GetHash() == hashForBytes(contents) {
		log.Debugf(ctx, "Matched already uploaded spec %s", task.specName())
		return nil
	}

	request := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:     task.specName(),
			MimeType: core.ProtobufMimeType("+zip"),
			Filename: task.fileName(),
			Contents: contents,
		},
		AllowMissing: true,
	}
	if task.baseURI != "" {
		request.ApiSpec.SourceUri = fmt.Sprintf("%s/%s", task.baseURI, task.apiPath())
	}

	response, err := task.client.UpdateApiSpec(ctx, request)
	if err != nil {
		log.FromContext(ctx).WithError(err).Errorf("Error %s [contents-length: %d]", task.specName(), len(contents))
	} else {
		log.Debugf(ctx, "Updated %s", response.Name)
	}

	return nil
}

func (task *uploadProtoTask) projectName() string {
	return fmt.Sprintf("projects/%s", task.projectID)
}

func (task *uploadProtoTask) apiName() string {
	return fmt.Sprintf("%s/locations/global/apis/%s", task.projectName(), task.apiID)
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
	contents, err := core.ZipArchiveOfPath(task.path, prefix)
	if err != nil {
		return nil, err
	}

	return contents.Bytes(), nil
}
