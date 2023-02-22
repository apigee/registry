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

package scoring

import (
	"testing"

	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/pkg/application/scoring"
)

func TestValidateScoreDefinition(t *testing.T) {
	tests := []struct {
		desc            string
		parent          string
		scoreDefinition *scoring.ScoreDefinition
		wantNumErr      int
	}{
		// No errors
		{
			desc:   "score formula",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 61,
									Max: 100,
								},
							},
							{
								Severity: scoring.Severity_WARNING,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 60,
								},
							},
						},
					},
				},
			},
		},
		{
			desc:   "rollup formula",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_RollupFormula{
					RollupFormula: &scoring.RollUpFormula{
						ScoreFormulas: []*scoring.ScoreFormula{
							{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.spec/artifacts/conformance-report",
								},
								ScoreExpression: "count(errors)",
								ReferenceId:     "lint_errors",
							},
							{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.spec/artifacts/conformance-report",
								},
								ScoreExpression: "count(warnings)",
								ReferenceId:     "lint_warnings",
							},
						},
						RollupExpression: "lint_errors+lint_warnings",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 61,
									Max: 100,
								},
							},
							{
								Severity: scoring.Severity_WARNING,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 60,
								},
							},
						},
					},
				},
			},
		},
		{
			desc:   "no thresholds",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
					},
				},
			},
		},
		{
			desc:   "only max limit",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MaxValue: 100,
					},
				},
			},
		},
		// Single errors
		{
			desc:   "target pattern error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "score formula error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.artifact/conformance-report", //error
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "rollup formula error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_RollupFormula{
					RollupFormula: &scoring.RollUpFormula{
						ScoreFormulas: []*scoring.ScoreFormula{
							{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.spec/artifacts/conformance-report",
								},
								ScoreExpression: "count(errors)",
								ReferenceId:     "lint_errors",
							},
							{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.artifact/conformance-report", //error
								},
								ScoreExpression: "count(warnings)",
								ReferenceId:     "lint_warnings",
							},
						},
						RollupExpression: "lint_errors+lint_warnings",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "percent threshold error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Percent{
					Percent: &scoring.PercentType{
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{ //error
									Min: 60,
									Max: 100,
								},
							},
							{
								Severity: scoring.Severity_WARNING,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 60,
								},
							},
						},
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "integer threshold error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{ //error
									Min: 62,
									Max: 100,
								},
							},
							{
								Severity: scoring.Severity_WARNING,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 60,
								},
							},
						},
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "integer threshold no limits",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 61,
									Max: 100,
								},
							},
							{
								Severity: scoring.Severity_WARNING,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 60,
								},
							},
						},
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "boolean threshold error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/audit-report",
						},
						ScoreExpression: "isApproved",
					},
				},
				Type: &scoring.ScoreDefinition_Boolean{
					Boolean: &scoring.BooleanType{
						Thresholds: []*scoring.BooleanThreshold{ //error
							{
								Severity: scoring.Severity_ALERT,
								Value:    false,
							},
						},
					},
				},
			},
			wantNumErr: 1,
		},
		// Combination errors
		{
			desc:   "target pattern and score formula error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.specs/artifacts/conformance-report", //error
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
					},
				},
			},
			// This is expected here since it is not possible to validate ScoreFormula patterns if there are errors in the targetResource pattern.
			wantNumErr: 1,
		},
		{
			desc:   "target pattern and threshold error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{ //error
									Min: 60,
									Max: 100,
								},
							},
						},
					},
				},
			},
			wantNumErr: 2,
		},
		{
			desc:   "score formula and threshold error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.specs/artifacts/conformance-report", //error
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{ //error
									Min: 60,
									Max: 100,
								},
							},
						},
					},
				},
			},
			wantNumErr: 2,
		},
		{
			desc:   "error in each component",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.specs/artifacts/conformance-report", //error
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{ //error
									Min: 60,
									Max: 100,
								},
							},
						},
					},
				},
			},
			// This is expected here since it is not possible to validate ScoreFormula patterns if there are errors in the targetResource pattern.
			wantNumErr: 2,
		},
		// Missing components
		{
			desc:   "missing target resource",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_OK,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 59,
								},
							},
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 60,
									Max: 100,
								},
							},
						},
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "missing rollup_formula.rollup_expression",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_RollupFormula{
					RollupFormula: &scoring.RollUpFormula{
						ScoreFormulas: []*scoring.ScoreFormula{
							{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.spec/artifacts/conformance-report",
								},
								ScoreExpression: "count(errors)",
								ReferenceId:     "lint_errors",
							},
							{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.spec/artifacts/conformance-report",
								},
								ScoreExpression: "count(warnings)",
								ReferenceId:     "lint_warnings",
							},
						},
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 61,
									Max: 100,
								},
							},
							{
								Severity: scoring.Severity_WARNING,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 60,
								},
							},
						},
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "missing rollup_formula.score_formulas",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_RollupFormula{
					RollupFormula: &scoring.RollUpFormula{
						RollupExpression: "lint_errors+lint_warnings",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 61,
									Max: 100,
								},
							},
							{
								Severity: scoring.Severity_WARNING,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 60,
								},
							},
						},
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "missing formula",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_OK,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 59,
								},
							},
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 60,
									Max: 100,
								},
							},
						},
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "missing type",
			parent: "projects/demo/locations/global",
			scoreDefinition: &scoring.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
			},
			wantNumErr: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := ValidateScoreDefinition(test.parent, test.scoreDefinition)
			if len(gotErrs) != test.wantNumErr {
				t.Errorf("ValidateScoreDefinition(%s, %v) returned unexpected no. of errors: want %d, got %s", test.parent, test.scoreDefinition, test.wantNumErr, gotErrs)
			}
		})
	}
}

