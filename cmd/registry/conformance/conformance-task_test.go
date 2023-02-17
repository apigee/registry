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
	want := &rpc.ConformanceReport{
		Id:         fmt.Sprintf("conformance-%s", styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
		GuidelineReportGroups: []*rpc.GuidelineReportGroup{
			{
				State:            rpc.Guideline_STATE_UNSPECIFIED,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
			},
			{
				State:            rpc.Guideline_PROPOSED,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
			},
			{
				State:            rpc.Guideline_ACTIVE,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
			},
			{
				State:            rpc.Guideline_DEPRECATED,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
			},
			{
				State:            rpc.Guideline_DISABLED,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
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
	want := &rpc.GuidelineReport{
		GuidelineId: styleguideId,
		RuleReportGroups: []*rpc.RuleReportGroup{
			{
				Severity:    rpc.Rule_SEVERITY_UNSPECIFIED,
				RuleReports: make([]*rpc.RuleReport, 0),
			},
			{
				Severity:    rpc.Rule_ERROR,
				RuleReports: make([]*rpc.RuleReport, 0),
			},
			{
				Severity:    rpc.Rule_WARNING,
				RuleReports: make([]*rpc.RuleReport, 0),
			},
			{
				Severity:    rpc.Rule_INFO,
				RuleReports: make([]*rpc.RuleReport, 0),
			},
			{
				Severity:    rpc.Rule_HINT,
				RuleReports: make([]*rpc.RuleReport, 0),
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
		linterResponse *rpc.LinterResponse
		linterMetadata *linterMetadata
		wantReport     *rpc.ConformanceReport
	}{
		// Test basic flow.
		{
			desc: "Normal case",
			linterResponse: &rpc.LinterResponse{
				Lint: &rpc.Lint{
					Name: "sample-result",
					Files: []*rpc.LintFile{
						{
							FilePath: "test-result-file",
							Problems: []*rpc.LintProblem{
								{
									Message:    "no-$ref-siblings violated",
									RuleId:     "no-$ref-siblings",
									Suggestion: "fix no-$ref-siblings",
									Location: &rpc.LintLocation{
										StartPosition: &rpc.LintPosition{LineNumber: 11, ColumnNumber: 25},
										EndPosition:   &rpc.LintPosition{LineNumber: 11, ColumnNumber: 32},
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
						guidelineRule: &rpc.Rule{
							Id:       "norefsiblings",
							Severity: rpc.Rule_ERROR,
						},
						guideline: &rpc.Guideline{
							Id:    "refproperties",
							State: rpc.Guideline_ACTIVE,
						},
					},
				},
			},
			wantReport: &rpc.ConformanceReport{
				Id:         fmt.Sprintf("conformance-%s", styleguideId),
				Kind:       "ConformanceReport",
				Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{State: rpc.Guideline_STATE_UNSPECIFIED},
					{State: rpc.Guideline_PROPOSED},
					{
						State: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:     "norefsiblings",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix no-$ref-siblings",
												Location: &rpc.LintLocation{
													StartPosition: &rpc.LintPosition{LineNumber: 11, ColumnNumber: 25},
													EndPosition:   &rpc.LintPosition{LineNumber: 11, ColumnNumber: 32},
												},
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{Severity: rpc.Rule_INFO},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{State: rpc.Guideline_DEPRECATED},
					{State: rpc.Guideline_DISABLED},
				},
			},
		},
		// Test: LinterResponse includes multiple rule violations. Validate that only the once configured in the styleguide show up in the final conformance report.
		{
			desc: "Multiple violations",
			linterResponse: &rpc.LinterResponse{
				Lint: &rpc.Lint{
					Name: "sample-result",
					Files: []*rpc.LintFile{
						{
							FilePath: "test-result-file",
							Problems: []*rpc.LintProblem{
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
						guidelineRule: &rpc.Rule{
							Id:       "operationdescription",
							Severity: rpc.Rule_ERROR,
						},
						guideline: &rpc.Guideline{
							Id:    "descriptionproperties",
							State: rpc.Guideline_ACTIVE,
						},
					},
				},
			},
			wantReport: &rpc.ConformanceReport{
				Id:         fmt.Sprintf("conformance-%s", styleguideId),
				Kind:       "ConformanceReport",
				Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{State: rpc.Guideline_STATE_UNSPECIFIED},
					{State: rpc.Guideline_PROPOSED},
					{
						State: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "descriptionproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:     "operationdescription",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix operation-description",
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{Severity: rpc.Rule_INFO},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{State: rpc.Guideline_DEPRECATED},
					{State: rpc.Guideline_DISABLED},
				},
			},
		},
		// Test: Multiple rules are defined in multiple guidelines, check conformance report gets generated accurately.
		{
			desc: "Multiple guidelines",
			linterResponse: &rpc.LinterResponse{
				Lint: &rpc.Lint{
					Name: "sample-result",
					Files: []*rpc.LintFile{
						{
							FilePath: "test-result-file",
							Problems: []*rpc.LintProblem{
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
						guidelineRule: &rpc.Rule{
							Id:       "operationdescription",
							Severity: rpc.Rule_ERROR,
						},
						guideline: &rpc.Guideline{
							Id:    "descriptionproperties",
							State: rpc.Guideline_ACTIVE,
						},
					},
					"tag-description": {
						guidelineRule: &rpc.Rule{
							Id:       "tagdescription",
							Severity: rpc.Rule_WARNING,
						},
						guideline: &rpc.Guideline{
							Id:    "descriptionproperties",
							State: rpc.Guideline_ACTIVE,
						},
					},
					"info-description": {
						guidelineRule: &rpc.Rule{
							Id:       "infodescription",
							Severity: rpc.Rule_WARNING,
						},
						guideline: &rpc.Guideline{
							Id:    "descriptionproperties",
							State: rpc.Guideline_ACTIVE,
						},
					},
					"no-$ref-siblings": {
						guidelineRule: &rpc.Rule{
							Id:       "norefsiblings",
							Severity: rpc.Rule_ERROR,
						},
						guideline: &rpc.Guideline{
							Id:    "refproperties",
							State: rpc.Guideline_PROPOSED,
						},
					},
				},
			},
			wantReport: &rpc.ConformanceReport{
				Id:         fmt.Sprintf("conformance-%s", styleguideId),
				Kind:       "ConformanceReport",
				Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{State: rpc.Guideline_STATE_UNSPECIFIED},
					{
						State: rpc.Guideline_PROPOSED,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:     "norefsiblings",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix no-$ref-siblings",
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{Severity: rpc.Rule_INFO},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{
						State: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "descriptionproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:     "operationdescription",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix operation-description",
											},
										},
									},
									{
										Severity: rpc.Rule_WARNING,
										RuleReports: []*rpc.RuleReport{
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
									{Severity: rpc.Rule_INFO},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{State: rpc.Guideline_DEPRECATED},
					{State: rpc.Guideline_DISABLED},
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
				protocmp.IgnoreFields(&rpc.RuleReport{}),
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
	linterResponse := &rpc.LinterResponse{
		Lint: &rpc.Lint{
			Name: "sample-result",
			Files: []*rpc.LintFile{
				{
					FilePath: "test-result-file",
					Problems: []*rpc.LintProblem{
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
				guidelineRule: &rpc.Rule{
					Id:       "operationdescription",
					Severity: rpc.Rule_ERROR,
				},
				guideline: &rpc.Guideline{
					Id:    "descriptionproperties",
					State: rpc.Guideline_ACTIVE,
				},
			},
		},
	}

	preexistingReport := &rpc.ConformanceReport{
		Id:         fmt.Sprintf("conformance-%s", styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
		GuidelineReportGroups: []*rpc.GuidelineReportGroup{
			{State: rpc.Guideline_STATE_UNSPECIFIED},
			{State: rpc.Guideline_PROPOSED},
			{
				State: rpc.Guideline_ACTIVE,
				GuidelineReports: []*rpc.GuidelineReport{
					{
						GuidelineId: "descriptionproperties",
						RuleReportGroups: []*rpc.RuleReportGroup{
							{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
							{
								Severity: rpc.Rule_ERROR,
								RuleReports: []*rpc.RuleReport{
									{
										RuleId:     "tagdescription",
										Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
										File:       "test-result-file",
										Suggestion: "fix tag-description",
									},
								},
							},
							{Severity: rpc.Rule_WARNING},
							{Severity: rpc.Rule_INFO},
							{Severity: rpc.Rule_HINT},
						},
					},
				},
			},
			{State: rpc.Guideline_DEPRECATED},
			{State: rpc.Guideline_DISABLED},
		},
	}

	guidelineReportsMap := map[string]int{
		"descriptionproperties": 0,
	}

	wantReport := &rpc.ConformanceReport{
		Id:         fmt.Sprintf("conformance-%s", styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
		GuidelineReportGroups: []*rpc.GuidelineReportGroup{
			{State: rpc.Guideline_STATE_UNSPECIFIED},
			{State: rpc.Guideline_PROPOSED},
			{
				State: rpc.Guideline_ACTIVE,
				GuidelineReports: []*rpc.GuidelineReport{
					{
						GuidelineId: "descriptionproperties",
						RuleReportGroups: []*rpc.RuleReportGroup{
							{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
							{
								Severity: rpc.Rule_ERROR,
								RuleReports: []*rpc.RuleReport{
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
							{Severity: rpc.Rule_WARNING},
							{Severity: rpc.Rule_INFO},
							{Severity: rpc.Rule_HINT},
						},
					},
				},
			},
			{State: rpc.Guideline_DEPRECATED},
			{State: rpc.Guideline_DISABLED},
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
		protocmp.IgnoreFields(&rpc.RuleReport{}),
		protocmp.Transform(),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}
	if !cmp.Equal(wantReport, preexistingReport, opts) {
		t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(wantReport, preexistingReport, opts))
	}
}
