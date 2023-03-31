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

package score

import (
	"context"

	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/cmd/registry/scoring"
	"github.com/apigee/registry/cmd/registry/tasks"
	scoring_message "github.com/apigee/registry/pkg/application/scoring"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func Command() *cobra.Command {
	var filter string
	var jobs int
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "score PATTERN",
		Short: "Compute scores for APIs and API specs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			args[0] = c.FQName(args[0])

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			// Initialize task queue.
			// Use the warnings queue to make sure that failure in one score calculation task doesn't abort the whole queue.
			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()

			inputPattern, err := patterns.ParseResourcePattern(args[0])
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("invalid pattern supplied in the args")
			}
			artifactClient := &scoring.RegistryArtifactClient{RegistryClient: client}

			scoreDefinitions, err := scoring.FetchScoreDefinitions(ctx, artifactClient, inputPattern.Project())
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatalf("Failed to get ScoreDefinitions")
			}
			// List resources based on the retrieved definitions
			for _, d := range scoreDefinitions {
				// Extract definition
				definition := &scoring_message.ScoreDefinition{}
				if err := proto.Unmarshal(d.GetContents(), definition); err != nil {
					log.FromContext(ctx).WithError(err).Errorf("Failed to unmarshal ScoreDefinition: %q", d.GetName())
					continue
				}
				mergedPattern, mergedFilter, err := scoring.GenerateCombinedPattern(definition.GetTargetResource(), inputPattern, filter)
				if err != nil {
					log.FromContext(ctx).WithError(err).Errorf("Skipping definition %q", d.GetName())
					continue
				}

				resources, err := patterns.ListResources(ctx, client, mergedPattern, mergedFilter)
				if err != nil || len(resources) == 0 {
					log.FromContext(ctx).WithError(err).Errorf("Skipping definition %q", d.GetName())
					continue
				}

				for _, r := range resources {
					taskQueue <- &computeScoreTask{
						client:      artifactClient,
						defArtifact: d,
						resource:    r,
						dryRun:      dryRun,
					}
				}
			}
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if set, computation results will only be printed and will not stored in the registry")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
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
