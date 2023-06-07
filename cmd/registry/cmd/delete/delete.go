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

package delete

import (
	"context"
	"errors"

	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
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
			// Create the visitor that will perform deletion.
			v := &deletionVisitor{
				registryClient: registryClient,
				adminClient:    adminClient,
				force:          force,
			}
			// Visit the selected resources.
			if err = visitor.Visit(ctx, v, visitor.VisitorOptions{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
				Pattern:        pattern,
				Filter:         filter,
			}); err != nil {
				return err
			}
			if len(v.tasks) == 0 {
				return errors.New("no resources found")
			}
			// Initialize task queue.
			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()
			// Delete all of the resources that were found.
			for _, task := range v.tasks {
				taskQueue <- task
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force deletion of child resources")
	return cmd
}

type deletionVisitor struct {
	registryClient connection.RegistryClient
	adminClient    connection.AdminClient
	force          bool
	tasks          []tasks.Task
}

func (v *deletionVisitor) enqueue(task tasks.Task) {
	v.tasks = append(v.tasks, task)
}

func (v *deletionVisitor) ProjectHandler() visitor.ProjectHandler {
	return func(ctx context.Context, message *rpc.Project) error {
		v.enqueue(&deleteProjectTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     v.adminClient,
			force:      v.force,
		})
		return nil
	}
}

func (v *deletionVisitor) ApiHandler() visitor.ApiHandler {
	return func(ctx context.Context, message *rpc.Api) error {
		v.enqueue(&deleteApiTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     v.registryClient,
			force:      v.force,
		})
		return nil
	}
}

func (v *deletionVisitor) VersionHandler() visitor.VersionHandler {
	return func(ctx context.Context, message *rpc.ApiVersion) error {
		v.enqueue(&deleteApiVersionTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     v.registryClient,
			force:      v.force,
		})
		return nil
	}
}

func (v *deletionVisitor) DeploymentHandler() visitor.DeploymentHandler {
	return func(ctx context.Context, message *rpc.ApiDeployment) error {
		v.enqueue(&deleteApiDeploymentTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     v.registryClient,
			force:      v.force,
		})
		return nil
	}
}

func (v *deletionVisitor) DeploymentRevisionHandler() visitor.DeploymentHandler {
	return func(ctx context.Context, message *rpc.ApiDeployment) error {
		v.enqueue(&deleteApiDeploymentRevisionTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     v.registryClient,
			force:      v.force,
		})
		return nil
	}
}

func (v *deletionVisitor) SpecHandler() visitor.SpecHandler {
	return func(ctx context.Context, message *rpc.ApiSpec) error {
		v.enqueue(&deleteApiSpecTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     v.registryClient,
			force:      v.force,
		})
		return nil
	}
}

func (v *deletionVisitor) SpecRevisionHandler() visitor.SpecHandler {
	return func(ctx context.Context, message *rpc.ApiSpec) error {
		v.enqueue(&deleteApiSpecRevisionTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     v.registryClient,
			force:      v.force,
		})
		return nil
	}
}

func (v *deletionVisitor) ArtifactHandler() visitor.ArtifactHandler {
	return func(ctx context.Context, message *rpc.Artifact) error {
		v.enqueue(&deleteArtifactTask{
			deleteTask: deleteTask{resourceName: message.Name},
			client:     v.registryClient,
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
