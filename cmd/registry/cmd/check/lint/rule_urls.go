package lint

import "strings"

// TODO: determine base URL
const baseURL = "https://github.com/apigee/registry/wiki/"

// A list of mapping functions, each of which returns the rule URL for
// the given rule name, and if not found, return an empty string.
// TODO: complete rule mappings
var ruleURLMappings = []func(string) string{
	coreRuleURL,
	hubRuleUrl,
}

func coreRuleURL(ruleName string) string {
	return groupUrl(ruleName, "core")
}

func hubRuleUrl(ruleName string) string {
	return groupUrl(ruleName, "hub")
}

func groupUrl(ruleName, groupName string) string {
	nameParts := strings.Split(ruleName, "::") // e.g., registry::0122::camel-case-uris -> ["registry", "0122", "camel-case-uris"]
	if len(nameParts) == 0 || nameParts[0] != groupName {
		return ""
	}
	path := strings.TrimPrefix(strings.Join(nameParts[1:], "/"), "0")
	return baseURL + path
}

func getRuleURL(ruleName string, nameURLMappings []func(string) string) string {
	for i := len(nameURLMappings) - 1; i >= 0; i-- {
		if url := nameURLMappings[i](ruleName); url != "" {
			return url
		}
	}
	return ""
}
