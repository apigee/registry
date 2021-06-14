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
	"log"
	"github.com/apigee/registry/cmd/registry/controller"
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
		if err!=nil {
			log.Fatal(err.Error())
		}

		actions, err := controller.ProcessManifest(manifest)
		if err!=nil {
			log.Fatal(err.Error())
		}

		log.Print("Actions:")
		for i, a := range actions {
			log.Printf("%d: %s", i, a)
		}
	},
} 