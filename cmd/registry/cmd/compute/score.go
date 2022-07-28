// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compute

import (
	"context"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/cmd/registry/scoring"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

func scoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "score",
		Short: "Compute scores for APIs and API specs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get filter from flags")
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get dry-run from flags")
			}

			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			// Initialize task queue.
			taskQueue, wait := core.WorkerPool(ctx, 64)
			defer wait()

			resources, err := patterns.ListResources(ctx, client, args[0], filter)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list resources")
			}

			artifactClient := &scoring.RegistryArtifactClient{RegistryClient: client}

			for _, r := range resources {
				// Fetch the ScoreDefinitions which can be applied to this resource
				scoreDefinitions, err := scoring.FetchScoreDefinitions(ctx, artifactClient, r.ResourceName())
				if err != nil {
					log.FromContext(ctx).WithError(err).Errorf("Skipping resource %q", r.ResourceName())
					continue
				}
				for _, d := range scoreDefinitions {
					taskQueue <- &computeScoreTask{
						client:      artifactClient,
						defArtifact: d,
						resource:    r,
						dryRun:      dryRun,
					}
				}
			}

			return
		},
	}
}

type computeScoreTask struct {
	client      *scoring.RegistryArtifactClient
	defArtifact *rpc.Artifact
	resource    patterns.ResourceInstance
	dryRun      bool
}

func (task *computeScoreTask) String() string {
	return "compute score " + task.resource.ResourceName().String()
}

func (task *computeScoreTask) Run(ctx context.Context) error {
	return scoring.CalculateScore(ctx, task.client, task.defArtifact, task.resource, task.dryRun)
}
