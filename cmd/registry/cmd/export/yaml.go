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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func yamlCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "yaml",
		Short: "Export a subtree of the registry to a YAML file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			var name string
			if len(args) > 0 {
				name = args[0]
			}

			if _, err := names.ParseProject(name); err == nil {
				patch.ExportProject(ctx, client, name)
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to export project YAML")
				}
			} else if api, err := names.ParseApi(name); err == nil {
				_, err = core.GetAPI(ctx, client, api, func(message *rpc.Api) {
					bytes, _, err := patch.ExportAPI(ctx, client, message)
					if err != nil {
						log.FromContext(ctx).WithError(err).Fatal("Failed to export API")
					} else {
						fmt.Println(string(bytes))
					}
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to export API YAML")
				}
			} else if artifact, err := names.ParseArtifact(name); err == nil {
				_, err = core.GetArtifact(ctx, client, artifact, false, func(message *rpc.Artifact) {
					bytes, _, err := patch.ExportArtifact(ctx, client, message)
					if err != nil {
						log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact")
					} else {
						fmt.Println(string(bytes))
					}
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
