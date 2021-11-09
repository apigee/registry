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

	discovery "github.com/google/gnostic/discovery"
	oas2 "github.com/google/gnostic/openapiv2"
	oas3 "github.com/google/gnostic/openapiv3"
)

func descriptorCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "descriptor",
		Short: "Compute descriptors of API specs",
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
				err = core.ListSpecs(ctx, client, spec, filter, func(spec *rpc.ApiSpec) {
					taskQueue <- &computeDescriptorTask{
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

type computeDescriptorTask struct {
	client   connection.Client
	specName string
}

func (task *computeDescriptorTask) String() string {
	return "compute descriptor " + task.specName
}

func (task *computeDescriptorTask) Run(ctx context.Context) error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpec(ctx, request)
	if err != nil {
		return err
	}
	name := spec.GetName()
	relation := "descriptor"
	log.Debugf(ctx, "Computing %s/artifacts/%s", name, relation)
	data, err := core.GetBytesForSpec(ctx, task.client, spec)
	if err != nil {
		return nil
	}
	subject := spec.GetName()
	var typeURL string
	var document proto.Message
	if core.IsOpenAPIv2(spec.GetMimeType()) {
		typeURL = "gnostic.openapiv2.Document"
		document, err = oas2.ParseDocument(data)
		if err != nil {
			return err
		}
	} else if core.IsOpenAPIv3(spec.GetMimeType()) {
		typeURL = "gnostic.openapiv3.Document"
		document, err = oas3.ParseDocument(data)
		if err != nil {
			return err
		}
	} else if core.IsDiscovery(spec.GetMimeType()) {
		typeURL = "gnostic.discoveryv1.Document"
		document, err = discovery.ParseDocument(data)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unable to generate descriptor for style %s", spec.GetMimeType())
	}
	messageData, err := proto.Marshal(document)
	if err != nil {
		return err
	}
	// TODO: consider gzipping descriptors to reduce size;
	// this will probably require some representation of compression type in the typeURL
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType(typeURL),
		Contents: messageData,
	}
	return core.SetArtifact(ctx, task.client, artifact)
}
