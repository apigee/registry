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
	client connection.RegistryClient
	name   string
}

func NewDeleteApiTask(client connection.RegistryClient, name string) *DeleteApiTask {
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
	return task.client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: task.name, Force: true})
}

// DeleteArtifactTask deletes a specified artifact.
type DeleteArtifactTask struct {
	client connection.RegistryClient
	name   string
}

func NewDeleteArtifactTask(client connection.RegistryClient, name string) *DeleteArtifactTask {
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
