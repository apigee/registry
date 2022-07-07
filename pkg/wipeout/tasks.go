// Copyright 2022 Google LLC. All Rights Reserved.
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

package wipeout

import (
	"context"

	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
)

// DeleteApiTask deletes a specified API.
type DeleteApiTask struct {
	client connection.Client
	name   string
}

func NewDeleteApiTask(client connection.Client, name string) *DeleteApiTask {
	return &DeleteApiTask{
		client: client,
		name:   name,
	}
}

func (task *DeleteApiTask) String() string {
	return "delete " + task.name
}

func (task *DeleteApiTask) Run(ctx context.Context) error {
	log.Debugf(ctx, "Deleting %s", task.name)
	return task.client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: task.name})
}

// DeleteVersionTask deletes a specified version.
type DeleteVersionTask struct {
	client connection.Client
	name   string
}

func NewDeleteVersionTask(client connection.Client, name string) *DeleteVersionTask {
	return &DeleteVersionTask{
		client: client,
		name:   name,
	}
}

func (task *DeleteVersionTask) String() string {
	return "delete " + task.name
}

func (task *DeleteVersionTask) Run(ctx context.Context) error {
	log.Debugf(ctx, "Deleting %s", task.name)
	return task.client.DeleteApiVersion(ctx, &rpc.DeleteApiVersionRequest{Name: task.name})
}

// DeleteSpecTask deletes a specified spec.
type DeleteSpecTask struct {
	client connection.Client
	name   string
}

func NewDeleteSpecTask(client connection.Client, name string) *DeleteSpecTask {
	return &DeleteSpecTask{
		client: client,
		name:   name,
	}
}

func (task *DeleteSpecTask) String() string {
	return "delete " + task.name
}

func (task *DeleteSpecTask) Run(ctx context.Context) error {
	log.Debugf(ctx, "Deleting %s", task.name)
	return task.client.DeleteApiSpec(ctx, &rpc.DeleteApiSpecRequest{Name: task.name})
}

// DeleteDeploymentTask deletes a specified deployment.
type DeleteDeploymentTask struct {
	client connection.Client
	name   string
}

func NewDeleteDeploymentTask(client connection.Client, name string) *DeleteDeploymentTask {
	return &DeleteDeploymentTask{
		client: client,
		name:   name,
	}
}

func (task *DeleteDeploymentTask) String() string {
	return "delete " + task.name
}

func (task *DeleteDeploymentTask) Run(ctx context.Context) error {
	log.Debugf(ctx, "Deleting %s", task.name)
	return task.client.DeleteApiDeployment(ctx, &rpc.DeleteApiDeploymentRequest{Name: task.name})
}

// DeleteArtifactTask deletes a specified artifact.
type DeleteArtifactTask struct {
	client connection.Client
	name   string
}

func NewDeleteArtifactTask(client connection.Client, name string) *DeleteArtifactTask {
	return &DeleteArtifactTask{
		client: client,
		name:   name,
	}
}

func (task *DeleteArtifactTask) String() string {
	return "delete " + task.name
}

func (task *DeleteArtifactTask) Run(ctx context.Context) error {
	log.Debugf(ctx, "Deleting %s", task.name)
	return task.client.DeleteArtifact(ctx, &rpc.DeleteArtifactRequest{Name: task.name})
}
