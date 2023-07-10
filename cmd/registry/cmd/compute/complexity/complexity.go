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

package complexity

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
		Use:   "complexity SPEC_REVISION",
		Short: "Compute complexity metrics of API specs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}
			args[0] = c.FQName(args[0])

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			// Initialize task queue.
			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()

			parsed, err := names.ParseSpecRevision(args[0])
			if err != nil {
				return err
			}

			if parsed.RevisionID == "" {
				err = visitor.ListSpecs(ctx, client, parsed.Spec(), 0, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
					taskQueue <- &computeComplexityTask{
						client:   client,
						specName: spec.Name,
						dryRun:   dryRun,
					}
					return nil
				})
			} else {
				err = visitor.ListSpecRevisions(ctx, client, parsed, 0, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
					taskQueue <- &computeComplexityTask{
						client:   client,
						specName: spec.Name,
						dryRun:   dryRun,
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
	var complexity *metrics.Complexity
	if mime.IsOpenAPIv2(spec.GetMimeType()) {
		document, err := oas2.ParseDocument(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		complexity = SummarizeOpenAPIv2Document(document)
	} else if mime.IsOpenAPIv3(spec.GetMimeType()) {
		document, err := oas3.ParseDocument(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		complexity = SummarizeOpenAPIv3Document(document)
	} else if mime.IsDiscovery(spec.GetMimeType()) {
		document, err := discovery.ParseDocument(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid Discovery: %s", task.specName)
			return nil
		}
		complexity = SummarizeDiscoveryDocument(document)
	} else if mime.IsProto(spec.GetMimeType()) && mime.IsZipArchive(spec.GetMimeType()) {
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
		MimeType: mime.MimeTypeForMessageType("gnostic.metrics.Complexity"),
		Contents: messageData,
	}
	return visitor.SetArtifact(ctx, task.client, artifact)
}
