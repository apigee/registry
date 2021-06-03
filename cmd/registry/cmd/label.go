// Copyright 2021 Google LLC. All Rights Reserved.
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
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"google.golang.org/genproto/protobuf/field_mask"
)

var labelFilter string
var labelOverwrite bool
var labelsToSet map[string]string
var labelsToClear []string

func init() {
	rootCmd.AddCommand(labelCmd)
	labelCmd.Flags().StringVar(&labelFilter, "filter", "", "Filter selected resources")
	labelCmd.Flags().BoolVar(&labelOverwrite, "overwrite", false, "Overwrite existing labels")
}

var labelCmd = &cobra.Command{
	Use:   "label RESOURCE KEY_1=VAL_1 ... KEY_N=VAL_N",
	Short: "Label resources in the API Registry",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		taskQueue := make(chan core.Task, 1024)
		workerCount := 64
		for i := 0; i < workerCount; i++ {
			core.WaitGroup().Add(1)
			go core.Worker(ctx, taskQueue)
		}

		labelsToClear = make([]string, 0)
		labelsToSet = make(map[string]string)
		for _, labeling := range args[1:] {
			if len(labeling) > 1 && strings.HasSuffix(labeling, "-") {
				labelsToClear = append(labelsToClear, strings.TrimSuffix(labeling, "-"))
			} else {
				pair := strings.Split(labeling, "=")
				if len(pair) != 2 {
					log.Fatalf("%q must have the form \"key=value\" (value can be empty) or \"key-\" (to remove the key)", labeling)
				}
				if pair[0] == "" {
					log.Fatalf("%q is invalid because it specifies an empty key", labeling)
				}
				labelsToSet[pair[0]] = pair[1]
			}
		}

		err = matchAndHandleLabelCmd(ctx, client, taskQueue, args[0])
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		close(taskQueue)
		core.WaitGroup().Wait()
	},
}

type labelApiTask struct {
	ctx    context.Context
	client connection.Client
	api    *rpc.Api
}

type labelVersionTask struct {
	ctx     context.Context
	client  connection.Client
	version *rpc.ApiVersion
}

type labelSpecTask struct {
	ctx    context.Context
	client connection.Client
	spec   *rpc.ApiSpec
}

func (task *labelApiTask) String() string {
	return "label " + task.api.Name
}

func (task *labelApiTask) Run() error {
	if task.api.Labels == nil {
		task.api.Labels = make(map[string]string)
	}
	if !labelOverwrite {
		for k, _ := range labelsToSet {
			if v, ok := task.api.Labels[k]; ok {
				return fmt.Errorf("%q already has a value (%s), and --overwrite is false", k, v)
			}
		}
	}
	for _, k := range labelsToClear {
		delete(task.api.Labels, k)
	}
	for k, v := range labelsToSet {
		task.api.Labels[k] = v
	}
	req := &rpc.UpdateApiRequest{
		Api: task.api,
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"labels"},
		},
	}
	_, err := task.client.UpdateApi(task.ctx, req)
	return err
}

func (task *labelVersionTask) String() string {
	return "label " + task.version.Name
}

func (task *labelVersionTask) Run() error {
	if task.version.Labels == nil {
		task.version.Labels = make(map[string]string)
	}
	for _, k := range labelsToClear {
		delete(task.version.Labels, k)
	}
	for k, v := range labelsToSet {
		task.version.Labels[k] = v
	}
	req := &rpc.UpdateApiVersionRequest{
		ApiVersion: task.version,
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"labels"},
		},
	}
	_, err := task.client.UpdateApiVersion(task.ctx, req)
	return err
}

func (task *labelSpecTask) String() string {
	return "label " + task.spec.Name
}

func (task *labelSpecTask) Run() error {
	if task.spec.Labels == nil {
		task.spec.Labels = make(map[string]string)
	}
	for _, k := range labelsToClear {
		delete(task.spec.Labels, k)
	}
	for k, v := range labelsToSet {
		task.spec.Labels[k] = v
	}
	req := &rpc.UpdateApiSpecRequest{
		ApiSpec: task.spec,
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"labels"},
		},
	}
	_, err := task.client.UpdateApiSpec(task.ctx, req)
	return err
}

func matchAndHandleLabelCmd(
	ctx context.Context,
	client connection.Client,
	taskQueue chan<- core.Task,
	name string,
) error {
	// First try to match collection names.
	if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		return labelAPIs(ctx, client, m, labelFilter, taskQueue)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		return labelVersions(ctx, client, m, labelFilter, taskQueue)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		return labelSpecs(ctx, client, m, labelFilter, taskQueue)
	}

	// Then try to match resource names.
	if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return labelAPIs(ctx, client, m, labelFilter, taskQueue)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return labelVersions(ctx, client, m, labelFilter, taskQueue)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return labelSpecs(ctx, client, m, labelFilter, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name %s", name)
	}
}

func labelAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListAPIs(ctx, client, segments, filterFlag, func(api *rpc.Api) {
		taskQueue <- &labelApiTask{
			ctx:    ctx,
			client: client,
			api:    api,
		}
	})
}

func labelVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListVersions(ctx, client, segments, filterFlag, func(version *rpc.ApiVersion) {
		taskQueue <- &labelVersionTask{
			ctx:     ctx,
			client:  client,
			version: version,
		}
	})
}

func labelSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListSpecs(ctx, client, segments, filterFlag, func(spec *rpc.ApiSpec) {
		taskQueue <- &labelSpecTask{
			ctx:    ctx,
			client: client,
			spec:   spec,
		}
	})
}
