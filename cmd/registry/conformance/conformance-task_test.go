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

	"github.com/apigee/registry/pkg/application/style"
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
	want := &style.ConformanceReport{
		Id:         fmt.Sprintf("conformance-%s", styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
		GuidelineReportGroups: []*style.GuidelineReportGroup{
			{
				State:            style.Guideline_STATE_UNSPECIFIED,
				GuidelineReports: make([]*style.GuidelineReport, 0),
			},
			{
				State:            style.Guideline_PROPOSED,
				GuidelineReports: make([]*style.GuidelineReport, 0),
			},
			{
				State:            style.Guideline_ACTIVE,
				GuidelineReports: make([]*style.GuidelineReport, 0),
			},
			{
				State:            style.Guideline_DEPRECATED,
				GuidelineReports: make([]*style.GuidelineReport, 0),
			},
			{
				State:            style.Guideline_DISABLED,
				GuidelineReports: make([]*style.GuidelineReport, 0),
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
	want := &style.GuidelineReport{
		GuidelineId: styleguideId,
		RuleReportGroups: []*style.RuleReportGroup{
			{
				Severity:    style.Rule_SEVERITY_UNSPECIFIED,
				RuleReports: make([]*style.RuleReport, 0),
			},
			{
				Severity:    style.Rule_ERROR,
				RuleReports: make([]*style.RuleReport, 0),
			},
			{
				Severity:    style.Rule_WARNING,
				RuleReports: make([]*style.RuleReport, 0),
			},
			{
				Severity:    style.Rule_INFO,
				RuleReports: make([]*style.RuleReport, 0),
			},
			{
				Severity:    style.Rule_HINT,
				RuleReports: make([]*style.RuleReport, 0),
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
		linterResponse *style.LinterResponse
		linterMetadata *linterMetadata
		wantReport     *style.ConformanceReport
	}{
		// Test basic flow.
		{
			desc: "Normal case",
			linterResponse: &style.LinterResponse{
				Lint: &style.Lint{
					Name: "sample-result",
					Files: []*style.LintFile{
						{
							FilePath: "test-result-file",
							Problems: []*style.LintProblem{
								{
									Message:    "no-$ref-siblings violated",
									RuleId:     "no-$ref-siblings",
									Suggestion: "fix no-$ref-siblings",
									Location: &style.LintLocation{
										StartPosition: &style.LintPosition{LineNumber: 11, ColumnNumber: 25},
										EndPosition:   &style.LintPosition{LineNumber: 11, ColumnNumber: 32},
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
						guidelineRule: &style.Rule{
							Id:       "norefsiblings",
							Severity: style.Rule_ERROR,
						},
						guideline: &style.Guideline{
							Id:    "refproperties",
							State: style.Guideline_ACTIVE,
						},
					},
				},
			},
			wantReport: &style.ConformanceReport{
				Id:         fmt.Sprintf("conformance-%s", styleguideId),
				Kind:       "ConformanceReport",
				Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
				GuidelineReportGroups: []*style.GuidelineReportGroup{
					{State: style.Guideline_STATE_UNSPECIFIED},
					{State: style.Guideline_PROPOSED},
					{
						State: style.Guideline_ACTIVE,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:     "norefsiblings",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix no-$ref-siblings",
												Location: &style.LintLocation{
													StartPosition: &style.LintPosition{LineNumber: 11, ColumnNumber: 25},
													EndPosition:   &style.LintPosition{LineNumber: 11, ColumnNumber: 32},
												},
											},
										},
									},
									{Severity: style.Rule_WARNING},
									{Severity: style.Rule_INFO},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{State: style.Guideline_DEPRECATED},
					{State: style.Guideline_DISABLED},
				},
			},
		},
		// Test: LinterResponse includes multiple rule violations. Validate that only the once configured in the styleguide show up in the final conformance report.
		{
			desc: "Multiple violations",
			linterResponse: &style.LinterResponse{
				Lint: &style.Lint{
					Name: "sample-result",
					Files: []*style.LintFile{
						{
							FilePath: "test-result-file",
							Problems: []*style.LintProblem{
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
						guidelineRule: &style.Rule{
							Id:       "operationdescription",
							Severity: style.Rule_ERROR,
						},
						guideline: &style.Guideline{
							Id:    "descriptionproperties",
							State: style.Guideline_ACTIVE,
						},
					},
				},
			},
			wantReport: &style.ConformanceReport{
				Id:         fmt.Sprintf("conformance-%s", styleguideId),
				Kind:       "ConformanceReport",
				Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
				GuidelineReportGroups: []*style.GuidelineReportGroup{
					{State: style.Guideline_STATE_UNSPECIFIED},
					{State: style.Guideline_PROPOSED},
					{
						State: style.Guideline_ACTIVE,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "descriptionproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:     "operationdescription",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix operation-description",
											},
										},
									},
									{Severity: style.Rule_WARNING},
									{Severity: style.Rule_INFO},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{State: style.Guideline_DEPRECATED},
					{State: style.Guideline_DISABLED},
				},
			},
		},
		// Test: Multiple rules are defined in multiple guidelines, check conformance report gets generated accurately.
		{
			desc: "Multiple guidelines",
			linterResponse: &style.LinterResponse{
				Lint: &style.Lint{
					Name: "sample-result",
					Files: []*style.LintFile{
						{
							FilePath: "test-result-file",
							Problems: []*style.LintProblem{
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
						guidelineRule: &style.Rule{
							Id:       "operationdescription",
							Severity: style.Rule_ERROR,
						},
						guideline: &style.Guideline{
							Id:    "descriptionproperties",
							State: style.Guideline_ACTIVE,
						},
					},
					"tag-description": {
						guidelineRule: &style.Rule{
							Id:       "tagdescription",
							Severity: style.Rule_WARNING,
						},
						guideline: &style.Guideline{
							Id:    "descriptionproperties",
							State: style.Guideline_ACTIVE,
						},
					},
					"info-description": {
						guidelineRule: &style.Rule{
							Id:       "infodescription",
							Severity: style.Rule_WARNING,
						},
						guideline: &style.Guideline{
							Id:    "descriptionproperties",
							State: style.Guideline_ACTIVE,
						},
					},
					"no-$ref-siblings": {
						guidelineRule: &style.Rule{
							Id:       "norefsiblings",
							Severity: style.Rule_ERROR,
						},
						guideline: &style.Guideline{
							Id:    "refproperties",
							State: style.Guideline_PROPOSED,
						},
					},
				},
			},
			wantReport: &style.ConformanceReport{
				Id:         fmt.Sprintf("conformance-%s", styleguideId),
				Kind:       "ConformanceReport",
				Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
				GuidelineReportGroups: []*style.GuidelineReportGroup{
					{State: style.Guideline_STATE_UNSPECIFIED},
					{
						State: style.Guideline_PROPOSED,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:     "norefsiblings",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix no-$ref-siblings",
											},
										},
									},
									{Severity: style.Rule_WARNING},
									{Severity: style.Rule_INFO},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{
						State: style.Guideline_ACTIVE,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "descriptionproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:     "operationdescription",
												Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
												File:       "test-result-file",
												Suggestion: "fix operation-description",
											},
										},
									},
									{
										Severity: style.Rule_WARNING,
										RuleReports: []*style.RuleReport{
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
									{Severity: style.Rule_INFO},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{State: style.Guideline_DEPRECATED},
					{State: style.Guideline_DISABLED},
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
				protocmp.IgnoreFields(&style.RuleReport{}),
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
	linterResponse := &style.LinterResponse{
		Lint: &style.Lint{
			Name: "sample-result",
			Files: []*style.LintFile{
				{
					FilePath: "test-result-file",
					Problems: []*style.LintProblem{
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
				guidelineRule: &style.Rule{
					Id:       "operationdescription",
					Severity: style.Rule_ERROR,
				},
				guideline: &style.Guideline{
					Id:    "descriptionproperties",
					State: style.Guideline_ACTIVE,
				},
			},
		},
	}

	preexistingReport := &style.ConformanceReport{
		Id:         fmt.Sprintf("conformance-%s", styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
		GuidelineReportGroups: []*style.GuidelineReportGroup{
			{State: style.Guideline_STATE_UNSPECIFIED},
			{State: style.Guideline_PROPOSED},
			{
				State: style.Guideline_ACTIVE,
				GuidelineReports: []*style.GuidelineReport{
					{
						GuidelineId: "descriptionproperties",
						RuleReportGroups: []*style.RuleReportGroup{
							{Severity: style.Rule_SEVERITY_UNSPECIFIED},
							{
								Severity: style.Rule_ERROR,
								RuleReports: []*style.RuleReport{
									{
										RuleId:     "tagdescription",
										Spec:       fmt.Sprintf("%s@%s", specName, revisionId),
										File:       "test-result-file",
										Suggestion: "fix tag-description",
									},
								},
							},
							{Severity: style.Rule_WARNING},
							{Severity: style.Rule_INFO},
							{Severity: style.Rule_HINT},
						},
					},
				},
			},
			{State: style.Guideline_DEPRECATED},
			{State: style.Guideline_DISABLED},
		},
	}

	guidelineReportsMap := map[string]int{
		"descriptionproperties": 0,
	}

	wantReport := &style.ConformanceReport{
		Id:         fmt.Sprintf("conformance-%s", styleguideId),
		Kind:       "ConformanceReport",
		Styleguide: fmt.Sprintf("projects/%s/locations/global/artifacts/%s", project, styleguideId),
		GuidelineReportGroups: []*style.GuidelineReportGroup{
			{State: style.Guideline_STATE_UNSPECIFIED},
			{State: style.Guideline_PROPOSED},
			{
				State: style.Guideline_ACTIVE,
				GuidelineReports: []*style.GuidelineReport{
					{
						GuidelineId: "descriptionproperties",
						RuleReportGroups: []*style.RuleReportGroup{
							{Severity: style.Rule_SEVERITY_UNSPECIFIED},
							{
								Severity: style.Rule_ERROR,
								RuleReports: []*style.RuleReport{
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
							{Severity: style.Rule_WARNING},
							{Severity: style.Rule_INFO},
							{Severity: style.Rule_HINT},
						},
					},
				},
			},
			{State: style.Guideline_DEPRECATED},
			{State: style.Guideline_DISABLED},
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
		protocmp.IgnoreFields(&style.RuleReport{}),
		protocmp.Transform(),
		cmpopts.SortSlices(func(a, b string) bool { return a < b }),
	}
	if !cmp.Equal(wantReport, preexistingReport, opts) {
		t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(wantReport, preexistingReport, opts))
	}
}
