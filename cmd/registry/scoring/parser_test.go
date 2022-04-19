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
	}{
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
			desc:   "no limits",
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
					Integer: &rpc.IntegerType{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			gotErrs := ValidateScoreDefinition(ctx, test.parent, test.scoreDefinition)
			if len(gotErrs) > 0 {
				t.Errorf("ValidateScoreDefinition() returned unexpected errors: %s", gotErrs)
			}
		})
	}
}

func TestValidateScoreDefinitionError(t *testing.T) {
	tests := []struct {
		desc            string
		parent          string
		scoreDefinition *rpc.ScoreDefinition
	}{
		{
			desc:   "target pattern error",
			parent: "projects/demo/locations/global",
			scoreDefinition: &rpc.ScoreDefinition{
				Id:   "test-score-definition",
				Kind: "ScoreDefinition",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/specs/-",
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
							Pattern: "$resource.artifact/conformance-report",
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
									Pattern: "$resource.artifact/conformance-report",
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
								Range: &rpc.NumberThreshold_NumberRange{
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
								Range: &rpc.NumberThreshold_NumberRange{
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
						Thresholds: []*rpc.BooleanThreshold{
							{
								Severity: rpc.Severity_ALERT,
								Value:    false,
							},
							{
								Severity: rpc.Severity_ALERT,
								Value:    false,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			gotErrs := ValidateScoreDefinition(ctx, test.parent, test.scoreDefinition)
			if len(gotErrs) == 0 {
				t.Errorf("expected ValidateScoreDefinition() to return errors")
			}
		})
	}
}

func TestValidateScoreFormula(t *testing.T) {
	tests := []struct {
		desc          string
		targetPattern *rpc.ResourcePattern
		scoreFormula  *rpc.ScoreFormula
	}{
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
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			targetName, _ := patterns.ParseResourcePattern(test.targetPattern.GetPattern())
			gotErrs := validateScoreFormula(targetName, test.scoreFormula)
			if len(gotErrs) > 0 {
				t.Errorf("validateScoreFormula() returned unexpected errors: %s", gotErrs)
			}
		})
	}
}

func TestValidateScoreFormulaError(t *testing.T) {
	tests := []struct {
		desc          string
		targetPattern *rpc.ResourcePattern
		scoreFormula  *rpc.ScoreFormula
	}{
		{
			desc: "no $resource reference",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-/artifacts/conformance-report",
				},
				ScoreExpression: "count(errors)",
			},
		},
		{
			desc: "invalid $resource reference",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-/specs/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.specs/artifacts/conformance-report",
				},
				ScoreExpression: "count(errors)",
			},
		},
		{
			desc: "invalid $resource wrt targetName",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/conformance-report",
				},
				ScoreExpression: "count(errors)",
			},
		},
		{
			desc: "missing artifact name",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "projects/demo/locations/global/apis/-/versions/-",
			},
			scoreFormula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/-",
				},
				ScoreExpression: "count(errors)",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			targetName, _ := patterns.ParseResourcePattern(test.targetPattern.GetPattern())
			gotErrs := validateScoreFormula(targetName, test.scoreFormula)
			if len(gotErrs) == 0 {
				t.Errorf("expected validateScoreFormula() to return errors")
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
	}{
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
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := validateNumberThresholds(test.thresholds, test.minValue, test.maxValue)
			if len(gotErrs) > 0 {
				t.Errorf("validateNumberThresholds() returned unexpected errors: %s", gotErrs)
			}
		})
	}

}

func TestValidateNumberThresholdsError(t *testing.T) {
	tests := []struct {
		desc       string
		minValue   int32
		maxValue   int32
		thresholds []*rpc.NumberThreshold
	}{
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
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := validateNumberThresholds(test.thresholds, test.minValue, test.maxValue)
			if len(gotErrs) == 0 {
				t.Errorf("expected validateNumberThresholds() to return errors")
			}
		})
	}
}

func TestValidateBooleanThresholds(t *testing.T) {
	tests := []struct {
		desc       string
		thresholds []*rpc.BooleanThreshold
	}{
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
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := validateBooleanThresholds(test.thresholds)
			if len(gotErrs) > 0 {
				t.Errorf("validateBooleanThresholds() returned unexpected errors: %s", gotErrs)
			}
		})
	}
}

func TestValidateBooleanThresholdsError(t *testing.T) {
	tests := []struct {
		desc       string
		thresholds []*rpc.BooleanThreshold
	}{
		{
			desc: "missing false",
			thresholds: []*rpc.BooleanThreshold{
				{
					Severity: rpc.Severity_OK,
					Value:    true,
				},
			},
		},
		{
			desc: "missing true",
			thresholds: []*rpc.BooleanThreshold{
				{
					Severity: rpc.Severity_WARNING,
					Value:    false,
				},
			},
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
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := validateBooleanThresholds(test.thresholds)
			if len(gotErrs) == 0 {
				t.Errorf("expected validateBooleanThresholds() to return errors")
			}
		})
	}
}
