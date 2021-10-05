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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/cmd/compute/conformance"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
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
				log.WithError(err).Fatal("Failed to get filter from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}

			name := args[0]

			// Ensure that the provided argument is a spec.
			var specSegments []string
			if specSegments = names.SpecRegexp().FindStringSubmatch(name); specSegments == nil {
				log.Fatalf("The provided argument %s does not match the regex of a spec", name)
			}
			spec, err := names.ParseSpec(name)
			if err != nil {
				log.WithError(err).Fatal("Invalid spec")
			}

			projectSegments := []string{"projects", spec.ProjectID}

			err = core.ListArtifacts(ctx, client, projectSegments, filter, true, func(artifact *rpc.Artifact) {
				// Only consider artifacts which have the styleguide mimetype.
				messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
				if err != nil {
					log.WithError(err).Debugf("Failed to get message type for MIME type %q", artifact.GetMimeType())
					return
				}

				if messageType != "google.cloud.apigee.registry.applications.v1alpha1.StyleGuide" {
					// Ignore any artifact that isn't a style guide.
					return
				}

				// Unmarshal the contents of the artifact into a style guide
				styleguide := &rpc.StyleGuide{}
				err = proto.Unmarshal(artifact.GetContents(), styleguide)
				if err != nil {
					log.WithError(err).Debugf("Unmarshal() to StyleGuide failed on artifact of type %s", messageType)
					return
				}

				computeConformanceForStyleGuide(ctx, client, styleguide, specSegments, filter)
			})

			if err != nil {
				log.WithError(err).Fatal("Failed to list artifacts")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	return cmd
}

// computeConformanceForStyleGuide computes and attaches conformance reports as
// artifacts to a spec or a collection of specs.
func computeConformanceForStyleGuide(ctx context.Context,
	client connection.Client,
	styleguide *rpc.StyleGuide,
	specSegments []string,
	filter string) {
	// A mapping between the linter name and the linter, and populate
	// all the rules that the linter should support.
	linterNameToLinter := make(map[string]conformance.Linter)

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
			linter_name := rule.GetLinter()

			// If the linter isn't initialized yet, initialize it.
			if _, ok := linterNameToLinter[linter_name]; !ok {
				linter, err := conformance.CreateLinter(rule.GetLinter())
				normalizedRuleName := canonicalizeRuleName(linter.GetName(), rule.GetLinterRulename())
				ruleNameToGuideline[normalizedRuleName] = guideline
				ruleNameToRule[normalizedRuleName] = rule
				if err != nil {
					// If the linter is unsupported, there is no reason
					// to prematurely exit. We can just ignore this specific
					// linter and log the message to the user.
					log.WithError(err).Debug("Failed to create linter")
				}
				linterNameToLinter[linter_name] = linter
			}

			// Register each rule specified by the style guide to the linter.
			linter := linterNameToLinter[linter_name]
			for _, allowedMimeType := range styleguide.MimeTypes {
				_ = linter.AddRule(allowedMimeType, rule.GetLinterRulename())
			}
		}
	}

	// Initialize task queue.
	taskQueue, wait := core.WorkerPool(ctx, 16)
	defer wait()

	// Generate tasks.
	err := core.ListSpecs(ctx, client, specSegments, filter, func(spec *rpc.ApiSpec) {
		// A list of linters that are used to lint this spec.
		linters := make([]conformance.Linter, 0)

		for _, linter := range linterNameToLinter {
			// If the linter supports the spec's mime type, then it can be used
			// to lint the spec.
			if linter.SupportsMimeType(spec.GetMimeType()) {
				linters = append(linters, linter)
			}
		}

		// Delegate the task of computing the conformance report for this spec
		// to the worker pool.
		if len(linters) > 0 {
			taskQueue <- &computeConformanceTask{
				client:              client,
				spec:                spec,
				linters:             linters,
				styleguideId:        styleguide.GetId(),
				ruleNameToGuideline: ruleNameToGuideline,
				ruleNameToRule:      ruleNameToRule,
			}
		}
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to list specs")
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
	linters             []conformance.Linter
	styleguideId        string
	ruleNameToGuideline map[string]*rpc.Guideline
	ruleNameToRule      map[string]*rpc.Rule
}

func (task *computeConformanceTask) String() string {
	return fmt.Sprintf("compute %s/artifacts/conformance-%s", task.spec.GetName(), task.styleguideId)
}

func (task *computeConformanceTask) Run(ctx context.Context) error {
	// If the spec passed is a compressed archive, then unzip it.
	if core.IsZipArchive(task.spec.GetMimeType()) {
		log.Debugf("Computing conformance report for zipped archive %s", task.spec.GetName())
		// Put the range of specs in a temporary directory
		root, err := ioutil.TempDir("", "registry-protos-")
		if err != nil {
			return err
		}
<<<<<<< HEAD

		// whenever we finish, delete the tmp directory
		defer os.RemoveAll(root)

		// For each file in the compresed archive, compute the
		// conformance report.
		filePaths, err := task.unzipSpecs(ctx, root)
		if err != nil {
			return err
		}

		conformanceReport := task.initializeConformanceReport()
		guidelineIdToGuidelineReport := make(map[string]*rpc.GuidelineReport)
		for _, filePath := range filePaths {
			// Debug the conformance report being computed
			unzippedSpecName := filepath.Base(filePath)
			if !strings.HasSuffix(unzippedSpecName, ".proto") {
				// Currently, only proto files are supported for linting in zipped folders.
				continue
			}

			log.Debugf("Adding conformance for spec %s into report.", unzippedSpecName)

=======

		// whenever we finish, delete the tmp directory
		defer os.RemoveAll(root)

		// For each file in the compresed archive, compute the
		// conformance report.
		filePaths, err := task.unzipSpecs(ctx, root)
		if err != nil {
			return err
		}

		conformanceReport := task.initializeConformanceReport()
		guidelineIdToGuidelineReport := make(map[string]*rpc.GuidelineReport)
		for _, filePath := range filePaths {
			// Debug the conformance report being computed
			unzippedSpecName := filepath.Base(filePath)
			log.Debugf("Adding conformance for spec %s into report.", unzippedSpecName)

>>>>>>> 831220d (Implement linting of Protos with API Linter Plugin)
			err = task.computeConformanceReport(ctx, filePath, conformanceReport, guidelineIdToGuidelineReport)

			// If computing the conformance report for a given spec fails, we should not
			// fail completely (because this doesn't imply that computing the conformance
			// report for another spec will fail, so we should continue.)
			if err != nil {
				log.Log.Errorf(
					"Failed to compute the conformance for spec %s: %s",
					unzippedSpecName,
					err,
				)
			}
		}

		return task.storeConformanceReport(ctx, conformanceReport)
	}

	log.Debugf("Computing conformance report for spec: %s", task.spec.GetName())
	// Get the spec's bytes
	data, err := core.GetBytesForSpec(ctx, task.client, task.spec)
	if err != nil {
		return err
	}

	// Put the spec in a temporary directory.
	root, err := ioutil.TempDir("", "registry-openapi-")
	if err != nil {
		return err
	}
	name := filepath.Base(task.spec.GetName())

	// Defer the deletion of the the temporary directory.
	defer os.RemoveAll(root)

	// Write the file to the temporary directory.
	filePath := filepath.Join(root, name)
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	conformanceReport := task.initializeConformanceReport()
	guidelineIdToGuidelineReport := make(map[string]*rpc.GuidelineReport)
	err = task.computeConformanceReport(ctx, filePath, conformanceReport, guidelineIdToGuidelineReport)
	if err != nil {
		return err
	}
<<<<<<< HEAD

	return task.storeConformanceReport(ctx, conformanceReport)
}

=======

	return task.storeConformanceReport(ctx, conformanceReport)
}

>>>>>>> 831220d (Implement linting of Protos with API Linter Plugin)
func (task *computeConformanceTask) computeConformanceReport(
	ctx context.Context,
	filePath string,
	conformanceReport *rpc.ConformanceReport,
	guidelineIdToGuidelineReport map[string]*rpc.GuidelineReport,
) error {
	// Iterate over all the linters, and lint.
	for _, linter := range task.linters {
		if linter == nil {
			return errors.New("linter is nil")
		}

		// Lint the directory containing the spec.
		lintProblems, err := linter.LintSpec(task.spec.GetMimeType(), filePath)

		// If a linter returned an error, we shouldn't stop linting completely and discard
		// the conformance report for this spec. We should log but still continue, because there
		// may still be useful information from other linters that we may be discarding.
		if err != nil {
<<<<<<< HEAD
=======
			fmt.Println("DUEENEE")
>>>>>>> 831220d (Implement linting of Protos with API Linter Plugin)
			log.Log.Errorf(
				"Linting the spec %s with the linter %s failed %s: %s",
				task.spec.GetName(),
				linter.GetName(),
				err,
			)
		}

		// Iterate over the list of problems returned by the linter.
		for _, problem := range lintProblems {
			normalizedRuleName := canonicalizeRuleName(linter.GetName(), problem.GetRuleId())
			guideline := task.ruleNameToGuideline[normalizedRuleName]
			rule := task.ruleNameToRule[normalizedRuleName]

			// Check if the guideline report for the guideline which contains this rule
			// has already been initialized. If it hasn't then create one.
			if _, ok := guidelineIdToGuidelineReport[guideline.GetId()]; !ok {
				report := task.initializeGuidelineReport()
				guidelineIdToGuidelineReport[guideline.GetId()] = report

				conformanceReport.GuidelineReportGroups[guideline.Status].GuidelineReports =
					append(conformanceReport.GuidelineReportGroups[guideline.Status].GuidelineReports, report)
			}

			// Add the rule report to the appropriate guideline report.
			guidelineReport := guidelineIdToGuidelineReport[guideline.GetId()]
			ruleReport := &rpc.RuleReport{
				RuleName:   rule.GetId(),
				SpecName:   task.spec.GetName(),
				Suggestion: problem.Suggestion,
				Location:   problem.Location,
			}
			guidelineReport.RuleReportGroups[rule.Severity].RuleReports =
				append(guidelineReport.RuleReportGroups[rule.Severity].RuleReports, ruleReport)
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

func (task *computeConformanceTask) initializeGuidelineReport() *rpc.GuidelineReport {
	// Create an empty guideline report.
	guidelineReport := &rpc.GuidelineReport{}

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
		MimeType: core.MimeTypeForMessageType("google.cloud.apigee.registry.applications.v1alpha1.Lint"),
		Contents: messageData,
	}
	return core.SetArtifact(ctx, task.client, artifact)
}

func (task *computeConformanceTask) unzipSpecs(
	ctx context.Context,
	temp_dir_root string) ([]string, error) {
	data, err := core.GetBytesForSpec(ctx, task.client, task.spec)
	if err != nil {
		return nil, err
	}

	// unzip the protos to the temp directory
	return core.UnzipArchiveToPath(data, temp_dir_root+"/protos")
}
