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
	"context"
	"errors"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
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
			taskQueue, wait := core.WorkerPool(ctx, jobs)
			defer wait()
			// Create the visitor that will perform exports.
			v := &exportVisitor{
				ctx:            ctx,
				registryClient: registryClient,
				adminClient:    adminClient,
				recursive:      recursive,
				root:           root,
				taskQueue:      taskQueue,
			}
			// Visit the selected resources.
			if err = visitor.Visit(ctx, v, visitor.VisitorOptions{
				RegistryClient:  registryClient,
				AdminClient:     adminClient,
				Pattern:         pattern,
				Filter:          filter,
				ImplicitProject: &rpc.Project{Name: "Implicit"},
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
	ctx            context.Context
	registryClient connection.RegistryClient
	adminClient    connection.AdminClient
	recursive      bool
	root           string
	count          int
	taskQueue      chan<- core.Task
}

func (h *exportVisitor) ProjectHandler() visitor.ProjectHandler {
	return func(message *rpc.Project) error {
		h.count++
		name, err := names.ParseProject(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportProject(h.ctx, h.registryClient, name, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) ApiHandler() visitor.ApiHandler {
	return func(message *rpc.Api) error {
		h.count++
		name, err := names.ParseApi(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPI(h.ctx, h.registryClient, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) VersionHandler() visitor.VersionHandler {
	return func(message *rpc.ApiVersion) error {
		h.count++
		name, err := names.ParseVersion(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPIVersion(h.ctx, h.registryClient, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) DeploymentHandler() visitor.DeploymentHandler {
	return func(message *rpc.ApiDeployment) error {
		h.count++
		name, err := names.ParseDeployment(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPIDeployment(h.ctx, h.registryClient, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) DeploymentRevisionHandler() visitor.DeploymentHandler {
	return func(message *rpc.ApiDeployment) error {
		return errors.New("exports of specific revisions are not supported")
	}
}

func (h *exportVisitor) SpecHandler() visitor.SpecHandler {
	return func(message *rpc.ApiSpec) error {
		h.count++
		name, err := names.ParseSpec(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPISpec(h.ctx, h.registryClient, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportVisitor) SpecRevisionHandler() visitor.SpecHandler {
	return func(message *rpc.ApiSpec) error {
		return errors.New("exports of specific revisions are not supported")
	}
}

func (h *exportVisitor) ArtifactHandler() visitor.ArtifactHandler {
	return func(message *rpc.Artifact) error {
		h.count++
		name, err := names.ParseArtifact(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportArtifact(h.ctx, h.registryClient, name, h.root, h.taskQueue)
	}
}
