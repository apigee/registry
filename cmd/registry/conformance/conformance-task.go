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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

func conformanceReportName(specName string, styleguideName string) string {
	return fmt.Sprintf("%s/artifacts/conformance-%s", specName, styleguideName)
}

func initializeConformanceReport(
	specName string,
	styleguideId string) *rpc.ConformanceReport {
	// Create an empty conformance report.
	conformanceReport := &rpc.ConformanceReport{
		Name:           conformanceReportName(specName, styleguideId),
		StyleguideName: styleguideId,
	}

	// Initialize guideline report groups.
	guidelineStatus := rpc.Guideline_Status(0)
	numStatuses := guidelineStatus.Descriptor().Values().Len()
	conformanceReport.GuidelineReportGroups = make([]*rpc.GuidelineReportGroup, numStatuses)
	for i := 0; i < numStatuses; i++ {
		conformanceReport.GuidelineReportGroups[i] = &rpc.GuidelineReportGroup{
			Status:           rpc.Guideline_Status(i),
			GuidelineReports: make([]*rpc.GuidelineReport, 0),
		}
	}

	return conformanceReport
}

func initializeGuidelineReport(guidelineID string) *rpc.GuidelineReport {
	// Create an empty guideline report.
	guidelineReport := &rpc.GuidelineReport{GuidelineId: guidelineID}

	// Initialize rule report groups.
	ruleSeverity := rpc.Rule_Severity(0)
	numSeverities := ruleSeverity.Descriptor().Values().Len()
	guidelineReport.RuleReportGroups = make([]*rpc.RuleReportGroup, numSeverities)
	for i := 0; i < numSeverities; i++ {
		guidelineReport.RuleReportGroups[i] = &rpc.RuleReportGroup{
			Severity:    rpc.Rule_Severity(i),
			RuleReports: make([]*rpc.RuleReport, 0),
		}
	}

	return guidelineReport
}

type ComputeConformanceTask struct {
	Client          connection.Client
	Spec            *rpc.ApiSpec
	LintersMetadata map[string]*linterMetadata
	StyleguideId    string
}

func (task *ComputeConformanceTask) String() string {
	return fmt.Sprintf("compute %s", conformanceReportName(task.Spec.GetName(), task.StyleguideId))
}

func (task *ComputeConformanceTask) Run(ctx context.Context) error {
	log.Debugf(ctx, "Computing conformance report %s", conformanceReportName(task.Spec.GetName(), task.StyleguideId))

	// Download spec

	data, err := core.GetBytesForSpec(ctx, task.Client, task.Spec)
	if err != nil {
		return err
	}
	// Put the spec in a temporary directory.
	root, err := ioutil.TempDir("", "registry-spec-")
	if err != nil {
		return err
	}
	name := filepath.Base(task.Spec.GetName())
	defer os.RemoveAll(root)

	if core.IsZipArchive(task.Spec.GetMimeType()) {
		_, err = core.UnzipArchiveToPath(data, root)
	} else {
		// Write the file to the temporary directory.
		err = ioutil.WriteFile(filepath.Join(root, name), data, 0644)
	}
	if err != nil {
		return err
	}

	// Generate conformance report
	conformanceReport := initializeConformanceReport(task.Spec.GetName(), task.StyleguideId)
	err = task.computeConformanceReport(ctx, root, conformanceReport)
	if err != nil {
		log.Errorf(ctx,
			"Failed to compute the conformance for spec %s: %s",
			task.Spec.GetName(),
			err,
		)
		return err
	}

	return task.storeConformanceReport(ctx, conformanceReport)
}

func (task *ComputeConformanceTask) computeConformanceReport(
	ctx context.Context,
	specDirectory string,
	conformanceReport *rpc.ConformanceReport,
) error {

	guidelineReportsMap := make(map[string]*rpc.GuidelineReport)

	// Iterate over all the linters, and lint.
	for _, metadata := range task.LintersMetadata {

		linterResponse, err := invokeLinter(ctx, specDirectory, metadata)
		// If a linter returned an error, we shouldn't stop linting completely across all linters and
		// discard the conformance report for this spec. We should log but still continue, because there
		// may still be useful information from other linters that we may be discarding.
		if err != nil {
			log.Errorf(ctx, "Linter error: %s", err)
			continue
		}

		// Process linterResponse to generate conformance report
		lintFiles := linterResponse.Lint.GetFiles()

		for _, lintFile := range lintFiles {
			lintProblems := lintFile.GetProblems()

			// Iterate over the list of problems returned by the linter.
			for _, problem := range lintProblems {
				ruleMetadata, ok := metadata.rulesMetadata[problem.GetRuleId()]
				if !ok {
					// If a problem the linter returned isn't one that we expect
					// then we should ignore it
					continue
				}

				guideline := ruleMetadata.guideline
				guidelineRule := ruleMetadata.guidelineRule

				// Check if the guideline report for the guideline which contains this rule
				// has already been initialized. If it hasn't then create one.
				guidelineReport, ok := guidelineReportsMap[guideline.GetId()]
				if !ok {
					guidelineReport = initializeGuidelineReport(guideline.GetId())

					guidelineReportsMap[guideline.GetId()] = guidelineReport
					conformanceReport.GuidelineReportGroups[guideline.Status].GuidelineReports =
						append(conformanceReport.GuidelineReportGroups[guideline.Status].GuidelineReports, guidelineReport)
				}

				// Add the rule report to the appropriate guideline report.
				ruleReport := &rpc.RuleReport{
					RuleId:     guidelineRule.GetId(),
					SpecName:   task.Spec.GetName(),
					FileName:   filepath.Base(lintFile.GetFilePath()),
					Suggestion: problem.Suggestion,
					Location:   problem.Location,
				}
				guidelineReport.RuleReportGroups[guidelineRule.Severity].RuleReports =
					append(guidelineReport.RuleReportGroups[guidelineRule.Severity].RuleReports, ruleReport)
			}
		}
	}

	return nil
}

func (task *ComputeConformanceTask) storeConformanceReport(
	ctx context.Context,
	conformanceReport *rpc.ConformanceReport) error {
	// Store the conformance report.
	messageData, err := proto.Marshal(conformanceReport)
	if err != nil {
		return err
	}

	artifact := &rpc.Artifact{
		Name:     conformanceReportName(task.Spec.GetName(), task.StyleguideId),
		MimeType: core.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.ConformanceReport"),
		Contents: messageData,
	}
	return core.SetArtifact(ctx, task.Client, artifact)
}
