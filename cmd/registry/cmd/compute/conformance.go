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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func conformanceCommand(ctx context.Context) *cobra.Command {
	var filter string

	cmd := &cobra.Command{
		Use:   "conformance",
		Short: "Compute lint results for API specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get filter from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			name := args[0]

			// Ensure that the provided argument is a spec.
			spec, err := names.ParseSpec(name)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatalf("The provided argument %s does not match the regex of a spec", name)
			}

			artifact, err := names.ParseArtifact(fmt.Sprintf("projects/%s/locations/%s/artifacts/-", spec.ProjectID, names.Location))
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Invalid project")
			}

			err = core.ListArtifacts(ctx, client, artifact, filter, true, func(artifact *rpc.Artifact) {
				// Only consider artifacts which have the styleguide mimetype.
				messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
				if err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Failed to get message type for MIME type %q", artifact.GetMimeType())
					return
				}

				if messageType != "google.cloud.apigeeregistry.applications.v1alpha1.StyleGuide" {
					// Ignore any artifact that isn't a style guide.
					return
				}

				// Unmarshal the contents of the artifact into a style guide
				styleguide := &rpc.StyleGuide{}
				err = proto.Unmarshal(artifact.GetContents(), styleguide)
				if err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Unmarshal() to StyleGuide failed on artifact of type %s", messageType)
					return
				}

				computeConformanceForStyleGuideWithPlugin(ctx, client, styleguide, spec, filter)
			})

			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list artifacts")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	return cmd
}

