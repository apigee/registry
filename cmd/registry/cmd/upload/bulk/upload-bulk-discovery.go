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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	rpcpb "github.com/apigee/registry/rpc"
	discovery "github.com/googleapis/gnostic/discovery"
	"github.com/nsf/jsondiff"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func init() {
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
			log.Fatal(err.Error())
		}
		if projectID == "" {
			log.Fatalf("Please specify a project_id")
		}
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}

		// create a queue for upload tasks and wait for the workers to finish after filling it.
		taskQueue := make(chan core.Task, 1024)
		for i := 0; i < 64; i++ {
			core.WaitGroup().Add(1)
			go core.Worker(ctx, taskQueue)
		}
		defer core.WaitGroup().Wait()
		defer close(taskQueue)

		core.EnsureProjectExists(ctx, client, projectID)
		discoveryResponse, err := discovery.FetchList()
		if err != nil {
			log.Fatal(err)
		}

		// Create an upload job for each API.
		for _, api := range discoveryResponse.APIs {
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
	},
}

type uploadDiscoveryTask struct {
	ctx       context.Context
	client    connection.Client
	path      string
	projectID string
	apiID     string
	versionID string
	specID    string
	document  *discovery.Document
}

func (task *uploadDiscoveryTask) String() string {
	return "upload discovery " + task.path
}

func (task *uploadDiscoveryTask) Run() error {
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

func (task *uploadDiscoveryTask) createAPI() error {
	if _, err := task.client.GetApi(task.ctx, &rpcpb.GetApiRequest{
		Name: task.apiName(),
	}); !core.NotFound(err) {
		return err // Returns nil when API is found without error.
	}

	response, err := task.client.CreateApi(task.ctx, &rpcpb.CreateApiRequest{
		Parent: task.projectName(),
		ApiId:  task.apiID,
		Api: &rpc.Api{
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

func (task *uploadDiscoveryTask) createVersion() error {
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

func (task *uploadDiscoveryTask) createSpec() error {
	contents, err := task.gzipContents()
	if err != nil {
		return err
	}

	if _, err := task.client.GetApiSpec(task.ctx, &rpcpb.GetApiSpecRequest{
		Name: task.specName(),
	}); !core.NotFound(err) {
		return err // Returns nil when spec is found without error.
	}

	response, err := task.client.CreateApiSpec(task.ctx, &rpcpb.CreateApiSpecRequest{
		Parent:    task.versionName(),
		ApiSpecId: task.specID,
		ApiSpec: &rpcpb.ApiSpec{
			MimeType:  core.DiscoveryMimeType("+gzip"),
			Filename:  "discovery.json",
			Contents:  contents,
			SourceUri: task.path,
		},
	})
	if err != nil {
		log.Printf("error %s: %s [contents-length: %d]", task.specName(), err.Error(), len(contents))
	} else {
		log.Printf("created %s", response.Name)
	}

	return nil
}

func (task *uploadDiscoveryTask) updateSpec() error {
	refSpec, err := task.client.GetApiSpec(task.ctx, &rpcpb.GetApiSpecRequest{
		Name: task.specName(),
	})
	if err != nil && !core.NotFound(err) {
		return err
	}

	refBytes, err := core.GetBytesForSpec(task.ctx, task.client, refSpec)
	if err != nil {
		return nil
	}

	docBytes, err := discovery.FetchDocumentBytes(task.path)
	if err != nil {
		return err
	}

	opts := jsondiff.DefaultJSONOptions()
	if diff, _ := jsondiff.Compare(refBytes, docBytes, &opts); diff == jsondiff.FullMatch {
		return nil
	}

	docZipped, err := core.GZippedBytes(docBytes)
	if err != nil {
		return err
	}

	response, err := task.client.UpdateApiSpec(task.ctx, &rpcpb.UpdateApiSpecRequest{
		ApiSpec: &rpcpb.ApiSpec{
			Name:     task.specName(),
			Contents: docZipped,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"contents"}},
	})
	if err != nil {
		log.Printf("error %s: %s [contents-length: %d]", task.specName(), err.Error(), len(docZipped))
	} else if response.RevisionId != refSpec.RevisionId {
		log.Printf("updated %s", response.Name)
	}

	return nil
}

func (task *uploadDiscoveryTask) projectName() string {
	return fmt.Sprintf("projects/%s", task.projectID)
}

func (task *uploadDiscoveryTask) apiName() string {
	return fmt.Sprintf("%s/apis/%s", task.projectName(), task.apiID)
}

func (task *uploadDiscoveryTask) versionName() string {
	return fmt.Sprintf("%s/versions/%s", task.apiName(), task.versionID)
}

func (task *uploadDiscoveryTask) specName() string {
	return fmt.Sprintf("%s/specs/%s", task.versionName(), task.specID)
}

func (task *uploadDiscoveryTask) gzipContents() ([]byte, error) {
	bytes, err := discovery.FetchDocumentBytes(task.path)
	if err != nil {
		return nil, err
	}

	return core.GZippedBytes(bytes)
}
