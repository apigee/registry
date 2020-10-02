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

	"github.com/apigee/registry/connection"
	"github.com/spf13/cobra"
)

func init() {
	exportCmd.AddCommand(exportVersionsCmd)
}

var exportVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Export vocabulary changes across a version history",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		log.Printf("%+v", client)
		// input is a sequence of versions of the same API
		// read a series of version-diff vocabularies from properties
		// with names vocabulary-new and vocabulary-deleted
		// create a spreadsheet with a tab for each vocabulary
		// a summary sheet might also be interesting with
		// the number of added and removed terms
	},
}
