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
	discovery_v1 "github.com/googleapis/gnostic/discovery"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var computeDescriptorCmd = &cobra.Command{
	Use:   "descriptor",
	Short: "Compute descriptors of API specs",
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
			err = core.ListSpecs(ctx, client, m, computeFilter, func(spec *rpc.ApiSpec) {
				taskQueue <- &computeDescriptorTask{
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

type computeDescriptorTask struct {
	ctx      context.Context
	client   connection.Client
	specName string
}

func (task *computeDescriptorTask) String() string {
	return "compute descriptor " + task.specName
}

func (task *computeDescriptorTask) Run() error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpec(task.ctx, request)
	if err != nil {
		return err
	}
	name := spec.GetName()
	relation := "descriptor"
	log.Printf("computing %s/artifacts/%s", name, relation)
	data, err := core.GetBytesForSpec(task.ctx, task.client, spec)
	if err != nil {
		return nil
	}
	subject := spec.GetName()
	var typeURL string
	var document proto.Message
	if core.IsOpenAPIv2(spec.GetMimeType()) {
		typeURL = "gnostic.openapiv2.Document"
		document, err = openapi_v2.ParseDocument(data)
		if err != nil {
			return err
		}
	} else if core.IsOpenAPIv3(spec.GetMimeType()) {
		typeURL = "gnostic.openapiv3.Document"
		document, err = openapi_v3.ParseDocument(data)
		if err != nil {
			return err
		}
	} else if core.IsDiscovery(spec.GetMimeType()) {
		typeURL = "gnostic.discoveryv1.Document"
		document, err = discovery_v1.ParseDocument(data)
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
	return core.SetArtifact(task.ctx, task.client, artifact)
}