func TestValidateReferencesInPattern(t *testing.T) {
	tests := []struct {
		desc          string
		targetPattern string
		pattern       string
		wantNumErr    int
	}{
		// No errors
		{
			desc:          "score formula",
			targetPattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			pattern:       "$resource.spec/artifacts/conformance-report",
		},
		// errors
		{
			desc:          "invalid $resource reference",
			targetPattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			pattern:       "$resource.specs/artifacts/conformance-report", //error
			wantNumErr:    1,
		},
		{
			desc:          "no $resource reference",
			targetPattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			pattern:       "apis/-/versions/-/specs/-/artifacts/conformance-report", //error
			wantNumErr:    1,
		},
		{
			desc:          "invalid $resource wrt targetName",
			targetPattern: "projects/demo/locations/global/apis/-/versions/-",
			pattern:       "$resource.spec/artifacts/conformance-report", //error
			wantNumErr:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			targetName, _ := patterns.ParseResourcePattern(test.targetPattern)
			gotErrs := validateReferencesInPattern(targetName, test.pattern)
			if len(gotErrs) != test.wantNumErr {
				t.Errorf("validateReferencesInPattern(%s, %s) returned unexpected no. of errors: want %d, got %s", targetName, test.pattern, test.wantNumErr, gotErrs)
			}
		})
	}
}

