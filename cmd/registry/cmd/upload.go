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
	"log"
	"sync"

	"github.com/apigee/registry/gapic"
	rpcpb "github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload API specifications.",
	Long:  `Upload API specifications.`,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}

func ensureProjectExists(ctx context.Context, client *gapic.RegistryClient, projectID string) {
	// if the project doesn't exist, create it
	req := &rpcpb.GetProjectRequest{Name: "projects/" + projectID}
	_, err := client.GetProject(ctx, req)
	if notFound(err) {
		req := &rpcpb.CreateProjectRequest{
			ProjectId: projectID,
		}
		_, err := client.CreateProject(ctx, req)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	}
}

// Runnable is a generic interface for a runnable operation
type Runnable interface {
	run() error
}

var wg sync.WaitGroup

func worker(ctx context.Context, jobChan <-chan Runnable) {
	defer wg.Done()
	for job := range jobChan {
		err := job.run()
		if err != nil {
			log.Printf("ERROR %s for job %+v", err.Error(), job)
		}
	}
}
