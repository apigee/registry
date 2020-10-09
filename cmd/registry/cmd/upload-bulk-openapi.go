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
)

func init() {
	uploadBulkCmd.AddCommand(uploadBulkOpenAPICmd)
	uploadBulkOpenAPICmd.Flags().String("project_id", "", "Project id.")
}

var uploadBulkOpenAPICmd = &cobra.Command{
	Use:   "openapi",
	Short: "Bulk-upload OpenAPI descriptions from a directory of specs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flagset := cmd.LocalFlags()
		projectID, err := flagset.GetString("project_id")
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		if projectID == "" {
			log.Fatalf("Please specify a project_id")
		}
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		core.EnsureProjectExists(ctx, client, projectID)
		for _, arg := range args {
			scanDirectoryForOpenAPI(projectID, arg)
		}
	},
}

func scanDirectoryForOpenAPI(projectID, directory string) {
	ctx := context.TODO()

	client, err := connection.NewClient(ctx)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(-1)
	}

	taskQueue := make(chan core.Task, 1024)

	workerCount := 64
	for i := 0; i < workerCount; i++ {
		core.WaitGroup().Add(1)
		go core.Worker(ctx, taskQueue)
	}

	// walk a directory hierarchy, uploading every API spec that matches a set of expected file names.
	err = filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if strings.HasSuffix(path, "swagger.yaml") || strings.HasSuffix(path, "swagger.json") {
				taskQueue <- &uploadOpenAPITask{
					ctx:       ctx,
					client:    client,
					projectID: projectID,
					path:      path,
					directory: directory,
					style:     "openapi/v2",
				}
			} else if strings.HasSuffix(path, "openapi.yaml") || strings.HasSuffix(path, "openapi.json") {
				taskQueue <- &uploadOpenAPITask{
					ctx:       ctx,
					client:    client,
					projectID: projectID,
					path:      path,
					directory: directory,
					style:     "openapi/v3",
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	close(taskQueue)
	core.WaitGroup().Wait()
}

type uploadOpenAPITask struct {
	ctx       context.Context
	client    connection.Client
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

func sanitize(name string) string {
	name = strings.Replace(name, " ", "-", -1)
	name = strings.Replace(name, ":", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	return name
}

func (task *uploadOpenAPITask) Run() error {
	var err error
	// Compute the API name from the path to the spec file.
	name := strings.TrimPrefix(task.path, task.directory+"/")
	parts := strings.Split(name, "/")
	if len(parts) < 3 {
		return fmt.Errorf("Invalid API path: %s", name)
	}
	task.apiID = sanitize(strings.Join(parts[0:len(parts)-2], "-"))
	task.apiID = strings.Replace(task.apiID, "/", "-", -1)
	task.apiOwner = parts[0]
	task.versionID = sanitize(parts[len(parts)-2])
	task.specID = sanitize(parts[len(parts)-1])
	log.Printf("^^ apis/%s/versions/%s/specs/%s", task.apiID, task.versionID, task.specID)
	// If the API does not exist, create it.
	err = task.createAPI()
	if err != nil {
		return err
	}
	// If the API version does not exist, create it.
	err = task.createVersion()
	if err != nil {
		return err
	}
	// If the API spec does not exist, create it.
	return task.createSpec()
}

func (task *uploadOpenAPITask) createAPI() error {
	request := &rpcpb.GetApiRequest{}
	request.Name = "projects/" + task.projectID + "/apis/" + task.apiID
	_, err := task.client.GetApi(task.ctx, request)
	if core.NotFound(err) {
		request := &rpcpb.CreateApiRequest{}
		request.Parent = "projects/" + task.projectID
		request.ApiId = task.apiID
		request.Api = &rpcpb.Api{}
		request.Api.DisplayName = task.apiID
		request.Api.Owner = task.apiOwner
		response, err := task.client.CreateApi(task.ctx, request)
		if err == nil {
			log.Printf("created %s", response.Name)
		} else if core.AlreadyExists(err) {
			log.Printf("found %s/apis/%s", request.Parent, request.ApiId)
		} else {
			log.Printf("error %s/apis/%s: %s",
				request.Parent, request.ApiId, err.Error())
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (task *uploadOpenAPITask) createVersion() error {
	request := &rpcpb.GetVersionRequest{}
	request.Name = "projects/" + task.projectID + "/apis/" + task.apiID + "/versions/" + task.versionID
	_, err := task.client.GetVersion(task.ctx, request)
	if core.NotFound(err) {
		request := &rpcpb.CreateVersionRequest{}
		request.Parent = "projects/" + task.projectID + "/apis/" + task.apiID
		request.VersionId = task.versionID
		request.Version = &rpcpb.Version{}
		response, err := task.client.CreateVersion(task.ctx, request)
		if err == nil {
			log.Printf("created %s", response.Name)
		} else if core.AlreadyExists(err) {
			log.Printf("found %s/versions/%s", request.Parent, request.VersionId)
		} else {
			log.Printf("error %s/versions/%s: %s",
				request.Parent, request.VersionId, err.Error())
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (task *uploadOpenAPITask) createSpec() error {
	filename := filepath.Base(task.path)

	request := &rpcpb.GetSpecRequest{}
	request.Name = "projects/" + task.projectID + "/apis/" + task.apiID +
		"/versions/" + task.versionID +
		"/specs/" + filename
	_, err := task.client.GetSpec(task.ctx, request)
	if core.NotFound(err) {
		fileBytes, err := ioutil.ReadFile(task.path)
		gzippedBytes, err := core.GZippedBytes(fileBytes)
		request := &rpcpb.CreateSpecRequest{}
		request.Parent = "projects/" + task.projectID + "/apis/" + task.apiID +
			"/versions/" + task.versionID
		request.SpecId = filename
		request.Spec = &rpcpb.Spec{}
		request.Spec.Style = task.style + "+gzip"
		request.Spec.Filename = filename
		request.Spec.Contents = gzippedBytes
		response, err := task.client.CreateSpec(task.ctx, request)
		if err == nil {
			log.Printf("created %s", response.Name)
		} else if core.AlreadyExists(err) {
			log.Printf("found %s/specs/%s", request.Parent, request.SpecId)
		} else {
			details := fmt.Sprintf("contents-length: %d", len(request.Spec.Contents))
			log.Printf("error %s/specs/%s: %s [%s]",
				request.Parent, request.SpecId, err.Error(), details)
		}
	} else if err != nil {
		return err
	}
	return nil
}
