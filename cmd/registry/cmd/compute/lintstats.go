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
	"sort"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	metrics "github.com/googleapis/gnostic/metrics"
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
				log.WithError(err).Fatal("Failed to get filter from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}

			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}

			// Generate tasks.
			name := args[0]

			err = matchAndHandleLintStatsCmd(ctx, client, adminClient, name, filter, linter)
			if err != nil {
				log.WithError(err).Fatal("Failed to match or handle command")
			}
		},
	}

	cmd.Flags().StringVar(&linter, "linter", "", "The name of the linter whose results will be used to compute stats (aip|spectral|gnostic)")
	_ = cmd.MarkFlagRequired("linter")
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
	spec names.Spec,
	filter string,
	linter string) error {

	return core.ListSpecs(ctx, client, spec, filter, func(spec *rpc.ApiSpec) {
		// Iterate through a collection of specs and evaluate each.
		log.Debug(spec.GetName())
		// get the lint results
		request := rpc.GetArtifactContentsRequest{
			Name: spec.Name + "/artifacts/" + lintRelation(linter),
		}
		contents, _ := client.GetArtifactContents(ctx, &request)
		if contents == nil {
			return // ignore missing results
		}

		messageType, err := core.MessageTypeForMimeType(contents.GetContentType())
		if err != nil {
			return
		}
		if messageType != "google.cloud.apigeeregistry.applications.v1alpha1.Lint" {
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
				Name: spec.Name + "/artifacts/complexity",
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

		_ = storeLintStatsArtifact(ctx, client, spec.GetName(), linter, lintStats)
	})
}

func computeLintStatsProjects(ctx context.Context,
	client *gapic.RegistryClient,
	adminClient *gapic.AdminClient,
	projectName names.Project,
	filter string,
	linter string) error {
	return core.ListProjects(ctx, adminClient, projectName, filter, func(project *rpc.Project) {
		project_stats := &rpc.LintStats{}

		if err := core.ListAPIs(ctx, client, projectName.Api(""), filter, func(api *rpc.Api) {
			aggregateLintStats(ctx, client, api.GetName(), linter, project_stats)
		}); err != nil {
			return
		}
		// Store the aggregate stats on this project
		_ = storeLintStatsArtifact(ctx, client, project.GetName()+"/locations/global", linter, project_stats)
		log.Debug(project.GetName())
	})
}

func computeLintStatsAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	apiName names.Api,
	filter string,
	linter string) error {

	return core.ListAPIs(ctx, client, apiName, filter, func(api *rpc.Api) {
		api_stats := &rpc.LintStats{}

		if err := core.ListVersions(ctx, client, apiName.Version(""), filter, func(version *rpc.ApiVersion) {
			aggregateLintStats(ctx, client, version.GetName(), linter, api_stats)
		}); err != nil {
			return
		}
		// Store the aggregate stats on this api
		_ = storeLintStatsArtifact(ctx, client, api.GetName(), linter, api_stats)
		log.Debug(api.GetName())
	})
}

func computeLintStatsVersions(ctx context.Context,
	client *gapic.RegistryClient,
	versionName names.Version,
	filter string,
	linter string) error {

	return core.ListVersions(ctx, client, versionName, filter, func(version *rpc.ApiVersion) {
		version_stats := &rpc.LintStats{}

		if err := core.ListSpecs(ctx, client, versionName.Spec(""), filter, func(spec *rpc.ApiSpec) {
			aggregateLintStats(ctx, client, spec.GetName(), linter, version_stats)
		}); err != nil {
			return
		}
		// Store the aggregate stats on this version
		_ = storeLintStatsArtifact(ctx, client, version.GetName(), linter, version_stats)
		log.Debug(version.GetName())
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
		MimeType: core.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.LintStats"),
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
		Name: name + "/artifacts/" + lintStatsRelation(linter),
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
	adminClient connection.AdminClient,
	name string,
	filter string,
	linter string,
) error {

	var err error
	// First try to match collection names, then try to match resource names.
	if project, err := names.ParseProjectCollection(name); err == nil {
		err = computeLintStatsProjects(ctx, client, adminClient, project, filter, linter)
	} else if api, err := names.ParseApiCollection(name); err == nil {
		err = computeLintStatsAPIs(ctx, client, api, filter, linter)
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		err = computeLintStatsVersions(ctx, client, version, filter, linter)
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		err = computeLintStatsSpecs(ctx, client, spec, filter, linter)
	} else if project, err := names.ParseProject(name); err == nil {
		err = computeLintStatsProjects(ctx, client, adminClient, project, filter, linter)
	} else if api, err := names.ParseApi(name); err == nil {
		err = computeLintStatsAPIs(ctx, client, api, filter, linter)
	} else if version, err := names.ParseVersion(name); err == nil {
		err = computeLintStatsVersions(ctx, client, version, filter, linter)
	} else if spec, err := names.ParseSpec(name); err == nil {
		err = computeLintStatsSpecs(ctx, client, spec, filter, linter)
	} else {
		// If nothing matched, return an error.
		return fmt.Errorf("unsupported argument: %s", name)
	}

	return err
}
