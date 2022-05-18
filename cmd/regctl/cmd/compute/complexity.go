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

	"github.com/apigee/registry/cmd/regctl/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	discovery "github.com/google/gnostic/discovery"
	metrics "github.com/google/gnostic/metrics"
	oas2 "github.com/google/gnostic/openapiv2"
	oas3 "github.com/google/gnostic/openapiv3"
)

func complexityCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "complexity",
		Short: "Compute complexity metrics of API specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
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

			spec, err := names.ParseSpec(args[0])
			if err != nil {
				return // TODO: Log an error.
			}

			// Iterate through a collection of specs and summarize each.
			err = core.ListSpecs(ctx, client, spec, filter, func(spec *rpc.ApiSpec) error {
				taskQueue <- &computeComplexityTask{
					client:   client,
					specName: spec.Name,
				}
				return nil
			})
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list specs")
			}
		},
	}
}

type computeComplexityTask struct {
	client   connection.Client
	specName string
}

func (task *computeComplexityTask) String() string {
	return "compute complexity " + task.specName
}

func (task *computeComplexityTask) Run(ctx context.Context) error {
	request := &rpc.GetApiSpecContentsRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpecContents(ctx, request)
	if err != nil {
		return err
	}
	relation := "complexity"
	log.Debugf(ctx, "Computing %s/artifacts/%s", task.specName, relation)
	var complexity *metrics.Complexity
	if core.IsOpenAPIv2(spec.GetContentType()) {
		document, err := oas2.ParseDocument(spec.GetData())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		complexity = core.SummarizeOpenAPIv2Document(document)
	} else if core.IsOpenAPIv3(spec.GetContentType()) {
		document, err := oas3.ParseDocument(spec.GetData())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		complexity = core.SummarizeOpenAPIv3Document(document)
	} else if core.IsDiscovery(spec.GetContentType()) {
		document, err := discovery.ParseDocument(spec.GetData())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid Discovery: %s", task.specName)
			return nil
		}
		complexity = core.SummarizeDiscoveryDocument(document)
	} else if core.IsProto(spec.GetContentType()) && core.IsZipArchive(spec.GetContentType()) {
		complexity, err = core.SummarizeZippedProtos(spec.GetData())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Error processing protos: %s", task.specName)
			return nil
		}
	} else {
		return fmt.Errorf("we don't know how to summarize %s", task.specName)
	}
	subject := task.specName
	messageData, _ := proto.Marshal(complexity)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("gnostic.metrics.Complexity"),
		Contents: messageData,
	}
	err = core.SetArtifact(ctx, task.client, artifact)
	if err != nil {
		return err
	}
	return nil
}
