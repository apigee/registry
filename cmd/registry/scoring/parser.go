package scoring

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/apigee/registry/rpc"

	"github.com/apigee/registry/cmd/registry/patterns"
)

func ValidateScoreDefinition(ctx context.Context, parent string, scoreDefinition *rpc.ScoreDefinition) []error {

	patternErrs := make([]error, 0)
	// target_resource.pattern should be a valid resource pattern
	targetName, err := patterns.ParseResourcePattern(fmt.Sprintf("%s/%s", parent, scoreDefinition.GetTargetResource().GetPattern()))
	if err != nil {
		patternErrs = append(patternErrs, err)
	}

	// TODO: Check for valid filter in target_resource

	formulaErrs := make([]error, 0)
	// Validate formula if target_resource.pattern was valid
	if len(patternErrs) == 0 {
		switch formula := scoreDefinition.GetFormula().(type) {
		case *rpc.ScoreDefinition_ScoreFormula:
			errs := validateScoreFormula(targetName, formula.ScoreFormula)
			formulaErrs = append(formulaErrs, errs...)
		case *rpc.ScoreDefinition_RollupFormula:
			for _, scoreFormula := range formula.RollupFormula.GetScoreFormulas() {
				errs := validateScoreFormula(targetName, scoreFormula)
				formulaErrs = append(formulaErrs, errs...)
			}
		}
	}

	thresholdErrs := make([]error, 0)
	// Validate threshold
	switch scoreType := scoreDefinition.GetType().(type) {
	case *rpc.ScoreDefinition_Percent:
		// minValue: 0 maxValue:100
		// validate that the set thresholds are within these bounds
		errs := validateNumberThresholds(scoreType.Percent.GetThresholds(), 0, 100)
		thresholdErrs = append(thresholdErrs, errs...)
	case *rpc.ScoreDefinition_Integer:
		// defaults if not set: minValue: 0 maxValue:0
		// validate that the set thresholds are within these bound
		errs := validateNumberThresholds(scoreType.Integer.GetThresholds(), scoreType.Integer.GetMinValue(), scoreType.Integer.GetMaxValue())
		thresholdErrs = append(thresholdErrs, errs...)
	case *rpc.ScoreDefinition_Boolean:
		errs := validateBooleanThresholds(scoreType.Boolean.GetThresholds())
		thresholdErrs = append(thresholdErrs, errs...)
	}

	errs := make([]error, len(patternErrs)+len(formulaErrs)+len(thresholdErrs))
	errs = append(errs, patternErrs...)
	errs = append(errs, formulaErrs...)
	errs = append(errs, thresholdErrs...)

	return errs
}

func validateScoreFormula(targetName patterns.ResourceName, scoreFormula *rpc.ScoreFormula) []error {
	errs := make([]error, 0)

	// Validation checks for score_formula.pattern.pattern
	pattern := scoreFormula.GetPattern().GetPattern()

	// pattern should have a valid $resource pattern
	_, entityType, err := patterns.GetReferenceEntityType(pattern)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid score_formula.pattern.pattern: %s", err))
	}
	// $resource should have valid entity reference wrt target_resource
	_, err = patterns.GetReferenceEntityValue(pattern, targetName)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid $resource reference: %s", err))
	}
	// pattern should always start with a $resource reference
	if entityType == "default" {
		errs = append(errs, fmt.Errorf("invalid score_formula.pattern.pattern: %q, must always start with '$resource.(api|version|spec|artifact)'", pattern))
	}
	// score_formula.pattern.pattern should not end with a "-"
	if strings.HasSuffix(scoreFormula.GetPattern().GetPattern(), "/-") {
		errs = append(errs, fmt.Errorf("invalid score_formula.pattern.pattern : %q, it should end with a name and not a \"-\"", pattern))
	}

	return errs
}

func validateNumberThresholds(thresholds []*rpc.NumberThreshold, minValue, maxValue int32) []error {
	errs := make([]error, 0)

	// sort the thresholds based on range.min
	sort.Slice(thresholds, func(i, j int) bool {
		return thresholds[i].GetRange().GetMin() < thresholds[j].GetRange().GetMin()
	})

	// This loop runs a pointer through the sorted ranges and checks for overlaps, missed values or out of bound scenarios.
	pointer := minValue - 1
	for i, t := range thresholds {
		if t.GetRange().GetMin() <= pointer {
			// overlap with previous range or out of bound of minValue if pointer == minValue - 1
			errs = append(errs, fmt.Errorf("overlap or out of bounds (<%d) in threshold value: {%v}", minValue, t))
		}

		if t.GetRange().GetMin() > pointer+1 {
			// missed values between current and previous range
			errs = append(errs, fmt.Errorf("missing coverage for some values in threshold value: {%v}", t))
		}

		if t.GetRange().GetMax() > maxValue {
			// out of bounds of max value
			errs = append(errs, fmt.Errorf("out of bounds (>%d) in threshold value: {%v}", maxValue, t))
		}

		pointer = t.GetRange().GetMax()

		// Check if the pointer reaches maxValue in the last iteration
		if i == len(thresholds)-1 && pointer != maxValue {
			// missed values from maxValue
			errs = append(errs, fmt.Errorf("missing coverage for max_value(%d) in threshold value: {%v}", maxValue, t))
		}
	}

	return errs
}

func validateBooleanThresholds(thresholds []*rpc.BooleanThreshold) []error {
	errs := make([]error, 0)

	var isFalseCovered, isTrueCovered bool
	for _, t := range thresholds {
		if isFalseCovered && !t.GetValue() {
			errs = append(errs, fmt.Errorf("duplicate entries for 'false' value"))
		}

		if isTrueCovered && t.GetValue() {
			errs = append(errs, fmt.Errorf("duplicate entries for 'true' value"))
		}

		isFalseCovered = !t.GetValue() || isFalseCovered
		isTrueCovered = t.GetValue() || isTrueCovered
	}

	if !isTrueCovered || !isFalseCovered {
		errs = append(errs, fmt.Errorf("missing coverage for one or both of the boolean values"))
	}
	return errs
}
