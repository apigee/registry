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

package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	rpcpb "github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func init() {
	uploadBulkCmd.AddCommand(uploadBulkOpenAPICmd)
	uploadBulkOpenAPICmd.Flags().String("project_id", "", "Project id.")
	uploadBulkOpenAPICmd.Flags().String("base_uri", "", "Base to use for setting source_uri fields of uploaded specs.")
}

var uploadBulkOpenAPICmd = &cobra.Command{
	Use:   "openapi",
	Short: "Bulk-upload OpenAPI descriptions from a directory of specs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		flagset := cmd.LocalFlags()
		projectID, err := flagset.GetString("project_id")
		if err != nil {
			log.Fatal(err.Error())
		}
		if projectID == "" {
			log.Fatal("Please specify a project_id")
		}
		baseURI, err := flagset.GetString("base_uri")
		if err != nil {
			log.Fatal(err.Error())
		}
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}
		core.EnsureProjectExists(ctx, client, projectID)
		for _, arg := range args {
			scanDirectoryForOpenAPI(projectID, baseURI, arg)
		}
	},
}

func scanDirectoryForOpenAPI(projectID, baseURI, directory string) {
	ctx := context.TODO()

	client, err := connection.NewClient(ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	// create a queue for upload tasks and wait for the workers to finish after filling it.
	taskQueue := make(chan core.Task, 1024)
	for i := 0; i < 64; i++ {
		core.WaitGroup().Add(1)
		go core.Worker(ctx, taskQueue)
	}
	defer core.WaitGroup().Wait()
	defer close(taskQueue)

	// walk a directory hierarchy, uploading every API spec that matches a set of expected file names.
	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		task := &uploadOpenAPITask{
			ctx:       ctx,
			client:    client,
			projectID: projectID,
			baseURI:   baseURI,
			path:      path,
			directory: directory,
		}

		switch {
		case strings.HasSuffix(path, "swagger.yaml"), strings.HasSuffix(path, "swagger.json"):
			task.style = "openapi/v2"
			taskQueue <- task
		case strings.HasSuffix(path, "openapi.yaml"), strings.HasSuffix(path, "openapi.json"):
			task.style = "openapi/v3"
			taskQueue <- task
		}

		return nil
	}); err != nil {
		log.Println(err)
	}
}

func sanitize(name string) string {
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, ":", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	return name
}

type uploadOpenAPITask struct {
	ctx       context.Context
	client    connection.Client
	baseURI   string
	path      string
	directory string
	style     string
	projectID string
	apiID     string // computed at runtime
	apiOwner  string // computed at runtime
	versionID string // computed at runtime
	specID    string // computed at runtime
}

func (task *uploadOpenAPITask) Name() string {
	return "upload openapi " + task.path
}

func (task *uploadOpenAPITask) Run() error {
	// Populate API path fields using the file's path.
	if err := task.populateFields(); err != nil {
		return err
	}
	log.Printf("^^ apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)

	// If the API does not exist, create it.
	if err := task.createAPI(); err != nil {
		return err
	}
	// If the API version does not exist, create it.
	if err := task.createVersion(); err != nil {
		return err
	}
	// If the API spec does not exist, create it.
	if err := task.createSpec(); err != nil {
		return err
	}
	// If the API spec needs a new revision, create it.
	return task.updateSpec()
}

func (task *uploadOpenAPITask) populateFields() error {
	parts := strings.Split(task.apiPath(), "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid API path: %s", task.apiPath())
	}

	task.apiOwner = parts[0]

	apiParts := parts[0 : len(parts)-2]
	apiPart := strings.ReplaceAll(strings.Join(apiParts, "-"), "/", "-")
	task.apiID = sanitize(apiPart)

	versionPart := parts[len(parts)-2]
	task.versionID = sanitize(versionPart)

	specPart := parts[len(parts)-1]
	task.specID = sanitize(specPart)

	return nil
}

func (task *uploadOpenAPITask) createAPI() error {
	if _, err := task.client.GetApi(task.ctx, &rpcpb.GetApiRequest{
		Name: task.apiName(),
	}); !core.NotFound(err) {
		return err // Returns nil when API is found without error.
	}

	response, err := task.client.CreateApi(task.ctx, &rpcpb.CreateApiRequest{
		Parent: task.projectName(),
		ApiId:  task.apiID,
		Api: &rpcpb.Api{
			DisplayName: task.apiID,
		},
	})
	if err != nil {
		log.Printf("error %s: %s", task.apiName(), err.Error())
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadOpenAPITask) createVersion() error {
	if _, err := task.client.GetApiVersion(task.ctx, &rpcpb.GetApiVersionRequest{
		Name: task.versionName(),
	}); !core.NotFound(err) {
		return err // Returns nil when version is found without error.
	}

	response, err := task.client.CreateApiVersion(task.ctx, &rpcpb.CreateApiVersionRequest{
		Parent:       task.apiName(),
		ApiVersionId: task.versionID,
		ApiVersion:   &rpcpb.ApiVersion{},
	})
	if err != nil {
		log.Printf("error %s: %s", task.versionName(), err.Error())
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadOpenAPITask) createSpec() error {
	contents, err := task.gzipContents()
	if err != nil {
		return err
	}

	if _, err = task.client.GetApiSpec(task.ctx, &rpcpb.GetApiSpecRequest{
		Name: task.specName(),
	}); !core.NotFound(err) {
		return err // Returns nil when spec is found without error.
	}

	request := &rpcpb.CreateApiSpecRequest{
		Parent:    task.versionName(),
		ApiSpecId: task.fileName(),
		ApiSpec: &rpcpb.ApiSpec{
			MimeType: fmt.Sprintf("%s+gzip", task.style),
			Filename: task.fileName(),
			Contents: contents,
		},
	}
	if task.baseURI != "" {
		request.ApiSpec.SourceUri = fmt.Sprintf("%s/%s", task.baseURI, task.apiPath())
	}

	response, err := task.client.CreateApiSpec(task.ctx, request)
	if err != nil {
		log.Printf("error %s: %s [contents-length: %d]", task.specName(), err.Error(), len(contents))
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadOpenAPITask) updateSpec() error {
	contents, err := task.gzipContents()
	if err != nil {
		return err
	}

	spec, err := task.client.GetApiSpec(task.ctx, &rpcpb.GetApiSpecRequest{
		Name: task.specName(),
	})
	if err != nil && !core.NotFound(err) {
		return err
	}

	request := &rpcpb.UpdateApiSpecRequest{
		ApiSpec: &rpcpb.ApiSpec{
			Name:     task.specName(),
			Contents: contents,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"contents"}},
	}

	response, err := task.client.UpdateApiSpec(task.ctx, request)
	if err != nil {
		log.Printf("error %s: %s [contents-length: %d]", request.ApiSpec.Name, err.Error(), len(contents))
	} else if response.RevisionId != spec.RevisionId {
		log.Printf("updated %s", response.Name)
	}

	return nil
}

func (task *uploadOpenAPITask) projectName() string {
	return fmt.Sprintf("projects/%s", task.projectID)
}

func (task *uploadOpenAPITask) apiName() string {
	return fmt.Sprintf("%s/apis/%s", task.projectName(), task.apiID)
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