func TestValidateScoreFormula(t *testing.T) {
	tests := []struct {
		desc          string
		targetPattern *scoring.ResourcePattern
		scoreFormula  *scoring.ScoreFormula
		wantNumErr    int
	}{
		// No errors
		{
			desc: "score formula",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/conformance-report",
				},
				ScoreExpression: "count(errors)",
			},
		},
		// Single errors
		{
			desc: "$resource error",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.specs/artifacts/conformance-report", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 1,
		},
		{
			desc: "missing artifact name",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/-", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 1,
		},
		{
			desc: "invalid reference_id",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/conformance-report",
				},
				ScoreExpression: "count(errors)",
				ReferenceId:     "num-errors",
			},
			wantNumErr: 1,
		},
		// Combination errors
		{
			desc: "invalid pattern and missing name",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.specs/-/artifacts/-", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 2,
		},
		{
			desc: "no reference and missing name",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-/artifacts/-", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 2,
		},
		{
			desc: "invalid reference and missing name",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-", //error
			},
			scoreFormula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/-",
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 2,
		},
		// missing components
		{
			desc: "missing artifact",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &scoring.ScoreFormula{
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 1,
		},
		{
			desc: "missing score expression",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/conformance-report",
				},
			},
			wantNumErr: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			targetName, _ := patterns.ParseResourcePattern(test.targetPattern.GetPattern())
			gotErrs := validateScoreFormula(targetName, test.scoreFormula)
			if len(gotErrs) != test.wantNumErr {
				t.Errorf("validateScoreFormula(%s, %v) returned unexpected no. of errors: want %d, got %s", targetName, test.scoreFormula, test.wantNumErr, gotErrs)
			}
		})
	}
}

func TestValidateNumberThresholds(t *testing.T) {
	tests := []struct {
		desc       string
		minValue   int32
		maxValue   int32
		thresholds []*scoring.NumberThreshold
		wantNumErr int
	}{
		// no errors
		{
			desc:     "percentage thresholds",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 61,
						Max: 100,
					},
				},
			},
		},
		{
			desc:     "integer thresholds",
			minValue: -50,
			maxValue: 50,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: -50,
						Max: -20,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: -19,
						Max: 10,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 11,
						Max: 50,
					},
				},
			},
		},
		{
			desc:     "descending score thresholds",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 61,
						Max: 100,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
			},
		},
		{
			desc:     "distributed severity",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 0,
						Max: 10,
					},
				},
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 91,
						Max: 100,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 11,
						Max: 30,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 71,
						Max: 90,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 31,
						Max: 70,
					},
				},
			},
		},
		// single errors
		{
			desc:     "missing range",
			minValue: -50,
			maxValue: 50,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
				},
			},
			wantNumErr: 2,
		},
		{
			desc:     "range.min greater than range.max",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Max: 0,
						Min: 50,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 51,
						Max: 100,
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:     "out of minValue bound",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: -1,
						Max: 30,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 61,
						Max: 100,
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:     "out of maxValue bound",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 61,
						Max: 101,
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:     "missing coverage for minValue",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 5,
						Max: 30,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 61,
						Max: 100,
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:     "missing coverage for maxValue",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 61,
						Max: 90,
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:     "missing coverage in between",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 0,
						Max: 29,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 31,
						Max: 59,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 61,
						Max: 100,
					},
				},
			},
			wantNumErr: 2,
		},
		{
			desc:     "overlap",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 30,
						Max: 60,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 60,
						Max: 100,
					},
				},
			},
			wantNumErr: 2,
		},
		// Combination errors
		{
			desc:     "out of min and max limits",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: -1,
						Max: 50,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 51,
						Max: 101,
					},
				},
			},
			wantNumErr: 2,
		},
		{
			desc:     "out of limits and overlap",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: -1,
						Max: 50,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 50,
						Max: 101,
					},
				},
			},
			wantNumErr: 2,
		},
		{
			desc:     "out of limits and missing coverage",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: -1,
						Max: 50,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 52,
						Max: 101,
					},
				},
			},
			wantNumErr: 2,
		},
		{
			desc:     "missing limits coverage",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 2,
						Max: 50,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 51,
						Max: 99,
					},
				},
			},
			wantNumErr: 2,
		},
		{
			desc:     "missing limits coverage and overlap",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 2,
						Max: 50,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 50,
						Max: 99,
					},
				},
			},
			wantNumErr: 3,
		},
		{
			desc:     "missing limits coverage and missing coverage",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 2,
						Max: 50,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 52,
						Max: 99,
					},
				},
			},
			wantNumErr: 3,
		},
		{
			desc:     "missing min coverage and out of max limit",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 2,
						Max: 50,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 51,
						Max: 101,
					},
				},
			},
			wantNumErr: 1,
		},
		{
			desc:     "overlap and missing coverage",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: scoring.Severity_WARNING,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 30,
						Max: 60,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 62,
						Max: 100,
					},
				},
			},
			wantNumErr: 2,
		},
		{
			desc:     "nested thresholds",
			minValue: 0,
			maxValue: 100,
			thresholds: []*scoring.NumberThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 0,
						Max: 100,
					},
				},
				{
					Severity: scoring.Severity_OK,
					Range: &scoring.NumberThreshold_NumberRange{
						Min: 20,
						Max: 50,
					},
				},
			},
			wantNumErr: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := validateNumberThresholds(test.thresholds, test.minValue, test.maxValue)
			if len(gotErrs) != test.wantNumErr {
				t.Errorf("validateNumberThresholds(%v, %d, %d) returned unexpected no. of errors: want %d, got %s", test.thresholds, test.minValue, test.maxValue, test.wantNumErr, gotErrs)
			}
		})
	}
}

