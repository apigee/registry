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
	"testing"

	"github.com/apigee/registry/pkg/artifacts"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

const project = "demo"
const specName = "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi"
const revisionId = "abcdef"
const styleguideId = "openapi-test"

// This test will catch any changes made to the original status values.
func TestInitializeConformanceReport(t *testing.T) {
	want := &artifacts.ConformanceReport{
		Id:         fmt.Sprintf("conformance-%s", styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
		GuidelineReportGroups: []*artifacts.GuidelineReportGroup{
			{
				State:            artifacts.Guideline_STATE_UNSPECIFIED,
				GuidelineReports: make([]*artifacts.GuidelineReport, 0),
			},
			{
				State:            artifacts.Guideline_PROPOSED,
				GuidelineReports: make([]*artifacts.GuidelineReport, 0),
			},
			{
				State:            artifacts.Guideline_ACTIVE,
				GuidelineReports: make([]*artifacts.GuidelineReport, 0),
			},
			{
				State:            artifacts.Guideline_DEPRECATED,
				GuidelineReports: make([]*artifacts.GuidelineReport, 0),
			},
			{
				State:            artifacts.Guideline_DISABLED,
				GuidelineReports: make([]*artifacts.GuidelineReport, 0),
			},
		},
	}

	got := initializeConformanceReport(specName, styleguideId, project)
	opts := cmp.Options{
		protocmp.Transform(),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}
	if !cmp.Equal(want, got, opts) {
		t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(want, got, opts))
	}
}

// This test will catch any changes made to the original severity values.
func TestInitializeGuidelineReport(t *testing.T) {
	want := &artifacts.GuidelineReport{
		GuidelineId: styleguideId,
		RuleReportGroups: []*artifacts.RuleReportGroup{
			{
				Severity:    artifacts.Rule_SEVERITY_UNSPECIFIED,
				RuleReports: make([]*artifacts.RuleReport, 0),
			},
			{
				Severity:    artifacts.Rule_ERROR,
				RuleReports: make([]*artifacts.RuleReport, 0),
			},
			{
				Severity:    artifacts.Rule_WARNING,
				RuleReports: make([]*artifacts.RuleReport, 0),
			},
			{
				Severity:    artifacts.Rule_INFO,
				RuleReports: make([]*artifacts.RuleReport, 0),
			},
			{
				Severity:    artifacts.Rule_HINT,
				RuleReports: make([]*artifacts.RuleReport, 0),
			},
		},
	}

	got := initializeGuidelineReport(styleguideId)
	opts := cmp.Options{
		protocmp.Transform(),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}
	if !cmp.Equal(want, got, opts) {
		t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(want, got, opts))
	}
}

