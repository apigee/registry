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

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/googleapis/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	discovery "github.com/googleapis/gnostic/discovery"
	metrics "github.com/googleapis/gnostic/metrics"
	oas2 "github.com/googleapis/gnostic/openapiv2"
	oas3 "github.com/googleapis/gnostic/openapiv3"
)

func vocabularyCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "vocabulary",
		Short: "Compute vocabularies of API specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.WithError(err).Fatal("Failed to get filter from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}
			// Initialize task queue.
			taskQueue, wait := core.WorkerPool(ctx, 64)
			defer wait()
			// Generate tasks.
			name := args[0]
			if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
				// Iterate through a collection of specs and summarize each.
				err = core.ListSpecs(ctx, client, m, filter, func(spec *rpc.ApiSpec) {
					taskQueue <- &computeVocabularyTask{
						client:   client,
						specName: spec.Name,
					}
				})
				if err != nil {
					log.WithError(err).Fatal("Failed to list specs")
				}
			}
		},
	}
}

type computeVocabularyTask struct {
	client   connection.Client
	specName string
}

func (task *computeVocabularyTask) String() string {
	return "compute vocabulary " + task.specName
}

func (task *computeVocabularyTask) Run(ctx context.Context) error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpec(ctx, request)
	if err != nil {
		return err
	}
	relation := "vocabulary"
	log.Debugf("Computing %s/artifacts/%s", spec.Name, relation)
	var vocab *metrics.Vocabulary
	if core.IsOpenAPIv2(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(ctx, task.client, spec)
		if err != nil {
			return nil
		}
		document, err := oas2.ParseDocument(data)
		if err != nil {
			log.WithError(err).Warnf("Invalid OpenAPI: %s", spec.Name)
			return nil
		}
		vocab = vocabulary.NewVocabularyFromOpenAPIv2(document)
	} else if core.IsOpenAPIv3(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(ctx, task.client, spec)
		if err != nil {
			return nil
		}
		document, err := oas3.ParseDocument(data)
		if err != nil {
			log.WithError(err).Warnf("Invalid OpenAPI: %s", spec.Name)
			return nil
		}
		vocab = vocabulary.NewVocabularyFromOpenAPIv3(document)
	} else if core.IsDiscovery(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(ctx, task.client, spec)
		if err != nil {
			return nil
		}
		document, err := discovery.ParseDocument(data)
		if err != nil {
			log.WithError(err).Warnf("Invalid Discovery: %s", spec.Name)
			return nil
		}
		vocab = vocabulary.NewVocabularyFromDiscovery(document)
	} else if core.IsProto(spec.GetMimeType()) && core.IsZipArchive(spec.GetMimeType()) {
		data, err := core.GetBytesForSpec(ctx, task.client, spec)
		if err != nil {
			return nil
		}
		vocab, err = core.NewVocabularyFromZippedProtos(data)
		if err != nil {
			log.WithError(err).Warnf("Error processing protos: %s", spec.Name)
			return nil
		}
	} else {
		return fmt.Errorf("we don't know how to summarize %s", spec.Name)
	}
	subject := spec.GetName()
	messageData, _ := proto.Marshal(vocab)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("gnostic.metrics.Vocabulary"),
		Contents: messageData,
	}
	err = core.SetArtifact(ctx, task.client, artifact)
	if err != nil {
		return err
	}
	return nil
}
