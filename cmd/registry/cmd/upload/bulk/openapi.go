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
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func openAPICommand(ctx context.Context) *cobra.Command {
	var baseURI string
	cmd := &cobra.Command{
		Use:   "openapi",
		Short: "Bulk-upload OpenAPI descriptions from a directory of specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectID, err := cmd.Flags().GetString("project-id")
			if err != nil {
				log.WithError(err).Fatal("Failed to get project-id from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}

			for _, arg := range args {
				scanDirectoryForOpenAPI(ctx, client, projectID, baseURI, arg)
			}
		},
	}

	cmd.Flags().StringVar(&baseURI, "base-uri", "", "Prefix to use for the source_uri field of each spec upload")
	return cmd
}

func scanDirectoryForOpenAPI(ctx context.Context, client connection.Client, projectID, baseURI, directory string) {
	// create a queue for upload tasks and wait for the workers to finish after filling it.
	taskQueue, wait := core.WorkerPool(ctx, 64)
	defer wait()

	// walk a directory hierarchy, uploading every API spec that matches a set of expected file names.
	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		task := &uploadOpenAPITask{
			client:    client,
			projectID: projectID,
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
		log.WithError(err).Debug("Failed to walk directory")
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
	client    connection.Client
	baseURI   string
	path      string
	directory string
	version   string
	projectID string
	apiID     string // computed at runtime
	versionID string // computed at runtime
	specID    string // computed at runtime
}

func (task *uploadOpenAPITask) String() string {
	return "upload openapi " + task.path
}

func (task *uploadOpenAPITask) Run(ctx context.Context) error {
	// Populate API path fields using the file's path.
	if err := task.populateFields(); err != nil {
		return err
	}
	log.Infof("Uploading apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)

	// If the API does not exist, create it.
	if err := task.createAPI(ctx); err != nil {
		return err
	}
	// If the API version does not exist, create it.
	if err := task.createVersion(ctx); err != nil {
		return err
	}
	// If the API spec does not exist, create it.
	if err := task.createSpec(ctx); err != nil {
		return err
	}
	// If the API spec needs a new revision, create it.
	return task.updateSpec(ctx)
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

	specPart := parts[len(parts)-1]
	task.specID = sanitize(specPart)

	return nil
}

func (task *uploadOpenAPITask) createAPI(ctx context.Context) error {
	if _, err := task.client.GetApi(ctx, &rpc.GetApiRequest{
		Name: task.apiName(),
	}); !core.NotFound(err) {
		return err // Returns nil when API is found without error.
	}

	response, err := task.client.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: task.projectName() + "/locations/global",
		ApiId:  task.apiID,
		Api: &rpc.Api{
			DisplayName: task.apiID,
		},
	})
	if err == nil {
		log.Debugf("Created %s", response.Name)
	} else if core.AlreadyExists(err) {
		log.Debugf("Found %s", task.apiName())
	} else {
		log.WithError(err).Debugf("Failed to create API %s", task.apiName())
		return fmt.Errorf("Failed to create %s, %s", task.apiName(), err)
	}

	return nil
}

func (task *uploadOpenAPITask) createVersion(ctx context.Context) error {
	if _, err := task.client.GetApiVersion(ctx, &rpc.GetApiVersionRequest{
		Name: task.versionName(),
	}); !core.NotFound(err) {
		return err // Returns nil when version is found without error.
	}

	response, err := task.client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
		Parent:       task.apiName(),
		ApiVersionId: task.versionID,
		ApiVersion:   &rpc.ApiVersion{},
	})
	if err != nil {
		log.WithError(err).Debugf("Failed to create version %s", task.versionName())
	} else {
		log.Debugf("Created %s", response.Name)
	}

	return nil
}

func (task *uploadOpenAPITask) createSpec(ctx context.Context) error {
	contents, err := task.gzipContents()
	if err != nil {
		return err
	}

	if _, err = task.client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: task.specName(),
	}); !core.NotFound(err) {
		return err // Returns nil when spec is found without error.
	}

	request := &rpc.CreateApiSpecRequest{
		Parent:    task.versionName(),
		ApiSpecId: task.fileName(),
		ApiSpec: &rpc.ApiSpec{
			MimeType: core.OpenAPIMimeType("+gzip", task.version),
			Filename: task.fileName(),
			Contents: contents,
		},
	}
	if task.baseURI != "" {
		request.ApiSpec.SourceUri = fmt.Sprintf("%s/%s", task.baseURI, task.apiPath())
	}

	response, err := task.client.CreateApiSpec(ctx, request)
	if err != nil {
		log.WithError(err).Debugf("Error %s [contents-length: %d]", task.specName(), len(contents))
	} else {
		log.Debugf("Created %s", response.Name)
	}

	return nil
}

func (task *uploadOpenAPITask) updateSpec(ctx context.Context) error {
	contents, err := task.gzipContents()
	if err != nil {
		return err
	}

	spec, err := task.client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: task.specName(),
	})
	if err != nil && !core.NotFound(err) {
		return err
	}

	request := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:     task.specName(),
			Contents: contents,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"contents"}},
	}

	response, err := task.client.UpdateApiSpec(ctx, request)
	if err != nil {
		log.WithError(err).Debugf("Error %s [contents-length: %d]", request.ApiSpec.Name, len(contents))
	} else if response.RevisionId != spec.RevisionId {
		log.Debugf("Updated %s", response.Name)
	}

	return nil
}

func (task *uploadOpenAPITask) projectName() string {
	return fmt.Sprintf("projects/%s", task.projectID)
}

func (task *uploadOpenAPITask) apiName() string {
	return fmt.Sprintf("%s/locations/global/apis/%s", task.projectName(), task.apiID)
}

func (task *uploadOpenAPITask) versionName() string {
	return fmt.Sprintf("%s/versions/%s", task.apiName(), task.versionID)
}

func (task *uploadOpenAPITask) specName() string {
	return fmt.Sprintf("%s/specs/%s", task.versionName(), filepath.Base(task.path))
}

func (task *uploadOpenAPITask) apiPath() string {
	prefix := task.directory + "/"
	return strings.TrimPrefix(task.path, prefix)
}

func (task *uploadOpenAPITask) fileName() string {
	return filepath.Base(task.path)
}

func (task *uploadOpenAPITask) gzipContents() ([]byte, error) {
	bytes, err := ioutil.ReadFile(task.path)
	if err != nil {
		return nil, err
	}

	return core.GZippedBytes(bytes)
}
