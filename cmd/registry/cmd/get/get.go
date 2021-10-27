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

package get

import (
	"context"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	var getContents bool
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get resources from the API Registry",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}

			var name string
			if len(args) > 0 {
				name = args[0]
			}

			if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
				_, err = core.GetProject(ctx, adminClient, m, core.PrintProjectDetail)
			} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
				_, err = core.GetAPI(ctx, client, m, core.PrintAPIDetail)
			} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
				_, err = core.GetVersion(ctx, client, m, core.PrintVersionDetail)
			} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
				if getContents {
					_, err = core.GetSpec(ctx, client, m, getContents, core.PrintSpecContents)
				} else {
					_, err = core.GetSpec(ctx, client, m, getContents, core.PrintSpecDetail)
				}
			} else if artifact, err := names.ParseArtifact(name); err == nil {
				if getContents {
					_, err = core.GetArtifact(ctx, client, artifact, getContents, core.PrintArtifactContents)
				} else {
					_, err = core.GetArtifact(ctx, client, artifact, getContents, core.PrintArtifactDetail)
				}
			} else {
				log.Debugf("Unsupported entity %+v", args)
			}
			if err != nil {
				log.WithError(err).Debugf("Failed to get resource")
			}
		},
	}

	cmd.Flags().BoolVar(&getContents, "contents", false, "Include resource contents if available")
	return cmd
}
