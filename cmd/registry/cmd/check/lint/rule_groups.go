package lint

import "fmt"

// A list of functions, each of which returns the group name for the given rule
// number and if no group is found, returns an empty string.
// List will be evaluated in the FILO order.
// TODO: complete rule groups
var ruleGroup = []func(int) string{
	registryGroup,
	hubGroup,
	scoreGroup,
}

func registryGroup(ruleNum int) string {
	if ruleNum > 0 && ruleNum < 1000 {
		return "registry"
	}
	return ""
}

func hubGroup(ruleNum int) string {
	if ruleNum >= 1000 && ruleNum < 1100 {
		return "apihub"
	}
	return ""
}

func scoreGroup(ruleNum int) string {
	if ruleNum >= 1100 && ruleNum < 1200 {
		return "score"
	}
	return ""
}

// getRuleGroup takes an rule number and returns the appropriate group.
// It panics if no group is found.
func getRuleGroup(ruleNum int, groups []func(int) string) string {
	for i := len(groups) - 1; i >= 0; i-- {
		if group := groups[i](ruleNum); group != "" {
			return group
		}
	}
	panic(fmt.Sprintf("Invalid rule number %d: no available group.", ruleNum))
}
