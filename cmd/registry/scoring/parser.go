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
		default:
			formulaErrs = append(formulaErrs, fmt.Errorf("missing formula, either 'score_formula' or 'rollup_formula' should be set"))
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
		minValue := scoreType.Integer.GetMinValue()
		maxValue := scoreType.Integer.GetMaxValue()
		// if minValue==maxValue, means the score can take only one value, in that case integer is not the correct type.
		// other types will be supported in the future (enums) to cover this case.
		if minValue >= maxValue {
			thresholdErrs = append(thresholdErrs, fmt.Errorf("invalid min_value and max_value, should be: min_value < max_value"))
		} else { // validate that the set thresholds are within minValue and maxValue limits
			errs := validateNumberThresholds(scoreType.Integer.GetThresholds(), minValue, maxValue)
			thresholdErrs = append(thresholdErrs, errs...)
		}
	case *rpc.ScoreDefinition_Boolean:
		errs := validateBooleanThresholds(scoreType.Boolean.GetThresholds())
		thresholdErrs = append(thresholdErrs, errs...)
	default:
		thresholdErrs = append(thresholdErrs, fmt.Errorf("missing type, either of 'percent', 'integer' or 'boolean' should be set"))
	}

	errs := make([]error, 0, len(patternErrs)+len(formulaErrs)+len(thresholdErrs))
	errs = append(errs, patternErrs...)
	errs = append(errs, formulaErrs...)
	errs = append(errs, thresholdErrs...)

	return errs
}

func validateScoreFormula(targetName patterns.ResourceName, scoreFormula *rpc.ScoreFormula) []error {
	errs := make([]error, 0)

	// Validation checks for score_formula.artifact.pattern
	pattern := scoreFormula.GetArtifact().GetPattern()

	// pattern should have a valid $resource pattern
	_, entityType, err := patterns.GetReferenceEntityType(pattern)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid score_formula.artifact.pattern: %s", err))
	} else if entityType == "default" {
		// pattern should always start with a $resource reference
		errs = append(errs, fmt.Errorf("invalid score_formula.artifact.pattern: %q, must always start with '$resource.(api|version|spec|artifact)'", pattern))
	} else if _, err = patterns.GetReferenceEntityValue(pattern, targetName); err != nil {
		// $resource should have valid entity reference wrt target_resource
		errs = append(errs, fmt.Errorf("invalid $resource reference in score_formula.artifact.pattern: %s", err))
	}

	// score_formula.pattern.pattern should not end with a "-"
	if strings.HasSuffix(scoreFormula.GetArtifact().GetPattern(), "/-") {
		errs = append(errs, fmt.Errorf("invalid score_formula.artifact.pattern : %q, it should end with a resourceID and not a \"-\"", pattern))
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
		min := t.GetRange().GetMin()
		max := t.GetRange().GetMax()

		// min==max is valid here, one specific value can have a dedicated severity.
		if min > max {
			// invalid min and max values, skip the remaining validation
			errs = append(errs, fmt.Errorf("invalid range.min: %d and range.max: %d values, should be range.min <= range.max", min, max))
			continue
		}

		if min <= pointer {
			// overlap with previous range or out of bound of minValue if pointer == minValue - 1
			errs = append(errs, fmt.Errorf("overlap or out of bounds (<%d) in threshold value: {%v}", minValue, t))
		}

		if min > pointer+1 {
			// missed values between current and previous range
			errs = append(errs, fmt.Errorf("missing coverage for some values in threshold value: {%v}", t))
		}

		if max > maxValue {
			// out of bounds of max value
			errs = append(errs, fmt.Errorf("out of bounds (>%d) in threshold value: {%v}", maxValue, t))
		}

		pointer = max

		// Check if the pointer reaches maxValue in the last iteration
		if i == len(thresholds)-1 && pointer < maxValue {
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
