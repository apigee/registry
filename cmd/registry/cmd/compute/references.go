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

func referencesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "references",
		Short: "Compute references of API specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.Fatalf("Failed to get filter from flags: %s", err)
			}

			ctx := context.Background()
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
				// Iterate through a collection of specs and compute references for each
				err = core.ListSpecs(ctx, client, m, filter, func(spec *rpc.ApiSpec) {
					taskQueue <- &computeReferencesTask{
						client:   client,
						specName: spec.Name,
					}
				})
				if err != nil {
					log.Fatalf("%s", err.Error())
				}
			}
			close(taskQueue)
			core.WaitGroup().Wait()
		},
	}
}

type computeReferencesTask struct {
	client   connection.Client
	specName string
}

func (task *computeReferencesTask) String() string {
	return "compute references " + task.specName
}

func (task *computeReferencesTask) Run(ctx context.Context) error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpec(ctx, request)
	if err != nil {
		return err
	}
	relation := "references"
	log.Printf("computing %s/properties/%s", spec.Name, relation)
	var references *rpc.References
	if core.IsProto(spec.MimeType) && core.IsZipArchive(spec.MimeType) {
		data, err := core.GetBytesForSpec(ctx, task.client, spec)
		if err != nil {
			return nil
		}
		references, err = core.NewReferencesFromZippedProtos(data)
		if err != nil {
			return fmt.Errorf("error processing protos: %s", spec.Name)
		}
	} else {
		return fmt.Errorf("we don't know how to compute references for %s of type %s", spec.Name, spec.MimeType)
	}
	subject := spec.Name
	messageData, _ := proto.Marshal(references)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("google.cloud.apigee.registry.applications.v1alpha1.References"),
		Contents: messageData,
	}
	err = core.SetArtifact(ctx, task.client, artifact)
	if err != nil {
		return err
	}
	return nil
}
