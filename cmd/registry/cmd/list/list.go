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

package list

import (
	"context"
	"fmt"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/servers/registry/names"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources in the API Registry",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}
			err = matchAndHandleListCmd(ctx, client, args[0], filter)
			if err != nil {
				log.WithError(err).Fatal("Failed to match or handle command")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	return cmd
}

func matchAndHandleListCmd(
	ctx context.Context,
	client connection.Client,
	name string,
	filter string,
) error {

	// First try to match collection names.
	if m := names.ProjectsRegexp().FindStringSubmatch(name); m != nil {
		return core.ListProjects(ctx, client, m, filter, core.PrintProject)
	} else if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		return core.ListAPIs(ctx, client, m, filter, core.PrintAPI)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		return core.ListVersions(ctx, client, m, filter, core.PrintVersion)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		return core.ListSpecs(ctx, client, m, filter, core.PrintSpec)
	} else if m := names.ArtifactsRegexp().FindStringSubmatch(name); m != nil {
		return core.ListArtifacts(ctx, client, m, filter, false, core.PrintArtifact)
	}

	// Then try to match resource names.
	if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
		return core.ListProjects(ctx, client, m, filter, core.PrintProject)
	} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return core.ListAPIs(ctx, client, m, filter, core.PrintAPI)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return core.ListVersions(ctx, client, m, filter, core.PrintVersion)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return core.ListSpecs(ctx, client, m, filter, core.PrintSpec)
	} else if m := names.ArtifactRegexp().FindStringSubmatch(name); m != nil {
		return core.ListArtifacts(ctx, client, m, filter, false, core.PrintArtifact)
	}

	// If nothing matched, return an error.
	return fmt.Errorf("unsupported argument: %s", name)
}
