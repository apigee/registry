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

package conformance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func conformanceReportId(styleguideId string) string {
	return fmt.Sprintf("conformance-%s", styleguideId)
}

func initializeConformanceReport(specName, styleguideId, project string) *style.ConformanceReport {
	// Create an empty conformance report.
	conformanceReport := &style.ConformanceReport{
		Id:         conformanceReportId(styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
	}

	// Initialize guideline report groups.
	guidelineState := style.Guideline_State(0)
	numStates := guidelineState.Descriptor().Values().Len()
	conformanceReport.GuidelineReportGroups = make([]*style.GuidelineReportGroup, numStates)
	for i := 0; i < numStates; i++ {
		conformanceReport.GuidelineReportGroups[i] = &style.GuidelineReportGroup{
			State:            style.Guideline_State(i),
			GuidelineReports: make([]*style.GuidelineReport, 0),
		}
	}

	return conformanceReport
}

func initializeGuidelineReport(guidelineID string) *style.GuidelineReport {
	// Create an empty guideline report.
	guidelineReport := &style.GuidelineReport{GuidelineId: guidelineID}

	// Initialize rule report groups.
	ruleSeverity := style.Rule_Severity(0)
	numSeverities := ruleSeverity.Descriptor().Values().Len()
	guidelineReport.RuleReportGroups = make([]*style.RuleReportGroup, numSeverities)
	for i := 0; i < numSeverities; i++ {
		guidelineReport.RuleReportGroups[i] = &style.RuleReportGroup{
			Severity:    style.Rule_Severity(i),
			RuleReports: make([]*style.RuleReport, 0),
		}
	}

	return guidelineReport
}

type ComputeConformanceTask struct {
	Client          connection.RegistryClient
	Spec            *rpc.ApiSpec
	LintersMetadata map[string]*linterMetadata
	StyleguideId    string
	DryRun          bool
}

func (task *ComputeConformanceTask) String() string {
	return fmt.Sprintf("compute %s/artifacts/%s", task.Spec.GetName(), conformanceReportId(task.StyleguideId))
}

func (task *ComputeConformanceTask) Run(ctx context.Context) error {
	log.Debugf(ctx, "Computing conformance report %s/artifacts/%s", task.Spec.GetName(), conformanceReportId(task.StyleguideId))

	root, err := WriteSpecForLinting(ctx, task.Client, task.Spec)
	if root != "" {
		defer os.RemoveAll(root)
	}
	if err != nil {
		return err
	}

	// Get project ID from spec name
	spec, err := names.ParseSpecRevision(task.Spec.GetName())
	if err != nil {
		return err
	}

	// Run the linters and compute conformance report
	conformanceReport := initializeConformanceReport(task.Spec.GetName(), task.StyleguideId, spec.ProjectID)
	guidelineReportsMap := make(map[string]int)
	for _, metadata := range task.LintersMetadata {
		linterResponse, err := RunLinter(ctx, root, metadata)
		// If a linter returned an error, we shouldn't stop linting completely across all linters and
		// discard the conformance report for this spec. We should log but still continue, because there
		// may still be useful information from other linters that we may be discarding.
		if err != nil {
			log.Errorf(ctx, "Linter error: %s", err)
			continue
		}

		task.computeConformanceReport(ctx, conformanceReport, guidelineReportsMap, linterResponse, metadata)
	}

	if task.DryRun {
		fmt.Println(protojson.Format((conformanceReport)))
		return nil
	}
	return task.storeConformanceReport(ctx, conformanceReport)
}

func (task *ComputeConformanceTask) computeConformanceReport(
	ctx context.Context,
	conformanceReport *style.ConformanceReport,
	guidelineReportsMap map[string]int,
	linterResponse *style.LinterResponse,
	linterMetadata *linterMetadata,
) {
	// Process linterResponse to generate conformance report
	lintFiles := linterResponse.Lint.GetFiles()

	for _, lintFile := range lintFiles {
		lintProblems := lintFile.GetProblems()

		// Iterate over the list of problems returned by the linter.
		for _, problem := range lintProblems {
			ruleMetadata, ok := linterMetadata.rulesMetadata[problem.GetRuleId()]
			if !ok {
				// If a problem the linter returned isn't one that we expect
				// then we should ignore it
				continue
			}

			guideline := ruleMetadata.guideline
			guidelineRule := ruleMetadata.guidelineRule

			// Check if the guideline report for the guideline which contains this rule
			// has already been initialized. If it hasn't then create one.
			reportIndex, ok := guidelineReportsMap[guideline.GetId()]
			if !ok {
				guidelineReport := initializeGuidelineReport(guideline.GetId())

				// Create a new entry in the conformance report
				guidelineGroup := conformanceReport.GuidelineReportGroups[guideline.GetState()]
				guidelineGroup.GuidelineReports = append(guidelineGroup.GuidelineReports, guidelineReport)

				// Store the index of this new entry in the map
				reportIndex = len(guidelineGroup.GuidelineReports) - 1
				guidelineReportsMap[guideline.GetId()] = reportIndex
			}

			ruleReport := &style.RuleReport{
				RuleId:      guidelineRule.GetId(),
				Spec:        fmt.Sprintf("%s@%s", task.Spec.GetName(), task.Spec.GetRevisionId()),
				File:        filepath.Base(lintFile.GetFilePath()),
				Suggestion:  problem.Suggestion,
				Location:    problem.Location,
				DisplayName: guidelineRule.GetDisplayName(),
				Description: guidelineRule.GetDescription(),
				DocUri:      guidelineRule.GetDocUri(),
			}
			// Add the rule report to the appropriate guideline report.
			guidelineGroup := conformanceReport.GuidelineReportGroups[guideline.GetState()]
			if reportIndex >= len(guidelineGroup.GuidelineReports) {
				log.Errorf(ctx, "Incorrect data in conformance report. Cannot attach entry for %s", guideline.GetId())
				continue
			}
			ruleGroup := guidelineGroup.GuidelineReports[reportIndex].RuleReportGroups[guidelineRule.GetSeverity()]
			ruleGroup.RuleReports = append(ruleGroup.RuleReports, ruleReport)
		}
	}
}

func (task *ComputeConformanceTask) storeConformanceReport(
	ctx context.Context,
	conformanceReport *style.ConformanceReport) error {
	// Store the conformance report.
	messageData, err := proto.Marshal(conformanceReport)
	if err != nil {
		return err
	}

	artifact := &rpc.Artifact{
		Name:     fmt.Sprintf("%s/artifacts/%s", task.Spec.GetName(), conformanceReportId(task.StyleguideId)),
		MimeType: mime.MimeTypeForKind("ConformanceReport"),
		Contents: messageData,
	}
	return visitor.SetArtifact(ctx, task.Client, artifact)
}
