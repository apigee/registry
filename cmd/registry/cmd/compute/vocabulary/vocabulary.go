// Copyright 2020 Google LLC.
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

package vocabulary

import (
	"context"
	"fmt"

	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/google/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	discovery "github.com/google/gnostic/discovery"
	metrics "github.com/google/gnostic/metrics"
	oas2 "github.com/google/gnostic/openapiv2"
	oas3 "github.com/google/gnostic/openapiv3"
)

func Command() *cobra.Command {
	var filter string
	var jobs int
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "vocabulary SPEC_REVISION",
		Short: "Compute vocabularies of API specs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}
			path := c.FQName(args[0])

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			// Initialize task queue.
			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()

			parsed, err := names.ParseSpecRevision(path)
			if err != nil {
				return err
			}

			// Iterate through a collection of specs and summarize each.
			if parsed.RevisionID == "" {
				err = visitor.ListSpecs(ctx, client, parsed.Spec(), 0, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
					taskQueue <- &computeVocabularyTask{
						client: client,
						spec:   spec,
						dryRun: dryRun,
					}
					return nil
				})
			} else {
				err = visitor.ListSpecRevisions(ctx, client, parsed, 0, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
					taskQueue <- &computeVocabularyTask{
						client: client,
						spec:   spec,
						dryRun: dryRun,
					}
					return nil
				})
			}
			return err
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if set, computation results will only be printed and will not stored in the registry")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
}

type computeVocabularyTask struct {
	client connection.RegistryClient
	spec   *rpc.ApiSpec
	dryRun bool
}

func (task *computeVocabularyTask) String() string {
	return "compute vocabulary " + task.spec.Name
}

func (task *computeVocabularyTask) Run(ctx context.Context) error {
	if err := visitor.FetchSpecContents(ctx, task.client, task.spec); err != nil {
		return err
	}

	log.Debugf(ctx, "Computing %s/artifacts/vocabulary", task.spec.Name)
	var vocab *metrics.Vocabulary

	if mime.IsOpenAPIv2(task.spec.GetMimeType()) {
		document, err := oas2.ParseDocument(task.spec.GetContents())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.spec.Name)
			return nil
		}
		vocab = vocabulary.NewVocabularyFromOpenAPIv2(document)
	} else if mime.IsOpenAPIv3(task.spec.GetMimeType()) {
		document, err := oas3.ParseDocument(task.spec.GetContents())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.spec.Name)
			return nil
		}
		vocab = vocabulary.NewVocabularyFromOpenAPIv3(document)
	} else if mime.IsDiscovery(task.spec.GetMimeType()) {
		document, err := discovery.ParseDocument(task.spec.GetContents())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid Discovery: %s", task.spec.Name)
			return nil
		}
		vocab = vocabulary.NewVocabularyFromDiscovery(document)
	} else if mime.IsProto(task.spec.GetMimeType()) && mime.IsZipArchive(task.spec.GetMimeType()) {
		var err error
		vocab, err = NewVocabularyFromZippedProtos(task.spec.GetContents())
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Error processing protos: %s", task.spec.Name)
			return nil
		}
	} else {
		return fmt.Errorf("we don't know how to compute the vocabulary of %s", task.spec.Name)
	}

	if task.dryRun {
		fmt.Println(protojson.Format((vocab)))
		return nil
	}

	messageData, err := proto.Marshal(vocab)
	if err != nil {
		return err
	}
	return visitor.SetArtifact(ctx, task.client, &rpc.Artifact{
		Name:     task.spec.Name + "/artifacts/vocabulary",
		MimeType: mime.MimeTypeForMessageType("gnostic.metrics.Vocabulary"),
		Contents: messageData,
	})
}
