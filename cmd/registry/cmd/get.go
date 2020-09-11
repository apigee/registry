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

	"github.com/apigee/registry/cmd/registry/tools"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
)

var getContents bool

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolVar(&getContents, "contents", false, "Get item contents (if applicable).")
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get entity values.",
	Long:  `Get entity values.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		var name string
		if len(args) > 0 {
			name = args[0]
		}

		if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
			_, err = tools.GetProject(ctx, client, m, tools.PrintProjectDetail)
		} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
			_, err = tools.GetAPI(ctx, client, m, tools.PrintAPIDetail)
		} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
			_, err = tools.GetVersion(ctx, client, m, tools.PrintVersionDetail)
		} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
			if getContents {
				_, err = tools.GetSpec(ctx, client, m, getContents, tools.PrintSpecContents)
			} else {
				_, err = tools.GetSpec(ctx, client, m, getContents, tools.PrintSpecDetail)
			}
		} else if m := names.PropertyRegexp().FindStringSubmatch(name); m != nil {
			_, err = tools.GetProperty(ctx, client, m, tools.PrintPropertyDetail)
		} else if m := names.LabelRegexp().FindStringSubmatch(name); m != nil {
			_, err = tools.GetLabel(ctx, client, m, tools.PrintLabelDetail)
		} else {
			log.Printf("Unsupported entity %+v", args)
		}
		if err != nil {
			log.Printf("%s", err.Error())
		}
	},
}
