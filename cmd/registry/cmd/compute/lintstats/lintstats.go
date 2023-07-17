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

package lintstats

import (
	"context"
	"fmt"
	"sort"

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	metrics "github.com/google/gnostic/metrics"
)

func lintRelation(linter string) string {
	return "lint-" + linter
}

func lintStatsRelation(linter string) string {
	return "lintstats-" + linter
}

func Command() *cobra.Command {
	var linter string
	var filter string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "lintstats RESOURCE",
		Short: "Compute summaries of linter runs",
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

			adminClient, err := connection.NewAdminClientWithSettings(ctx, c)
			if err != nil {
				return err
			}

			return matchAndHandleLintStatsCmd(ctx, client, adminClient, args[0], filter, linter, dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if set, computation results will only be printed and will not stored in the registry")
	cmd.Flags().StringVar(&linter, "linter", "", "the name of the linter whose results will be used to compute stats (aip|spectral|gnostic)")
	_ = cmd.MarkFlagRequired("linter")
	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	return cmd
}

func computeLintStats(lint *style.Lint) *style.LintStats {
	problemCounts := make([]*style.LintProblemCount, 0)
	for _, file := range lint.Files {
		for _, problem := range file.Problems {
			var problemCount *style.LintProblemCount
			for _, pc := range problemCounts {
				if pc.RuleId == problem.RuleId {
					problemCount = pc
					break
				}
			}
			if problemCount == nil {
				problemCount = &style.LintProblemCount{
					Count:      0,
					RuleId:     problem.RuleId,
					RuleDocUri: problem.RuleDocUri,
				}
				problemCounts = append(problemCounts, problemCount)
			}
			problemCount.Count++
		}
	}
	// sort results in decreasing order of count
	sort.Slice(problemCounts, func(i, j int) bool {
		return problemCounts[i].Count > problemCounts[j].Count
	})
	return &style.LintStats{ProblemCounts: problemCounts}
}

func computeLintStatsSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	spec names.Spec,
	filter string,
	linter string,
	dryRun bool) error {
	return visitor.ListSpecs(ctx, client, spec, 0, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
		// Iterate through a collection of specs and evaluate each.
		log.Debug(ctx, spec.GetName())
		// get the lint results
		request := rpc.GetArtifactContentsRequest{
			Name: spec.Name + "/artifacts/" + lintRelation(linter),
		}
		contents, _ := client.GetArtifactContents(ctx, &request)
		if contents == nil {
			return nil // ignore missing results
		}

		messageType, err := mime.MessageTypeForMimeType(contents.GetContentType())
		if err != nil {
			return nil
		}
		if messageType != "google.cloud.apigeeregistry.applications.v1alpha1.Lint" {
			return nil // ignore unexpected message types
		}

		lint := &style.Lint{}
		err = patch.UnmarshalContents(contents.GetData(), contents.GetContentType(), lint)
		if err != nil {
			return nil
		}

		// generate the stats from the result by counting problems
		lintStats := computeLintStats(lint)

		{
			// Calculate the operation and schema count
			request := rpc.GetArtifactContentsRequest{
				Name: spec.Name + "/artifacts/complexity",
			}
			contents, _ := client.GetArtifactContents(ctx, &request)
			if contents == nil {
				return nil // ignore missing results
			}
			complexity := &metrics.Complexity{}
			err := patch.UnmarshalContents(contents.GetData(), contents.GetContentType(), complexity)
			if err != nil {
				return nil
			}

			lintStats.OperationCount = complexity.GetDeleteCount() +
				complexity.GetPutCount() + complexity.GetGetCount() + complexity.GetPostCount()

			lintStats.SchemaCount = complexity.GetSchemaCount()
		}

		if dryRun {
			fmt.Println(protojson.Format(lintStats))
		} else {
			_ = storeLintStatsArtifact(ctx, client, spec.GetName(), linter, lintStats)
		}
		return nil
	})
}

func computeLintStatsProjects(ctx context.Context,
	client *gapic.RegistryClient,
	adminClient *gapic.AdminClient,
	projectName names.Project,
	filter string,
	linter string,
	dryRun bool) error {
	return visitor.ListProjects(ctx, adminClient, projectName, nil, 0, filter, func(ctx context.Context, project *rpc.Project) error {
		project_stats := &style.LintStats{}

		if err := visitor.ListAPIs(ctx, client, projectName.Api(""), 0, filter, func(ctx context.Context, api *rpc.Api) error {
			aggregateLintStats(ctx, client, api.GetName(), linter, project_stats)
			return nil
		}); err != nil {
			return nil
		}
		if dryRun {
			fmt.Println(protojson.Format((project_stats)))
		} else {
			// Store the aggregate stats on this project
			_ = storeLintStatsArtifact(ctx, client, project.GetName()+"/locations/global", linter, project_stats)
			log.Debug(ctx, project.GetName())
		}
		return nil
	})
}

func computeLintStatsAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	apiName names.Api,
	filter string,
	linter string,
	dryRun bool) error {
	return visitor.ListAPIs(ctx, client, apiName, 0, filter, func(ctx context.Context, api *rpc.Api) error {
		api_stats := &style.LintStats{}

		if err := visitor.ListVersions(ctx, client, apiName.Version(""), 0, filter, func(ctx context.Context, version *rpc.ApiVersion) error {
			aggregateLintStats(ctx, client, version.GetName(), linter, api_stats)
			return nil
		}); err != nil {
			return nil
		}

		if dryRun {
			fmt.Println(protojson.Format((api_stats)))
		} else {
			// Store the aggregate stats on this api
			_ = storeLintStatsArtifact(ctx, client, api.GetName(), linter, api_stats)
			log.Debug(ctx, api.GetName())
		}
		return nil
	})
}

func computeLintStatsVersions(ctx context.Context,
	client *gapic.RegistryClient,
	versionName names.Version,
	filter string,
	linter string,
	dryRun bool) error {
	return visitor.ListVersions(ctx, client, versionName, 0, filter, func(ctx context.Context, version *rpc.ApiVersion) error {
		stats := &style.LintStats{}
		if err := visitor.ListSpecs(ctx, client, versionName.Spec(""), 0, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
			aggregateLintStats(ctx, client, spec.GetName(), linter, stats)
			return nil
		}); err != nil {
			return nil
		}

		if dryRun {
			fmt.Println(protojson.Format((stats)))
		} else {
			// Store the aggregate stats on this version
			_ = storeLintStatsArtifact(ctx, client, version.GetName(), linter, stats)
			log.Debug(ctx, version.GetName())
		}
		return nil
	})
}

func storeLintStatsArtifact(ctx context.Context,
	client *gapic.RegistryClient,
	subject string,
	linter string,
	lintStats *style.LintStats) error {
	// store the lintstats artifact
	relation := lintStatsRelation(linter)
	messageData, _ := proto.Marshal(lintStats)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: mime.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.LintStats"),
		Contents: messageData,
	}
	return visitor.SetArtifact(ctx, client, artifact)
}

func aggregateLintStats(ctx context.Context,
	client connection.RegistryClient,
	name string,
	linter string,
	aggregateStats *style.LintStats) {
	// Calculate the operation and schema count
	request := rpc.GetArtifactContentsRequest{
		Name: name + "/artifacts/" + lintStatsRelation(linter),
	}
	contents, _ := client.GetArtifactContents(ctx, &request)
	if contents == nil {
		return // ignore missing results
	}
	stats := &style.LintStats{}
	err := patch.UnmarshalContents(contents.GetData(), contents.GetContentType(), stats)
	if err != nil {
		return
	}

	aggregateStats.OperationCount += stats.OperationCount
	aggregateStats.SchemaCount += stats.SchemaCount
	aggregateStats.ProblemCounts = append(aggregateStats.ProblemCounts, stats.ProblemCounts...)
}

func matchAndHandleLintStatsCmd(
	ctx context.Context,
	client connection.RegistryClient,
	adminClient connection.AdminClient,
	name string,
	filter string,
	linter string,
	dryRun bool,
) error {
	// First try to match collection names, then try to match resource names.
	if project, err := names.ParseProjectCollection(name); err == nil {
		return computeLintStatsProjects(ctx, client, adminClient, project, filter, linter, dryRun)
	} else if api, err := names.ParseApiCollection(name); err == nil {
		return computeLintStatsAPIs(ctx, client, api, filter, linter, dryRun)
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		return computeLintStatsVersions(ctx, client, version, filter, linter, dryRun)
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		return computeLintStatsSpecs(ctx, client, spec, filter, linter, dryRun)
	} else if project, err := names.ParseProject(name); err == nil {
		return computeLintStatsProjects(ctx, client, adminClient, project, filter, linter, dryRun)
	} else if api, err := names.ParseApi(name); err == nil {
		return computeLintStatsAPIs(ctx, client, api, filter, linter, dryRun)
	} else if version, err := names.ParseVersion(name); err == nil {
		return computeLintStatsVersions(ctx, client, version, filter, linter, dryRun)
	} else if spec, err := names.ParseSpec(name); err == nil {
		return computeLintStatsSpecs(ctx, client, spec, filter, linter, dryRun)
	} else {
		// If nothing matched, return an error.
		return fmt.Errorf("unsupported argument: %s", name)
	}
}
