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
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
)

// This collects one (arbitrary) error that occurs during wipeout.
// It's a global and gets assigned from concurrent goroutines,
// but if it fails, it fails.
var _err error

// Wipeout deletes all resources in a project using the specified number of parallel worker jobs.
func Wipeout(ctx context.Context, client connection.Client, projectid string, jobs int) error {
	log.Infof(ctx, "Deleting everything in project %s", projectid)
	project := "projects/" + projectid + "/locations/global"
	// Wipeout resources in groups to ensure that children are deleted before parents.
	{
		log.Infof(ctx, "Deleting artifacts")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutArtifacts(ctx, client, taskQueue, project+"/apis/-/versions/-/specs/-")
		wipeoutArtifacts(ctx, client, taskQueue, project+"/apis/-/versions/-")
		wipeoutArtifacts(ctx, client, taskQueue, project+"/apis/-/deployments/-")
		wipeoutArtifacts(ctx, client, taskQueue, project+"/apis/-")
		wipeoutArtifacts(ctx, client, taskQueue, project)
		wait()
	}
	{
		log.Infof(ctx, "Deleting specs")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutApiSpecs(ctx, client, taskQueue, project+"/apis/-/versions/-")
		wait()
	}
	{
		log.Infof(ctx, "Deleting versions")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutApiVersions(ctx, client, taskQueue, project+"/apis/-")
		wait()
	}
	{
		log.Infof(ctx, "Deleting deployments")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutApiDeployments(ctx, client, taskQueue, project+"/apis/-")
		wait()
	}
	{
		log.Infof(ctx, "Deleting apis")
		taskQueue, wait := core.WorkerPool(ctx, jobs)
		wipeoutApis(ctx, client, taskQueue, project)
		wait()
	}
	log.Infof(ctx, "Wipeout complete with err=%+v", _err)
	return _err
}

type ResourceKind int

const (
	API ResourceKind = iota
	Version
	Spec
	Deployment
	Artifact
)

type deleteResourceTask struct {
	client connection.Client
	name   string
	kind   ResourceKind
}

func (task *deleteResourceTask) String() string {
	return "delete " + task.name
}

func (task *deleteResourceTask) Run(ctx context.Context) error {
	var err error
	switch task.kind {
	case API:
		log.Infof(ctx, "Deleting %s", task.name)
		err = task.client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: task.name})
	case Version:
		log.Infof(ctx, "Deleting %s", task.name)
		err = task.client.DeleteApiVersion(ctx, &rpc.DeleteApiVersionRequest{Name: task.name})
	case Spec:
		log.Infof(ctx, "Deleting %s", task.name)
		err = task.client.DeleteApiSpec(ctx, &rpc.DeleteApiSpecRequest{Name: task.name})
	case Deployment:
		log.Infof(ctx, "Deleting %s", task.name)
		err = task.client.DeleteApiDeployment(ctx, &rpc.DeleteApiDeploymentRequest{Name: task.name})
	case Artifact:
		log.Infof(ctx, "Deleting %s", task.name)
		err = task.client.DeleteArtifact(ctx, &rpc.DeleteArtifactRequest{Name: task.name})
	default:
		log.Infof(ctx, "Unknown resource type %s", task.name)
	}
	if err != nil {
		log.FromContext(ctx).WithError(err).Errorf("Deletion %s failed", task.name)
		_err = err
	}
	return nil
}

func wipeoutArtifacts(ctx context.Context, registryClient connection.Client, taskQueue chan<- core.Task, parent string) {
	it := registryClient.ListArtifacts(ctx, &rpc.ListArtifactsRequest{Parent: parent})
	for artifact, err := it.Next(); err == nil; artifact, err = it.Next() {
		taskQueue <- &deleteResourceTask{
			client: registryClient,
			name:   artifact.Name,
			kind:   Artifact,
		}
	}
}

func wipeoutApiDeployments(ctx context.Context, registryClient connection.Client, taskQueue chan<- core.Task, parent string) {
	it := registryClient.ListApiDeployments(ctx, &rpc.ListApiDeploymentsRequest{Parent: parent})
	for deployment, err := it.Next(); err == nil; deployment, err = it.Next() {
		taskQueue <- &deleteResourceTask{
			client: registryClient,
			name:   deployment.Name,
			kind:   Deployment,
		}
	}
}

func wipeoutApiSpecs(ctx context.Context, registryClient connection.Client, taskQueue chan<- core.Task, parent string) {
	it := registryClient.ListApiSpecs(ctx, &rpc.ListApiSpecsRequest{Parent: parent})
	for spec, err := it.Next(); err == nil; spec, err = it.Next() {
		taskQueue <- &deleteResourceTask{
			client: registryClient,
			name:   spec.Name,
			kind:   Spec,
		}
	}
}

func wipeoutApiVersions(ctx context.Context, registryClient connection.Client, taskQueue chan<- core.Task, parent string) {
	it := registryClient.ListApiVersions(ctx, &rpc.ListApiVersionsRequest{Parent: parent})
	for version, err := it.Next(); err == nil; version, err = it.Next() {
		taskQueue <- &deleteResourceTask{
			client: registryClient,
			name:   version.Name,
			kind:   Version,
		}
	}
}

func wipeoutApis(ctx context.Context, registryClient connection.Client, taskQueue chan<- core.Task, parent string) {
	it := registryClient.ListApis(ctx, &rpc.ListApisRequest{Parent: parent})
	for api, err := it.Next(); err == nil; api, err = it.Next() {
		taskQueue <- &deleteResourceTask{
			client: registryClient,
			name:   api.Name,
			kind:   API,
		}
	}
}
