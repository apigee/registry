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
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/connection"
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
			client, err := connection.NewRegistryClientWithSettings(ctx, c)
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
			h := &exportHandler{
				ctx:         ctx,
				client:      client,
				adminClient: adminClient,
				pattern:     pattern,
				filter:      filter,
				recursive:   recursive,
				root:        root,
				taskQueue:   taskQueue,
			}
			if err = h.traverse(); err != nil {
				return err
			}
			if h.count == 0 {
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

type exportHandler struct {
	ctx         context.Context
	client      connection.RegistryClient
	adminClient connection.AdminClient
	pattern     string
	filter      string
	recursive   bool
	root        string
	count       int
	taskQueue   chan<- core.Task
}

func (h *exportHandler) traverse() error {
	// Define aliases to simplify the subsequent code.
	pattern := h.pattern
	ctx := h.ctx
	client := h.client
	adminClient := h.adminClient
	filter := h.filter

	// First try to match collection names.
	if project, err := names.ParseProjectCollection(pattern); err == nil {
		return core.ListProjects(ctx, adminClient, project, filter, h.projectHandler())
	} else if api, err := names.ParseApiCollection(pattern); err == nil {
		return core.ListAPIs(ctx, client, api, filter, h.apiHandler())
	} else if deployment, err := names.ParseDeploymentCollection(pattern); err == nil {
		return core.ListDeployments(ctx, client, deployment, filter, h.apiDeploymentHandler())
	} else if _, err := names.ParseDeploymentRevisionCollection(pattern); err == nil {
		return errors.New("exports of specific revisions are not supported")
	} else if version, err := names.ParseVersionCollection(pattern); err == nil {
		return core.ListVersions(ctx, client, version, filter, h.apiVersionHandler())
	} else if spec, err := names.ParseSpecCollection(pattern); err == nil {
		return core.ListSpecs(ctx, client, spec, filter, false, h.apiSpecHandler())
	} else if _, err := names.ParseSpecRevisionCollection(pattern); err == nil {
		return errors.New("exports of specific revisions are not supported")
	} else if artifact, err := names.ParseArtifactCollection(pattern); err == nil {
		return core.ListArtifacts(ctx, client, artifact, filter, false, h.artifactHandler())
	}

	// Then try to match resource names containing wildcards, these also are treated as collections.
	if strings.Contains(pattern, "/-") || strings.Contains(pattern, "@-") {
		if project, err := names.ParseProject(pattern); err == nil {
			return core.ListProjects(ctx, adminClient, project, filter, h.projectHandler())
		} else if api, err := names.ParseApi(pattern); err == nil {
			return core.ListAPIs(ctx, client, api, filter, h.apiHandler())
		} else if deployment, err := names.ParseDeployment(pattern); err == nil {
			return core.ListDeployments(ctx, client, deployment, filter, h.apiDeploymentHandler())
		} else if _, err := names.ParseDeploymentRevision(pattern); err == nil {
			return errors.New("exports of specific revisions are not supported")
		} else if version, err := names.ParseVersion(pattern); err == nil {
			return core.ListVersions(ctx, client, version, filter, h.apiVersionHandler())
		} else if spec, err := names.ParseSpec(pattern); err == nil {
			return core.ListSpecs(ctx, client, spec, filter, false, h.apiSpecHandler())
		} else if _, err := names.ParseSpecRevision(pattern); err == nil {
			return errors.New("exports of specific revisions are not supported")
		} else if artifact, err := names.ParseArtifact(pattern); err == nil {
			return core.ListArtifacts(ctx, client, artifact, filter, false, h.artifactHandler())
		}
		return fmt.Errorf("unsupported pattern %+v", pattern)
	}

	// If we get here, name designates an individual resource to be displayed.
	// So if a filter was specified, that's an error.
	if filter != "" {
		return errors.New("--filter must not be specified for a non-collection resource")
	}

	if project, err := names.ParseProject(pattern); err == nil {
		return core.GetProject(ctx, adminClient, project, true, h.projectHandler())
	} else if api, err := names.ParseApi(pattern); err == nil {
		return core.GetAPI(ctx, client, api, h.apiHandler())
	} else if deployment, err := names.ParseDeployment(pattern); err == nil {
		return core.GetDeployment(ctx, client, deployment, h.apiDeploymentHandler())
	} else if _, err := names.ParseDeploymentRevision(pattern); err == nil {
		return errors.New("exports of specific revisions are not supported")
	} else if version, err := names.ParseVersion(pattern); err == nil {
		return core.GetVersion(ctx, client, version, h.apiVersionHandler())
	} else if spec, err := names.ParseSpec(pattern); err == nil {
		return core.GetSpec(ctx, client, spec, false, h.apiSpecHandler())
	} else if _, err := names.ParseSpecRevision(pattern); err == nil {
		return errors.New("exports of specific revisions are not supported")
	} else if artifact, err := names.ParseArtifact(pattern); err == nil {
		return core.GetArtifact(ctx, client, artifact, false, h.artifactHandler())
	} else {
		return fmt.Errorf("unsupported pattern %+v", pattern)
	}
}

func (h *exportHandler) projectHandler() func(message *rpc.Project) error {
	return func(message *rpc.Project) error {
		h.count++
		name, err := names.ParseProject(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportProject(h.ctx, h.client, name, h.root, h.taskQueue)
	}
}

func (h *exportHandler) apiHandler() func(message *rpc.Api) error {
	return func(message *rpc.Api) error {
		h.count++
		name, err := names.ParseApi(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPI(h.ctx, h.client, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportHandler) apiVersionHandler() func(message *rpc.ApiVersion) error {
	return func(message *rpc.ApiVersion) error {
		h.count++
		name, err := names.ParseVersion(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPIVersion(h.ctx, h.client, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportHandler) apiDeploymentHandler() func(message *rpc.ApiDeployment) error {
	return func(message *rpc.ApiDeployment) error {
		h.count++
		name, err := names.ParseDeployment(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPIDeployment(h.ctx, h.client, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportHandler) apiSpecHandler() func(message *rpc.ApiSpec) error {
	return func(message *rpc.ApiSpec) error {
		h.count++
		name, err := names.ParseSpec(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportAPISpec(h.ctx, h.client, name, h.recursive, h.root, h.taskQueue)
	}
}

func (h *exportHandler) artifactHandler() func(message *rpc.Artifact) error {
	return func(message *rpc.Artifact) error {
		h.count++
		name, err := names.ParseArtifact(message.Name)
		if err != nil {
			return err
		}
		return patch.ExportArtifact(h.ctx, h.client, name, h.root, h.taskQueue)
	}
}
