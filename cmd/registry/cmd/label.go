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

const labelFieldName = "labels"
const labelCommandName = "label"

func init() {
	rootCmd.AddCommand(labelCmd())
}

func labelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s RESOURCE KEY_1=VAL_1 ... KEY_N=VAL_N", labelCommandName),
		Short: fmt.Sprintf("%s resources in the API Registry", strings.Title(labelCommandName)),
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			flagset := cmd.LocalFlags()
			filter, err := flagset.GetString("filter")
			if err != nil {
				log.Fatalf("Failed to get filter string from flags: %s", err)
			}
			overwrite := false
			overwrite, err = flagset.GetBool("overwrite")
			if err != nil {
				log.Fatalf("Failed to get overwrite boolean from flags: %s", err)
			}

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

			valuesToClear := make([]string, 0)
			valuesToSet := make(map[string]string)
			for _, operation := range args[1:] {
				if len(operation) > 1 && strings.HasSuffix(operation, "-") {
					valuesToClear = append(valuesToClear, strings.TrimSuffix(operation, "-"))
				} else {
					pair := strings.Split(operation, "=")
					if len(pair) != 2 {
						log.Fatalf("%q must have the form \"key=value\" (value can be empty) or \"key-\" (to remove the key)", operation)
					}
					if pair[0] == "" {
						log.Fatalf("%q is invalid because it specifies an empty key", operation)
					}
					valuesToSet[pair[0]] = pair[1]
				}
			}
			labeling := &core.Labeling{Overwrite: overwrite, Set: valuesToSet, Clear: valuesToClear}

			err = matchAndHandleLabelCmd(ctx, client, taskQueue, args[0], filter, labeling)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}

			close(taskQueue)
			core.WaitGroup().Wait()
		},
	}
	cmd.Flags().String("filter", "", "Filter selected resources")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing labels")
	return cmd
}

func matchAndHandleLabelCmd(
	ctx context.Context,
	client connection.Client,
	taskQueue chan<- core.Task,
	name string,
	filter string,
	labeling *core.Labeling,
) error {
	// First try to match collection names.
	if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		return labelAPIs(ctx, client, m, filter, labeling, taskQueue)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		return labelVersions(ctx, client, m, filter, labeling, taskQueue)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		return labelSpecs(ctx, client, m, filter, labeling, taskQueue)
	}

	// Then try to match resource names.
	if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return labelAPIs(ctx, client, m, filter, labeling, taskQueue)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return labelVersions(ctx, client, m, filter, labeling, taskQueue)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return labelSpecs(ctx, client, m, filter, labeling, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name %s", name)
	}
}

func labelAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListAPIs(ctx, client, segments, filterFlag, func(api *rpc.Api) {
		taskQueue <- &labelApiTask{
			ctx:      ctx,
			client:   client,
			api:      api,
			labeling: labeling,
		}
	})
}

func labelVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListVersions(ctx, client, segments, filterFlag, func(version *rpc.ApiVersion) {
		taskQueue <- &labelVersionTask{
			ctx:      ctx,
			client:   client,
			version:  version,
			labeling: labeling,
		}
	})
}

func labelSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	labeling *core.Labeling,
	taskQueue chan<- core.Task) error {
	return core.ListSpecs(ctx, client, segments, filterFlag, func(spec *rpc.ApiSpec) {
		taskQueue <- &labelSpecTask{
			ctx:      ctx,
			client:   client,
			spec:     spec,
			labeling: labeling,
		}
	})
}

type labelApiTask struct {
	ctx      context.Context
	client   connection.Client
	api      *rpc.Api
	labeling *core.Labeling
}

func (task *labelApiTask) String() string {
	return labelCommandName + " " + task.api.Name
}

func (task *labelApiTask) Run() error {
	var err error
	task.api.Labels, err = task.labeling.Apply(task.api.Labels)
	if err != nil {
		return err
	}
	_, err = task.client.UpdateApi(task.ctx,
		&rpc.UpdateApiRequest{
			Api: task.api,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{labelFieldName},
			},
		})
	return err
}

type labelVersionTask struct {
	ctx      context.Context
	client   connection.Client
	version  *rpc.ApiVersion
	labeling *core.Labeling
}

func (task *labelVersionTask) String() string {
	return labelCommandName + " " + task.version.Name
}

func (task *labelVersionTask) Run() error {
	var err error
	task.version.Labels, err = task.labeling.Apply(task.version.Labels)
	if err != nil {
		return err
	}
	_, err = task.client.UpdateApiVersion(task.ctx,
		&rpc.UpdateApiVersionRequest{
			ApiVersion: task.version,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{labelFieldName},
			},
		})
	return err
}

type labelSpecTask struct {
	ctx      context.Context
	client   connection.Client
	spec     *rpc.ApiSpec
	labeling *core.Labeling
}

func (task *labelSpecTask) String() string {
	return labelCommandName + " " + task.spec.Name
}

func (task *labelSpecTask) Run() error {
	var err error
	task.spec.Labels, err = task.labeling.Apply(task.spec.Labels)
	if err != nil {
		return err
	}
	_, err = task.client.UpdateApiSpec(task.ctx,
		&rpc.UpdateApiSpecRequest{
			ApiSpec: task.spec,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{labelFieldName},
			},
		})
	return err
}
