// Copyright 2022 Google LLC.
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

	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/rpc"
)

// Wipeout deletes all resources in a project using the specified number of parallel worker jobs.
// Errors are currently not returned but are logged by the task queue as fatal errors.
func Wipeout(ctx context.Context, client connection.RegistryClient, projectID string, jobs int) {
	log.Debugf(ctx, "Deleting everything in project %s", projectID)
	project := "projects/" + projectID + "/locations/global"
	{
		log.Debugf(ctx, "Deleting apis")
		taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
		wipeoutApis(ctx, client, taskQueue, project)
		wait()
	}
	{
		log.Debugf(ctx, "Deleting artifacts")
		taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
		wipeoutArtifacts(ctx, client, taskQueue, project)
		wait()
	}
	log.Debugf(ctx, "Wipeout complete")
}

func wipeoutArtifacts(ctx context.Context, client connection.RegistryClient, taskQueue chan<- tasks.Task, parent string) {
	it := client.ListArtifacts(ctx, &rpc.ListArtifactsRequest{Parent: parent})
	names := make([]string, 0)
	for artifact, err := it.Next(); err == nil; artifact, err = it.Next() {
		names = append(names, artifact.Name)
	}
	for _, name := range names {
		taskQueue <- NewDeleteArtifactTask(client, name)
	}
}

func wipeoutApis(ctx context.Context, client connection.RegistryClient, taskQueue chan<- tasks.Task, parent string) {
	it := client.ListApis(ctx, &rpc.ListApisRequest{Parent: parent})
	names := make([]string, 0)
	for api, err := it.Next(); err == nil; api, err = it.Next() {
		names = append(names, api.Name)
	}
	for _, name := range names {
		taskQueue <- NewDeleteApiTask(client, name)
	}
}
