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
	"log"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources from the API Registry",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}

			// Initialize task queue.
			taskQueue, wait := core.WorkerPool(ctx, 64)
			defer wait()

			err = matchAndHandleDeleteCmd(ctx, client, taskQueue, args[0], filter)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	return cmd
}

type deleteTask struct {
	client       connection.Client
	resourceName string
	resourceKind string
}

func (task *deleteTask) String() string {
	return "delete " + task.resourceName
}

func (task *deleteTask) Run(ctx context.Context) error {
	log.Printf("deleting %s %s", task.resourceKind, task.resourceName)
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
	client connection.Client,
	taskQueue chan<- core.Task,
	name string,
	filter string,
) error {
	if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return deleteAPIs(ctx, client, m, filter, taskQueue)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return deleteVersions(ctx, client, m, filter, taskQueue)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return deleteSpecs(ctx, client, m, filter, taskQueue)
	} else if m := names.ArtifactRegexp().FindStringSubmatch(name); m != nil {
		return deleteArtifacts(ctx, client, m, filter, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name: see the 'apg registry delete-' subcommands for alternatives")
	}
}

func deleteAPIs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListAPIs(ctx, client, segments, filterFlag, func(api *rpc.Api) {
		taskQueue <- &deleteTask{
			client:       client,
			resourceName: api.Name,
			resourceKind: "api",
		}
	})
}

func deleteVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListVersions(ctx, client, segments, filterFlag, func(version *rpc.ApiVersion) {
		taskQueue <- &deleteTask{
			client:       client,
			resourceName: version.Name,
			resourceKind: "version",
		}
	})
}

func deleteSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListSpecs(ctx, client, segments, filterFlag, func(spec *rpc.ApiSpec) {
		taskQueue <- &deleteTask{
			client:       client,
			resourceName: spec.Name,
			resourceKind: "spec",
		}
	})
}

func deleteArtifacts(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListArtifacts(ctx, client, segments, filterFlag, false, func(artifact *rpc.Artifact) {
		taskQueue <- &deleteTask{
			client:       client,
			resourceName: artifact.Name,
			resourceKind: "artifact",
		}
	})
}
