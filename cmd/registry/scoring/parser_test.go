package scoring

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/rpc"
)

func TestValidateScoreDefinition(t *testing.T) {
	tests := []struct {
		desc            string
		parent          string
		scoreDefinition *rpc.ScoreDefinition
		wantNumErr      int
	}{
		// No errors
		{
			desc:   "score formula",
			parent: "projects/demo/locations/global",
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*rpc.NumberThreshold{
							{
								Severity: rpc.Severity_ALERT,
								Range: &rpc.NumberThreshold_NumberRange{
									Min: 61,
									Max: 100,
								},
							},
							{
								Severity: rpc.Severity_WARNING,
								Range: &rpc.NumberThreshold_NumberRange{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_RollupFormula{
					RollupFormula: &rpc.RollUpFormula{
						ScoreFormulas: []*rpc.ScoreFormula{
							{
								Artifact: &rpc.ResourcePattern{
									Pattern: "$resource.spec/artifacts/conformance-report",
								},
								ScoreExpression: "count(errors)",
								ReferenceId:     "lint_errors",
							},
							{
								Artifact: &rpc.ResourcePattern{
									Pattern: "$resource.spec/artifacts/conformance-report",
								},
								ScoreExpression: "count(warnings)",
								ReferenceId:     "lint_warnings",
							},
						},
						RollupExpression: "lint_errors+lint_warnings",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*rpc.NumberThreshold{
							{
								Severity: rpc.Severity_ALERT,
								Range: &rpc.NumberThreshold_NumberRange{
									Min: 61,
									Max: 100,
								},
							},
							{
								Severity: rpc.Severity_WARNING,
								Range: &rpc.NumberThreshold_NumberRange{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						MinValue: 0,
						MaxValue: 100,
					},
				},
			},
		},
		{
			desc:   "only max limit",
			parent: "projects/demo/locations/global",
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						MaxValue: 100,
					},
				},
			},
		},
		// Single errors
		{
			desc:   "target pattern error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.artifact/conformance-report", //error
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_RollupFormula{
					RollupFormula: &rpc.RollUpFormula{
						ScoreFormulas: []*rpc.ScoreFormula{
							{
								Artifact: &rpc.ResourcePattern{
									Pattern: "$resource.spec/artifacts/conformance-report",
								},
								ScoreExpression: "count(errors)",
								ReferenceId:     "lint_errors",
							},
							{
								Artifact: &rpc.ResourcePattern{
									Pattern: "$resource.artifact/conformance-report", //error
								},
								ScoreExpression: "count(warnings)",
								ReferenceId:     "lint_warnings",
							},
						},
						RollupExpression: "lint_errors+lint_warnings",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Percent{
					Percent: &rpc.PercentType{
						Thresholds: []*rpc.NumberThreshold{
							{
								Severity: rpc.Severity_ALERT,
								Range: &rpc.NumberThreshold_NumberRange{ //error
									Min: 60,
									Max: 100,
								},
							},
							{
								Severity: rpc.Severity_WARNING,
								Range: &rpc.NumberThreshold_NumberRange{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*rpc.NumberThreshold{
							{
								Severity: rpc.Severity_ALERT,
								Range: &rpc.NumberThreshold_NumberRange{ //error
									Min: 62,
									Max: 100,
								},
							},
							{
								Severity: rpc.Severity_WARNING,
								Range: &rpc.NumberThreshold_NumberRange{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						Thresholds: []*rpc.NumberThreshold{
							{
								Severity: rpc.Severity_ALERT,
								Range: &rpc.NumberThreshold_NumberRange{
									Min: 61,
									Max: 100,
								},
							},
							{
								Severity: rpc.Severity_WARNING,
								Range: &rpc.NumberThreshold_NumberRange{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/audit-report",
						},
						ScoreExpression: "isApproved",
					},
				},
				Type: &rpc.ScoreDefinition_Boolean{
					Boolean: &rpc.BooleanType{
						Thresholds: []*rpc.BooleanThreshold{ //error
							{
								Severity: rpc.Severity_ALERT,
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.specs/artifacts/conformance-report", //error
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-report",
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*rpc.NumberThreshold{
							{
								Severity: rpc.Severity_ALERT,
								Range: &rpc.NumberThreshold_NumberRange{ //error
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.specs/artifacts/conformance-report", //error
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*rpc.NumberThreshold{
							{
								Severity: rpc.Severity_ALERT,
								Range: &rpc.NumberThreshold_NumberRange{ //error
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.specs/artifacts/conformance-report", //error
						},
						ScoreExpression: "count(errors)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*rpc.NumberThreshold{
							{
								Severity: rpc.Severity_ALERT,
								Range: &rpc.NumberThreshold_NumberRange{ //error
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
		// Missing oneof components
		{
			desc:   "missing formula",
			parent: "projects/demo/locations/global",
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/specs/-", //error
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*rpc.NumberThreshold{
							{
								Severity: rpc.Severity_OK,
								Range: &rpc.NumberThreshold_NumberRange{
									Min: 0,
									Max: 59,
								},
							},
							{
								Severity: rpc.Severity_ALERT,
								Range: &rpc.NumberThreshold_NumberRange{
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
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
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
			ctx := context.Background()
			gotErrs := ValidateScoreDefinition(ctx, test.parent, test.scoreDefinition)
			if len(gotErrs) != test.wantNumErr {
				t.Errorf("ValidateScoreDefinition() returned unexpected no. of errors: want %d, got %s", test.wantNumErr, gotErrs)
			}
		})
	}
}

func TestValidateScoreFormula(t *testing.T) {
	tests := []struct {
		desc          string
		targetPattern *rpc.ResourcePattern
		scoreFormula  *rpc.ScoreFormula
		wantNumErr    int
	}{
		// No errors
		{
			desc: "score formula",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/conformance-report",
				},
				ScoreExpression: "count(errors)",
			},
		},
		// Single errors
		{
			desc: "no $resource reference",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-/artifacts/conformance-report", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 1,
		},
		{
			desc: "invalid $resource reference",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.specs/artifacts/conformance-report", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 1,
		},
		{
			desc: "invalid $resource wrt targetName",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/conformance-report", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 1,
		},
		{
			desc: "missing artifact name",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/-", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 1,
		},
		// Combination errors
		{
			desc: "invalid pattern and missing name",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.specs/-/artifacts/-", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 2,
		},
		{
			desc: "no reference and missing name",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-/artifacts/-", //error
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 2,
		},
		{
			desc: "invalid reference and missing name",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-", //error
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/-",
				},
				ScoreExpression: "count(errors)",
			},
			wantNumErr: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			targetName, _ := patterns.ParseResourcePattern(test.targetPattern.GetPattern())
			gotErrs := validateScoreFormula(targetName, test.scoreFormula)
			if len(gotErrs) != test.wantNumErr {
				t.Errorf("validateScoreFormula() returned unexpected no. of errors: want %d, got %s", test.wantNumErr, gotErrs)
			}
		})
	}
}

func TestValidateNumberThresholds(t *testing.T) {
	tests := []struct {
		desc       string
		minValue   int32
		maxValue   int32
		thresholds []*rpc.NumberThreshold
		wantNumErr int
	}{
		// no errors
		{
			desc:     "percentage thresholds",
			minValue: 0,
			maxValue: 100,
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: -50,
						Max: -20,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: -19,
						Max: 10,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 61,
						Max: 100,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 0,
						Max: 10,
					},
				},
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 91,
						Max: 100,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 11,
						Max: 30,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 71,
						Max: 90,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
				},
			},
			wantNumErr: 2,
		},
		{
			desc:     "range.min greater than range.max",
			minValue: 0,
			maxValue: 100,
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Max: 0,
						Min: 50,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: -1,
						Max: 30,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 5,
						Max: 30,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 31,
						Max: 60,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 0,
						Max: 29,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 31,
						Max: 59,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 30,
						Max: 60,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: -1,
						Max: 50,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: -1,
						Max: 50,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: -1,
						Max: 50,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 2,
						Max: 50,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 2,
						Max: 50,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 2,
						Max: 50,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 2,
						Max: 50,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 0,
						Max: 30,
					},
				},
				{
					Severity: rpc.Severity_WARNING,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 30,
						Max: 60,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
			thresholds: []*rpc.NumberThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Range: &rpc.NumberThreshold_NumberRange{
						Min: 0,
						Max: 100,
					},
				},
				{
					Severity: rpc.Severity_OK,
					Range: &rpc.NumberThreshold_NumberRange{
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
				t.Errorf("validateNumberThresholds() returned unexpected no. of errors: want %d, got %s", test.wantNumErr, gotErrs)
			}
		})
	}

}

func TestValidateBooleanThresholds(t *testing.T) {
	tests := []struct {
		desc       string
		thresholds []*rpc.BooleanThreshold
		wantNumErr int
	}{
		// no errors
		{
			desc: "normal case",
			thresholds: []*rpc.BooleanThreshold{
				{
					Severity: rpc.Severity_ALERT,
					Value:    false,
				},
				{
					Severity: rpc.Severity_OK,
					Value:    true,
				},
			},
		},
		{
			desc: "same severity",
			thresholds: []*rpc.BooleanThreshold{
				{
					Severity: rpc.Severity_OK,
					Value:    false,
				},
				{
					Severity: rpc.Severity_OK,
					Value:    true,
				},
			},
		},
		// single errors
		{
			desc: "missing false",
			thresholds: []*rpc.BooleanThreshold{
				{
					Severity: rpc.Severity_OK,
					Value:    true,
				},
			},
			wantNumErr: 1,
		},
		{
			desc: "missing true",
			thresholds: []*rpc.BooleanThreshold{
				{
					Severity: rpc.Severity_WARNING,
					Value:    false,
				},
			},
			wantNumErr: 1,
		},
		{
			desc: "duplicate entries",
			thresholds: []*rpc.BooleanThreshold{
				{
					Severity: rpc.Severity_OK,
					Value:    true,
				},
				{
					Severity: rpc.Severity_WARNING,
					Value:    true,
				},
				{
					Severity: rpc.Severity_ALERT,
					Value:    false,
				},
			},
			wantNumErr: 1,
		},
		// combination errors
		{
			desc: "missing and duplicate entries",
			thresholds: []*rpc.BooleanThreshold{
				{
					Severity: rpc.Severity_OK,
					Value:    true,
				},
				{
					Severity: rpc.Severity_WARNING,
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
				t.Errorf("validateBooleanThresholds() returned unexpected no. of errors: want %d, got %s", test.wantNumErr, gotErrs)
			}
		})
	}
}
