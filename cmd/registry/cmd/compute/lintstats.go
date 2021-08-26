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
	"log"
	"sort"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	metrics "github.com/googleapis/gnostic/metrics"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func lintStatsRelation(linter string) string {
	return "lintstats-" + linter
}

func lintStatsCommand(ctx context.Context) *cobra.Command {
	var linter string
	cmd := &cobra.Command{
		Use:   "lintstats",
		Short: "Compute summaries of linter runs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.Fatalf("Failed to get filter from flags: %s", err)
			}

			ctx := context.Background()
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}

			// Generate tasks.
			name := args[0]

			err = matchAndHandleLintStatsCmd(ctx, client, name, filter, linter)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		},
	}

	cmd.Flags().StringVar(&linter, "linter", "", "The name of the linter whose results will be used to compute stats (aip|spectral|gnostic)")
	cmd.MarkFlagRequired("linter")
	return cmd
}

func computeLintStats(lint *rpc.Lint) *rpc.LintStats {
	problemCounts := make([]*rpc.LintProblemCount, 0)
	for _, file := range lint.Files {
		for _, problem := range file.Problems {
			var problemCount *rpc.LintProblemCount
			for _, pc := range problemCounts {
				if pc.RuleId == problem.RuleId {
					problemCount = pc
					break
				}
			}
			if problemCount == nil {
				problemCount = &rpc.LintProblemCount{
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
	return &rpc.LintStats{ProblemCounts: problemCounts}
}

func computeLintStatsSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filter string,
	linter string) error {

	return core.ListSpecs(ctx, client, segments, filter, func(spec *rpc.ApiSpec) {
		// Iterate through a collection of specs and evaluate each.
		fmt.Printf("%s\n", spec.GetName())
		// get the lint results
		request := rpc.GetArtifactContentsRequest{
			Name: spec.Name + "/artifacts/" + lintRelation(linter) + "/contents",
		}
		contents, _ := client.GetArtifactContents(ctx, &request)
		if contents == nil {
			return // ignore missing results
		}

		messageType, err := core.MessageTypeForMimeType(contents.GetContentType())
		if err != nil {
			return
		}
		if messageType != "google.cloud.apigee.registry.applications.v1alpha1.Lint" {
			return // ignore unexpected message types
		}

		lint := &rpc.Lint{}
		err = proto.Unmarshal(contents.GetData(), lint)
		if err != nil {
			return
		}

		// generate the stats from the result by counting problems
		lintStats := computeLintStats(lint)

		{
			// Calculate the operation and schema count
			request := rpc.GetArtifactContentsRequest{
				Name: spec.Name + "/artifacts/complexity/contents",
			}
			contents, _ := client.GetArtifactContents(ctx, &request)
			if contents == nil {
				return // ignore missing results
			}
			complexity := &metrics.Complexity{}
			err := proto.Unmarshal(contents.GetData(), complexity)
			if err != nil {
				return
			}

			lintStats.OperationCount = complexity.GetDeleteCount() +
				complexity.GetPutCount() + complexity.GetGetCount() + complexity.GetPostCount()

			lintStats.SchemaCount = complexity.GetSchemaCount()
		}

		storeLintStatsArtifact(ctx, client, spec.GetName(), linter, lintStats)
	})
}

func computeLintStatsProjects(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filter string,
	linter string) error {
	return core.ListProjects(ctx, client, segments, filter, func(project *rpc.Project) {
		if project_segments :=
			names.ProjectRegexp().FindStringSubmatch(project.GetName()); project_segments != nil {
			project_stats := &rpc.LintStats{}

			if err := core.ListAPIs(ctx, client, segments, filter, func(api *rpc.Api) {
				aggregateLintStats(ctx, client, api.GetName(), linter, project_stats)
			}); err != nil {
				return
			}
			// Store the aggregate stats on this api
			storeLintStatsArtifact(ctx, client, project.GetName(), linter, project_stats)
		}
		fmt.Printf("%s\n", project.GetName())
	})
}

func computeLintStatsAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filter string,
	linter string) error {

	return core.ListAPIs(ctx, client, segments, filter, func(api *rpc.Api) {
		if api_segments :=
			names.ApiRegexp().FindStringSubmatch(api.GetName()); api_segments != nil {
			api_stats := &rpc.LintStats{}

			if err := core.ListVersions(ctx, client, segments, filter, func(version *rpc.ApiVersion) {
				aggregateLintStats(ctx, client, version.GetName(), linter, api_stats)
			}); err != nil {
				return
			}
			// Store the aggregate stats on this api
			storeLintStatsArtifact(ctx, client, api.GetName(), linter, api_stats)
		}
		fmt.Printf("%s\n", api.GetName())
	})
}

func computeLintStatsVersions(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filter string,
	linter string) error {

	return core.ListVersions(ctx, client, segments, filter, func(version *rpc.ApiVersion) {
		if version_name_segments :=
			names.VersionRegexp().FindStringSubmatch(version.GetName()); version_name_segments != nil {

			version_stats := &rpc.LintStats{}

			if err := core.ListSpecs(ctx, client, segments, filter, func(spec *rpc.ApiSpec) {
				aggregateLintStats(ctx, client, spec.GetName(), linter, version_stats)
			}); err != nil {
				return
			}
			// Store the aggregate stats on this version
			storeLintStatsArtifact(ctx, client, version.GetName(), linter, version_stats)
		}
		fmt.Printf("%s\n", version.GetName())
	})
}

func storeLintStatsArtifact(ctx context.Context,
	client *gapic.RegistryClient,
	subject string,
	linter string,
	lintStats *rpc.LintStats) error {
	// store the lintstats artifact
	relation := lintStatsRelation(linter)
	messageData, _ := proto.Marshal(lintStats)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("google.cloud.apigee.registry.applications.v1alpha1.LintStats"),
		Contents: messageData,
	}
	return core.SetArtifact(ctx, client, artifact)
}

func aggregateLintStats(ctx context.Context,
	client connection.Client,
	name string,
	linter string,
	aggregateStats *rpc.LintStats) {
	// Calculate the operation and schema count
	request := rpc.GetArtifactContentsRequest{
		Name: name + "/artifacts/" + lintStatsRelation(linter) + "/contents",
	}
	contents, _ := client.GetArtifactContents(ctx, &request)
	if contents == nil {
		return // ignore missing results
	}
	stats := &rpc.LintStats{}
	err := proto.Unmarshal(contents.GetData(), stats)
	if err != nil {
		return
	}

	aggregateStats.OperationCount += stats.OperationCount
	aggregateStats.SchemaCount += stats.SchemaCount
	aggregateStats.ProblemCounts = append(aggregateStats.ProblemCounts, stats.ProblemCounts...)
}

func matchAndHandleLintStatsCmd(
	ctx context.Context,
	client connection.Client,
	name string,
	filter string,
	linter string,
) error {

	var err error
	// First try to match collection names, then try to match resource names.
	if m := names.ProjectsRegexp().FindStringSubmatch(name); m != nil {
		err = computeLintStatsProjects(ctx, client, m, filter, linter)
	} else if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		err = computeLintStatsAPIs(ctx, client, m, filter, linter)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		err = computeLintStatsVersions(ctx, client, m, filter, linter)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		err = computeLintStatsSpecs(ctx, client, m, filter, linter)
	} else if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
		err = computeLintStatsProjects(ctx, client, m, filter, linter)
	} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		err = computeLintStatsAPIs(ctx, client, m, filter, linter)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		err = computeLintStatsVersions(ctx, client, m, filter, linter)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		err = computeLintStatsSpecs(ctx, client, m, filter, linter)
	} else {
		// If nothing matched, return an error.
		return fmt.Errorf("unsupported argument: %s", name)
	}

	return err
}
