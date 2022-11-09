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
	"fmt"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var filter string
	var jobs int
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources from the API Registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			args[0] = c.FQName(args[0])

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			// Initialize task queue.
			taskQueue, wait := core.WorkerPool(ctx, jobs)
			defer wait()

			err = matchAndHandleDeleteCmd(ctx, client, taskQueue, args[0], filter)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to match or handle command")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	cmd.Flags().IntVar(&jobs, "jobs", 10, "Number of actions to perform concurrently")
	return cmd
}

type deleteTask struct {
	client       connection.RegistryClient
	resourceName string
	resourceKind string
}

func (task *deleteTask) String() string {
	return "delete " + task.resourceName
}

func (task *deleteTask) Run(ctx context.Context) error {
	log.Debugf(ctx, "Deleting %s %s", task.resourceKind, task.resourceName)
	switch task.resourceKind {
	case "api":
		return task.client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: task.resourceName})
	case "version":
		return task.client.DeleteApiVersion(ctx, &rpc.DeleteApiVersionRequest{Name: task.resourceName})
	case "spec":
		return task.client.DeleteApiSpec(ctx, &rpc.DeleteApiSpecRequest{Name: task.resourceName})
	case "artifact":
		return task.client.DeleteArtifact(ctx, &rpc.DeleteArtifactRequest{Name: task.resourceName})
	default:
		return nil
	}
}

func matchAndHandleDeleteCmd(
	ctx context.Context,
	client connection.RegistryClient,
	taskQueue chan<- core.Task,
	name string,
	filter string,
) error {
	if api, err := names.ParseApi(name); err == nil {
		return deleteAPIs(ctx, client, api, filter, taskQueue)
	} else if version, err := names.ParseVersion(name); err == nil {
		return deleteVersions(ctx, client, version, filter, taskQueue)
	} else if spec, err := names.ParseSpec(name); err == nil {
		return deleteSpecs(ctx, client, spec, filter, taskQueue)
	} else if artifact, err := names.ParseArtifact(name); err == nil {
		return deleteArtifacts(ctx, client, artifact, filter, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name: see the 'registry rpc delete-' subcommands for alternatives")
	}
}

func deleteAPIs(
	ctx context.Context,
	client *gapic.RegistryClient,
	api names.Api,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListAPIs(ctx, client, api, filterFlag, func(api *rpc.Api) error {
		taskQueue <- &deleteTask{
			client:       client,
			resourceName: api.Name,
			resourceKind: "api",
		}
		return nil
	})
}

func deleteVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	version names.Version,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListVersions(ctx, client, version, filterFlag, func(version *rpc.ApiVersion) error {
		taskQueue <- &deleteTask{
			client:       client,
			resourceName: version.Name,
			resourceKind: "version",
		}
		return nil
	})
}

func deleteSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	spec names.Spec,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListSpecs(ctx, client, spec, filterFlag, func(spec *rpc.ApiSpec) error {
		taskQueue <- &deleteTask{
			client:       client,
			resourceName: spec.Name,
			resourceKind: "spec",
		}
		return nil
	})
}

func deleteArtifacts(
	ctx context.Context,
	client *gapic.RegistryClient,
	artifact names.Artifact,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListArtifacts(ctx, client, artifact, filterFlag, false, func(artifact *rpc.Artifact) error {
		taskQueue <- &deleteTask{
			client:       client,
			resourceName: artifact.Name,
			resourceKind: "artifact",
		}
		return nil
	})
}
