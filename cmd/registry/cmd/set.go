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
)

var setFilter string
var setLabelID string

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.Flags().StringVar(&setFilter, "filter", "", "Filter selected resources")
	setCmd.Flags().StringVar(&setLabelID, "label_id", "", "Label to set on selected resources")
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set labels and artifacts on resources in the API Registry",
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

		err = matchAndHandleSetCmd(ctx, client, taskQueue, args[0])
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		close(taskQueue)
		core.WaitGroup().Wait()
	},
}

type setTask struct {
	ctx          context.Context
	client       connection.Client
	resourceName string
	resourceKind string
}

func (task *setTask) Name() string {
	return "set " + task.resourceKind + " " + task.resourceName
}

func (task *setTask) Run() error {
	if setLabelID != "" {
		// todo: set labels
	}
	return nil
}

func matchAndHandleSetCmd(
	ctx context.Context,
	client connection.Client,
	taskQueue chan core.Task,
	name string,
) error {
	if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
		return setProjects(ctx, client, m, setFilter, taskQueue)
	} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return setAPIs(ctx, client, m, setFilter, taskQueue)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return setVersions(ctx, client, m, setFilter, taskQueue)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return setSpecs(ctx, client, m, setFilter, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name %s", name)
	}
}

func setProjects(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan core.Task) error {
	return core.ListProjects(ctx, client, segments, filterFlag, func(project *rpc.Project) {
		taskQueue <- &setTask{
			ctx:          ctx,
			client:       client,
			resourceName: project.Name,
			resourceKind: "project",
		}
	})
}

func setAPIs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan core.Task) error {
	return core.ListAPIs(ctx, client, segments, filterFlag, func(api *rpc.Api) {
		taskQueue <- &setTask{
			ctx:          ctx,
			client:       client,
			resourceName: api.Name,
			resourceKind: "api",
		}
	})
}

func setVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan core.Task) error {
	return core.ListVersions(ctx, client, segments, filterFlag, func(version *rpc.ApiVersion) {
		taskQueue <- &setTask{
			ctx:          ctx,
			client:       client,
			resourceName: version.Name,
			resourceKind: "version",
		}
	})
}

func setSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan core.Task) error {
	return core.ListSpecs(ctx, client, segments, filterFlag, func(spec *rpc.ApiSpec) {
		taskQueue <- &setTask{
			ctx:          ctx,
			client:       client,
			resourceName: spec.Name,
			resourceKind: "spec",
		}
	})
}
