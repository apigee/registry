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
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/proto"
)

func conformanceReportId(styleguideId string) string {
	return fmt.Sprintf("conformance-%s", styleguideId)
}

func initializeConformanceReport(specName, styleguideId, project string) *rpc.ConformanceReport {
	// Create an empty conformance report.
	conformanceReport := &rpc.ConformanceReport{
		Id:         conformanceReportId(styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
	}

	// Initialize guideline report groups.
	guidelineState := rpc.Guideline_State(0)
	numStates := guidelineState.Descriptor().Values().Len()
	conformanceReport.GuidelineReportGroups = make([]*rpc.GuidelineReportGroup, numStates)
	for i := 0; i < numStates; i++ {
		conformanceReport.GuidelineReportGroups[i] = &rpc.GuidelineReportGroup{
			State:            rpc.Guideline_State(i),
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

	data, err := core.GetBytesForSpec(ctx, task.Client, task.Spec)
	if err != nil {
		return err
	}
	// Put the spec in a temporary directory.
	root, err := os.MkdirTemp("", "registry-spec-")
	if err != nil {
		return err
	}
	name := filepath.Base(task.Spec.GetName())
	defer os.RemoveAll(root)

	if core.IsZipArchive(task.Spec.GetMimeType()) {
		_, err = core.UnzipArchiveToPath(data, root)
	} else {
		// Write the file to the temporary directory.
		err = os.WriteFile(filepath.Join(root, name), data, 0644)
	}
	if err != nil {
		return err
	}

	// Get project
	spec, err := names.ParseSpecRevision(task.Spec.GetName())
	if err != nil {
		return err
	}

	// Run the linters and compute conformance report
	conformanceReport := initializeConformanceReport(task.Spec.GetName(), task.StyleguideId, spec.ProjectID)
	guidelineReportsMap := make(map[string]int)
	for _, metadata := range task.LintersMetadata {
		linterResponse, err := task.invokeLinter(ctx, root, metadata)
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
		core.PrintMessage(conformanceReport)
		return nil
	}
	return task.storeConformanceReport(ctx, conformanceReport)
}

func (task *ComputeConformanceTask) invokeLinter(
	ctx context.Context,
	specDirectory string,
	metadata *linterMetadata) (*rpc.LinterResponse, error) {
	// Formulate the request.
	requestBytes, err := proto.Marshal(&rpc.LinterRequest{
		SpecDirectory: specDirectory,
		RuleIds:       metadata.rules,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed marshaling linterRequest, Error: %s ", err)
	}

	executableName := getLinterBinaryName(metadata.name)
	cmd := exec.Command(executableName)
	cmd.Stdin = bytes.NewReader(requestBytes)
	cmd.Stderr = os.Stderr

	pluginStartTime := time.Now()
	// Run the linter.
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Running the plugin %s return error: %s", executableName, err)
	}

	pluginElapsedTime := time.Since(pluginStartTime)
	log.Debugf(ctx, "Plugin %s ran in time %s", executableName, pluginElapsedTime)

	// Unmarshal the output bytes into a response object. If there's a failure, log and continue.
	linterResponse := &rpc.LinterResponse{}
	err = proto.Unmarshal(output, linterResponse)
	if err != nil {
		return nil, fmt.Errorf("Failed unmarshalling LinterResponse (plugins must write log messages to stderr, not stdout): %s", err)
	}

	// Check if there were any errors in the plugin.
	if len(linterResponse.GetErrors()) > 0 {
		return nil, fmt.Errorf("Plugin %s encountered errors: %v", executableName, linterResponse.GetErrors())
	}

	return linterResponse, nil
}

func (task *ComputeConformanceTask) computeConformanceReport(
	ctx context.Context,
	conformanceReport *rpc.ConformanceReport,
	guidelineReportsMap map[string]int,
	linterResponse *rpc.LinterResponse,
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

			ruleReport := &rpc.RuleReport{
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
	conformanceReport *rpc.ConformanceReport) error {
	// Store the conformance report.
	messageData, err := proto.Marshal(conformanceReport)
	if err != nil {
		return err
	}

	artifact := &rpc.Artifact{
		Name:     fmt.Sprintf("%s/artifacts/%s", task.Spec.GetName(), conformanceReportId(task.StyleguideId)),
		MimeType: patch.MimeTypeForKind("ConformanceReport"),
		Contents: messageData,
	}
	return core.SetArtifact(ctx, task.Client, artifact)
}
