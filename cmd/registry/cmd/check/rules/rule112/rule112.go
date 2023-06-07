// Copyright 2023 Google LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Each resource can have multiple labels, up to a maximum of 64.
// Each label must be a key-value pair.
// Keys have a minimum length of 1 character and a maximum length of 63 characters, and cannot be empty.
// Values can be empty, and have a maximum length of 63 characters.
// Keys and values can contain only lowercase letters, numeric characters, underscores, and dashes.
// All characters must use UTF-8 encoding, and international characters are allowed.
// The key portion of a label must be unique within a single resource. However,
// you can use the same key with multiple resources.
// Keys must start with a lowercase letter or international character.
package rule112

import (
	"context"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
)

const ruleNum = 112
const fieldName = "labels"

var ruleName = lint.NewRuleName(ruleNum, "labels-format")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		labels,
	)
}

var labels = &lint.FieldRule{
	Name: ruleName,
	OnlyIf: func(resource lint.Resource, field string) bool {
		return field == fieldName
	},
	ApplyToField: func(ctx context.Context, resource lint.Resource, field string, value interface{}) []*check.Problem {
		labels := value.(map[string]string)
		if len(labels) == 0 {
			return nil
		}
		var probs []*check.Problem
		for k, v := range labels {
			probs = append(probs, checkLabel(k, v)...)
		}
		if len(labels) > 64 {
			probs = append(probs, &check.Problem{
				Severity:   check.Problem_ERROR,
				Message:    `Maximum number of labels is 64.`,
				Suggestion: `Delete some entries.`,
			})
		}

		return probs
	},
}

func checkLabel(k string, v string) []*check.Problem {
	var probs []*check.Problem
	if r, _ := utf8.DecodeRuneInString(k); r == utf8.RuneError || !unicode.In(r, unicode.Ll, unicode.Lo) {
		probs = append(probs, &check.Problem{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf(`Key %q has illegal first character %q.`, k, r),
			Suggestion: `Fix key.`,
		})
	} else if ok, r := validLabelRunes(k); !ok {
		probs = append(probs, &check.Problem{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf(`Key %q contains illegal character %q.`, k, r),
			Suggestion: `Fix key.`,
		})
	}
	if count := utf8.RuneCountInString(k); count > 63 {
		probs = append(probs, &check.Problem{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf(`Key %q exceeds max length of 63 characters.`, k),
			Suggestion: `Fix key.`,
		})
	}
	if ok, r := validLabelRunes(v); !ok {
		probs = append(probs, &check.Problem{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf(`Value for key %q contains illegal character %q.`, k, r),
			Suggestion: `Fix value.`,
		})
	}
	if count := utf8.RuneCountInString(v); count > 63 {
		probs = append(probs, &check.Problem{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf(`Value for key %q exceeds max length of 63 characters.`, k),
			Suggestion: `Fix value.`,
		})
	}
	return probs
}

func validLabelRunes(s string) (bool, rune) {
	for _, r := range s {
		if !unicode.In(r, unicode.Ll, unicode.Lo, unicode.N) && r != '_' && r != '-' {
			return false, r
		}
	}
	return true, 0
}
