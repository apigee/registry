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
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func protosCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "protos",
		Short: "Bulk-upload Protocol Buffer descriptions from a directory of specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			flagset := cmd.LocalFlags()
			projectID, err := flagset.GetString("project_id")
			if err != nil {
				log.Fatal(err.Error())
			}
			if projectID == "" {
				log.Fatalf("Please specify a project_id")
			}
			baseURI, err := flagset.GetString("base_uri")
			if err != nil {
				log.Fatal(err.Error())
			}
			ctx := context.Background()
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.Fatal(err.Error())
			}
			core.EnsureProjectExists(ctx, client, projectID)
			for _, arg := range args {
				scanDirectoryForProtos(ctx, client, projectID, baseURI, arg)
			}
		},
	}

	cmd.Flags().String("project_id", "", "Project id.")
	cmd.Flags().String("base_uri", "", "Base to use for setting source_uri fields of uploaded specs.")
	return cmd
}

func scanDirectoryForProtos(ctx context.Context, client connection.Client, projectID, baseURI, directory string) {
	// create a queue for upload tasks and wait for the workers to finish after filling it.
	taskQueue := make(chan core.Task, 1024)
	for i := 0; i < 64; i++ {
		core.WaitGroup().Add(1)
		go core.Worker(ctx, taskQueue)
	}
	defer core.WaitGroup().Wait()
	defer close(taskQueue)

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
		log.Println(err)
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
	log.Printf("^^ apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)

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

func (task *uploadProtoTask) populateFields() {
	parts := strings.Split(task.apiPath(), "/")
	apiParts := parts[0 : len(parts)-1]

	task.apiID = strings.ReplaceAll(strings.Join(apiParts, "-"), "/", "-")
	task.versionID = parts[len(parts)-1]
	task.specID = task.fileName()
}

func (task *uploadProtoTask) createAPI(ctx context.Context) error {
	if _, err := task.client.GetApi(ctx, &rpc.GetApiRequest{
		Name: task.apiName(),
	}); !core.NotFound(err) {
		return err // Returns nil when API is found without error.
	}

	response, err := task.client.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: task.projectName(),
		ApiId:  task.apiID,
		Api:    &rpc.Api{},
	})
	if err != nil {
		log.Printf("error %s: %s", task.apiName(), err.Error())
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadProtoTask) createVersion(ctx context.Context) error {
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
		log.Printf("error %s: %s", task.versionName(), err.Error())
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadProtoTask) createSpec(ctx context.Context) error {
	contents, err := task.zipContents()
	if err != nil {
		return err
	}

	if _, err := task.client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: task.specName(),
	}); !core.NotFound(err) {
		return err // Returns nil when spec is found without error.
	}

	request := &rpc.CreateApiSpecRequest{
		Parent:    task.versionName(),
		ApiSpecId: task.fileName(),
		ApiSpec: &rpc.ApiSpec{
			MimeType: core.ProtobufMimeType("+zip"),
			Filename: task.fileName(),
			Contents: contents,
		},
	}
	if task.baseURI != "" {
		request.ApiSpec.SourceUri = fmt.Sprintf("%s/%s", task.baseURI, task.apiPath())
	}

	response, err := task.client.CreateApiSpec(ctx, request)
	if err != nil {
		log.Printf("error %s: %s [contents-length: %d]", task.specName(), err.Error(), len(contents))
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadProtoTask) updateSpec(ctx context.Context) error {
	contents, err := task.zipContents()
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
		log.Printf("error %s: %s [contents-length: %d]", request.ApiSpec.Name, err.Error(), len(contents))
	} else if response.RevisionId != spec.RevisionId {
		log.Printf("updated %s", response.Name)
	}

	return nil
}

func (task *uploadProtoTask) projectName() string {
	return fmt.Sprintf("projects/%s", task.projectID)
}

func (task *uploadProtoTask) apiName() string {
	return fmt.Sprintf("%s/apis/%s", task.projectName(), task.apiID)
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
