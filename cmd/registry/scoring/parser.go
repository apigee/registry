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

	"github.com/apigee/registry/pkg/application/scoring"
	"github.com/apigee/registry/pkg/names"

	"github.com/apigee/registry/cmd/registry/patterns"
)

func ValidateScoreDefinition(parent string, scoreDefinition *scoring.ScoreDefinition) []error {
	totalErrs := make([]error, 0)

	// target_resource.pattern should be a valid resource pattern
	targetName, err := patterns.ParseResourcePattern(fmt.Sprintf("%s/%s", parent, scoreDefinition.GetTargetResource().GetPattern()))
	if err != nil {
		totalErrs = append(totalErrs, err)
	}

	// Validate formula if there were no errors in target_resource
	if len(totalErrs) == 0 {
		switch formula := scoreDefinition.GetFormula().(type) {
		case *scoring.ScoreDefinition_ScoreFormula:
			errs := validateScoreFormula(targetName, formula.ScoreFormula)
			totalErrs = append(totalErrs, errs...)
		case *scoring.ScoreDefinition_RollupFormula:
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
	case *scoring.ScoreDefinition_Percent:
		// minValue: 0 maxValue:100
		// validate that the set thresholds are within these bounds
		errs := validateNumberThresholds(scoreType.Percent.GetThresholds(), 0, 100)
		totalErrs = append(totalErrs, errs...)
	case *scoring.ScoreDefinition_Integer:
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
	case *scoring.ScoreDefinition_Boolean:
		errs := validateBooleanThresholds(scoreType.Boolean.GetThresholds())
		totalErrs = append(totalErrs, errs...)
	default:
		totalErrs = append(totalErrs, fmt.Errorf("missing type, either of 'percent', 'integer' or 'boolean' should be set"))
	}

	return totalErrs
}

func ValidateScoreCardDefinition(parent string, scoreCardDefinition *scoring.ScoreCardDefinition) []error {
	totalErrs := make([]error, 0)

	// target_resource.pattern should be a valid resource pattern
	targetName, err := patterns.ParseResourcePattern(fmt.Sprintf("%s/%s", parent, scoreCardDefinition.GetTargetResource().GetPattern()))
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

func validateScoreFormula(targetName patterns.ResourceName, scoreFormula *scoring.ScoreFormula) []error {
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

func validateNumberThresholds(thresholds []*scoring.NumberThreshold, minValue, maxValue int32) []error {
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

func validateBooleanThresholds(thresholds []*scoring.BooleanThreshold) []error {
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

func GenerateCombinedPattern(targetPattern *scoring.ResourcePattern, inputPatternName patterns.ResourceName, inputFilter string) (string, string, error) {
	projectID := strings.Split(inputPatternName.Project(), "/")[1]
	// Generate a common list pattern based on the supplied input and the targetPattern in the definition
	targetPatternName, err := patterns.ParseResourcePattern(fmt.Sprintf("projects/%s/locations/global/%s", projectID, targetPattern.GetPattern()))
	if err != nil {
		return "", "", fmt.Errorf("invalid targetPattern in ScoreDefinition: %s", err)
	}
	fmt.Printf("%s", targetPatternName.String())

	// Merge the two patterns into one
	switch tp := targetPatternName.(type) {
	case patterns.SpecName:
		// Check if targetPattern and inputPattern match in type
		ip, ok := inputPatternName.(patterns.SpecName)
		if !ok {
			return "", "", fmt.Errorf("input pattern %q does not match with target pattern %q", ip, tp)
		}

		// Merge the patters together
		mergedPatternName := patterns.SpecName{
			Name: names.Spec{
				ProjectID: projectID,
			},
		}
		mergedApi, err := findCommonPattern(tp.Name.ApiID, ip.Name.ApiID)
		if err != nil {
			return "", "", fmt.Errorf("cannot find common pattern between %q and %q", tp.String(), ip.String())
		}
		mergedPatternName.Name.ApiID = mergedApi
		mergedVersion, err := findCommonPattern(tp.Name.VersionID, ip.Name.VersionID)
		if err != nil {
			return "", "", fmt.Errorf("cannot find common pattern between %q and %q", tp.String(), ip.String())
		}
		mergedPatternName.Name.VersionID = mergedVersion
		mergedSpec, err := findCommonPattern(tp.Name.SpecID, ip.Name.SpecID)
		if err != nil {
			return "", "", fmt.Errorf("cannot find common pattern between %q and %q", tp.String(), ip.String())
		}
		mergedPatternName.Name.SpecID = mergedSpec

		// Merge the filters together
		mergedFilter := generateCommonFilter(targetPattern.GetFilter(), inputFilter)

		return mergedPatternName.String(), mergedFilter, nil
	case patterns.VersionName:
		// Check if targetPattern and inputPattern match in type
		ip, ok := inputPatternName.(patterns.VersionName)
		if !ok {
			return "", "", fmt.Errorf("input pattern %q does not match with target pattern %q", ip, tp)
		}

		// Merge the patters together
		mergedPatternName := patterns.VersionName{
			Name: names.Version{
				ProjectID: projectID,
			},
		}
		mergedApi, err := findCommonPattern(tp.Name.ApiID, ip.Name.ApiID)
		if err != nil {
			return "", "", fmt.Errorf("cannot find common pattern between %q and %q", tp.String(), ip.String())
		}
		mergedPatternName.Name.ApiID = mergedApi
		mergedVersion, err := findCommonPattern(tp.Name.VersionID, ip.Name.VersionID)
		if err != nil {
			return "", "", fmt.Errorf("cannot find common pattern between %q and %q", tp.String(), ip.String())
		}
		mergedPatternName.Name.VersionID = mergedVersion

		// Merge the filters together
		mergedFilter := generateCommonFilter(targetPattern.GetFilter(), inputFilter)

		return mergedPatternName.String(), mergedFilter, nil
	case patterns.ApiName:
		// Check if targetPattern and inputPattern match in type
		ip, ok := inputPatternName.(patterns.ApiName)
		if !ok {
			return "", "", fmt.Errorf("input pattern %q does not match with target pattern %q", ip, tp)
		}

		// Merge the patters together
		mergedPatternName := patterns.ApiName{
			Name: names.Api{
				ProjectID: projectID,
			},
		}
		mergedApi, err := findCommonPattern(tp.Name.ApiID, ip.Name.ApiID)
		if err != nil {
			return "", "", fmt.Errorf("cannot find common pattern between %q and %q", tp.String(), ip.String())
		}
		mergedPatternName.Name.ApiID = mergedApi

		// Merge the filters together
		mergedFilter := generateCommonFilter(targetPattern.GetFilter(), inputFilter)

		return mergedPatternName.String(), mergedFilter, nil
	default:
		return "", "", fmt.Errorf("unsupported pattern in either targetPattern %q or inputPattern %q", targetPatternName.String(), inputPatternName.String())
	}
}

func generateCommonFilter(a, b string) string {
	if a == b {
		return a
	} else if a != "" && b != "" {
		return fmt.Sprintf("(%s) && (%s)", a, b)
	} else if a != "" {
		return a
	}
	return b
}

func findCommonPattern(a, b string) (string, error) {
	if a == "" || b == "" {
		return "", fmt.Errorf("cannot have empty name")
	}
	if a == b {
		return a, nil
	}
	if a == "-" && b != "-" {
		return b, nil
	}
	if a != "-" && b == "-" {
		return a, nil
	}
	return "", fmt.Errorf("cannot find common pattern")
}
