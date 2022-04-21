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
			thresholdErrs = append(thresholdErrs, fmt.Errorf("invalid min_value(%d) and max_value(%d), min_value shoud be less than max_value", minValue, maxValue))
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

	if len(thresholds) == 0 {
		// no error returned since thresholds are optional
		return []error{}
	}

	errs := make([]error, 0)

	// keep track of the endpoints of the ranges
	globalRangeMax := minValue
	globalRangeMin := maxValue

	for _, t := range thresholds {
		// Check for validity of the range values
		rangeMin := t.GetRange().GetMin()
		rangeMax := t.GetRange().GetMax()

		// min==max is valid here, one specific value can have a dedicated severity.
		if rangeMin > rangeMax {
			errs = append(errs, fmt.Errorf("invalid range [%d, %d]: range.min cannot be greater than range.max", rangeMin, rangeMax))
			continue
		}

		// Check for out of limits conditions
		if rangeMax > maxValue || rangeMax < minValue {
			errs = append(errs, fmt.Errorf("invalid range.max [%d, %d]: range.max(%d) should be within min_value(%d) and max_value(%d) limits", rangeMin, rangeMax, rangeMax, minValue, maxValue))
		}
		if rangeMin < minValue || rangeMin > maxValue {
			errs = append(errs, fmt.Errorf("invalid range [%d, %d]: range.min(%d) should be within min_value(%d) and max_value(%d) limits", rangeMin, rangeMax, rangeMin, minValue, maxValue))
		}

		// Update global min and max
		if rangeMin <= globalRangeMin {
			globalRangeMin = rangeMin
		}
		if rangeMax >= globalRangeMax {
			globalRangeMax = rangeMax
		}
	}

	// Don't check for missing coverage or threshold overlaps unless all the ranges are valid and in bounds.
	if len(errs) > 0 {
		return errs
	}

	// sort the thresholds based on range.min
	sort.Slice(thresholds, func(i, j int) bool {
		return thresholds[i].GetRange().GetMin() < thresholds[j].GetRange().GetMin()
	})

	// Check for overlaps and gaps
	for i := 0; i < len(thresholds)-1; i++ {
		left, right := thresholds[i].GetRange(), thresholds[i+1].GetRange()
		if left.GetMax() < right.GetMin()-1 {
			errs = append(errs, fmt.Errorf("incomplete coverage: missing coverage between %d and %d", left.GetMax(), right.GetMin()))
		} else if left.GetMax() > right.GetMin()-1 {
			errs = append(errs, fmt.Errorf("invalid thresholds [%d, %d] and [%d, %d]: thresholds must not overlap", left.GetMin(), left.GetMax(), right.GetMin(), right.GetMax()))
		}
	}

	// Finally check for gaps at the endpoints
	if globalRangeMax < maxValue {
		errs = append(errs, fmt.Errorf("incomplete coverage: missing coverage between %d and %d", maxValue, globalRangeMax))
	}
	if globalRangeMin > minValue {
		errs = append(errs, fmt.Errorf("incomplete coverage: missing coverage between %d and %d", globalRangeMin, minValue))
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
