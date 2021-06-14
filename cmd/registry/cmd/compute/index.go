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

package compute

import (
	"context"
	"fmt"
	"log"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

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
			err = core.ListSpecs(ctx, client, m, computeFilter, func(spec *rpc.ApiSpec) {
				taskQueue <- &computeIndexTask{
					ctx:      ctx,
					client:   client,
					specName: spec.Name,
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

type computeIndexTask struct {
	ctx      context.Context
	client   connection.Client
	specName string
}

func (task *computeIndexTask) String() string {
	return "compute index " + task.specName
}

func (task *computeIndexTask) Run() error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpec(task.ctx, request)
	if err != nil {
		return err
	}
	data, err := core.GetBytesForSpec(task.ctx, task.client, spec)
	if err != nil {
		return nil
	}
	relation := "index"
	log.Printf("computing %s/artifacts/%s", spec.Name, relation)
	var index *rpc.Index
	if core.IsProto(spec.GetMimeType()) && core.IsZipArchive(spec.GetMimeType()) {
		index, err = core.NewIndexFromZippedProtos(data)
		if err != nil {
			return fmt.Errorf("error processing protos: %s", spec.Name)
		}
	} else {
		return fmt.Errorf("we don't know how to compute the index of %s", spec.Name)
	}
	subject := spec.GetName()
	messageData, _ := proto.Marshal(index)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("google.cloud.apigee.registry.applications.v1alpha1.Index"),
		Contents: messageData,
	}
	err = core.SetArtifact(task.ctx, task.client, artifact)
	if err != nil {
		return err
	}
	return nil
}
