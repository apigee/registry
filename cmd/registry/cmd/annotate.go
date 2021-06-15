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

var annotateKeyFilter string
var annotateKeyOverwrite bool
var annotateKeysToSet map[string]string
var annotateKeysToClear []string

const annotateFieldName = "annotations"
const annotateCommandName = "annotate"

func init() {
	rootCmd.AddCommand(annotateCmd)
	annotateCmd.Flags().StringVar(&annotateKeyFilter, "filter", "", "Filter selected resources")
	annotateCmd.Flags().BoolVar(&annotateKeyOverwrite, "overwrite", false, "Overwrite existing annotations")
}

var annotateCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s RESOURCE KEY_1=VAL_1 ... KEY_N=VAL_N", annotateCommandName),
	Short: fmt.Sprintf("%s resources in the API Registry", strings.Title(annotateCommandName)),
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

		annotateKeysToClear = make([]string, 0)
		annotateKeysToSet = make(map[string]string)
		for _, operation := range args[1:] {
			if len(operation) > 1 && strings.HasSuffix(operation, "-") {
				annotateKeysToClear = append(annotateKeysToClear, strings.TrimSuffix(operation, "-"))
			} else {
				pair := strings.Split(operation, "=")
				if len(pair) != 2 {
					log.Fatalf("%q must have the form \"key=value\" (value can be empty) or \"key-\" (to remove the key)", operation)
				}
				if pair[0] == "" {
					log.Fatalf("%q is invalid because it specifies an empty key", operation)
				}
				annotateKeysToSet[pair[0]] = pair[1]
			}
		}

		err = matchAndHandleAnnotateCmd(ctx, client, taskQueue, args[0])
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		close(taskQueue)
		core.WaitGroup().Wait()
	},
}

func matchAndHandleAnnotateCmd(
	ctx context.Context,
	client connection.Client,
	taskQueue chan<- core.Task,
	name string,
) error {
	// First try to match collection names.
	if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		return annotateAPIs(ctx, client, m, annotateKeyFilter, taskQueue)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		return annotateVersions(ctx, client, m, annotateKeyFilter, taskQueue)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		return annotateSpecs(ctx, client, m, annotateKeyFilter, taskQueue)
	}

	// Then try to match resource names.
	if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		return annotateAPIs(ctx, client, m, annotateKeyFilter, taskQueue)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		return annotateVersions(ctx, client, m, annotateKeyFilter, taskQueue)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		return annotateSpecs(ctx, client, m, annotateKeyFilter, taskQueue)
	} else {
		return fmt.Errorf("unsupported resource name %s", name)
	}
}

func annotateAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListAPIs(ctx, client, segments, filterFlag, func(api *rpc.Api) {
		taskQueue <- &annotateApiTask{
			ctx:    ctx,
			client: client,
			api:    api,
		}
	})
}

func annotateVersions(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListVersions(ctx, client, segments, filterFlag, func(version *rpc.ApiVersion) {
		taskQueue <- &annotateVersionTask{
			ctx:     ctx,
			client:  client,
			version: version,
		}
	})
}

func annotateSpecs(
	ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	taskQueue chan<- core.Task) error {
	return core.ListSpecs(ctx, client, segments, filterFlag, func(spec *rpc.ApiSpec) {
		taskQueue <- &annotateSpecTask{
			ctx:    ctx,
			client: client,
			spec:   spec,
		}
	})
}

type annotateApiTask struct {
	ctx    context.Context
	client connection.Client
	api    *rpc.Api
}

func (task *annotateApiTask) String() string {
	return annotateCommandName + " " + task.api.Name
}

func (task *annotateApiTask) Run() error {
	var err error
	task.api.Annotations, err = core.UpdateMap(task.api.Annotations, annotateKeyOverwrite, annotateKeysToSet, annotateKeysToClear)
	if err != nil {
		return err
	}
	_, err = task.client.UpdateApi(task.ctx,
		&rpc.UpdateApiRequest{
			Api: task.api,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{annotateFieldName},
			},
		})
	return err
}

type annotateVersionTask struct {
	ctx     context.Context
	client  connection.Client
	version *rpc.ApiVersion
}

func (task *annotateVersionTask) String() string {
	return annotateCommandName + " " + task.version.Name
}

func (task *annotateVersionTask) Run() error {
	var err error
	task.version.Annotations, err = core.UpdateMap(task.version.Annotations, annotateKeyOverwrite, annotateKeysToSet, annotateKeysToClear)
	if err != nil {
		return err
	}
	_, err = task.client.UpdateApiVersion(task.ctx,
		&rpc.UpdateApiVersionRequest{
			ApiVersion: task.version,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{annotateFieldName},
			},
		})
	return err
}

type annotateSpecTask struct {
	ctx    context.Context
	client connection.Client
	spec   *rpc.ApiSpec
}

func (task *annotateSpecTask) String() string {
	return annotateCommandName + " " + task.spec.Name
}

func (task *annotateSpecTask) Run() error {
	var err error
	task.spec.Annotations, err = core.UpdateMap(task.spec.Annotations, annotateKeyOverwrite, annotateKeysToSet, annotateKeysToClear)
	if err != nil {
		return err
	}
	_, err = task.client.UpdateApiSpec(task.ctx,
		&rpc.UpdateApiSpecRequest{
			ApiSpec: task.spec,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{annotateFieldName},
			},
		})
	return err
}
