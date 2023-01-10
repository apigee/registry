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

package delete

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var filter string
	var jobs int
	var force bool

	cmd := &cobra.Command{
		Use:   "delete PATTERN",
		Short: "Delete resources from the API Registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			args[0] = c.FQName(args[0])
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
			h := &deleteHandler{
				ctx:         ctx,
				client:      client,
				adminClient: adminClient,
				name:        args[0],
				filter:      filter,
				force:       force,
				taskQueue:   taskQueue,
			}
			err = h.traverse()
			if err != nil {
				return err
			}
			if h.count == 0 {
				return errors.New("no resources found")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().IntVar(&jobs, "jobs", 10, "number of actions to perform concurrently")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion of child resources")
	return cmd
}

type deleteHandler struct {
	ctx         context.Context
	client      connection.RegistryClient
	adminClient connection.AdminClient
	name        string
	filter      string
	force       bool
	count       int
	taskQueue   chan<- core.Task
}

func (h *deleteHandler) traverse() error {
	// Define aliases to simplify the subsequent code.
	name := h.name
	ctx := h.ctx
	client := h.client
	adminClient := h.adminClient
	filter := h.filter

	// First try to match collection names.
	if project, err := names.ParseProjectCollection(name); err == nil {
		return core.ListProjects(ctx, adminClient, project, filter, h.projectHandler())
	} else if api, err := names.ParseApiCollection(name); err == nil {
		return core.ListAPIs(ctx, client, api, filter, h.apiHandler())
	} else if deployment, err := names.ParseDeploymentCollection(name); err == nil {
		return core.ListDeployments(ctx, client, deployment, filter, h.apiDeploymentHandler())
	} else if rev, err := names.ParseDeploymentRevisionCollection(name); err == nil {
		return core.ListDeploymentRevisions(ctx, client, rev, filter, h.apiDeploymentRevisionHandler())
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		return core.ListVersions(ctx, client, version, filter, h.apiVersionHandler())
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		return core.ListSpecs(ctx, client, spec, filter, false, h.apiSpecHandler())
	} else if rev, err := names.ParseSpecRevisionCollection(name); err == nil {
		return core.ListSpecRevisions(ctx, client, rev, filter, false, h.apiSpecRevisionHandler())
	} else if artifact, err := names.ParseArtifactCollection(name); err == nil {
		return core.ListArtifacts(ctx, client, artifact, filter, false, h.artifactHandler())
	}

	// Then try to match resource names containing wildcards, these also are treated as collections.
	if strings.Contains(name, "/-") || strings.Contains(name, "@-") {
		if project, err := names.ParseProject(name); err == nil {
			return core.ListProjects(ctx, adminClient, project, filter, h.projectHandler())
		} else if api, err := names.ParseApi(name); err == nil {
			return core.ListAPIs(ctx, client, api, filter, h.apiHandler())
		} else if deployment, err := names.ParseDeployment(name); err == nil {
			return core.ListDeployments(ctx, client, deployment, filter, h.apiDeploymentHandler())
		} else if rev, err := names.ParseDeploymentRevision(name); err == nil {
			return core.ListDeploymentRevisions(ctx, client, rev, filter, h.apiDeploymentRevisionHandler())
		} else if version, err := names.ParseVersion(name); err == nil {
			return core.ListVersions(ctx, client, version, filter, h.apiVersionHandler())
		} else if spec, err := names.ParseSpec(name); err == nil {
			return core.ListSpecs(ctx, client, spec, filter, false, h.apiSpecHandler())
		} else if rev, err := names.ParseSpecRevision(name); err == nil {
			return core.ListSpecRevisions(ctx, client, rev, filter, false, h.apiSpecRevisionHandler())
		} else if artifact, err := names.ParseArtifact(name); err == nil {
			return core.ListArtifacts(ctx, client, artifact, filter, false, h.artifactHandler())
		}
		return fmt.Errorf("unsupported pattern %+v", name)
	}

	// If we get here, name designates an individual resource to be displayed.
	// So if a filter was specified, that's an error.
	if filter != "" {
		return errors.New("--filter must not be specified for a non-collection resource")
	}

	if project, err := names.ParseProject(name); err == nil {
		return core.GetProject(ctx, adminClient, project, h.projectHandler())
	} else if api, err := names.ParseApi(name); err == nil {
		return core.GetAPI(ctx, client, api, h.apiHandler())
	} else if deployment, err := names.ParseDeployment(name); err == nil {
		return core.GetDeployment(ctx, client, deployment, h.apiDeploymentHandler())
	} else if deployment, err := names.ParseDeploymentRevision(name); err == nil {
		return core.GetDeploymentRevision(ctx, client, deployment, h.apiDeploymentRevisionHandler())
	} else if version, err := names.ParseVersion(name); err == nil {
		return core.GetVersion(ctx, client, version, h.apiVersionHandler())
	} else if spec, err := names.ParseSpec(name); err == nil {
		return core.GetSpec(ctx, client, spec, false, h.apiSpecHandler())
	} else if spec, err := names.ParseSpecRevision(name); err == nil {
		return core.GetSpecRevision(ctx, client, spec, false, h.apiSpecRevisionHandler())
	} else if artifact, err := names.ParseArtifact(name); err == nil {
		return core.GetArtifact(ctx, client, artifact, false, h.artifactHandler())
	} else {
		return fmt.Errorf("unsupported pattern %+v", name)
	}
}

func (h *deleteHandler) enqueue(task core.Task) {
	h.count++
	h.taskQueue <- task
}

func (h *deleteHandler) projectHandler() func(message *rpc.Project) error {
	return func(message *rpc.Project) error {
		h.enqueue(&deleteProjectTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     h.adminClient,
			force:      h.force,
		})
		return nil
	}
}

