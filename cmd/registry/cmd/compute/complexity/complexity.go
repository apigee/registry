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

package complexity

import (
	"context"
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/types"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	discovery "github.com/google/gnostic/discovery"
	metrics "github.com/google/gnostic/metrics"
	oas2 "github.com/google/gnostic/openapiv2"
	oas3 "github.com/google/gnostic/openapiv3"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complexity",
		Short: "Compute complexity metrics of API specs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			args[0] = c.FQName(args[0])

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
			taskQueue, wait := core.WorkerPool(ctx, jobs)
			defer wait()

			parsed, err := names.ParseSpecRevision(args[0])
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed parse")
			}

			if parsed.RevisionID == "" {
				err = visitor.ListSpecs(ctx, client, parsed.Spec(), filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
					taskQueue <- &computeComplexityTask{
						client:   client,
						specName: spec.Name,
						dryRun:   dryRun,
					}
					return nil
				})
			} else {
				err = visitor.ListSpecRevisions(ctx, client, parsed, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
					taskQueue <- &computeComplexityTask{
						client:   client,
						specName: spec.Name,
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
	cmd.Flags().String("filter", "", "Filter selected resources")
	cmd.Flags().Bool("dry-run", false, "if set, computation results will only be printed and will not stored in the registry")
	cmd.Flags().Int("jobs", 10, "Number of actions to perform concurrently")
	return cmd
}

type computeComplexityTask struct {
	client   connection.RegistryClient
	specName string
	dryRun   bool
}

func (task *computeComplexityTask) String() string {
	return "compute complexity " + task.specName
}

func (task *computeComplexityTask) Run(ctx context.Context) error {
	specName, err := names.ParseSpecRevision(task.specName)
	if err != nil {
		return err
	}
	var spec *rpc.ApiSpec
	if err = visitor.GetSpecRevision(ctx, task.client, specName, true, func(ctx context.Context, s *rpc.ApiSpec) error {
		spec = s
		return nil
	}); err != nil {
		return err
	}

	relation := "complexity"
	log.Debugf(ctx, "Computing %s/artifacts/%s", task.specName, relation)
	contents := spec.GetContents()
	if strings.Contains(spec.GetMimeType(), "+gzip") {
		if contents, err = core.GUnzippedBytes(contents); err != nil {
			return err
		}
	}
	var complexity *metrics.Complexity
	if types.IsOpenAPIv2(spec.GetMimeType()) {
		document, err := oas2.ParseDocument(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		complexity = SummarizeOpenAPIv2Document(document)
	} else if types.IsOpenAPIv3(spec.GetMimeType()) {
		document, err := oas3.ParseDocument(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		complexity = SummarizeOpenAPIv3Document(document)
	} else if types.IsDiscovery(spec.GetMimeType()) {
		document, err := discovery.ParseDocument(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid Discovery: %s", task.specName)
			return nil
		}
		complexity = SummarizeDiscoveryDocument(document)
	} else if types.IsProto(spec.GetMimeType()) && types.IsZipArchive(spec.GetMimeType()) {
		complexity, err = SummarizeZippedProtos(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Error processing protos: %s", task.specName)
			return nil
		}
	} else {
		return fmt.Errorf("we don't know how to summarize %s", task.specName)
	}

	if task.dryRun {
		fmt.Println(protojson.Format(complexity))
		return nil
	}
	subject := task.specName
	messageData, _ := proto.Marshal(complexity)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: types.MimeTypeForMessageType("gnostic.metrics.Complexity"),
		Contents: messageData,
	}
	return core.SetArtifact(ctx, task.client, artifact)
}
