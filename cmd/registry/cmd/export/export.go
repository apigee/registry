// Copyright 2020 Google LLC.
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
	"context"
	"errors"

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var filter string
	var recursive bool
	var jobs int
	var root string
	cmd := &cobra.Command{
		Use:   "export PATTERN",
		Short: "Export resources from the API Registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}
			pattern := c.FQName(args[0])
			registryClient, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			adminClient, err := connection.NewAdminClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			// Initialize task queue.
			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()
			// Create the visitor that will perform exports.
			v := &exportVisitor{
				registryClient: registryClient,
				adminClient:    adminClient,
				recursive:      recursive,
				root:           root,
				taskQueue:      taskQueue,
			}
			// Visit the selected resources.
			patternName, err := names.Parse(pattern)
			if err != nil {
				return err
			}
			if err = visitor.Visit(ctx, v, visitor.VisitorOptions{
				RegistryClient:  registryClient,
				AdminClient:     adminClient,
				Pattern:         pattern,
				Filter:          filter,
				ImplicitProject: &rpc.Project{Name: patternName.Project().String()},
			}); err != nil {
				return err
			}
			if v.count == 0 {
				return errors.New("no resources found")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "include child resources in export")
	cmd.Flags().StringVar(&root, "root", "", "root directory for export")
	return cmd
}

type exportVisitor struct {
	registryClient connection.RegistryClient
	adminClient    connection.AdminClient
	recursive      bool
	root           string
	count          int
	taskQueue      chan<- tasks.Task
}

func (h *exportVisitor) ProjectHandler() visitor.ProjectHandler {
	return func(ctx context.Context, message *rpc.Project) error {
		h.count++
		name, err := names.ParseProject(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportProject(ctx, h.registryClient, h.adminClient, name, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) ApiHandler() visitor.ApiHandler {
	return func(ctx context.Context, message *rpc.Api) error {
		h.count++
		name, err := names.ParseApi(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPI(ctx, h.registryClient, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) VersionHandler() visitor.VersionHandler {
	return func(ctx context.Context, message *rpc.ApiVersion) error {
		h.count++
		name, err := names.ParseVersion(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPIVersion(ctx, h.registryClient, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) DeploymentHandler() visitor.DeploymentHandler {
	return func(ctx context.Context, message *rpc.ApiDeployment) error {
		h.count++
		name, err := names.ParseDeployment(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPIDeployment(ctx, h.registryClient, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) DeploymentRevisionHandler() visitor.DeploymentHandler {
	return func(ctx context.Context, message *rpc.ApiDeployment) error {
		return errors.New("exports of specific revisions are not supported")
	}
}

func (h *exportVisitor) SpecHandler() visitor.SpecHandler {
	return func(ctx context.Context, message *rpc.ApiSpec) error {
		h.count++
		name, err := names.ParseSpec(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPISpec(ctx, h.registryClient, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) SpecRevisionHandler() visitor.SpecHandler {
	return func(ctx context.Context, message *rpc.ApiSpec) error {
		return errors.New("exports of specific revisions are not supported")
	}
}

func (h *exportVisitor) ArtifactHandler() visitor.ArtifactHandler {
	return func(ctx context.Context, message *rpc.Artifact) error {
		h.count++
		name, err := names.ParseArtifact(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportArtifact(ctx, h.registryClient, name, h.root, h.taskQueue)
	}
}