// computeConformanceForStyleGuide computes and attaches conformance reports as
// artifacts to a spec or a collection of specs.
func computeConformanceForStyleGuideWithPlugin(ctx context.Context,
	client connection.Client,
	styleguide *rpc.StyleGuide,
	spec names.Spec,
	filter string) {
	// A mapping between the linter name and the the names of the
	// rules that the linter should support. These names are names
	// that the linter recognizes i.e. the linter rule names, not
	// the style guide rule names.
	linterNameToRules := make(map[string][]string)

	// A mapping between the canonicalized rule name to the rule
	// object associated with it.
	ruleNameToRule := make(map[string]*rpc.Rule)

	// A mapping between the canonicalized guidline name to the
	// guideline object associated with it.
	ruleNameToGuideline := make(map[string]*rpc.Guideline)

	// Iterate through all the guidelines of the style guide.
	for _, guideline := range styleguide.GetGuidelines() {

		// Iterate through all the rules of the style guide.
		for _, rule := range guideline.GetRules() {

			// Get the name of the linter associated with the rule.
			linterName := rule.GetLinter()

			// If the linter isn't initialized yet, initialize it.
			normalizedRuleName := canonicalizeRuleName(linterName, rule.GetLinterRulename())
			ruleNameToGuideline[normalizedRuleName] = guideline
			ruleNameToRule[normalizedRuleName] = rule
			linterNameToRules[linterName] = append(linterNameToRules[linterName], rule.GetLinterRulename())
		}
	}

	// Initialize task queue.
	taskQueue, wait := core.WorkerPool(ctx, 16)
	defer wait()

	// Generate tasks.
	err := core.ListSpecs(ctx, client, spec, filter, func(spec *rpc.ApiSpec) {
		// Delegate the task of computing the conformance report for this spec
		// to the worker pool.
		taskQueue <- &computeConformanceTask{
			client:              client,
			spec:                spec,
			linterNameToRules:   linterNameToRules,
			styleguideId:        styleguide.GetId(),
			ruleNameToGuideline: ruleNameToGuideline,
			ruleNameToRule:      ruleNameToRule,
		}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to list specs")
	}
}

// canonicalizeRuleName normalizes a rule name according to a linter.
// This allows multiple linters to have the same rule name.
func canonicalizeRuleName(linterName string, ruleName string) string {
	return fmt.Sprintf("%s_%s", linterName, ruleName)
}

func conformanceRelation(styleguideName string) string {
	return "conformance-" + styleguideName
}

type computeConformanceTask struct {
	client              connection.Client
	spec                *rpc.ApiSpec
	linterNameToRules   map[string][]string
	styleguideId        string
	ruleNameToGuideline map[string]*rpc.Guideline
	ruleNameToRule      map[string]*rpc.Rule
}

func (task *computeConformanceTask) String() string {
	return fmt.Sprintf("compute %s/artifacts/conformance-%s", task.spec.GetName(), task.styleguideId)
}

func (task *computeConformanceTask) Run(ctx context.Context) error {
	log.Debugf(ctx, "Computing conformance report for spec: %s", task.spec.GetName())
	// Get the spec's bytes
	data, err := core.GetBytesForSpec(ctx, task.client, task.spec)
	if err != nil {
		return err
	}

	// Put the spec in a temporary directory.
	root, err := ioutil.TempDir("", "registry-spec-")
	if err != nil {
		return err
	}
	name := filepath.Base(task.spec.GetName())

	// Defer the deletion of the the temporary directory.
	defer os.RemoveAll(root)

	if core.IsZipArchive(task.spec.GetMimeType()) {
		// unzip to the temp directory
		_, err = core.UnzipArchiveToPath(data, root)
	} else {
		// Write the file to the temporary directory.
		err = ioutil.WriteFile(filepath.Join(root, name), data, 0644)
	}
	if err != nil {
		return err
	}

	conformanceReport := task.initializeConformanceReport()
	guidelineIdToGuidelineReport := make(map[string]*rpc.GuidelineReport)
	err = task.computeConformanceReport(ctx, root, conformanceReport, guidelineIdToGuidelineReport)
	if err != nil {
		log.Errorf(ctx,
			"Failed to compute the conformance for spec %s: %s",
			task.spec.GetName(),
			err,
		)
		return err
	}

	return task.storeConformanceReport(ctx, conformanceReport)
}

func getLinterBinaryName(linterName string) string {
	return "registry-lint-" + linterName
}

func createLinterRequest(specDirectory string, ruleIds []string) *rpc.LinterRequest {
	return &rpc.LinterRequest{
		SpecDirectory: specDirectory,
		RuleIds:       ruleIds,
	}
}

func (task *computeConformanceTask) computeConformanceReport(
	ctx context.Context,
	specDirectory string,
	conformanceReport *rpc.ConformanceReport,
	guidelineIdToGuidelineReport map[string]*rpc.GuidelineReport,
) error {
	// Iterate over all the linters, and lint.
	for linterName, ruleIds := range task.linterNameToRules {
		// Get the executable name.
		executableName := getLinterBinaryName(linterName)

		// Formulate the request.
		requestBytes, err := proto.Marshal(createLinterRequest(specDirectory, ruleIds))
		if err != nil {
			return err
		}

		// Run the linter.
		cmd := exec.Command(executableName)
		cmd.Stdin = bytes.NewReader(requestBytes)
		cmd.Stderr = os.Stderr
		pluginStartTime := time.Now()
		output, err := cmd.Output()

		// If a linter returned an error, we shouldn't stop linting completely across all linters and
		// discard the conformance report for this spec. We should log but still continue, because there
		// may still be useful information from other linters that we may be discarding.
		if err != nil {
			log.Errorf(ctx, "Running the plugin %s return error: %s", executableName, err)
		}

		pluginElapsedTime := time.Since(pluginStartTime)
		log.Debugf(ctx, "Plugin %s ran in time %s", executableName, pluginElapsedTime)

		// Unmarshal the output bytes into a response object. If there's a failure, log and continue.
		linterResponse := &rpc.LinterResponse{}
		err = proto.Unmarshal(output, linterResponse)
		if err != nil {
			log.Errorf(ctx,
				"Invalid plugin response (plugins must write log messages to stderr, not stdout) : %s",
				err,
			)
			continue
		}

		// Check if there was any errors with the plugin.
		if len(linterResponse.GetErrors()) > 0 {
			for _, err := range linterResponse.GetErrors() {
				log.Errorf(ctx, err)
			}
			continue
		}

		lint := linterResponse.Lint
		lintFiles := lint.GetFiles()

		for _, lintFile := range lintFiles {
			lintProblems := lintFile.GetProblems()

			// Iterate over the list of problems returned by the linter.
			for _, problem := range lintProblems {
				normalizedRuleName := canonicalizeRuleName(linterName, problem.GetRuleId())
				if _, ok := task.ruleNameToGuideline[normalizedRuleName]; !ok {
					// If a problem the linter returned isn't one that we expect
					// then we should ignore it
					continue
				}
				guideline := task.ruleNameToGuideline[normalizedRuleName]
				rule := task.ruleNameToRule[normalizedRuleName]

				// Check if the guideline report for the guideline which contains this rule
				// has already been initialized. If it hasn't then create one.
				if _, ok := guidelineIdToGuidelineReport[guideline.GetId()]; !ok {
					report := task.initializeGuidelineReport(guideline.GetId())
					guidelineIdToGuidelineReport[guideline.GetId()] = report

					conformanceReport.GuidelineReportGroups[guideline.Status].GuidelineReports =
						append(conformanceReport.GuidelineReportGroups[guideline.Status].GuidelineReports, report)
				}

				// Add the rule report to the appropriate guideline report.
				guidelineReport := guidelineIdToGuidelineReport[guideline.GetId()]
				ruleReport := &rpc.RuleReport{
					RuleId:     rule.GetId(),
					SpecName:   task.spec.GetName(),
					FileName:   filepath.Base(lintFile.GetFilePath()),
					Suggestion: problem.Suggestion,
					Location:   problem.Location,
				}
				guidelineReport.RuleReportGroups[rule.Severity].RuleReports =
					append(guidelineReport.RuleReportGroups[rule.Severity].RuleReports, ruleReport)
			}
		}
	}

	return nil
}

func (task *computeConformanceTask) initializeConformanceReport() *rpc.ConformanceReport {
	// Create an empty conformance report.
	conformanceReport := &rpc.ConformanceReport{
		Name:           task.spec.GetName() + "/artifacts/" + conformanceRelation(task.styleguideId),
		StyleguideName: task.styleguideId,
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

func (task *computeConformanceTask) initializeGuidelineReport(guidelineID string) *rpc.GuidelineReport {
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

func (task *computeConformanceTask) storeConformanceReport(
	ctx context.Context,
	conformanceReport *rpc.ConformanceReport) error {
	// Store the conformance report.
	subject := task.spec.GetName()
	messageData, err := proto.Marshal(conformanceReport)
	if err != nil {
		return err
	}

	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + conformanceRelation(task.styleguideId),
		MimeType: core.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.ConformanceReport"),
		Contents: messageData,
	}
	return core.SetArtifact(ctx, task.client, artifact)
}
