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
	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
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
				log.WithError(err).Fatal("Failed to get client")
			}

			var name string
			if len(args) > 0 {
				name = args[0]
			}

			if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
				_, err := core.GetProject(ctx, client, m, func(message *rpc.Project) {
					core.ExportYAMLForProject(ctx, client, message)
				})
				if err != nil {
					log.WithError(err).Fatal("Failed to export project YAML")
				}
			} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
				_, err = core.GetAPI(ctx, client, m, func(message *rpc.Api) {
					core.ExportYAMLForAPI(ctx, client, message)
				})
				if err != nil {
					log.WithError(err).Fatal("Failed to export API YAML")
				}
			} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
				_, err = core.GetVersion(ctx, client, m, func(message *rpc.ApiVersion) {
					core.ExportYAMLForVersion(ctx, client, message)
				})
				if err != nil {
					log.WithError(err).Fatal("Failed to export version YAML")
				}
			} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
				_, err = core.GetSpec(ctx, client, m, false, func(message *rpc.ApiSpec) {
					core.ExportYAMLForSpec(ctx, client, message)
				})
				if err != nil {
					log.WithError(err).Fatal("Failed to export spec YAML")
				}
			} else {
				log.Fatalf("Unsupported entity %+s", name)
			}
		},
	}
}
