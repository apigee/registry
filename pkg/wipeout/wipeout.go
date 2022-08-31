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
	// Wipeout resources in groups to ensure that children are deleted before parents.
	{
		log.Debugf(ctx, "Deleting artifacts")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutArtifacts(ctx, client, taskQueue, project+"/apis/-/versions/-/specs/-")
		wipeoutArtifacts(ctx, client, taskQueue, project+"/apis/-/versions/-")
		wipeoutArtifacts(ctx, client, taskQueue, project+"/apis/-/deployments/-")
		wipeoutArtifacts(ctx, client, taskQueue, project+"/apis/-")
		wipeoutArtifacts(ctx, client, taskQueue, project)
		wait()
	}
	{
		log.Debugf(ctx, "Deleting specs")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutApiSpecs(ctx, client, taskQueue, project+"/apis/-/versions/-")
		wait()
	}
	{
		log.Debugf(ctx, "Deleting versions")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutApiVersions(ctx, client, taskQueue, project+"/apis/-")
		wait()
	}
	{
		log.Debugf(ctx, "Deleting deployments")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutApiDeployments(ctx, client, taskQueue, project+"/apis/-")
		wait()
	}
	{
		log.Debugf(ctx, "Deleting apis")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutApis(ctx, client, taskQueue, project)
		wait()
	}
	log.Debugf(ctx, "Wipeout complete")
}

func wipeoutArtifacts(ctx context.Context, client connection.RegistryClient, taskQueue chan<- core.Task, parent string) {
	it := client.ListArtifacts(ctx, &rpc.ListArtifactsRequest{Parent: parent})
	names := make([]string, 0)
	for artifact, err := it.Next(); err == nil; artifact, err = it.Next() {
		names = append(names, artifact.Name)
	}
	for _, name := range names {
		taskQueue <- NewDeleteArtifactTask(client, name)
	}
}

func wipeoutApiDeployments(ctx context.Context, client connection.RegistryClient, taskQueue chan<- core.Task, parent string) {
	it := client.ListApiDeployments(ctx, &rpc.ListApiDeploymentsRequest{Parent: parent})
	names := make([]string, 0)
	for deployment, err := it.Next(); err == nil; deployment, err = it.Next() {
		names = append(names, deployment.Name)
	}
	for _, name := range names {
		taskQueue <- NewDeleteDeploymentTask(client, name)
	}
}

func wipeoutApiSpecs(ctx context.Context, client connection.RegistryClient, taskQueue chan<- core.Task, parent string) {
	it := client.ListApiSpecs(ctx, &rpc.ListApiSpecsRequest{Parent: parent})
	names := make([]string, 0)
	for spec, err := it.Next(); err == nil; spec, err = it.Next() {
		names = append(names, spec.Name)
	}
	for _, name := range names {
		taskQueue <- NewDeleteSpecTask(client, name)
	}
}

func wipeoutApiVersions(ctx context.Context, client connection.RegistryClient, taskQueue chan<- core.Task, parent string) {
	it := client.ListApiVersions(ctx, &rpc.ListApiVersionsRequest{Parent: parent})
	names := make([]string, 0)
	for version, err := it.Next(); err == nil; version, err = it.Next() {
		names = append(names, version.Name)
	}
	for _, name := range names {
		taskQueue <- NewDeleteVersionTask(client, name)
	}
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
