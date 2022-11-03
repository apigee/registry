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
	"fmt"
	"sort"
	"strings"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/filtering"

	"github.com/apigee/registry/cmd/registry/patterns"
)

func ValidateScoreDefinition(parent string, scoreDefinition *rpc.ScoreDefinition) []error {
	totalErrs := make([]error, 0)

	// target_resource.pattern should be a valid resource pattern
	targetName, err := patterns.ParseResourcePattern(fmt.Sprintf("%s/%s", parent, scoreDefinition.GetTargetResource().GetPattern()))
	if err != nil {
		totalErrs = append(totalErrs, err)
	}

	// target_resource.filter should be a valid filter on the resource
	err = validateFilter(targetName, scoreDefinition.GetTargetResource().GetFilter())
	if err != nil {
		totalErrs = append(totalErrs, err)
	}

	// Validate formula if there were no errors in target_resource
	if len(totalErrs) == 0 {
		switch formula := scoreDefinition.GetFormula().(type) {
		case *rpc.ScoreDefinition_ScoreFormula:
			errs := validateScoreFormula(targetName, formula.ScoreFormula)
			totalErrs = append(totalErrs, errs...)
		case *rpc.ScoreDefinition_RollupFormula:
			if len(formula.RollupFormula.GetScoreFormulas()) == 0 {
				totalErrs = append(totalErrs, fmt.Errorf("missing rollup_formula.score_formulas"))
			}
			for _, scoreFormula := range formula.RollupFormula.GetScoreFormulas() {
				errs := validateScoreFormula(targetName, scoreFormula)
				totalErrs = append(totalErrs, errs...)
			}
			if formula.RollupFormula.GetRollupExpression() == "" {
				totalErrs = append(totalErrs, fmt.Errorf("missing rollup_formula.rollup_expression"))
			}
		default:
			totalErrs = append(totalErrs, fmt.Errorf("missing formula, either 'score_formula' or 'rollup_formula' should be set"))
		}
	}

	// Validate threshold
	switch scoreType := scoreDefinition.GetType().(type) {
	case *rpc.ScoreDefinition_Percent:
		// minValue: 0 maxValue:100
		// validate that the set thresholds are within these bounds
		errs := validateNumberThresholds(scoreType.Percent.GetThresholds(), 0, 100)
		totalErrs = append(totalErrs, errs...)
	case *rpc.ScoreDefinition_Integer:
		// defaults if not set: minValue: 0 maxValue:0
		minValue := scoreType.Integer.GetMinValue()
		maxValue := scoreType.Integer.GetMaxValue()
		// if minValue==maxValue, means the score can take only one value, in that case integer is not the correct type.
		// other types will be supported in the future (enums) to cover this case.
		if minValue >= maxValue {
			totalErrs = append(totalErrs, fmt.Errorf("invalid min_value(%d) and max_value(%d), min_value should be less than max_value", minValue, maxValue))
		} else { // validate that the set thresholds are within minValue and maxValue limits
			errs := validateNumberThresholds(scoreType.Integer.GetThresholds(), minValue, maxValue)
			totalErrs = append(totalErrs, errs...)
		}
	case *rpc.ScoreDefinition_Boolean:
		errs := validateBooleanThresholds(scoreType.Boolean.GetThresholds())
		totalErrs = append(totalErrs, errs...)
	default:
		totalErrs = append(totalErrs, fmt.Errorf("missing type, either of 'percent', 'integer' or 'boolean' should be set"))
	}

	return totalErrs
}

func ValidateScoreCardDefinition(parent string, scoreCardDefinition *rpc.ScoreCardDefinition) []error {
	totalErrs := make([]error, 0)

	// target_resource.pattern should be a valid resource pattern
	targetName, err := patterns.ParseResourcePattern(fmt.Sprintf("%s/%s", parent, scoreCardDefinition.GetTargetResource().GetPattern()))
	if err != nil {
		totalErrs = append(totalErrs, err)
	}

	// target_resource.filter should be a valid filter on the resource
	err = validateFilter(targetName, scoreCardDefinition.GetTargetResource().GetFilter())
	if err != nil {
		totalErrs = append(totalErrs, err)
	}

	scorePatterns := scoreCardDefinition.GetScorePatterns()

	// Check if score_patterns are set
	if len(scorePatterns) == 0 {
		totalErrs = append(totalErrs, fmt.Errorf("missing score_patterns"))
		return totalErrs
	}

	// Validate score_patterns only if target_resource.pattern is valid
	if len(totalErrs) == 0 {
		for _, pattern := range scorePatterns {
			// Should have valid $resource references
			errs := validateReferencesInPattern(targetName, pattern)
			totalErrs = append(totalErrs, errs...)

			// Should not end with a "-"
			if strings.HasSuffix(pattern, "/-") {
				totalErrs = append(totalErrs, fmt.Errorf("invalid score_pattern : %q, it should end with a resourceID and not a \"-\"", pattern))
			}
		}
	}

	return totalErrs
}

