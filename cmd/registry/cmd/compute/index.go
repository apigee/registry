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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func indexCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "Compute indexes of API specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get filter from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			// Initialize task queue.
			taskQueue, wait := core.WorkerPool(ctx, 64)
			defer wait()
			// Generate tasks.
			name := args[0]
			if spec, err := names.ParseSpec(name); err == nil {
				// Iterate through a collection of specs and summarize each.
				err = core.ListSpecs(ctx, client, spec, filter, func(spec *rpc.ApiSpec) {
					taskQueue <- &computeIndexTask{
						client:   client,
						specName: spec.Name,
					}
				})
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to list specs")
				}
			}
		},
	}
}

type computeIndexTask struct {
	client   connection.Client
	specName string
}

func (task *computeIndexTask) String() string {
	return "compute index " + task.specName
}

func (task *computeIndexTask) Run(ctx context.Context) error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpec(ctx, request)
	if err != nil {
		return err
	}
	data, err := core.GetBytesForSpec(ctx, task.client, spec)
	if err != nil {
		return nil
	}
	relation := "index"
	log.Debugf(ctx, "Computing %s/artifacts/%s", spec.Name, relation)
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
		MimeType: core.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.Index"),
		Contents: messageData,
	}
	err = core.SetArtifact(ctx, task.client, artifact)
	if err != nil {
		return err
	}
	return nil
}
