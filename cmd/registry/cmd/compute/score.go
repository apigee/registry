package compute

import (
	"context"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/cmd/registry/scoring"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

func scoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "score (wip)",
		Short: "Compute scores for APIs and API specs",
		Args:  cobra.ExactArgs(1),
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

			resources, err := patterns.ListResources(ctx, client, args[0], filter)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list resources")
			}

			for _, r := range resources {
				// Fetch the ScoreDefinitions which can be applied to this resource
				scoreDefinitions, err := scoring.FetchScoreDefinitions(ctx, client, r.ResourceName())
				if err != nil {
					log.FromContext(ctx).WithError(err).Errorf("Skipping resource %q", r.ResourceName())
					continue
				}
				for _, d := range scoreDefinitions {
					taskQueue <- &computeScoreTask{
						ctx:        ctx,
						client:     client,
						definition: d,
						resource:   r,
					}
				}
			}

			return
		},
	}
}

type computeScoreTask struct {
	ctx        context.Context
	client     connection.Client
	definition *rpc.ScoreDefinition
	resource   patterns.ResourceInstance
}

func (task *computeScoreTask) String() string {
	return "compute score " + task.resource.ResourceName().String()
}

func (task *computeScoreTask) Run(ctx context.Context) error {
	return scoring.CalculateScore(task.ctx, task.client, task.definition, task.resource)
}
