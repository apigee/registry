package breakingChangeDetector

import (
	"regexp"
	"strconv"

	"github.com/apigee/registry/rpc"
)
type detectionTypeList struct {
	breakingChanges detectionList
	nonBreakingChanges detectionList
}
type detectionList struct {
	additions []detectionPattern
	deletions []detectionPattern
	modifications []detectionPattern
}

type detectionPattern struct {
	detectionWeight int
	PositiveMatchRegex *regexp.Regexp
	NegativeMatchRegex *regexp.Regexp
}

func GetBreakingChanges(diff *rpc.Diff) *rpc.Changes {
	changePatterns := getChangePatterns()
	changesProto, _ := compareChangesToPatterns(changePatterns, diff)
	return changesProto
}

func getChangePatterns()(detectionTypeList){
	breakingChanges := detectionList{
		additions: []detectionPattern{
			{
		detectionWeight: 1,
		PositiveMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.)+(required)"),
			},
		},

		deletions: []detectionPattern{
			{
		detectionWeight: 1,
		PositiveMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.)"),
		NegativeMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.)+(required)"),
			},
		},

		modifications: []detectionPattern{
			{
		detectionWeight: 1,
		PositiveMatchRegex: regexp.MustCompile("(components.)+(.|)+(schemas)+(.|)+(type)"),
					},
		},
	}
	nonBreakingChanges := detectionList{
		additions: []detectionPattern{
			{
		detectionWeight: 1,
		PositiveMatchRegex: regexp.MustCompile("(info.)+(.)"),
			},
		},

		deletions: []detectionPattern{
			{
		detectionWeight: 1,
		PositiveMatchRegex: regexp.MustCompile("(info.)+(.)"),
			},
		},

		modifications: []detectionPattern{
			{
			detectionWeight: 1,
			PositiveMatchRegex: regexp.MustCompile("(info.)+(.)"),
			},
		},
	}
	return detectionTypeList{breakingChanges:breakingChanges, nonBreakingChanges:nonBreakingChanges,}
}

func compareChangesToPatterns(d detectionTypeList, diff *rpc.Diff) (*rpc.Changes, error){
	allChanges := &rpc.Changes{
		BreakingChanges: &rpc.Diff{
			Additions:     []string{},
			Deletions:     []string{},
			Modifications: make(map[string]*rpc.Diff_ValueChange),
		},
		NonBreakingChanges:&rpc.Diff{
			Additions:     []string{},
			Deletions:     []string{},
			Modifications: make(map[string]*rpc.Diff_ValueChange),
		},
		UnknownChanges:&rpc.Diff{
			Additions:     []string{},
			Deletions:     []string{},
			Modifications: make(map[string]*rpc.Diff_ValueChange),
		},
	}
	additions := diff.GetAdditions()
	for _, addition := range additions{
			if fitsAnyPattern(d.breakingChanges.additions, addition){
				allChanges.BreakingChanges.Additions = append(allChanges.BreakingChanges.Additions, addition)
				continue
			}
			if fitsAnyPattern(d.nonBreakingChanges.additions, addition){
				allChanges.NonBreakingChanges.Additions = append(allChanges.NonBreakingChanges.Additions, addition)
				continue
			}
			allChanges.UnknownChanges.Additions = append(allChanges.UnknownChanges.Additions, addition)
	}
	deletions := diff.GetDeletions()
	for _, deletion := range deletions{
		if fitsAnyPattern(d.breakingChanges.deletions, deletion){
			allChanges.BreakingChanges.Deletions = append(allChanges.BreakingChanges.Deletions, deletion)
			continue
		}
		if fitsAnyPattern(d.nonBreakingChanges.deletions, deletion){
			allChanges.NonBreakingChanges.Deletions = append(allChanges.NonBreakingChanges.Deletions, deletion)
			continue
		}
		allChanges.UnknownChanges.Deletions = append(allChanges.UnknownChanges.Deletions, deletion)
	}
	modifications := diff.GetModifications()
	for modification, mod_value := range modifications {
		if fitsAnyPattern(d.breakingChanges.modifications, modification){
			allChanges.BreakingChanges.Modifications[modification] = mod_value
			continue
		}
		if fitsAnyPattern(d.nonBreakingChanges.modifications, modification){
			allChanges.NonBreakingChanges.Modifications[modification] = mod_value
			continue
		}
		if isNonStringValue(mod_value.To) || isNonStringValue(mod_value.From){
			if !isSameType(mod_value.To, mod_value.From){
				allChanges.BreakingChanges.Modifications[modification] = mod_value
				continue
			}
	}
	allChanges.UnknownChanges.Modifications[modification] = mod_value
	}
	return allChanges, nil
}

func isSameType(valueOne, valueTwo string) (bool){
	if valueOne == valueTwo{
		return true
	}
	if _, err := strconv.ParseInt(valueOne,10, 64); err == nil{
		if _, err := strconv.ParseInt(valueTwo,10, 64); err == nil{
			return true
		}
	}
	if _, err := strconv.ParseFloat(valueOne, 64); err == nil{
		if _, err := strconv.ParseFloat(valueTwo, 64); err == nil{
			return  true
		}
	}
	if _, err := strconv.ParseBool(valueOne); err == nil{
		if _, err := strconv.ParseBool(valueOne); err == nil{
			return true
		}
	}
	return  false
}

func isNonStringValue(value string) (bool){
	if _, err := strconv.ParseFloat(value, 64); err == nil{
		return true
		}
	if _, err := strconv.ParseInt(value,10, 64); err == nil{
		return true
		}
	if _, err := strconv.ParseBool(value); err == nil{
		return true
		}
	return false
}

func fitsAnyPattern(patterns []detectionPattern, change string)(bool){
	for _, pattern := range patterns{
		if pattern.PositiveMatchRegex == nil {
			if !pattern.NegativeMatchRegex.MatchString(change){
				return true
			}
		}
		if pattern.NegativeMatchRegex == nil {
			if pattern.PositiveMatchRegex.MatchString(change){
				return true
			}
		}
		if pattern.NegativeMatchRegex.MatchString(change){
			continue
		}
		if pattern.PositiveMatchRegex.MatchString(change){
			return true
		}
	}
	return false
}