func TestValidateBooleanThresholds(t *testing.T) {
	tests := []struct {
		desc       string
		thresholds []*scoring.BooleanThreshold
		wantNumErr int
	}{
		// no errors
		{
			desc: "normal case",
			thresholds: []*scoring.BooleanThreshold{
				{
					Severity: scoring.Severity_ALERT,
					Value:    false,
				},
				{
					Severity: scoring.Severity_OK,
					Value:    true,
				},
			},
		},
		{
			desc: "same severity",
			thresholds: []*scoring.BooleanThreshold{
				{
					Severity: scoring.Severity_OK,
					Value:    false,
				},
				{
					Severity: scoring.Severity_OK,
					Value:    true,
				},
			},
		},
		// single errors
		{
			desc: "missing false",
			thresholds: []*scoring.BooleanThreshold{
				{
					Severity: scoring.Severity_OK,
					Value:    true,
				},
			},
			wantNumErr: 1,
		},
		{
			desc: "missing true",
			thresholds: []*scoring.BooleanThreshold{
				{
					Severity: scoring.Severity_WARNING,
					Value:    false,
				},
			},
			wantNumErr: 1,
		},
		{
			desc: "duplicate entries",
			thresholds: []*scoring.BooleanThreshold{
				{
					Severity: scoring.Severity_OK,
					Value:    true,
				},
				{
					Severity: scoring.Severity_WARNING,
					Value:    true,
				},
				{
					Severity: scoring.Severity_ALERT,
					Value:    false,
				},
			},
			wantNumErr: 1,
		},
		// combination errors
		{
			desc: "missing and duplicate entries",
			thresholds: []*scoring.BooleanThreshold{
				{
					Severity: scoring.Severity_OK,
					Value:    true,
				},
				{
					Severity: scoring.Severity_WARNING,
					Value:    true,
				},
			},
			wantNumErr: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := validateBooleanThresholds(test.thresholds)
			if len(gotErrs) != test.wantNumErr {
				t.Errorf("validateBooleanThresholds(%v) returned unexpected no. of errors: want %d, got %s", test.thresholds, test.wantNumErr, gotErrs)
			}
		})
	}
}

