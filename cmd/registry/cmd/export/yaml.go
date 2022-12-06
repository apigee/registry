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
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

func yamlCommand() *cobra.Command {
	var jobs int
	var nested bool
	cmd := &cobra.Command{
		Use:   "yaml RESOURCE",
		Short: "Export a subtree of the registry as YAML",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}
			args[0] = c.FQName(args[0])

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}

			taskQueue, wait := core.WorkerPool(ctx, jobs)
			defer wait()

			if project, err := names.ParseProject(args[0]); err == nil {
				err = patch.ExportProject(ctx, client, project, taskQueue)
				if err != nil {
					return err
				}
			} else if api, err := names.ParseApi(c.FQName(args[0])); err == nil {
				err = core.GetAPI(ctx, client, api, func(message *rpc.Api) error {
					bytes, _, err := patch.ExportAPI(ctx, client, message, nested)
					if err != nil {
						return err
					}
					_, err = cmd.OutOrStdout().Write(bytes)
					return err
				})
				if err != nil {
					return err
				}
			} else if version, err := names.ParseVersion(c.FQName(args[0])); err == nil {
				err = core.GetVersion(ctx, client, version, func(message *rpc.ApiVersion) error {
					bytes, _, err := patch.ExportAPIVersion(ctx, client, message, nested)
					if err != nil {
						return err
					}
					_, err = cmd.OutOrStdout().Write(bytes)
					return err
				})
				if err != nil {
					return err
				}
			} else if spec, err := names.ParseSpec(c.FQName(args[0])); err == nil {
				err = core.GetSpec(ctx, client, spec, false, func(message *rpc.ApiSpec) error {
					bytes, _, err := patch.ExportAPISpec(ctx, client, message, nested)
					if err != nil {
						return err
					}
					_, err = cmd.OutOrStdout().Write(bytes)
					return err
				})
				if err != nil {
					return err
				}
			} else if deployment, err := names.ParseDeployment(c.FQName(args[0])); err == nil {
				err = core.GetDeployment(ctx, client, deployment, func(message *rpc.ApiDeployment) error {
					bytes, _, err := patch.ExportAPIDeployment(ctx, client, message, nested)
					if err != nil {
						return err
					}
					_, err = cmd.OutOrStdout().Write(bytes)
					return err
				})
				if err != nil {
					return err
				}
			} else if artifact, err := names.ParseArtifact(c.FQName(args[0])); err == nil {
				err = core.GetArtifact(ctx, client, artifact, false, func(message *rpc.Artifact) error {
					bytes, _, err := patch.ExportArtifact(ctx, client, message)
					if err != nil {
						return err
					}
					_, err = cmd.OutOrStdout().Write(bytes)
					return err
				})
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Unsupported entity %+s", args[0])
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "Number of file exports to perform simultaneously")
	cmd.Flags().BoolVarP(&nested, "nested", "n", false, "Nest child resources in parents")
	return cmd
}
