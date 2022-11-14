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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources in the API Registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			args[0] = c.FQName(args[0])

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			adminClient, err := connection.NewAdminClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			err = matchAndHandleListCmd(ctx, client, adminClient, args[0], filter)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to match or handle command")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	return cmd
}

func matchAndHandleListCmd(
	ctx context.Context,
	client connection.RegistryClient,
	adminClient connection.AdminClient,
	name string,
	filter string,
) error {
	// First try to match collection names.
	if project, err := names.ParseProjectCollection(name); err == nil {
		return core.ListProjects(ctx, adminClient, project, filter, core.PrintProject)
	} else if api, err := names.ParseApiCollection(name); err == nil {
		return core.ListAPIs(ctx, client, api, filter, core.PrintAPI)
	} else if deployment, err := names.ParseDeploymentCollection(name); err == nil {
		return core.ListDeployments(ctx, client, deployment, filter, core.PrintDeployment)
	} else if rev, err := names.ParseDeploymentRevisionCollection(name); err == nil {
		return core.ListDeploymentRevisions(ctx, client, rev, filter, core.PrintDeployment)
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		return core.ListVersions(ctx, client, version, filter, core.PrintVersion)
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		return core.ListSpecs(ctx, client, spec, filter, core.PrintSpec)
	} else if rev, err := names.ParseSpecRevisionCollection(name); err == nil {
		return core.ListSpecRevisions(ctx, client, rev, filter, core.PrintSpec)
	} else if artifact, err := names.ParseArtifactCollection(name); err == nil {
		return core.ListArtifacts(ctx, client, artifact, filter, false, core.PrintArtifact)
	}

	// Then try to match resource names.
	if project, err := names.ParseProject(name); err == nil {
		return core.ListProjects(ctx, adminClient, project, filter, core.PrintProject)
	} else if api, err := names.ParseApi(name); err == nil {
		return core.ListAPIs(ctx, client, api, filter, core.PrintAPI)
	} else if deployment, err := names.ParseDeployment(name); err == nil {
		return core.ListDeployments(ctx, client, deployment, filter, core.PrintDeployment)
	} else if rev, err := names.ParseDeploymentRevision(name); err == nil {
		return core.ListDeploymentRevisions(ctx, client, rev, filter, core.PrintDeployment)
	} else if version, err := names.ParseVersion(name); err == nil {
		return core.ListVersions(ctx, client, version, filter, core.PrintVersion)
	} else if spec, err := names.ParseSpec(name); err == nil {
		return core.ListSpecs(ctx, client, spec, filter, core.PrintSpec)
	} else if rev, err := names.ParseSpecRevision(name); err == nil {
		return core.ListSpecRevisions(ctx, client, rev, filter, core.PrintSpec)
	} else if artifact, err := names.ParseArtifact(name); err == nil {
		return core.ListArtifacts(ctx, client, artifact, filter, false, core.PrintArtifact)
	}

	// If nothing matched, return an error.
	return fmt.Errorf("unsupported argument: %s", name)
}