func TestValidateScoreCardDefinition(t *testing.T) {
	tests := []struct {
		desc                string
		parent              string
		scoreCardDefinition *scoring.ScoreCardDefinition
		wantNumErr          int
	}{
		// No errors
		{
			desc:   "simple scorecard definition",
			parent: "projects/demo/locations/global",
			scoreCardDefinition: &scoring.ScoreCardDefinition{
				Id:   "test-scorecard-definition",
				Kind: "ScoreCardDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
					Filter:  "name.contains('openapi.yaml')",
				},
				ScorePatterns: []string{
					"$resource.spec/artifacts/score-lint-error",
					"$resource.spec/artifacts/score-lang-reuse",
					"$resource.spec/artifacts/score-security-audit",
					"$resource.spec/artifacts/score-accuracy",
				},
			},
		},
		// errors
		{
			desc:   "invalid target_resource pattern",
			parent: "projects/demo/locations/global",
			scoreCardDefinition: &scoring.ScoreCardDefinition{
				Id:   "test-scorecard-definition",
				Kind: "ScoreCardDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				ScorePatterns: []string{
					"$resource.spec/artifacts/score-lint-error",
					"$resource.spec/artifacts/score-lang-reuse",
					"$resource.spec/artifacts/score-security-audit",
					"$resource.spec/artifacts/score-accuracy",
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "missing score_patterns",
			parent: "projects/demo/locations/global",
			scoreCardDefinition: &scoring.ScoreCardDefinition{
				Id:   "test-scorecard-definition",
				Kind: "ScoreCardDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				// error
			},
			wantNumErr: 1,
		},
		{
			desc:   "invalid target_resource and missing score_patterns",
			parent: "projects/demo/locations/global",
			scoreCardDefinition: &scoring.ScoreCardDefinition{
				Id:   "test-scorecard-definition",
				Kind: "ScoreCardDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				//error
			},
			wantNumErr: 2,
		},
		{
			desc:   "invalid score_pattern $resource",
			parent: "projects/demo/locations/global",
			scoreCardDefinition: &scoring.ScoreCardDefinition{
				Id:   "test-scorecard-definition",
				Kind: "ScoreCardDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				ScorePatterns: []string{
					"$resource.specs/artifacts/score-lint-error", //error
					"$resource.spec/artifacts/score-lang-reuse",
					"$resource.spec/artifacts/score-security-audit",
					"$resource.spec/artifacts/score-accuracy",
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "invalid score_pattern no artifactID",
			parent: "projects/demo/locations/global",
			scoreCardDefinition: &scoring.ScoreCardDefinition{
				Id:   "test-scorecard-definition",
				Kind: "ScoreCardDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				ScorePatterns: []string{
					"$resource.spec/artifacts/-", //error
				},
			},
			wantNumErr: 1,
		},
		{
			desc:   "multiple invalid score_pattern",
			parent: "projects/demo/locations/global",
			scoreCardDefinition: &scoring.ScoreCardDefinition{
				Id:   "test-scorecard-definition",
				Kind: "ScoreCardDefinition",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				ScorePatterns: []string{
					"$resource.specs/artifacts/score-lint-error",     //error
					"$resource.specs/artifacts/score-lang-reuse",     //error
					"$resource.specs/artifacts/score-security-audit", //error
					"$resource.specs/artifacts/score-accuracy",       //error
				},
			},
			wantNumErr: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := ValidateScoreCardDefinition(test.parent, test.scoreCardDefinition)
			if len(gotErrs) != test.wantNumErr {
				t.Errorf("ValidateScoreCardDefinition(%s, %v) returned unexpected no. of errors: want %d, got %s", test.parent, test.scoreCardDefinition, test.wantNumErr, gotErrs)
			}
		})
	}
}

