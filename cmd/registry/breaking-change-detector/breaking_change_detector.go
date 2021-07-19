package breakingChangeDetector

import (
	"regexp"

	"github.com/apigee/registry/rpc"
)

type detectionList struct {
	additions []detectionPattern
	deletions []detectionPattern
	modifications []detectionPattern
}

type detectionPattern struct {
	detectionWeight int
	detectionPatternPositive string
	detectionPatternNegative string
}

func getChangePatterns()(detectionList, detectionList){

	breakingChanges := detectionList{

		additions: []detectionPattern{
			{
		detectionWeight: 1,
		detectionPatternPositive: "(components.)+(.|)+(schemas)+(.)+(required)",
		detectionPatternNegative: "",
			},
		},

		deletions: []detectionPattern{
			{
		detectionWeight: 1,
		detectionPatternPositive: "(components.)+(.|)+(schemas)+(.)",
		detectionPatternNegative: "(components.)+(.|)+(schemas)+(.)+(required)",
			},
		},
		modifications: []detectionPattern{},

	}

	nonBreakingChanges := detectionList{

		additions: []detectionPattern{
			{
		detectionWeight: 1,
		detectionPatternPositive: "",
		detectionPatternNegative: "",
			},
		},

		deletions: []detectionPattern{
			{
		detectionWeight: 1,
		detectionPatternPositive: "",
		detectionPatternNegative: "",
			},
		},
		modifications: []detectionPattern{},
	}

	return breakingChanges, nonBreakingChanges
}

func GetBreakingChanges(diff *rpc.Diff) *rpc.Changes {

	breakingChanges, nonBreakingChanges := getChangePatterns()
	breakingChangesProto := compareChangesToPatterns(breakingChanges, diff)
	nonBreakingChangesProto := compareChangesToPatterns(nonBreakingChanges, diff)

	changesProto := &rpc.Changes{
		BreakingChanges: breakingChangesProto,
		NonbreakingChanges: nonBreakingChangesProto,
	}

	return changesProto
}

func compareChangesToPatterns(d detectionList, diff *rpc.Diff) *rpc.Diff{
	changesProto := &rpc.Diff{
		Additions:     []string{},
		Deletions:     []string{},
		Modifications: make(map[string]*rpc.Diff_ValueChange),
	}

	additions := diff.GetAdditions()
	for _, addition := range additions{
			if fitsPatterns(d.additions, addition){
				changesProto.Additions = append(changesProto.Additions, addition)
			}
	}

	deletions := diff.GetDeletions()
	for _, deletion := range deletions{
			if fitsPatterns(d.deletions, deletion){
				changesProto.Deletions = append(changesProto.Deletions, deletion)
			}
	}

	return changesProto
}

func fitsPatterns(patterns []detectionPattern, change string)(bool){
	for _, pattern := range patterns{
		re := regexp.MustCompile(pattern.detectionPatternPositive)
		if match := re.FindString(change); match == "" {
			continue
		}
		re = regexp.MustCompile(pattern.detectionPatternNegative)
		if match := re.FindString(change); match == "" {
			return true
		}
	}
	return false
}