func TestComputeConformanceReport(t *testing.T) {
	tests := []struct {
		desc           string
		linterResponse *artifacts.LinterResponse
		linterMetadata *linterMetadata
		wantReport     *artifacts.ConformanceReport
	}{
		// Test basic flow.
		{
			desc: "Normal case",
			linterResponse: &artifacts.LinterResponse{
				Lint: &artifacts.Lint{
					Name: "sample-result",
					Files: []*artifacts.LintFile{
						{
							FilePath: "test-result-file",
							Problems: []*artifacts.LintProblem{
								{
									Message:    "no-$ref-siblings violated",
									RuleId:     "no-$ref-siblings",
									Suggestion: "fix no-$ref-siblings",
									Location: &artifacts.LintLocation{
										StartPosition: &artifacts.LintPosition{LineNumber: 11, ColumnNumber: 25},
										EndPosition:   &artifacts.LintPosition{LineNumber: 11, ColumnNumber: 32},
									},
								},
							},
						},
					},
				},
			},
			linterMetadata: &linterMetadata{
				name:  "sample-linter",
				rules: []string{"no-$ref-siblings"},
				rulesMetadata: map[string]*ruleMetadata{
					"no-$ref-siblings": {
						guidelineRule: &artifacts.Rule{
							Id:       "norefsiblings",
							Severity: artifacts.Rule_ERROR,
						},
						guideline: &artifacts.Guideline{
							Id:    "refproperties",
							State: artifacts.Guideline_ACTIVE,
						},
					},
				},
			},
			wantReport: &artifacts.ConformanceReport{
				Id:         fmt.Sprintf("conformance-%s", styleguideId),
				Kind:       "ConformanceReport",
				Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
				GuidelineReportGroups: []*artifacts.GuidelineReportGroup{
					{State: artifacts.Guideline_STATE_UNSPECIFIED},
					{State: artifacts.Guideline_PROPOSED},
					{
						State: artifacts.Guideline_ACTIVE,
						GuidelineReports: []*artifacts.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*artifacts.RuleReportGroup{
									{Severity: artifacts.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: artifacts.Rule_ERROR,
										RuleReports: []*artifacts.RuleReport{
											{
												RuleId:     "norefsiblings",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix no-$ref-siblings",
												Location: &artifacts.LintLocation{
													StartPosition: &artifacts.LintPosition{LineNumber: 11, ColumnNumber: 25},
													EndPosition:   &artifacts.LintPosition{LineNumber: 11, ColumnNumber: 32},
												},
											},
										},
									},
									{Severity: artifacts.Rule_WARNING},
									{Severity: artifacts.Rule_INFO},
									{Severity: artifacts.Rule_HINT},
								},
							},
						},
					},
					{State: artifacts.Guideline_DEPRECATED},
					{State: artifacts.Guideline_DISABLED},
				},
			},
		},
		// Test: LinterResponse includes multiple rule violations. Validate that only the once configured in the styleguide show up in the final conformance report.
		{
			desc: "Multiple violations",
			linterResponse: &artifacts.LinterResponse{
				Lint: &artifacts.Lint{
					Name: "sample-result",
					Files: []*artifacts.LintFile{
						{
							FilePath: "test-result-file",
							Problems: []*artifacts.LintProblem{
								{
									Message:    "tag-description violated",
									RuleId:     "tag-description",
									Suggestion: "fix tag-description",
								},
								{
									Message:    "operation-description violated",
									RuleId:     "operation-description",
									Suggestion: "fix operation-description",
								},
							},
						},
					},
				},
			},
			linterMetadata: &linterMetadata{
				name:  "sample-linter",
				rules: []string{"operation-description"},
				rulesMetadata: map[string]*ruleMetadata{
					"operation-description": {
						guidelineRule: &artifacts.Rule{
							Id:       "operationdescription",
							Severity: artifacts.Rule_ERROR,
						},
						guideline: &artifacts.Guideline{
							Id:    "descriptionproperties",
							State: artifacts.Guideline_ACTIVE,
						},
					},
				},
			},
			wantReport: &artifacts.ConformanceReport{
				Id:         fmt.Sprintf("conformance-%s", styleguideId),
				Kind:       "ConformanceReport",
				Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
				GuidelineReportGroups: []*artifacts.GuidelineReportGroup{
					{State: artifacts.Guideline_STATE_UNSPECIFIED},
					{State: artifacts.Guideline_PROPOSED},
					{
						State: artifacts.Guideline_ACTIVE,
						GuidelineReports: []*artifacts.GuidelineReport{
							{
								GuidelineId: "descriptionproperties",
								RuleReportGroups: []*artifacts.RuleReportGroup{
									{Severity: artifacts.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: artifacts.Rule_ERROR,
										RuleReports: []*artifacts.RuleReport{
											{
												RuleId:     "operationdescription",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix operation-description",
											},
										},
									},
									{Severity: artifacts.Rule_WARNING},
									{Severity: artifacts.Rule_INFO},
									{Severity: artifacts.Rule_HINT},
								},
							},
						},
					},
					{State: artifacts.Guideline_DEPRECATED},
					{State: artifacts.Guideline_DISABLED},
				},
			},
		},
		// Test: Multiple rules are defined in multiple guidelines, check conformance report gets generated accurately.
		{
			desc: "Multiple guidelines",
			linterResponse: &artifacts.LinterResponse{
				Lint: &artifacts.Lint{
					Name: "sample-result",
					Files: []*artifacts.LintFile{
						{
							FilePath: "test-result-file",
							Problems: []*artifacts.LintProblem{
								{
									Message:    "tag-description violated",
									RuleId:     "tag-description",
									Suggestion: "fix tag-description",
								},
								{
									Message:    "operation-description violated",
									RuleId:     "operation-description",
									Suggestion: "fix operation-description",
								},
								{
									Message:    "info-description violated",
									RuleId:     "info-description",
									Suggestion: "fix info-description",
								},
								{
									Message:    "description-contains-no-tags violated",
									RuleId:     "description-contains-no-tags",
									Suggestion: "fix description-contains-no-tags",
								},
								{
									Message:    "no-$ref-siblings violated",
									RuleId:     "no-$ref-siblings",
									Suggestion: "fix no-$ref-siblings",
								},
							},
						},
					},
				},
			},
			linterMetadata: &linterMetadata{
				name:  "sample-linter",
				rules: []string{"operation-description", "tag-description", "info-description", "no-$ref-siblings"},
				rulesMetadata: map[string]*ruleMetadata{
					"operation-description": {
						guidelineRule: &artifacts.Rule{
							Id:       "operationdescription",
							Severity: artifacts.Rule_ERROR,
						},
						guideline: &artifacts.Guideline{
							Id:    "descriptionproperties",
							State: artifacts.Guideline_ACTIVE,
						},
					},
					"tag-description": {
						guidelineRule: &artifacts.Rule{
							Id:       "tagdescription",
							Severity: artifacts.Rule_WARNING,
						},
						guideline: &artifacts.Guideline{
							Id:    "descriptionproperties",
							State: artifacts.Guideline_ACTIVE,
						},
					},
					"info-description": {
						guidelineRule: &artifacts.Rule{
							Id:       "infodescription",
							Severity: artifacts.Rule_WARNING,
						},
						guideline: &artifacts.Guideline{
							Id:    "descriptionproperties",
							State: artifacts.Guideline_ACTIVE,
						},
					},
					"no-$ref-siblings": {
						guidelineRule: &artifacts.Rule{
							Id:       "norefsiblings",
							Severity: artifacts.Rule_ERROR,
						},
						guideline: &artifacts.Guideline{
							Id:    "refproperties",
							State: artifacts.Guideline_PROPOSED,
						},
					},
				},
			},
			wantReport: &artifacts.ConformanceReport{
				Id:         fmt.Sprintf("conformance-%s", styleguideId),
				Kind:       "ConformanceReport",
				Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
				GuidelineReportGroups: []*artifacts.GuidelineReportGroup{
					{State: artifacts.Guideline_STATE_UNSPECIFIED},
					{
						State: artifacts.Guideline_PROPOSED,
						GuidelineReports: []*artifacts.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*artifacts.RuleReportGroup{
									{Severity: artifacts.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: artifacts.Rule_ERROR,
										RuleReports: []*artifacts.RuleReport{
											{
												RuleId:     "norefsiblings",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix no-$ref-siblings",
											},
										},
									},
									{Severity: artifacts.Rule_WARNING},
									{Severity: artifacts.Rule_INFO},
									{Severity: artifacts.Rule_HINT},
								},
							},
						},
					},
					{
						State: artifacts.Guideline_ACTIVE,
						GuidelineReports: []*artifacts.GuidelineReport{
							{
								GuidelineId: "descriptionproperties",
								RuleReportGroups: []*artifacts.RuleReportGroup{
									{Severity: artifacts.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: artifacts.Rule_ERROR,
										RuleReports: []*artifacts.RuleReport{
											{
												RuleId:     "operationdescription",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix operation-description",
											},
										},
									},
									{
										Severity: artifacts.Rule_WARNING,
										RuleReports: []*artifacts.RuleReport{
											{
												RuleId:     "tagdescription",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix tag-description",
											},
											{
												RuleId:     "infodescription",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix info-description",
											},
										},
									},
									{Severity: artifacts.Rule_INFO},
									{Severity: artifacts.Rule_HINT},
								},
							},
						},
					},
					{State: artifacts.Guideline_DEPRECATED},
					{State: artifacts.Guideline_DISABLED},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			task := &ComputeConformanceTask{
				Spec: &rpc.ApiSpec{
					Name:       specName,
					RevisionId: "abcdef",
				},
				StyleguideId: styleguideId,
			}

			gotReport := initializeConformanceReport(task.Spec.GetName(), task.StyleguideId, project)
			guidelineReportsMap := make(map[string]int)
			task.computeConformanceReport(ctx, gotReport, guidelineReportsMap, test.linterResponse, test.linterMetadata)

			opts := cmp.Options{
				protocmp.IgnoreFields(&artifacts.RuleReport{}),
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.wantReport, gotReport, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantReport, gotReport, opts))
			}
		})
	}
}