func (h *deleteHandler) apiHandler() func(message *rpc.Api) error {
	return func(message *rpc.Api) error {
		h.enqueue(&deleteApiTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     h.client,
			force:      h.force,
		})
		return nil
	}
}

func (h *deleteHandler) apiVersionHandler() func(message *rpc.ApiVersion) error {
	return func(message *rpc.ApiVersion) error {
		h.enqueue(&deleteApiVersionTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     h.client,
			force:      h.force,
		})
		return nil
	}
}

func (h *deleteHandler) apiDeploymentHandler() func(message *rpc.ApiDeployment) error {
	return func(message *rpc.ApiDeployment) error {
		h.enqueue(&deleteApiDeploymentTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     h.client,
			force:      h.force,
		})
		return nil
	}
}

func (h *deleteHandler) apiDeploymentRevisionHandler() func(message *rpc.ApiDeployment) error {
	return func(message *rpc.ApiDeployment) error {
		h.enqueue(&deleteApiDeploymentRevisionTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     h.client,
			force:      h.force,
		})
		return nil
	}
}

func (h *deleteHandler) apiSpecHandler() func(message *rpc.ApiSpec) error {
	return func(message *rpc.ApiSpec) error {
		h.enqueue(&deleteApiSpecTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     h.client,
			force:      h.force,
		})
		return nil
	}
}

func (h *deleteHandler) apiSpecRevisionHandler() func(message *rpc.ApiSpec) error {
	return func(message *rpc.ApiSpec) error {
		h.enqueue(&deleteApiSpecRevisionTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     h.client,
			force:      h.force,
		})
		return nil
	}
}

func (h *deleteHandler) artifactHandler() func(message *rpc.Artifact) error {
	return func(message *rpc.Artifact) error {
		h.enqueue(&deleteArtifactTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     h.client,
		})
		return nil
	}
}

type deleteTask struct {
	resourceName string
}

func (task *deleteTask) String() string {
	return "delete " + task.resourceName
}

func (task *deleteTask) log(ctx context.Context) {
	log.Debugf(ctx, "Deleting %s", task.resourceName)
}

type deleteProjectTask struct {
	deleteTask
	client connection.AdminClient
	force  bool
}

func (task *deleteProjectTask) Run(ctx context.Context) error {
	task.log(ctx)
	return task.client.DeleteProject(ctx, &rpc.DeleteProjectRequest{Name: task.resourceName, Force: task.force})
}

type deleteApiTask struct {
	deleteTask
	client connection.RegistryClient
	force  bool
}

func (task *deleteApiTask) Run(ctx context.Context) error {
	task.log(ctx)
	return task.client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: task.resourceName, Force: task.force})
}

type deleteApiVersionTask struct {
	deleteTask
	client connection.RegistryClient
	force  bool
}

func (task *deleteApiVersionTask) Run(ctx context.Context) error {
	task.log(ctx)
	return task.client.DeleteApiVersion(ctx, &rpc.DeleteApiVersionRequest{Name: task.resourceName, Force: task.force})
}

type deleteApiSpecTask struct {
	deleteTask
	client connection.RegistryClient
	force  bool
}

func (task *deleteApiSpecTask) Run(ctx context.Context) error {
	task.log(ctx)
	return task.client.DeleteApiSpec(ctx, &rpc.DeleteApiSpecRequest{Name: task.resourceName, Force: task.force})
}

type deleteApiSpecRevisionTask struct {
	deleteTask
	client connection.RegistryClient
	force  bool
}

func (task *deleteApiSpecRevisionTask) Run(ctx context.Context) error {
	task.log(ctx)
	_, err := task.client.DeleteApiSpecRevision(ctx, &rpc.DeleteApiSpecRevisionRequest{Name: task.resourceName})
	return err
}

type deleteApiDeploymentTask struct {
	deleteTask
	client connection.RegistryClient
	force  bool
}

func (task *deleteApiDeploymentTask) Run(ctx context.Context) error {
	task.log(ctx)
	return task.client.DeleteApiDeployment(ctx, &rpc.DeleteApiDeploymentRequest{Name: task.resourceName, Force: task.force})
}

type deleteApiDeploymentRevisionTask struct {
	deleteTask
	client connection.RegistryClient
	force  bool
}

func (task *deleteApiDeploymentRevisionTask) Run(ctx context.Context) error {
	task.log(ctx)
	_, err := task.client.DeleteApiDeploymentRevision(ctx, &rpc.DeleteApiDeploymentRevisionRequest{Name: task.resourceName})
	return err
}

type deleteArtifactTask struct {
	deleteTask
	client connection.RegistryClient
}

func (task *deleteArtifactTask) Run(ctx context.Context) error {
	task.log(ctx)
	return task.client.DeleteArtifact(ctx, &rpc.DeleteArtifactRequest{Name: task.resourceName})
}
