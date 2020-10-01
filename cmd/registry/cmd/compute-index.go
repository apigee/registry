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
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func init() {
	computeCmd.AddCommand(computeIndexCmd)
}

var computeIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Compute indexes of API specs",
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
		if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
			// Iterate through a collection of specs and summarize each.
			err = core.ListSpecs(ctx, client, m, computeFilter, func(spec *rpc.Spec) {
				taskQueue <- &computeIndexTask{
					ctx:      ctx,
					client:   client,
					specName: spec.Name,
				}
			})
			close(taskQueue)
			core.WaitGroup().Wait()
		}
	},
}

type computeIndexTask struct {
	ctx      context.Context
	client   connection.Client
	specName string
}

func (task *computeIndexTask) Run() error {
	request := &rpc.GetSpecRequest{
		Name: task.specName,
		View: rpc.SpecView_FULL,
	}
	spec, err := task.client.GetSpec(task.ctx, request)
	if err != nil {
		return err
	}
	relation := "index"
	log.Printf("computing %s/properties/%s", spec.Name, relation)
	var index *rpc.Index
	if spec.GetStyle() == "proto+zip" {
		index, err = core.NewIndexFromZippedProtos(spec.GetContents())
		if err != nil {
			return fmt.Errorf("error processing protos: %s", spec.Name)
		}
	} else {
		return fmt.Errorf("we don't know how to compute the index of %s", spec.Name)
	}
	subject := spec.GetName()
	messageData, err := proto.Marshal(index)
	property := &rpc.Property{
		Subject:  subject,
		Relation: relation,
		Name:     subject + "/properties/" + relation,
		Value: &rpc.Property_MessageValue{
			MessageValue: &any.Any{
				TypeUrl: "google.cloud.apigee.registry.v1alpha1.Index",
				Value:   messageData,
			},
		},
	}
	err = core.SetProperty(task.ctx, task.client, property)
	if err != nil {
		return err
	}
	return nil
}
