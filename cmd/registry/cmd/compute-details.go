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
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/genproto/protobuf/field_mask"
)

func init() {
	computeCmd.AddCommand(computeDetailsCmd)
}

var computeDetailsCmd = &cobra.Command{
	Use:   "details",
	Short: "Compute details about APIs from information in their specs.",
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
		// Generate tasks.
		name := args[0]
		if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
			// Iterate through a collection of specs and summarize each.
			err = core.ListAPIs(ctx, client, m, computeFilter, func(api *rpc.Api) {
				taskQueue <- &computeDetailsTask{
					ctx:     ctx,
					client:  client,
					apiName: api.Name,
				}
			})
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			close(taskQueue)
			core.WaitGroup().Wait()
		}
	},
}

type computeDetailsTask struct {
	ctx     context.Context
	client  connection.Client
	apiName string
}

func (task *computeDetailsTask) Name() string {
	return "compute details " + task.apiName
}

func (task *computeDetailsTask) Run() error {
	m := names.SpecRegexp().FindStringSubmatch(task.apiName + "/versions/-/specs/-")
	specs := make([]*rpc.Spec, 0)
	core.ListSpecs(task.ctx, task.client, m, "", func(spec *rpc.Spec) {
		specs = append(specs, spec)
	})
	// use the last (latest) spec
	if len(specs) == 0 {
		return nil
	}
	spec := specs[len(specs)-1]
	var err error
	m = names.SpecRegexp().FindStringSubmatch(spec.Name)
	spec, err = core.GetSpec(task.ctx, task.client, m, true, nil)
	if err != nil {
		return nil
	}
	var title string
	var description string
	if strings.HasPrefix(spec.GetStyle(), "openapi/v2") {
		data, err := core.GetBytesForSpec(spec)
		if err != nil {
			return nil
		}
		document, err := openapi_v2.ParseDocument(data)
		if document == nil && err != nil {
			return fmt.Errorf("invalid OpenAPI v2: %s", spec.Name)
		}
		if document.Info != nil {
			title = document.Info.Title
			description = document.Info.Description
		}
	} else if strings.HasPrefix(spec.GetStyle(), "openapi/v3") {
		data, err := core.GetBytesForSpec(spec)
		if err != nil {
			return nil
		}
		document, err := openapi_v3.ParseDocument(data)
		if document == nil && err != nil {
			return fmt.Errorf("invalid OpenAPI v3: %s", spec.Name)
		}
		if document.Info != nil {
			title = document.Info.Title
			description = document.Info.Description
		}
	} else {
		return fmt.Errorf("we don't know how to compute the title of %s", task.apiName)
	}
	if len(description) > 256 {
		description = description[0:256]
	}
	request := &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:        task.apiName,
			DisplayName: title,
			Description: description,
		},
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{"display_name", "description"},
		},
	}
	_, err = task.client.UpdateApi(task.ctx, request)
	return err
}
