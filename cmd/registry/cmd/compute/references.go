// Copyright 2021 Google LLC. All Rights Reserved.
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
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func referencesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "references",
		Short: "Compute references of API specs",
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

			// Iterate through a collection of specs and compute references for each
			if parsed.RevisionID == "" {
				err = core.ListSpecs(ctx, client, parsed.Spec(), filter, func(spec *rpc.ApiSpec) error {
					taskQueue <- &computeReferencesTask{
						client:   client,
						specName: spec.Name,
						dryRun:   dryRun,
					}
					return nil
				})
			} else {
				err = core.ListSpecRevisions(ctx, client, parsed, filter, func(spec *rpc.ApiSpec) error {
					taskQueue <- &computeReferencesTask{
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

type computeReferencesTask struct {
	client   connection.RegistryClient
	specName string
	dryRun   bool
}

func (task *computeReferencesTask) String() string {
	return "compute references " + task.specName
}

func (task *computeReferencesTask) Run(ctx context.Context) error {
	contents, err := task.client.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
		Name: task.specName,
	})
	if err != nil {
		return err
	}

	log.Debugf(ctx, "Computing %s/artifacts/references", task.specName)

	var references *rpc.References
	if core.IsProto(contents.GetContentType()) && core.IsZipArchive(contents.GetContentType()) {
		references, err = core.NewReferencesFromZippedProtos(contents.GetData())
		if err != nil {
			return fmt.Errorf("error processing protos: %s", task.specName)
		}
	} else {
		return fmt.Errorf("we don't know how to compute references for %s of type %s", task.specName, contents.GetContentType())
	}

	if task.dryRun {
		core.PrintMessage(references)
		return nil
	}

	messageData, _ := proto.Marshal(references)
	artifact := &rpc.Artifact{
		Name:     task.specName + "/artifacts/references",
		MimeType: core.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.References"),
		Contents: messageData,
	}

	return core.SetArtifact(ctx, task.client, artifact)
}
