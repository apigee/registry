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
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/google/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	discovery "github.com/google/gnostic/discovery"
	metrics "github.com/google/gnostic/metrics"
	oas2 "github.com/google/gnostic/openapiv2"
	oas3 "github.com/google/gnostic/openapiv3"
)

func vocabularyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "vocabulary",
		Short: "Compute vocabularies of API specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			path := c.FQName(args[0])

			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get filter from flags")
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get dry-run from flags")
			}

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			// Initialize task queue.
			jobs, err := cmd.Flags().GetInt("jobs")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get jobs from flags")
			}
			taskQueue, wait := core.WorkerPoolWithWarnings(ctx, jobs)
			defer wait()

			parsed, err := names.ParseSpecRevision(path)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed parse")
			}

			// Iterate through a collection of specs and summarize each.
			if parsed.RevisionID == "" {
				err = core.ListSpecs(ctx, client, parsed.Spec(), filter, func(spec *rpc.ApiSpec) error {
					taskQueue <- &computeVocabularyTask{
						client:   client,
						specName: spec.GetName(),
						dryRun:   dryRun,
					}
					return nil
				})
			} else {
				err = core.ListSpecRevisions(ctx, client, parsed, filter, func(spec *rpc.ApiSpec) error {
					taskQueue <- &computeVocabularyTask{
						client:   client,
						specName: spec.GetName(),
						dryRun:   dryRun,
					}
					return nil
				})
			}
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list specs")
			}
		},
	}
}

type computeVocabularyTask struct {
	client   connection.RegistryClient
	specName string
	dryRun   bool
}

func (task *computeVocabularyTask) String() string {
	return "compute vocabulary " + task.specName
}

func (task *computeVocabularyTask) Run(ctx context.Context) error {
	contents, err := task.client.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
		Name: task.specName,
	})
	if err != nil {
		return err
	}

	log.Debugf(ctx, "Computing %s/artifacts/vocabulary", task.specName)
	var vocab *metrics.Vocabulary

	if core.IsOpenAPIv2(contents.GetContentType()) {
		document, err := oas2.ParseDocument(contents.GetData())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		vocab = vocabulary.NewVocabularyFromOpenAPIv2(document)
	} else if core.IsOpenAPIv3(contents.GetContentType()) {
		document, err := oas3.ParseDocument(contents.GetData())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		vocab = vocabulary.NewVocabularyFromOpenAPIv3(document)
	} else if core.IsDiscovery(contents.GetContentType()) {
		document, err := discovery.ParseDocument(contents.GetData())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid Discovery: %s", task.specName)
			return nil
		}
		vocab = vocabulary.NewVocabularyFromDiscovery(document)
	} else if core.IsProto(contents.GetContentType()) && core.IsZipArchive(contents.GetContentType()) {
		vocab, err = core.NewVocabularyFromZippedProtos(contents.GetData())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Error processing protos: %s", task.specName)
			return nil
		}
	} else {
		return fmt.Errorf("we don't know how to summarize %s", task.specName)
	}

	if task.dryRun {
		core.PrintMessage(vocab)
		return nil
	}

	messageData, err := proto.Marshal(vocab)
	if err != nil {
		return err
	}
	return core.SetArtifact(ctx, task.client, &rpc.Artifact{
		Name:     task.specName + "/artifacts/vocabulary",
		MimeType: core.MimeTypeForMessageType("gnostic.metrics.Vocabulary"),
		Contents: messageData,
	})
}