func validateReferencesInPattern(targetName patterns.ResourceName, pattern string) []error {
	errs := make([]error, 0)

	// pattern should have a valid $resource reference
	_, entityType, err := patterns.GetReferenceEntityType(pattern)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid pattern: %q, %s", pattern, err))
	} else if entityType == "default" {
		// pattern should always start with a $resource reference
		errs = append(errs, fmt.Errorf("invalid pattern: %q, must always start with '$resource.(api|version|spec|artifact)'", pattern))
	} else if _, err = patterns.GetReferenceEntityValue(pattern, targetName); err != nil {
		// $resource should have valid entity reference wrt target_resource
		errs = append(errs, fmt.Errorf("invalid pattern: %q, invalid $resource reference in pattern: %s", pattern, err))
	}

	return errs
}

func validateFilter(targetName patterns.ResourceName, filter string) error {
	switch targetName.(type) {
	case patterns.SpecName:
		_, err := filtering.NewFilter(filter, filtering.SpecFields)
		if err != nil {
			return fmt.Errorf("invalid filter: %q, %s", filter, err)
		}
	case patterns.VersionName:
		_, err := filtering.NewFilter(filter, filtering.VersionFields)
		if err != nil {
			return fmt.Errorf("invalid filter: %q, %s", filter, err)
		}
	case patterns.ApiName:
		_, err := filtering.NewFilter(filter, filtering.ApiFields)
		if err != nil {
			return fmt.Errorf("invalid filter: %q, %s", filter, err)
		}
	case patterns.ArtifactName:
		_, err := filtering.NewFilter(filter, filtering.ArtifactFields)
		if err != nil {
			return fmt.Errorf("invalid filter: %q, %s", filter, err)
		}
	}
	return nil
}

