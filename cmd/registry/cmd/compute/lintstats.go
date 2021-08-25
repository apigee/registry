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
	return &rpc.LintStats{ProblemCounts: problemCounts, OperationAndSchemaCount: 1}
}

func computeLintStatsSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filter string,
	linter string) (*rpc.LintStats, error) {

	ruleIdToProblemCounts := make(map[string]*rpc.LintProblemCount)
	var operationAndSchemaCount int32 = 0

	if err := core.ListSpecs(ctx, client, segments, filter, func(spec *rpc.ApiSpec) {
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

			lintStats.OperationAndSchemaCount = 1 + complexity.GetDeleteCount() +
				complexity.GetPutCount() + complexity.GetGetCount() +
				complexity.GetPostCount() + complexity.GetSchemaCount()
		}

		storeLintStatsArtifact(ctx, client, spec.GetName(), linter, lintStats)

		aggregateLintStats(ruleIdToProblemCounts, lintStats)
		operationAndSchemaCount += lintStats.GetOperationAndSchemaCount()
	}); err != nil {
		return nil, err
	}

	return constructLintStats(operationAndSchemaCount, ruleIdToProblemCounts), nil
}

func constructLintStats(operationAndSchemaCount int32,
	ruleIdToProblemCounts map[string]*rpc.LintProblemCount) *rpc.LintStats {
	problemCounts := make([]*rpc.LintProblemCount, 0)
	for _, problemCount := range ruleIdToProblemCounts {
		problemCounts = append(problemCounts, problemCount)
	}
	sort.Slice(problemCounts, func(i, j int) bool {
		return problemCounts[i].Count > problemCounts[j].Count
	})

	return &rpc.LintStats{
		OperationAndSchemaCount: operationAndSchemaCount,
		ProblemCounts:           problemCounts,
	}
}

func computeLintStatsProjects(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filter string,
	linter string) (*rpc.LintStats, error) {
	ruleIdToProblemCounts := make(map[string]*rpc.LintProblemCount)
	var operationAndSchemaCount int32 = 0
	if err := core.ListProjects(ctx, client, segments, filter, func(project *rpc.Project) {
		if project_segments :=
			names.ProjectRegexp().FindStringSubmatch(project.GetName()); project_segments != nil {

			aggregateStats, err := computeLintStatsAPIs(ctx, client, project_segments, filter, linter)
			if err != nil {
				return
			}

			// Store the aggregate stats on this project
			storeLintStatsArtifact(ctx, client, project.GetName(), linter, aggregateStats)

			// Aggregate the stats to pass back up
			aggregateLintStats(ruleIdToProblemCounts, aggregateStats)
			operationAndSchemaCount += aggregateStats.GetOperationAndSchemaCount()
		}
		fmt.Printf("%s\n", project.GetName())
	}); err != nil {
		return nil, err
	}

	return constructLintStats(operationAndSchemaCount, ruleIdToProblemCounts), nil
}

func computeLintStatsAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filter string,
	linter string) (*rpc.LintStats, error) {

	ruleIdToProblemCounts := make(map[string]*rpc.LintProblemCount)
	var operationAndSchemaCount int32 = 0
	fmt.Println(segments)
	if err := core.ListAPIs(ctx, client, segments, filter, func(api *rpc.Api) {
		if api_segments :=
			names.ApiRegexp().FindStringSubmatch(api.GetName()); api_segments != nil {

			aggregateStats, err := computeLintStatsVersions(ctx, client, api_segments, filter, linter)
			if err != nil {
				return
			}
			// Store the aggregate stats on this api
			storeLintStatsArtifact(ctx, client, api.GetName(), linter, aggregateStats)

			// Aggregate the stats to pass back up
			aggregateLintStats(ruleIdToProblemCounts, aggregateStats)
			operationAndSchemaCount += aggregateStats.GetOperationAndSchemaCount()
		}
		fmt.Printf("%s\n", api.GetName())
	}); err != nil {
		return nil, err
	}

	return constructLintStats(operationAndSchemaCount, ruleIdToProblemCounts), nil
}

func computeLintStatsVersions(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filter string,
	linter string) (*rpc.LintStats, error) {

	ruleIdToProblemCounts := make(map[string]*rpc.LintProblemCount)
	var operationAndSchemaCount int32 = 0
	if err := core.ListVersions(ctx, client, segments, filter, func(version *rpc.ApiVersion) {
		if version_name_segments :=
			names.VersionRegexp().FindStringSubmatch(version.GetName()); version_name_segments != nil {

			aggregateStats, err := computeLintStatsSpecs(ctx, client, version_name_segments, filter, linter)
			if err != nil {
				return
			}
			// Store the aggregate stats on this version
			storeLintStatsArtifact(ctx, client, version.GetName(), linter, aggregateStats)

			// Aggregate the stats to pass back up
			aggregateLintStats(ruleIdToProblemCounts, aggregateStats)
			operationAndSchemaCount += aggregateStats.GetOperationAndSchemaCount()
		}
		fmt.Printf("%s\n", version.GetName())
	}); err != nil {
		return nil, err
	}

	return constructLintStats(operationAndSchemaCount, ruleIdToProblemCounts), nil
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

func aggregateLintStats(ruleIdToProblemCounts map[string]*rpc.LintProblemCount,
	lintStats *rpc.LintStats) {
	for _, problemCount := range lintStats.GetProblemCounts() {
		if _, ok := ruleIdToProblemCounts[problemCount.GetRuleId()]; !ok {
			ruleIdToProblemCounts[problemCount.GetRuleId()] =
				&rpc.LintProblemCount{
					RuleId:     problemCount.GetRuleId(),
					RuleDocUri: problemCount.GetRuleDocUri(),
				}
		}
		ruleIdToProblemCounts[problemCount.GetRuleId()].Count++
	}
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
		fmt.Println("Here")
		_, err = computeLintStatsProjects(ctx, client, m, filter, linter)
	} else if m := names.ApisRegexp().FindStringSubmatch(name); m != nil {
		_, err = computeLintStatsAPIs(ctx, client, m, filter, linter)
	} else if m := names.VersionsRegexp().FindStringSubmatch(name); m != nil {
		_, err = computeLintStatsVersions(ctx, client, m, filter, linter)
	} else if m := names.SpecsRegexp().FindStringSubmatch(name); m != nil {
		_, err = computeLintStatsSpecs(ctx, client, m, filter, linter)
	} else if m := names.ProjectRegexp().FindStringSubmatch(name); m != nil {
		_, err = computeLintStatsProjects(ctx, client, m, filter, linter)
	} else if m := names.ApiRegexp().FindStringSubmatch(name); m != nil {
		_, err = computeLintStatsAPIs(ctx, client, m, filter, linter)
	} else if m := names.VersionRegexp().FindStringSubmatch(name); m != nil {
		_, err = computeLintStatsVersions(ctx, client, m, filter, linter)
	} else if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
		_, err = computeLintStatsSpecs(ctx, client, m, filter, linter)
	} else {
		// If nothing matched, return an error.
		return fmt.Errorf("unsupported argument: %s", name)
	}

	return err
}
