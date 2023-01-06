package lint

import (
	"testing"
)

func TestGroups(t *testing.T) {
	tests := []struct {
		name   string
		ruleID int
		group  string
	}{
		{"registry", 1, "registry"},
		{"registry", 999, "registry"},
		{"hub", 1000, "hub"},
		{"hub", 1099, "hub"},
		{"controller", 1100, "controller"},
		{"controller", 1199, "controller"},
		{"score", 1200, "score"},
		{"score", 1299, "score"},
		{"style", 1300, "style"},
		{"style", 1399, "style"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := getRuleGroup(test.ruleID, ruleGroup); got != test.group {
				t.Errorf("ruleID(%d) got %s, but want %s", test.ruleID, got, test.group)
			}
		})
	}
}

func TestGetRuleGroupPanic(t *testing.T) {
	var groups []func(int) string
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("getRuleGroup did not panic")
		}
	}()
	getRuleGroup(0, groups)
}
