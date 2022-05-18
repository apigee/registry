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

package export

import (
	"fmt"

	"github.com/apigee/registry/cmd/regctl/core"
	"github.com/apigee/registry/cmd/regctl/patch"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

func yamlCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "yaml",
		Short: "Export a subtree of the registry to a YAML file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			var name string
			if len(args) > 0 {
				name = args[0]
			}

			if project, err := names.ParseProject(name); err == nil {
				err = patch.ExportProject(ctx, client, project)
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to export project YAML")
				}
			} else if api, err := names.ParseApi(name); err == nil {
				err = core.GetAPI(ctx, client, api, func(message *rpc.Api) error {
					bytes, _, err := patch.ExportAPI(ctx, client, message)
					if err != nil {
						log.FromContext(ctx).WithError(err).Fatal("Failed to export API")
					}
					fmt.Println(string(bytes))
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to export API YAML")
				}
			} else if artifact, err := names.ParseArtifact(name); err == nil {
				err = core.GetArtifact(ctx, client, artifact, false, func(message *rpc.Artifact) error {
					bytes, _, err := patch.ExportArtifact(ctx, client, message)
					if err != nil {
						log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact")
					}

					fmt.Println(string(bytes))
					return nil
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact YAML")
				}
			} else {
				log.Fatalf(ctx, "Unsupported entity %+s", name)
			}
		},
	}
}
