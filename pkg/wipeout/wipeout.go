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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
)

// Wipeout deletes all resources in a project using the specified number of parallel worker jobs.
// Errors are currently not returned but are logged by the task queue as fatal errors.
func Wipeout(ctx context.Context, client connection.RegistryClient, projectID string, jobs int) {
	log.Debugf(ctx, "Deleting everything in project %s", projectID)
	project := "projects/" + projectID + "/locations/global"

	log.Debugf(ctx, "Deleting apis")
	taskQueue, wait := core.WorkerPool(ctx, jobs)
	wipeoutApis(ctx, client, taskQueue, project)
	wait()

	log.Debugf(ctx, "Wipeout complete")
}

func wipeoutApis(ctx context.Context, client connection.RegistryClient, taskQueue chan<- core.Task, parent string) {
	it := client.ListApis(ctx, &rpc.ListApisRequest{Parent: parent})
	names := make([]string, 0)
	for api, err := it.Next(); err == nil; api, err = it.Next() {
		names = append(names, api.Name)
	}
	for _, name := range names {
		taskQueue <- NewDeleteApiTask(client, name)
	}
}

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
