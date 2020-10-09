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
	"log"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	rpcpb "github.com/apigee/registry/rpc"
	discovery "github.com/googleapis/gnostic/discovery"
	"github.com/spf13/cobra"
)

func init() {
	uploadBulkCmd.AddCommand(uploadBulkDiscoveryCmd)
	uploadBulkDiscoveryCmd.Flags().String("project_id", "", "Project id.")
}

var uploadBulkDiscoveryCmd = &cobra.Command{
	Use:   "discovery",
	Short: "Bulk-upload API Discovery documents from the Google API Discovery service",
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
		taskQueue := make(chan core.Task, 1024)

		workerCount := 64
		for i := 0; i < workerCount; i++ {
			core.WaitGroup().Add(1)
			go core.Worker(ctx, taskQueue)
		}

		core.EnsureProjectExists(ctx, client, projectID)
		// Get the list of specs.
		listResponse, err := discovery.FetchList()
		if err != nil {
			log.Fatalf("%+v", err)
		}
		// Create an upload job for each API.
		for _, api := range listResponse.APIs {
			taskQueue <- &uploadDiscoveryTask{
				ctx:       ctx,
				client:    client,
				path:      api.DiscoveryRestURL,
				projectID: projectID,
				apiID:     api.Name,
				versionID: api.Version,
				specID:    "discovery.json",
			}
		}
		close(taskQueue)
		core.WaitGroup().Wait()
	},
}

type uploadDiscoveryTask struct {
	ctx       context.Context
	client    connection.Client
	path      string
	projectID string
	apiID     string // computed at runtime
	versionID string // computed at runtime
	specID    string // computed at runtime
	fileBytes []byte
	document  *discovery.Document
}

func (task *uploadDiscoveryTask) Name() string {
	return "upload discovery " + task.path
}

func (task *uploadDiscoveryTask) Run() error {
	var err error
	// Fetch the discovery description of the API.
	task.fileBytes, err = discovery.FetchDocumentBytes(task.path)
	if err != nil {
		return err
	}
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

func (task *uploadDiscoveryTask) createAPI() error {
	request := &rpcpb.GetApiRequest{}
	request.Name = "projects/" + task.projectID + "/apis/" + task.apiID
	_, err := task.client.GetApi(task.ctx, request)
	if core.NotFound(err) {
		request := &rpcpb.CreateApiRequest{}
		request.Parent = "projects/" + task.projectID
		request.ApiId = task.apiID
		request.Api = &rpcpb.Api{}
		request.Api.DisplayName = task.apiID
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

func (task *uploadDiscoveryTask) createVersion() error {
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

func (task *uploadDiscoveryTask) createSpec() error {
	request := &rpcpb.GetSpecRequest{}
	request.Name = "projects/" + task.projectID + "/apis/" + task.apiID +
		"/versions/" + task.versionID +
		"/specs/" + task.specID
	_, err := task.client.GetSpec(task.ctx, request)
	if core.NotFound(err) {
		fileBytes := task.fileBytes
		// compress the spec before uploading it
		gzippedBytes, err := core.GZippedBytes(fileBytes)
		request := &rpcpb.CreateSpecRequest{}
		request.Parent = "projects/" + task.projectID + "/apis/" + task.apiID +
			"/versions/" + task.versionID
		request.SpecId = task.specID
		request.Spec = &rpcpb.Spec{}
		request.Spec.Style = "discovery" + "+gzip"
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