// Test the scenario where there are preexisting entries in the conformance report from other linters.
func TestPreExistingConformanceReport(t *testing.T) {
	linterResponse := &artifacts.LinterResponse{
		Lint: &artifacts.Lint{
			Name: "sample-result",
			Files: []*artifacts.LintFile{
				{
					FilePath: "test-result-file",
					Problems: []*artifacts.LintProblem{
						{
							Message:    "operation-description violated",
							RuleId:     "operation-description",
							Suggestion: "fix operation-description",
						},
					},
				},
			},
		},
	}

	linterMetadata := &linterMetadata{
		name:  "sample-linter",
		rules: []string{"operation-description"},
		rulesMetadata: map[string]*ruleMetadata{
			"operation-description": {
				guidelineRule: &artifacts.Rule{
					Id:       "operationdescription",
					Severity: artifacts.Rule_ERROR,
				},
				guideline: &artifacts.Guideline{
					Id:    "descriptionproperties",
					State: artifacts.Guideline_ACTIVE,
				},
			},
		},
	}

	preexistingReport := &artifacts.ConformanceReport{
		Id:         fmt.Sprintf("conformance-%s", styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
		GuidelineReportGroups: []*artifacts.GuidelineReportGroup{
			{State: artifacts.Guideline_STATE_UNSPECIFIED},
			{State: artifacts.Guideline_PROPOSED},
			{
				State: artifacts.Guideline_ACTIVE,
				GuidelineReports: []*artifacts.GuidelineReport{
					{
						GuidelineId: "descriptionproperties",
						RuleReportGroups: []*artifacts.RuleReportGroup{
							{Severity: artifacts.Rule_SEVERITY_UNSPECIFIED},
							{
								Severity: artifacts.Rule_ERROR,
								RuleReports: []*artifacts.RuleReport{
									{
										RuleId:     "tagdescription",
										Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
										File:       "test-result-file",
										Suggestion: "fix tag-description",
									},
								},
							},
							{Severity: artifacts.Rule_WARNING},
							{Severity: artifacts.Rule_INFO},
							{Severity: artifacts.Rule_HINT},
						},
					},
				},
			},
			{State: artifacts.Guideline_DEPRECATED},
			{State: artifacts.Guideline_DISABLED},
		},
	}

	guidelineReportsMap := map[string]int{
		"descriptionproperties": 0,
	}

	wantReport := &artifacts.ConformanceReport{
		Id:         fmt.Sprintf("conformance-%s", styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
		GuidelineReportGroups: []*artifacts.GuidelineReportGroup{
			{State: artifacts.Guideline_STATE_UNSPECIFIED},
			{State: artifacts.Guideline_PROPOSED},
			{
				State: artifacts.Guideline_ACTIVE,
				GuidelineReports: []*artifacts.GuidelineReport{
					{
						GuidelineId: "descriptionproperties",
						RuleReportGroups: []*artifacts.RuleReportGroup{
							{Severity: artifacts.Rule_SEVERITY_UNSPECIFIED},
							{
								Severity: artifacts.Rule_ERROR,
								RuleReports: []*artifacts.RuleReport{
									{
										RuleId:     "tagdescription",
										Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
										File:       "test-result-file",
										Suggestion: "fix tag-description",
									},
									{
										RuleId:     "operationdescription",
										Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
										File:       "test-result-file",
										Suggestion: "fix operation-description",
									},
								},
							},
							{Severity: artifacts.Rule_WARNING},
							{Severity: artifacts.Rule_INFO},
							{Severity: artifacts.Rule_HINT},
						},
					},
				},
			},
			{State: artifacts.Guideline_DEPRECATED},
			{State: artifacts.Guideline_DISABLED},
		},
	}

	ctx := context.Background()

	task := &ComputeConformanceTask{
		Spec: &rpc.ApiSpec{
			Name:       specName,
			RevisionId: revisionId,
		},
		StyleguideId: styleguideId,
	}

	task.computeConformanceReport(ctx, preexistingReport, guidelineReportsMap, linterResponse, linterMetadata)

	opts := cmp.Options{
		protocmp.IgnoreFields(&artifacts.RuleReport{}),
		protocmp.Transform(),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}
	if !cmp.Equal(wantReport, preexistingReport, opts) {
		t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(wantReport, preexistingReport, opts))
	}
}
