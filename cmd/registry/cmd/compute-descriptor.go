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
	"github.com/golang/protobuf/ptypes/any"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func init() {
	computeCmd.AddCommand(computeDescriptorCmd)
}

var computeDescriptorCmd = &cobra.Command{
	Use:   "descriptor",
	Short: "Compute the descriptor of API specs.",
	Long:  `Compute the descriptor of API specs.`,
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
			err = core.ListSpecs(ctx, client, m, computeFilter, func(spec *rpc.Spec) {
				taskQueue <- &computeDescriptorTask{
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

type computeDescriptorTask struct {
	ctx      context.Context
	client   connection.Client
	specName string
}

func (task *computeDescriptorTask) Run() error {
	request := &rpc.GetSpecRequest{
		Name: task.specName,
		View: rpc.SpecView_FULL,
	}
	spec, err := task.client.GetSpec(task.ctx, request)
	if err != nil {
		return err
	}
	name := spec.GetName()
	log.Printf("computing descriptor %s", name)
	data, err := core.GetBytesForSpec(spec)
	if err != nil {
		return nil
	}
	subject := spec.GetName()
	relation := "descriptor"
	var typeURL string
	var document proto.Message
	if strings.HasPrefix(spec.GetStyle(), "openapi/v2") {
		typeURL = "gnostic.openapiv2.Document"
		document, err = openapi_v2.ParseDocument(data)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(spec.GetStyle(), "openapi/v3") {
		typeURL = "gnostic.openapiv3.Document"
		document, err = openapi_v3.ParseDocument(data)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unable to generate descriptor for style %s", spec.GetStyle())
	}
	messageData, err := proto.Marshal(document)
	if err != nil {
		return err
	}
	property := &rpc.Property{
		Subject:  subject,
		Relation: relation,
		Name:     subject + "/properties/" + relation,
		Value: &rpc.Property_MessageValue{
			MessageValue: &any.Any{
				TypeUrl: typeURL,
				Value:   messageData,
			},
		},
	}
	return core.SetProperty(task.ctx, task.client, property)
}
