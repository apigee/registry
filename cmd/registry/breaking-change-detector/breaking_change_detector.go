package breakingchangedetector

import (
	"regexp"

	"github.com/apigee/registry/rpc"
)

type detectionPattern struct {
	PositiveMatchRegex *regexp.Regexp
	NegativeMatchRegex *regexp.Regexp
}

var (
	unsafeAdds = []detectionPattern{
		{
			PositiveMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.)+(required)"),
		},
	}

	unsafeDeletes = []detectionPattern{
		{
			PositiveMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.)"),
			NegativeMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.)+(required)"),
		},
		{
			PositiveMatchRegex: regexp.MustCompile("(paths.)+(.|)"),
			NegativeMatchRegex: regexp.MustCompile("((paths.)+(.|)+(tags))+(|)+((paths.)+(.|)+(description))"),
		},
	}

	unsafeMods = []detectionPattern{
		{
			PositiveMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.|)+(type)"),
		},
		{
			PositiveMatchRegex: regexp.MustCompile("(paths.)+(.|)+(type)"),
			NegativeMatchRegex: regexp.MustCompile("((paths.)+(.|)+(tags))+(|)+((paths.)+(.|)+(description))"),
		},
	}

	safeAdds = []detectionPattern{
		{
			PositiveMatchRegex: regexp.MustCompile("(info.)+(.)"),
		},
		{
			PositiveMatchRegex: regexp.MustCompile("(tags.)+(.)"),
		},
	}

	safeDeletes = []detectionPattern{
		{
			PositiveMatchRegex: regexp.MustCompile("(info.)+(.)"),
		},
		{
			PositiveMatchRegex: regexp.MustCompile("(tags.)+(.)"),
		},
	}

	safeMods = []detectionPattern{
		{
			PositiveMatchRegex: regexp.MustCompile("(info.)+(.)"),
		},
		{
			PositiveMatchRegex: regexp.MustCompile("(tags.)+(.)"),
		},
	}
)

// GetBreakingChanges compares each change in a diff Proto to the relavant change type detection Patterns.
// Each change is then catgorized as breaking, nonbreaking, or unknown.
func GetBreakingChanges(diff *rpc.Diff) *rpc.ClassifiedChanges {
	return &rpc.ClassifiedChanges{
		BreakingChanges: getBreakingChanges(diff),
		NonBreakingChanges: getNonBreakingChanges(diff),
		UnknownChanges: getUnknownChanges(diff),
	}
}

func getBreakingChanges(diff *rpc.Diff) *rpc.Diff {
	breakingChanges := &rpc.Diff{
		Additions:     []string{},
		Deletions:     []string{},
		Modifications: map[string]*rpc.Diff_ValueChange{},
	}
	for _, addition := range diff.GetAdditions() {
		if fitsAnyPattern(unsafeAdds, addition) {
			breakingChanges.Additions = append(breakingChanges.Additions, addition)
		}
	}

	for _, deletion := range diff.GetDeletions() {
		if fitsAnyPattern(unsafeDeletes, deletion) {
			breakingChanges.Deletions = append(breakingChanges.Deletions, deletion)
		}
	}

	for modification, modValue := range diff.GetModifications() {
		if fitsAnyPattern(unsafeMods, modification) {
			breakingChanges.Modifications[modification] = modValue
		}
	}
	return breakingChanges
}

func getNonBreakingChanges(diff *rpc.Diff) *rpc.Diff {
	nonBreakingChanges := &rpc.Diff{
		Additions:     []string{},
		Deletions:     []string{},
		Modifications: map[string]*rpc.Diff_ValueChange{},
	}
	for _, addition := range diff.GetAdditions() {
		if fitsAnyPattern(safeAdds, addition) {
			nonBreakingChanges.Additions = append(nonBreakingChanges.Additions, addition)
		}
	}
	for _, deletion := range diff.GetDeletions() {
		if fitsAnyPattern(safeDeletes, deletion) {
			nonBreakingChanges.Deletions = append(nonBreakingChanges.Deletions, deletion)
		}
	}

	for modification, modValue := range diff.GetModifications() {
		if fitsAnyPattern(safeMods, modification) {
			nonBreakingChanges.Modifications[modification] = modValue
		}
	}
	return nonBreakingChanges
}

func getUnknownChanges(diff *rpc.Diff) *rpc.Diff {
	unknownChanges := &rpc.Diff{
		Additions:     []string{},
		Deletions:     []string{},
		Modifications: map[string]*rpc.Diff_ValueChange{},
	}
	for _, addition := range diff.GetAdditions() {
		if !fitsAnyPattern(safeAdds, addition) && !fitsAnyPattern(unsafeAdds, addition) {
			unknownChanges.Additions = append(unknownChanges.Additions, addition)
		}
	}

	for _, deletion := range diff.GetDeletions() {
		if !fitsAnyPattern(safeDeletes, deletion) && !fitsAnyPattern(unsafeDeletes, deletion) {
			unknownChanges.Deletions = append(unknownChanges.Deletions, deletion)
		}
	}

	for modification, modValue := range diff.GetModifications() {
		if !fitsAnyPattern(safeMods, modification) && !fitsAnyPattern(unsafeMods, modification) {
			unknownChanges.Modifications[modification] = modValue
		}
	}
	return unknownChanges
}

func fitsAnyPattern(patterns []detectionPattern, change string) bool {
	for _, pattern := range patterns {
		if p := pattern.NegativeMatchRegex; p != nil && p.MatchString(change) {
			continue
		}
		if p := pattern.PositiveMatchRegex; p != nil && p.MatchString(change) {
			return true
		}
	}
	return false
}
