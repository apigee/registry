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

var commandIsLabel bool
var keyFilter string
var keyOverwrite bool
var keysToSet map[string]string
var keysToClear []string

// fieldName returns the name of the resource field to be modified.
func fieldName() string {
	if commandIsLabel {
		return "labels"
	}
	return "annotations"
}

// fieldName returns the name of the command being handled.
func commandName() string {
	if commandIsLabel {
		return "label"
	}
	return "annotate"
}

// updateMap updates the map containing the labels or annotations to be modified.
func updateMap(m map[string]string) (map[string]string, error) {
	if m == nil {
		m = make(map[string]string)
	}
	if !keyOverwrite {
		for k, _ := range keysToSet {
			if v, ok := m[k]; ok {
				return nil, fmt.Errorf("%q already has a value (%s), and --overwrite is false", k, v)
			}
		}
	}
	for _, k := range keysToClear {
		delete(m, k)
	}
	for k, v := range keysToSet {
		m[k] = v
	}
	return m, nil
}

func init() {
	var labelCmd = labelOrAnnotateCommand(true)
	rootCmd.AddCommand(labelCmd)
	labelCmd.Flags().StringVar(&keyFilter, "filter", "", "Filter selected resources")
	labelCmd.Flags().BoolVar(&keyOverwrite, "overwrite", false, "Overwrite existing labels")

	var annotateCmd = labelOrAnnotateCommand(false)
	rootCmd.AddCommand(annotateCmd)
	annotateCmd.Flags().StringVar(&keyFilter, "filter", "", "Filter selected resources")
	annotateCmd.Flags().BoolVar(&keyOverwrite, "overwrite", false, "Overwrite existing annotations")
}

func labelOrAnnotateCommand(commandIsLabel bool) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s RESOURCE KEY_1=VAL_1 ... KEY_N=VAL_N", commandName()),
		Short: fmt.Sprintf("%s resources in the API Registry", strings.Title(commandName())),
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

			keysToClear = make([]string, 0)
			keysToSet = make(map[string]string)
			for _, operation := range args[1:] {
				if len(operation) > 1 && strings.HasSuffix(operation, "-") {
					keysToClear = append(keysToClear, strings.TrimSuffix(operation, "-"))
				} else {
					pair := strings.Split(operation, "=")
					if len(pair) != 2 {
						log.Fatalf("%q must have the form \"key=value\" (value can be empty) or \"key-\" (to remove the key)", operation)
					}
					if pair[0] == "" {
						log.Fatalf("%q is invalid because it specifies an empty key", operation)
					}
					keysToSet[pair[0]] = pair[1]
				}
			}

			err = matchAndHandleLabelOrAnnotateCmd(ctx, client, taskQueue, args[0])
			if err != nil {
				log.Fatalf("%s", err.Error())
			}

			close(taskQueue)
			core.WaitGroup().Wait()
		},
	}
}

type labelOrAnnotateApiTask struct {
	ctx    context.Context
	client connection.Client
	api    *rpc.Api
}

func (task *labelOrAnnotateApiTask) String() string {
	return commandName() + " " + task.api.Name
}

func (task *labelOrAnnotateApiTask) Run() error {
	var err error
	if commandIsLabel {
		task.api.Labels, err = updateMap(task.api.Labels)
	} else {
		task.api.Annotations, err = updateMap(task.api.Annotations)
	}
	if err != nil {
		return err
	}
	_, err = task.client.UpdateApi(task.ctx,
		&rpc.UpdateApiRequest{
			Api: task.api,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{fieldName()},
			},
		})
	return err
}

type labelOrAnnotateVersionTask struct {
	ctx     context.Context
	client  connection.Client
	version *rpc.ApiVersion
}

func (task *labelOrAnnotateVersionTask) String() string {
	return commandName() + " " + task.version.Name
}

func (task *labelOrAnnotateVersionTask) Run() error {
	var err error
	if commandIsLabel {
		task.version.Labels, err = updateMap(task.version.Labels)
	} else {
		task.version.Annotations, err = updateMap(task.version.Annotations)
	}
	if err != nil {
		return err
	}
	_, err = task.client.UpdateApiVersion(task.ctx,
		&rpc.UpdateApiVersionRequest{
			ApiVersion: task.version,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{fieldName()},
			},
		})
	return err
}

type labelOrAnnotateSpecTask struct {
	ctx    context.Context
	client connection.Client
	spec   *rpc.ApiSpec
}

func (task *labelOrAnnotateSpecTask) String() string {
	return commandName() + " " + task.spec.Name
}

func (task *labelOrAnnotateSpecTask) Run() error {
	var err error
	if commandIsLabel {
		task.spec.Labels, err = updateMap(task.spec.Labels)
	} else {
		task.spec.Annotations, err = updateMap(task.spec.Annotations)
	}
	if err != nil {
		return err
	}
	_, err = task.client.UpdateApiSpec(task.ctx,
		&rpc.UpdateApiSpecRequest{
			ApiSpec: task.spec,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{fieldName()},
			},
		})
	return err
}

func matchAndHandleLabelOrAnnotateCmd(
	ctx context.Context,
	client connection.Client,
	taskQueue chan<- core.Task,
	name string,
) error {
	// First try to match collection names.
	if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		return labelOrAnnotateAPIs(ctx, client, m, keyFilter, taskQueue)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		return labelOrAnnotateVersions(ctx, client, m, keyFilter, taskQueue)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		return labelOrAnnotateSpecs(ctx, client, m, keyFilter, taskQueue)
	}

	// Then try to match resource names.
	if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return labelOrAnnotateAPIs(ctx, client, m, keyFilter, taskQueue)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return labelOrAnnotateVersions(ctx, client, m, keyFilter, taskQueue)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return labelOrAnnotateSpecs(ctx, client, m, keyFilter, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name %s", name)
	}
}

func labelOrAnnotateAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListAPIs(ctx, client, segments, filterFlag, func(api *rpc.Api) {
		taskQueue <- &labelOrAnnotateApiTask{
			ctx:    ctx,
			client: client,
			api:    api,
		}
	})
}

func labelOrAnnotateVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListVersions(ctx, client, segments, filterFlag, func(version *rpc.ApiVersion) {
		taskQueue <- &labelOrAnnotateVersionTask{
			ctx:     ctx,
			client:  client,
			version: version,
		}
	})
}

func labelOrAnnotateSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListSpecs(ctx, client, segments, filterFlag, func(spec *rpc.ApiSpec) {
		taskQueue <- &labelOrAnnotateSpecTask{
			ctx:    ctx,
			client: client,
			spec:   spec,
		}
	})
}