func TestGenerateCombinedPattern(t *testing.T) {
	tests := []struct {
		desc          string
		targetPattern *scoring.ResourcePattern
		inputPattern  string
		wantPattern   string
		wantErr       bool
	}{
		{
			desc: "spec pattern spec input",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			wantPattern:  "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
		{
			desc: "spec pattern collection input",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
			wantPattern:  "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
		},
		{
			desc: "specific api match spec input",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/petstore/versions/-/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			wantPattern:  "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
		{
			desc: "specific api match collection input",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/petstore/versions/-/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
			wantPattern:  "projects/pattern-test/locations/global/apis/petstore/versions/-/specs/-",
		},
		{
			desc: "specific api no match spec input",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/test/versions/-/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			wantErr:      true,
		},
		{
			desc: "specific version match spec input",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/1.0.0/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			wantPattern:  "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
		{
			desc: "specific version match collection input",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/1.0.0/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
			wantPattern:  "projects/pattern-test/locations/global/apis/-/versions/1.0.0/specs/-",
		},
		{
			desc: "specific version no match",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/2.0.0/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			wantErr:      true,
		},
		{
			desc: "specific spec match spec input",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/openapi",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			wantPattern:  "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
		{
			desc: "specific spec match collection input",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/openapi",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
			wantPattern:  "projects/pattern-test/locations/global/apis/-/versions/-/specs/openapi",
		},
		{
			desc: "specific spec no match",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/swagger.yaml",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			wantErr:      true,
		},
		{
			desc: "artifact pattern error",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-/artifacts/lint-spectral",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			wantErr:      true,
		},
		{
			desc: "target and resource mismatch",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0",
			wantErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			inputPatternName, err := patterns.ParseResourcePattern(test.inputPattern)
			if err != nil {
				t.Errorf("invalid test.inputPattern %q", test.inputPattern)
			}
			gotPattern, _, gotErr := GenerateCombinedPattern(test.targetPattern, inputPatternName, "")
			if gotPattern != test.wantPattern {
				t.Errorf("GenerateCombinedPattern(%s, %s,\"\") returned unexpected output, want: %q, got: %q ", test.targetPattern, inputPatternName.String(), test.wantPattern, gotPattern)
			}

			if test.wantErr && gotErr == nil {
				t.Errorf("GenerateCombinedPattern(%s, %s, \"\") did not return an error", test.targetPattern, inputPatternName.String())
			}

			if !test.wantErr && gotErr != nil {
				t.Errorf("GenerateCombinedPattern() returned unexpected error: %s", gotErr)
			}
		})
	}
}

func TestGenerateCombinedPatternFilters(t *testing.T) {
	tests := []struct {
		desc          string
		targetPattern *scoring.ResourcePattern
		inputPattern  string
		inputFilter   string
		wantFilter    string
	}{
		{
			desc: "input filter only",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
			inputFilter:  "mime_type.contains('openapi')",
			wantFilter:   "mime_type.contains('openapi')",
		},
		{
			desc: "target filter only",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
				Filter:  "mime_type.contains('openapi')",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
			inputFilter:  "",
			wantFilter:   "mime_type.contains('openapi')",
		},
		{
			desc: "target and input filter identical",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
				Filter:  "mime_type.contains('openapi')",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
			inputFilter:  "mime_type.contains('openapi')",
			wantFilter:   "mime_type.contains('openapi')",
		},
		{
			desc: "target and input filter empty",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
				Filter:  "",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
			inputFilter:  "",
			wantFilter:   "",
		},
		{
			desc: "target and input filter different",
			targetPattern: &scoring.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
				Filter:  "mime_type.contains('openapi')",
			},
			inputPattern: "projects/pattern-test/locations/global/apis/-/versions/-/specs/-",
			inputFilter:  "mime_type.contains('protobuf')",
			wantFilter:   "(mime_type.contains('openapi')) && (mime_type.contains('protobuf'))",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			inputPatternName, err := patterns.ParseResourcePattern(test.inputPattern)
			if err != nil {
				t.Errorf("invalid test.inputPattern %q", test.inputPattern)
			}
			_, gotFilter, _ := GenerateCombinedPattern(test.targetPattern, inputPatternName, test.inputFilter)
			if gotFilter != test.wantFilter {
				t.Errorf("GenerateCombinedPattern(%s, %s,\"\") returned unexpected output, want: %q, got: %q ", test.targetPattern, inputPatternName.String(), test.wantFilter, gotFilter)
			}
		})
	}
}
