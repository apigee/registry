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
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
)

var listFilter string

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listFilter, "filter", "", "Filter option to send with list calls")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources in the API Registry",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		err = matchAndHandleListCmd(ctx, client, args[0])
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	},
}

func matchAndHandleListCmd(
	ctx context.Context,
	client connection.Client,
	name string,
) error {

	// First try to match collection names.
	if m := names.ProjectsRegexp().FindStringSubmatch(name); m != nil {
		return core.ListProjects(ctx, client, m, listFilter, core.PrintProject)
	} else if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		return core.ListAPIs(ctx, client, m, listFilter, core.PrintAPI)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		return core.ListVersions(ctx, client, m, listFilter, core.PrintVersion)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		return core.ListSpecs(ctx, client, m, listFilter, core.PrintSpec)
	} else if m := names.ArtifactsRegexp().FindStringSubmatch(name); m != nil {
		return core.ListArtifacts(ctx, client, m, listFilter, false, core.PrintArtifact)
	}

	// Then try to match resource names.
	if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
		return core.ListProjects(ctx, client, m, listFilter, core.PrintProject)
	} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return core.ListAPIs(ctx, client, m, listFilter, core.PrintAPI)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return core.ListVersions(ctx, client, m, listFilter, core.PrintVersion)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return core.ListSpecs(ctx, client, m, listFilter, core.PrintSpec)
	} else if m := names.ArtifactRegexp().FindStringSubmatch(name); m != nil {
		return core.ListArtifacts(ctx, client, m, listFilter, false, core.PrintArtifact)
	}

	// If nothing matched, return an error.
	return fmt.Errorf("unsupported argument: %s", name)
}
