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

package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/apigee/registry/cmd/registry/controller"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/spf13/cobra"
)

func init() {
	controllerCmd.AddCommand(controllerUpdateCmd)
}

var controllerUpdateCmd = &cobra.Command{
	Use:   "update FILENAME",
	Short: "Generate a list of commands to update the registry state (experimental)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		manifestPath := args[0]
		if manifestPath == "" {
			log.Fatal("Please provide manifest_path")
		}

		manifest, err := controller.ReadManifest(manifestPath)
		if err != nil {
			log.Fatal(err.Error())
		}

		ctx := context.Background()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Print("Generating list of actions...")
		actions, err := controller.ProcessManifest(ctx, client, manifest)
		if err != nil {
			log.Fatal(err.Error())
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

		return
	},
}
