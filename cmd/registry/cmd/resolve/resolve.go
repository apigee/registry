// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resolve

import (
	"context"
	"fmt"
	"log"

	"github.com/apigee/registry/cmd/registry/controller"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve MANIFEST_RESOURCE",
		Short: "resolve the dependencies and update the registry state (experimental)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			manifestName := args[0]
			if manifestName == "" {
				log.Fatal("Please provide the manifest resource name")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.Fatal(err.Error())
			}

			manifest, err := controller.FetchManifest(ctx, client, manifestName)
			if err != nil {
				log.Fatal(err.Error())
			}

			projectID, err := core.ProjectID(manifestName)
			if err != nil {
				log.Fatalf("Error while extracting project_id: %s", err.Error())
			}

			log.Print("Generating the list of actions...")
			actions, err := controller.ProcessManifest(ctx, client, projectID, manifest)
			if err != nil {
				log.Fatal(err.Error())
			}

			if len(actions) == 0 {
				log.Printf("No actions needed. The registry is already in a resolved state.")
				return
			}

			log.Printf("Generated %d actions. Starting Execution...", len(actions))

			taskQueue := make(chan core.Task, 1024)
			for i := 0; i < 64; i++ {
				core.WaitGroup().Add(1)
				go core.Worker(ctx, taskQueue)
			}
			defer core.WaitGroup().Wait()
			defer close(taskQueue)

			// Submit tasks to taskQueue
			for i, a := range actions {
				taskQueue <- &controller.ExecCommandTask{
					Action: a,
					TaskID: fmt.Sprintf("task%d", i),
				}
			}
		},
	}
	return cmd
}
