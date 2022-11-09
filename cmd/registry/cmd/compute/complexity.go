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
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
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
				err = core.ListSpecs(ctx, client, parsed.Spec(), filter, func(spec *rpc.ApiSpec) error {
					taskQueue <- &computeComplexityTask{
						client:   client,
						specName: spec.Name,
						dryRun:   dryRun,
					}
					return nil
				})
			} else {
				err = core.ListSpecRevisions(ctx, client, parsed, filter, func(spec *rpc.ApiSpec) error {
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
	if err = core.GetSpecRevision(ctx, task.client, specName, true, func(s *rpc.ApiSpec) error {
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
	if core.IsOpenAPIv2(spec.GetMimeType()) {
		document, err := oas2.ParseDocument(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		complexity = core.SummarizeOpenAPIv2Document(document)
	} else if core.IsOpenAPIv3(spec.GetMimeType()) {
		document, err := oas3.ParseDocument(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid OpenAPI: %s", task.specName)
			return nil
		}
		complexity = core.SummarizeOpenAPIv3Document(document)
	} else if core.IsDiscovery(spec.GetMimeType()) {
		document, err := discovery.ParseDocument(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Invalid Discovery: %s", task.specName)
			return nil
		}
		complexity = core.SummarizeDiscoveryDocument(document)
	} else if core.IsProto(spec.GetMimeType()) && core.IsZipArchive(spec.GetMimeType()) {
		complexity, err = core.SummarizeZippedProtos(contents)
		if err != nil {
			log.FromContext(ctx).WithError(err).Errorf("Error processing protos: %s", task.specName)
			return nil
		}
	} else {
		return fmt.Errorf("we don't know how to summarize %s", task.specName)
	}

	if task.dryRun {
		core.PrintMessage(complexity)
		return nil
	}
	subject := task.specName
	messageData, _ := proto.Marshal(complexity)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("gnostic.metrics.Complexity"),
		Contents: messageData,
	}
	return core.SetArtifact(ctx, task.client, artifact)
}
