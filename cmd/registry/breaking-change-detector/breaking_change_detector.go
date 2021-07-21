package breakingchangedetector

import (
	"regexp"
	"strconv"

	"github.com/apigee/registry/rpc"
)

type detectionTypeList struct {
	breakingChanges    detectionList
	nonBreakingChanges detectionList
}
type detectionList struct {
	additions     []detectionPattern
	deletions     []detectionPattern
	modifications []detectionPattern
}

type detectionPattern struct {
	detectionWeight    int
	PositiveMatchRegex *regexp.Regexp
	NegativeMatchRegex *regexp.Regexp
}

// GetBreakingChanges compares each change in a diff Proto to the relavant change type detection Patterns.
// Each change is then catgorized as breaking, nonbreaking, or unknown.
func GetBreakingChanges(diff *rpc.Diff) *rpc.Changes {
	changePatterns := getChangePatterns()
	changesProto := compareChangesToPatterns(changePatterns, diff)
	return changesProto
}

// getChangePatterns Intalizes the Breaking and NonBreaking change Patterns.
func getChangePatterns() detectionTypeList {
	breakingChanges := detectionList{
		additions: []detectionPattern{
			{
				detectionWeight:    1,
				PositiveMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.)+(required)"),
			},
		},

		deletions: []detectionPattern{
			{
				detectionWeight:    1,
				PositiveMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.)"),
				NegativeMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.)+(required)"),
			},
		},

		modifications: []detectionPattern{
			{
				detectionWeight:    1,
				PositiveMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.|)+(type)"),
			},
		},
	}
	nonBreakingChanges := detectionList{
		additions: []detectionPattern{
			{
				detectionWeight:    1,
				PositiveMatchRegex: regexp.MustCompile("(info.)+(.)"),
			},
		},

		deletions: []detectionPattern{
			{
				detectionWeight:    1,
				PositiveMatchRegex: regexp.MustCompile("(info.)+(.)"),
			},
		},

		modifications: []detectionPattern{
			{
				detectionWeight:    1,
				PositiveMatchRegex: regexp.MustCompile("(info.)+(.)"),
			},
		},
	}
	return detectionTypeList{breakingChanges: breakingChanges, nonBreakingChanges: nonBreakingChanges}
}

func compareChangesToPatterns(d detectionTypeList, diff *rpc.Diff) *rpc.Changes {
	allChanges := &rpc.Changes{
		BreakingChanges: &rpc.Diff{
			Additions:     []string{},
			Deletions:     []string{},
			Modifications: make(map[string]*rpc.Diff_ValueChange),
		},
		NonBreakingChanges: &rpc.Diff{
			Additions:     []string{},
			Deletions:     []string{},
			Modifications: make(map[string]*rpc.Diff_ValueChange),
		},
		UnknownChanges: &rpc.Diff{
			Additions:     []string{},
			Deletions:     []string{},
			Modifications: make(map[string]*rpc.Diff_ValueChange),
		},
	}
	updateChangeStatusAdditions(allChanges, d, diff)
	updateChangeStatusDeletions(allChanges, d, diff)
	updateChangeStatusModifications(allChanges, d, diff)

	return allChanges
}

func updateChangeStatusAdditions(allChanges *rpc.Changes, d detectionTypeList, diff *rpc.Diff) {
	additions := diff.GetAdditions()
	for _, addition := range additions {
		if fitsAnyPattern(d.breakingChanges.additions, addition) {
			allChanges.BreakingChanges.Additions = append(allChanges.BreakingChanges.Additions, addition)
			continue
		}
		if fitsAnyPattern(d.nonBreakingChanges.additions, addition) {
			allChanges.NonBreakingChanges.Additions = append(allChanges.NonBreakingChanges.Additions, addition)
			continue
		}
		allChanges.UnknownChanges.Additions = append(allChanges.UnknownChanges.Additions, addition)
	}
}

func updateChangeStatusDeletions(allChanges *rpc.Changes, d detectionTypeList, diff *rpc.Diff) {
	deletions := diff.GetDeletions()
	for _, deletion := range deletions {
		if fitsAnyPattern(d.breakingChanges.deletions, deletion) {
			allChanges.BreakingChanges.Deletions = append(allChanges.BreakingChanges.Deletions, deletion)
			continue
		}
		if fitsAnyPattern(d.nonBreakingChanges.deletions, deletion) {
			allChanges.NonBreakingChanges.Deletions = append(allChanges.NonBreakingChanges.Deletions, deletion)
			continue
		}
		allChanges.UnknownChanges.Deletions = append(allChanges.UnknownChanges.Deletions, deletion)
	}
}

func updateChangeStatusModifications(allChanges *rpc.Changes, d detectionTypeList, diff *rpc.Diff) {
	modifications := diff.GetModifications()
	for modification, modValue := range modifications {
		if fitsAnyPattern(d.breakingChanges.modifications, modification) {
			allChanges.BreakingChanges.Modifications[modification] = modValue
			continue
		}
		if fitsAnyPattern(d.nonBreakingChanges.modifications, modification) {
			allChanges.NonBreakingChanges.Modifications[modification] = modValue
			continue
		}
		if isNonStringValue(modValue.To) || isNonStringValue(modValue.From) {
			if !isSameType(modValue.To, modValue.From) {
				allChanges.BreakingChanges.Modifications[modification] = modValue
				continue
			}
		}
		allChanges.UnknownChanges.Modifications[modification] = modValue
	}
}

func isSameType(valueOne, valueTwo string) bool {
	if valueOne == valueTwo {
		return true
	}
	if _, err := strconv.ParseInt(valueOne, 10, 64); err == nil {
		if _, err := strconv.ParseInt(valueTwo, 10, 64); err == nil {
			return true
		}
	}
	if _, err := strconv.ParseFloat(valueOne, 64); err == nil {
		if _, err := strconv.ParseFloat(valueTwo, 64); err == nil {
			return true
		}
	}
	if _, err := strconv.ParseBool(valueOne); err == nil {
		if _, err := strconv.ParseBool(valueOne); err == nil {
			return true
		}
	}
	return false
}

func isNonStringValue(value string) bool {
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return true
	}
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return true
	}
	if _, err := strconv.ParseBool(value); err == nil {
		return true
	}
	return false
}

func fitsAnyPattern(patterns []detectionPattern, change string) bool {
	for _, pattern := range patterns {
		if pattern.PositiveMatchRegex == nil {
			if !pattern.NegativeMatchRegex.MatchString(change) {
				return true
			}
		}
		if pattern.NegativeMatchRegex == nil {
			if pattern.PositiveMatchRegex.MatchString(change) {
				return true
			}
		}
		if pattern.NegativeMatchRegex.MatchString(change) {
			continue
		}
		if pattern.PositiveMatchRegex.MatchString(change) {
			return true
		}
	}
	return false
}
