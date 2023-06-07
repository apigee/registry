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
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	"gopkg.in/yaml.v3"
)

const openAPISpecID = "openapi"

func openAPICommand() *cobra.Command {
	var baseURI string
	var jobs int
	cmd := &cobra.Command{
		Use:   "openapi DIRECTORY",
		Short: "Upload OpenAPI descriptions from a directory of specs",
		Args:  cobra.MinimumNArgs(1),
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
				scanDirectoryForOpenAPI(ctx, client, parent, baseURI, path, taskQueue)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "project ID to use for each upload (deprecated)")
	cmd.Flags().StringVar(&parent, "parent", "", "parent for the upload (projects/PROJECT/locations/LOCATION)")
	cmd.Flags().StringVar(&baseURI, "base-uri", "", "prefix to use for the source_uri field of each spec upload")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
}

func scanDirectoryForOpenAPI(ctx context.Context, client connection.RegistryClient, parent, baseURI, directory string, taskQueue chan<- tasks.Task) {
	// walk a directory hierarchy, uploading every API spec that matches a set of expected file names.
	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		task := &uploadOpenAPITask{
			client:    client,
			parent:    parent,
			baseURI:   baseURI,
			path:      path,
			directory: directory,
		}

		switch {
		case strings.HasSuffix(path, "swagger.yaml"), strings.HasSuffix(path, "swagger.json"):
			task.version = "2"
			taskQueue <- task
		case strings.HasSuffix(path, "openapi.yaml"), strings.HasSuffix(path, "openapi.json"):
			task.version = "3"
			taskQueue <- task
		}

		return nil
	}); err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to walk directory")
	}
}

// sanitize converts a name into a "safe" form for use as an identifier
func sanitize(name string) string {
	// identifiers are lower-case
	name = strings.ToLower(name)
	// certain characters that are used as separators are replaced with dashes
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, ":", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	name = strings.Replace(name, "+", "-", -1)
	// any remaining unsupported characters are removed
	pattern := regexp.MustCompile("[^a-z0-9-.]")
	name = pattern.ReplaceAllString(name, "")
	return name
}

type uploadOpenAPITask struct {
	client    connection.RegistryClient
	baseURI   string
	path      string
	directory string
	version   string
	parent    string
	apiID     string // computed at runtime
	versionID string // computed at runtime
	contents  []byte
	document  PartialOpenAPIDocument
}

func (task *uploadOpenAPITask) String() string {
	return "upload openapi " + task.path
}

func (task *uploadOpenAPITask) Run(ctx context.Context) error {
	// Populate API path fields using the file's path.
	if err := task.populateFields(); err != nil {
		log.FromContext(ctx).WithError(err).Debugf("Failed to import API %s", task.apiName())
		return nil
	}
	log.Infof(ctx, "Uploading apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, openAPISpecID)

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

func (task *uploadOpenAPITask) populateFields() error {
	parts := strings.Split(task.apiPath(), "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid API path: %s", task.apiPath())
	}

	apiParts := parts[0 : len(parts)-2]
	apiPart := strings.ReplaceAll(strings.Join(apiParts, "-"), "/", "-")
	task.apiID = sanitize(apiPart)

	versionPart := parts[len(parts)-2]
	task.versionID = sanitize(versionPart)

	var err error
	task.contents, err = os.ReadFile(task.path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(task.contents, &(task.document))
}

func (task *uploadOpenAPITask) createAPI(ctx context.Context) error {
	// Create an API if needed (or update an existing one)
	response, err := task.client.UpdateApi(ctx, &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:        task.apiName(),
			DisplayName: task.apiID,
			Description: task.document.Info.Title,
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

func (task *uploadOpenAPITask) createVersion(ctx context.Context) error {
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

func (task *uploadOpenAPITask) createOrUpdateSpec(ctx context.Context) error {
	// Use the spec size and hash to avoid unnecessary uploads.
	spec, err := task.client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: task.specName(),
	})

	if err == nil && int(spec.GetSizeBytes()) == len(task.contents) && spec.GetHash() == hashForBytes(task.contents) {
		log.Debugf(ctx, "Matched already uploaded spec %s", task.specName())
		return nil
	}

	gzippedContents, err := compress.GZippedBytes(task.contents)
	if err != nil {
		return err
	}

	request := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:     task.specName(),
			MimeType: mime.OpenAPIMimeType("+gzip", task.version),
			Filename: task.fileName(),
			Contents: gzippedContents,
		},
		AllowMissing: true,
	}
	if task.baseURI != "" {
		request.ApiSpec.SourceUri = fmt.Sprintf("%s/%s", task.baseURI, task.apiPath())
	}

	response, err := task.client.UpdateApiSpec(ctx, request)
	if err != nil {
		log.FromContext(ctx).WithError(err).Debugf("Error %s [contents-length: %d]", task.specName(), len(task.contents))
	} else {
		log.Debugf(ctx, "Updated %s", response.Name)
	}

	return nil
}

func (task *uploadOpenAPITask) apiName() string {
	return fmt.Sprintf("%s/apis/%s", task.parent, task.apiID)
}

func (task *uploadOpenAPITask) versionName() string {
	return fmt.Sprintf("%s/versions/%s", task.apiName(), task.versionID)
}

func (task *uploadOpenAPITask) specName() string {
	return fmt.Sprintf("%s/specs/%s", task.versionName(), openAPISpecID)
}

func (task *uploadOpenAPITask) apiPath() string {
	prefix := task.directory + "/"
	return strings.TrimPrefix(task.path, prefix)
}

func (task *uploadOpenAPITask) fileName() string {
	return filepath.Base(task.path)
}

func hashForBytes(b []byte) string {
	h := sha256.New()
	_, _ = h.Write(b)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// A subset of the OpenAPI document useful for adding an API to the registry
type PartialOpenAPIDocument struct {
	Swagger string             `yaml:"swagger"`
	OpenAPI string             `yaml:"openapi"`
	Info    PartialOpenAPIInfo `yaml:"info"`
}

// A subset of the OpenAPI info structure useful for adding an API to the registry
type PartialOpenAPIInfo struct {
	Title string `yaml:"title"`
}
