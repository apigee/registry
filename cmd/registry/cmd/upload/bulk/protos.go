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
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

func protosCommand(ctx context.Context) *cobra.Command {
	var baseURI string
	cmd := &cobra.Command{
		Use:   "protos",
		Short: "Bulk-upload Protocol Buffer descriptions from a directory of specs",
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
				scanDirectoryForProtos(ctx, client, projectID, baseURI, arg)
			}
		},
	}

	cmd.Flags().StringVar(&baseURI, "base-uri", "", "Prefix to use for the source_uri field of each proto upload")
	return cmd
}

func scanDirectoryForProtos(ctx context.Context, client connection.Client, projectID, baseURI, directory string) {
	// create a queue for upload tasks and wait for the workers to finish after filling it.
	taskQueue, wait := core.WorkerPool(ctx, 64)
	defer wait()

	dirPattern := regexp.MustCompile("v.*[1-9]+.*")
	if err := filepath.Walk(directory, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-matching directories.
		filename := path.Base(filepath)
		if !info.IsDir() || !dirPattern.MatchString(filename) {
			return nil
		}

		taskQueue <- &uploadProtoTask{
			client:    client,
			baseURI:   baseURI,
			projectID: projectID,
			path:      filepath,
			directory: directory,
		}

		return nil
	}); err != nil {
		log.WithError(err).Debug("Failed to walk directory")
	}
}

type uploadProtoTask struct {
	client    connection.Client
	baseURI   string
	projectID string
	path      string
	directory string
	apiID     string // computed at runtime
	versionID string // computed at runtime
	specID    string // computed at runtime
}

func (task *uploadProtoTask) String() string {
	return "upload proto " + task.path
}

func (task *uploadProtoTask) Run(ctx context.Context) error {
	// Populate API path fields using the file's path.
	task.populateFields()
	log.Infof("Uploading apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)

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
	apiParts := parts[0 : len(parts)-1]
	apiPart := strings.ReplaceAll(strings.Join(apiParts, "-"), "/", "-")
	task.apiID = sanitize(apiPart)

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
			DisplayName: task.apiID,
		},
		AllowMissing: true,
	})
	if err == nil {
		log.Debugf("Updated %s", response.Name)
	} else {
		log.WithError(err).Debugf("Failed to create API %s", task.apiName())
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
		log.Debugf("Updated %s", response.Name)
	} else {
		log.WithError(err).Debugf("Failed to create version %s", task.versionName())
	}

	return nil
}

func (task *uploadProtoTask) createOrUpdateSpec(ctx context.Context) error {
	contents, err := task.zipContents()
	if err != nil {
		return err
	}

	// Use the spec size and hash to avoid unnecessary uploads.
	if spec, err := task.client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: task.specName(),
	}); err == nil {
		if int(spec.GetSizeBytes()) == len(contents) {
			hash := hashForBytes(contents)
			if spec.GetHash() == hash {
				log.Debugf("Matched already uploaded spec %s", task.specName())
				return nil
			}
		}
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
		log.WithError(err).Errorf("Error %s [contents-length: %d]", task.specName(), len(contents))
	} else {
		log.Debugf("Updated %s", response.Name)
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
	return "protos.zip"
}

func (task *uploadProtoTask) zipContents() ([]byte, error) {
	prefix := task.directory + "/"
	contents, err := core.ZipArchiveOfPath(task.path, prefix)
	if err != nil {
		return nil, err
	}

	return contents.Bytes(), nil
}
