package conformance

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	// "github.com/apigee/registry/connection"
)

const specName = "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml"
const styleguideId = "openapi-test"

// This test will catch any changes made to the original status values.
func TestInitializeConformanceReport(t *testing.T) {
	specName := specName
	styleguideId := styleguideId
	want := &rpc.ConformanceReport{
		Name:           fmt.Sprintf("%s/artifacts/conformance-%s", specName, styleguideId),
		StyleguideName: styleguideId,
		GuidelineReportGroups: []*rpc.GuidelineReportGroup{
			{
				Status:           rpc.Guideline_STATUS_UNSPECIFIED,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
			},
			{
				Status:           rpc.Guideline_PROPOSED,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
			},
			{
				Status:           rpc.Guideline_ACTIVE,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
			},
			{
				Status:           rpc.Guideline_DEPRECATED,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
			},
			{
				Status:           rpc.Guideline_DISABLED,
				GuidelineReports: make([]*rpc.GuidelineReport, 0),
			},
		},
	}

	got := initializeConformanceReport(specName, styleguideId)
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
	guidelineId := styleguideId
	want := &rpc.GuidelineReport{
		GuidelineId: guidelineId,
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

	got := initializeGuidelineReport(guidelineId)
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
							Id:     "refproperties",
							Status: rpc.Guideline_ACTIVE,
						},
					},
				},
			},
			wantReport: func() *rpc.ConformanceReport {
				conformance := InitReport(t)
				conformance.Name = fmt.Sprintf("%s/artifacts/conformance-%s", specName, styleguideId)
				conformance.StyleguideName = styleguideId

				ruleReportGroups := InitRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					{
						RuleId:     "norefsiblings",
						SpecName:   specName,
						FileName:   "test-result-file",
						Suggestion: "fix no-$ref-siblings",
						Location: &rpc.LintLocation{
							StartPosition: &rpc.LintPosition{LineNumber: 11, ColumnNumber: 25},
							EndPosition:   &rpc.LintPosition{LineNumber: 11, ColumnNumber: 32},
						},
					},
				}

				//Populate the expected guideline reports
				guidelineReports := []*rpc.GuidelineReport{
					{
						GuidelineId:      "refproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_ACTIVE].GuidelineReports = guidelineReports

				return conformance
			}(),
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
							Id:     "descriptionproperties",
							Status: rpc.Guideline_ACTIVE,
						},
					},
				},
			},
			wantReport: func() *rpc.ConformanceReport {
				conformance := InitReport(t)
				conformance.Name = fmt.Sprintf("%s/artifacts/conformance-%s", specName, styleguideId)
				conformance.StyleguideName = styleguideId

				ruleReportGroups := InitRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					{
						RuleId:     "operationdescription",
						SpecName:   specName,
						FileName:   "test-result-file",
						Suggestion: "fix operation-description",
					},
				}

				//Populate the expected guideline reports
				guidelineReports := []*rpc.GuidelineReport{
					{
						GuidelineId:      "descriptionproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_ACTIVE].GuidelineReports = guidelineReports

				return conformance
			}(),
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
							Id:     "descriptionproperties",
							Status: rpc.Guideline_ACTIVE,
						},
					},
					"tag-description": {
						guidelineRule: &rpc.Rule{
							Id:       "tagdescription",
							Severity: rpc.Rule_WARNING,
						},
						guideline: &rpc.Guideline{
							Id:     "descriptionproperties",
							Status: rpc.Guideline_ACTIVE,
						},
					},
					"info-description": {
						guidelineRule: &rpc.Rule{
							Id:       "infodescription",
							Severity: rpc.Rule_WARNING,
						},
						guideline: &rpc.Guideline{
							Id:     "descriptionproperties",
							Status: rpc.Guideline_ACTIVE,
						},
					},
					"no-$ref-siblings": {
						guidelineRule: &rpc.Rule{
							Id:       "norefsiblings",
							Severity: rpc.Rule_ERROR,
						},
						guideline: &rpc.Guideline{
							Id:     "refproperties",
							Status: rpc.Guideline_PROPOSED,
						},
					},
				},
			},
			wantReport: func() *rpc.ConformanceReport {
				conformance := InitReport(t)
				conformance.Name = fmt.Sprintf("%s/artifacts/conformance-%s", specName, styleguideId)
				conformance.StyleguideName = styleguideId

				ruleReportGroups := InitRuleReportGroups(t)

				// Populate the expected severity entries
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					{
						RuleId:     "operationdescription",
						SpecName:   specName,
						FileName:   "test-result-file",
						Suggestion: "fix operation-description",
					},
				}

				ruleReportGroups[rpc.Rule_WARNING].RuleReports = []*rpc.RuleReport{
					{
						RuleId:     "tagdescription",
						SpecName:   specName,
						FileName:   "test-result-file",
						Suggestion: "fix tag-description",
					},
					{
						RuleId:     "infodescription",
						SpecName:   specName,
						FileName:   "test-result-file",
						Suggestion: "fix info-description",
					},
				}

				//Populate the expected guideline reports
				guidelineReports := []*rpc.GuidelineReport{
					{
						GuidelineId:      "descriptionproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}
				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_ACTIVE].GuidelineReports = guidelineReports

				ruleReportGroups = InitRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					{
						RuleId:     "norefsiblings",
						SpecName:   specName,
						FileName:   "test-result-file",
						Suggestion: "fix no-$ref-siblings",
					},
				}

				//Populate the expected guideline reports
				guidelineReports = []*rpc.GuidelineReport{
					{
						GuidelineId:      "refproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_PROPOSED].GuidelineReports = guidelineReports

				return conformance
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			task := &ComputeConformanceTask{
				Spec:         &rpc.ApiSpec{Name: specName},
				StyleguideId: styleguideId,
			}

			gotReport := initializeConformanceReport(task.Spec.GetName(), task.StyleguideId)
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

// Test the scenario where there are pre-existing entries in the conformance report from other linters.
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
					Id:     "descriptionproperties",
					Status: rpc.Guideline_ACTIVE,
				},
			},
		},
	}

	preexistingReport := func() *rpc.ConformanceReport {
		conformance := InitReport(t)
		conformance.Name = fmt.Sprintf("%s/artifacts/conformance-%s", specName, styleguideId)
		conformance.StyleguideName = styleguideId

		// Populate the preexisting data
		ruleReportGroups := InitRuleReportGroups(t)

		// Populate the expected severity entry
		ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
			{
				RuleId:     "tagdescription",
				SpecName:   specName,
				FileName:   "test-result-file",
				Suggestion: "fix tag-description",
			},
		}

		//Populate the expected guideline reports
		guidelineReports := []*rpc.GuidelineReport{
			{
				GuidelineId:      "descriptionproperties",
				RuleReportGroups: ruleReportGroups,
			},
		}

		//Populate the expected status entry
		conformance.GuidelineReportGroups[rpc.Guideline_ACTIVE].GuidelineReports = guidelineReports

		return conformance
	}()

	guidelineReportsMap := map[string]int{
		"descriptionproperties": 0,
	}

	wantReport := func() *rpc.ConformanceReport {
		conformance := InitReport(t)
		conformance.Name = fmt.Sprintf("%s/artifacts/conformance-%s", specName, styleguideId)
		conformance.StyleguideName = styleguideId

		ruleReportGroups := InitRuleReportGroups(t)

		// Populate the expected severity entry
		ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
			{
				RuleId:     "tagdescription",
				SpecName:   specName,
				FileName:   "test-result-file",
				Suggestion: "fix tag-description",
			},
			{
				RuleId:     "operationdescription",
				SpecName:   specName,
				FileName:   "test-result-file",
				Suggestion: "fix operation-description",
			},
		}

		//Populate the expected guideline reports
		guidelineReports := []*rpc.GuidelineReport{
			{
				GuidelineId:      "descriptionproperties",
				RuleReportGroups: ruleReportGroups,
			},
		}

		//Populate the expected status entry
		conformance.GuidelineReportGroups[rpc.Guideline_ACTIVE].GuidelineReports = guidelineReports

		return conformance
	}()

	ctx := context.Background()

	task := &ComputeConformanceTask{
		Spec:         &rpc.ApiSpec{Name: specName},
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