func validateScoreFormula(targetName patterns.ResourceName, scoreFormula *rpc.ScoreFormula) []error {
	errs := make([]error, 0)

	// Validation checks for score_formula.artifact.pattern
	pattern := scoreFormula.GetArtifact().GetPattern()

	// Should have valid $resource references
	patternErrs := validateReferencesInPattern(targetName, pattern)
	errs = append(errs, patternErrs...)

	// Should not end with a "-"
	if strings.HasSuffix(pattern, "/-") {
		errs = append(errs, fmt.Errorf("invalid score_formula.artifact.pattern : %q, it should end with a resourceID and not a \"-\"", pattern))
	}

	if scoreFormula.GetScoreExpression() == "" {
		errs = append(errs, fmt.Errorf("missing score_formula.score_expression"))
	}

	if refId := scoreFormula.GetReferenceId(); refId != "" && strings.Contains(refId, "-") {
		errs = append(errs, fmt.Errorf("invalid score_formula.reference_id: %s, it should not contain hyphens '-'", refId))
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
			errs = append(errs, fmt.Errorf("invalid range [%d, %d]: range.max(%d) should be within min_value(%d) and max_value(%d) limits", rangeMin, rangeMax, rangeMax, minValue, maxValue))
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

func matchResourceWithTarget(targetPattern *rpc.ResourcePattern, resourceInstance patterns.ResourceInstance, project string) error {
	targetPatternName, err := patterns.ParseResourcePattern(fmt.Sprintf("%s/%s", project, targetPattern.GetPattern()))
	if err != nil {
		return err
	}

	switch tp := targetPatternName.(type) {
	case patterns.SpecName:
		// Check if targetPattern and resource match in type
		r, ok := resourceInstance.(patterns.SpecResource)
		if !ok {
			return fmt.Errorf("resource %q doesn't match target pattern %q", r, tp)
		}

		// Convert ResourceName to SpecName
		rName, ok := resourceInstance.ResourceName().(patterns.SpecName)
		if !ok {
			return fmt.Errorf("resourceName %q doesn't match target pattern %q", rName, tp)
		}

		// Check if the individual entities match
		if tp.Name.ApiID != "-" && tp.Name.ApiID != rName.Name.ApiID {
			return fmt.Errorf("api mismatch in resource %q and target pattern %v", resourceInstance.ResourceName().String(), targetPattern)
		}
		if tp.Name.VersionID != "-" && tp.Name.VersionID != rName.Name.VersionID {
			return fmt.Errorf("version mismatch in resource %q and target pattern %v", resourceInstance.ResourceName().String(), targetPattern)
		}
		if tp.Name.SpecID != "-" && tp.Name.SpecID != rName.Name.SpecID {
			return fmt.Errorf("spec mismatch in resource %q and target pattern %v", resourceInstance.ResourceName().String(), targetPattern)
		}

		// Check if the filter matches with the resource
		filterStr := targetPattern.GetFilter()
		filter, err := filtering.NewFilter(filterStr, filtering.SpecFields)
		if err != nil {
			return fmt.Errorf("invalid filter %q: %s", filterStr, err)
		}

		rMap, err := filtering.SpecMapFromMessage(r.Spec)
		if err != nil {
			return fmt.Errorf("internal error while applying filter: %s", err)
		}

		match, err := filter.Matches(rMap)
		if err != nil {
			return fmt.Errorf("failed applying filter %q: %s", filterStr, err)
		}

		if !match {
			return fmt.Errorf("target pattern filter %q does not match with resource %q", filterStr, resourceInstance.ResourceName().String())
		}
	case patterns.VersionName:
		// Check if targetPattern and resource match in type
		r, ok := resourceInstance.(patterns.VersionResource)
		if !ok {
			return fmt.Errorf("resource %q doesn't match target pattern %q", r, tp)
		}

		// Convert ResourceName to VersionName
		rName, ok := resourceInstance.ResourceName().(patterns.VersionName)
		if !ok {
			return fmt.Errorf("resourceName %q doesn't match target pattern %q", rName, tp)
		}

		// Check if the individual entities match
		if tp.Name.ApiID != "-" && tp.Name.ApiID != rName.Name.ApiID {
			return fmt.Errorf("api mismatch in resource %s and target pattern %v", resourceInstance.ResourceName().String(), targetPattern)
		}
		if tp.Name.VersionID != "-" && tp.Name.VersionID != rName.Name.VersionID {
			return fmt.Errorf("version mismatch in resource %s and target pattern %v", resourceInstance.ResourceName().String(), targetPattern)
		}

		// Check if the filter matches with the resource
		filterStr := targetPattern.GetFilter()
		filter, err := filtering.NewFilter(filterStr, filtering.VersionFields)
		if err != nil {
			return fmt.Errorf("invalid filter %q: %s", filterStr, err)
		}

		rMap, err := filtering.VersionMapFromMessage(r.Version)
		if err != nil {
			return fmt.Errorf("internal error while applying filter: %s", err)
		}

		match, err := filter.Matches(rMap)
		if err != nil {
			return fmt.Errorf("failed applying filter %q: %s", filterStr, err)
		}

		if !match {
			return fmt.Errorf("target pattern filter %q does not match with resource %q", filterStr, resourceInstance.ResourceName().String())
		}
	case patterns.ApiName:
		// Check if targetPattern and resource match in type
		r, ok := resourceInstance.(patterns.ApiResource)
		if !ok {
			return fmt.Errorf("resource %s doesn't match target pattern %s", r, tp)
		}

		// Convert ResourceName to ApiName
		rName, ok := resourceInstance.ResourceName().(patterns.ApiName)
		if !ok {
			return fmt.Errorf("resourceName %q doesn't match target pattern %q", rName, tp)
		}

		// Check if the individual entities match
		if tp.Name.ApiID != "-" && tp.Name.ApiID != rName.Name.ApiID {
			return fmt.Errorf("api mismatch in resource %s and target pattern %v", resourceInstance.ResourceName().String(), targetPattern)
		}

		// Check if the filter matches with the resource
		filterStr := targetPattern.GetFilter()
		filter, err := filtering.NewFilter(filterStr, filtering.ApiFields)
		if err != nil {
			return fmt.Errorf("invalid filter %q: %s", filterStr, err)
		}

		rMap, err := filtering.ApiMapFromMessage(r.Api)
		if err != nil {
			return fmt.Errorf("internal error while applying filter: %s", err)
		}

		match, err := filter.Matches(rMap)
		if err != nil {
			return fmt.Errorf("failed applying filter %q: %s", filterStr, err)
		}

		if !match {
			return fmt.Errorf("target pattern filter %q does not match with resource %q", filterStr, resourceInstance.ResourceName().String())
		}
	default:
		return fmt.Errorf("unsupported resource type %T", targetPatternName)
	}

	return nil
}
