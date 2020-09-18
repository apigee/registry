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

package cmd

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
	"google.golang.org/grpc/status"
)

var deleteFilter string

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVar(&deleteFilter, "filter", "", "Filter resources to delete")
}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete matching entities and their children.",
	Long:  "Delete matching entities and their children.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		// Initialize task queue.
		taskQueue := make(chan core.Task, 1024)
		workerCount := 64
		for i := 0; i < workerCount; i++ {
			core.WaitGroup().Add(1)
			go core.Worker(ctx, taskQueue)
		}

		err = matchAndHandleDeleteCmd(ctx, client, taskQueue, args[0])
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				log.Fatalf("%s", err.Error())
			} else {
				log.Fatalf("%s", st.Message())
			}
		}

		close(taskQueue)
		core.WaitGroup().Wait()
	},
}

type deleteTask struct {
	ctx          context.Context
	client       connection.Client
	resourceName string
	resourceKind string
}

func (task *deleteTask) Run() error {
	log.Printf("deleting %s %s", task.resourceKind, task.resourceName)
	switch task.resourceKind {
	case "api":
		return task.client.DeleteApi(task.ctx, &rpc.DeleteApiRequest{Name: task.resourceName})
	case "version":
		return task.client.DeleteVersion(task.ctx, &rpc.DeleteVersionRequest{Name: task.resourceName})
	case "spec":
		return task.client.DeleteSpec(task.ctx, &rpc.DeleteSpecRequest{Name: task.resourceName})
	case "property":
		return task.client.DeleteProperty(task.ctx, &rpc.DeletePropertyRequest{Name: task.resourceName})
	case "label":
		return task.client.DeleteLabel(task.ctx, &rpc.DeleteLabelRequest{Name: task.resourceName})
	default:
		return nil
	}
}

func matchAndHandleDeleteCmd(
	ctx context.Context,
	client connection.Client,
	taskQueue chan core.Task,
	name string,
) error {
	if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return deleteAPIs(ctx, client, m, deleteFilter, taskQueue)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return deleteVersions(ctx, client, m, deleteFilter, taskQueue)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return deleteSpecs(ctx, client, m, deleteFilter, taskQueue)
	} else if m := names.PropertyRegexp().FindStringSubmatch(name); m != nil {
		return deleteProperties(ctx, client, m, deleteFilter, taskQueue)
	} else if m := names.LabelRegexp().FindStringSubmatch(name); m != nil {
		return deleteLabels(ctx, client, m, deleteFilter, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name: see the 'apg registry delete-' subcommands for alternatives")
	}
}

func deleteAPIs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan core.Task) error {
	return core.ListAPIs(ctx, client, segments, filterFlag, func(api *rpc.Api) {
		taskQueue <- &deleteTask{
			ctx:          ctx,
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
	taskQueue chan core.Task) error {
	return core.ListVersions(ctx, client, segments, filterFlag, func(version *rpc.Version) {
		taskQueue <- &deleteTask{
			ctx:          ctx,
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
	taskQueue chan core.Task) error {
	return core.ListSpecs(ctx, client, segments, filterFlag, func(spec *rpc.Spec) {
		taskQueue <- &deleteTask{
			ctx:          ctx,
			client:       client,
			resourceName: spec.Name,
			resourceKind: "spec",
		}
	})
}

func deleteProperties(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan core.Task) error {
	return core.ListProperties(ctx, client, segments, filterFlag, func(property *rpc.Property) {
		taskQueue <- &deleteTask{
			ctx:          ctx,
			client:       client,
			resourceName: property.Name,
			resourceKind: "property",
		}
	})
}

func deleteLabels(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan core.Task) error {
	return core.ListLabels(ctx, client, segments, filterFlag, func(label *rpc.Label) {
		taskQueue <- &deleteTask{
			ctx:          ctx,
			client:       client,
			resourceName: label.Name,
			resourceKind: "label",
		}
	})
}
