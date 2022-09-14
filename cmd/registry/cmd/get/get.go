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
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var getContents bool
	var getRawContents bool
	var getPrintedContents bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get resources from the API Registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			args[0] = c.FQName(args[0])
			if getContents && getRawContents || getContents && getPrintedContents || getRawContents && getPrintedContents {
				log.FromContext(ctx).Fatal("Please use at most one of --print, --raw, and --contents.")
			}
			if getContents {
				getPrintedContents = true
				log.FromContext(ctx).Warn("--contents is deprecated, please use --print or --raw instead.")
			}

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			adminClient, err := connection.NewAdminClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			var err2 error
			if project, err := names.ParseProject(args[0]); err == nil {
				err2 = core.GetProject(ctx, adminClient, project, core.PrintProjectDetail)
			} else if api, err := names.ParseApi(args[0]); err == nil {
				err2 = core.GetAPI(ctx, client, api, core.PrintAPIDetail)
			} else if deployment, err := names.ParseDeployment(args[0]); err == nil {
				err2 = core.GetDeployment(ctx, client, deployment, core.PrintDeploymentDetail)
			} else if deployment, err := names.ParseDeploymentRevision(args[0]); err == nil {
				err2 = core.GetDeploymentRevision(ctx, client, deployment, core.PrintDeploymentDetail)
			} else if version, err := names.ParseVersion(args[0]); err == nil {
				err2 = core.GetVersion(ctx, client, version, core.PrintVersionDetail)
			} else if spec, err := names.ParseSpec(args[0]); err == nil {
				// for specs, these options are synonymous
				if getPrintedContents || getRawContents {
					err2 = core.GetSpec(ctx, client, spec, true, core.WriteSpecContents)
				} else {
					err2 = core.GetSpec(ctx, client, spec, false, core.PrintSpecDetail)
				}
			} else if spec, err := names.ParseSpecRevision(args[0]); err == nil {
				// for specs, these options are synonymous
				if getPrintedContents || getRawContents {
					err2 = core.GetSpecRevision(ctx, client, spec, true, core.WriteSpecContents)
				} else {
					err2 = core.GetSpecRevision(ctx, client, spec, false, core.PrintSpecDetail)
				}
			} else if artifact, err := names.ParseArtifact(args[0]); err == nil {
				if getPrintedContents {
					err2 = core.GetArtifact(ctx, client, artifact, true, core.PrintArtifactContents)
				} else if getRawContents {
					err2 = core.GetArtifact(ctx, client, artifact, true, core.WriteArtifactContents)
				} else {
					err2 = core.GetArtifact(ctx, client, artifact, false, core.PrintArtifactDetail)
				}
			} else {
				log.Debugf(ctx, "Unsupported entity %+v", args)
			}
			if err2 != nil {
				log.FromContext(ctx).WithError(err2).Debugf("Failed to get resource")
			}
		},
	}

	cmd.Flags().BoolVar(&getContents, "contents", false, "Get resource contents if available (deprecated)")
	cmd.Flags().BoolVar(&getRawContents, "raw", false, "Get raw resource contents if available")
	cmd.Flags().BoolVar(&getPrintedContents, "print", false, "Print resource contents if available")
	return cmd
}
