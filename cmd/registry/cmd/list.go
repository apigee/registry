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

	"github.com/apigee/registry/cmd/registry/tools"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
)

var filterFlag string

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&filterFlag, "filter", "", "Filter option to send with list calls")
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources in the Registry.",
	Long:  "List resources in the Registry.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		err = matchAndHandle(ctx, client, args[0])
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				log.Fatalf("%s", err.Error())
			} else {
				log.Fatalf("%s", st.Message())
			}
		}
	},
}

func matchAndHandle(
	ctx context.Context,
	client connection.Client,
	name string,
) error {

	// First try to match collection names.
	if m := names.ProjectsRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListProjects(ctx, client, m, filterFlag, tools.PrintProject)
	} else if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListAPIs(ctx, client, m, filterFlag, tools.PrintAPI)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListVersions(ctx, client, m, filterFlag, tools.PrintVersion)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListSpecs(ctx, client, m, filterFlag, tools.PrintSpec)
	} else if m := names.PropertiesRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListProperties(ctx, client, m, filterFlag, tools.PrintProperty)
	} else if m := names.LabelsRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListLabels(ctx, client, m, filterFlag, tools.PrintLabel)
	}

	// Then try to match resource names.
	if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListProjects(ctx, client, m, filterFlag, tools.PrintProject)
	} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListAPIs(ctx, client, m, filterFlag, tools.PrintAPI)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListVersions(ctx, client, m, filterFlag, tools.PrintVersion)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListSpecs(ctx, client, m, filterFlag, tools.PrintSpec)
	} else if m := names.PropertyRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListProperties(ctx, client, m, filterFlag, tools.PrintProperty)
	} else if m := names.LabelRegexp().FindStringSubmatch(name); m != nil {
		return tools.ListLabels(ctx, client, m, filterFlag, tools.PrintLabel)
	}

	// If nothing matched, return an error.
	return fmt.Errorf("unsupported argument: %s", name)
}
