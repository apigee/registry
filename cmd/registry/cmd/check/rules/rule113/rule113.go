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

// Each resource can have multiple annotations, up to a maximum total size of 256k.
// Each annotation must be a key-value pair.
// Keys have a minimum length of 1 character and a maximum length of 63 characters, and cannot be empty.
// Values can be empty, and have no maximum length (other than total size limit).
// Keys can contain only lowercase letters, numeric characters, underscores, and dashes.
// All characters must use UTF-8 encoding, and international characters are allowed.
// Values can contain any characters.
// The key portion of a label must be unique within a single resource. However,
// you can use the same key with multiple resources.
// Keys must start with a lowercase letter or international character.
package rule113

import (
	"context"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
)

const ruleNum = 113
const fieldName = "annotations"

const totalSizeLimit int = 256 * (1 << 10) // 256 kB

var ruleName = lint.NewRuleName(ruleNum, "annotations-format")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		annotations,
	)
}

var annotations = &lint.FieldRule{
	Name: ruleName,
	OnlyIf: func(resource lint.Resource, field string) bool {
		return field == fieldName
	},
	ApplyToField: func(ctx context.Context, resource lint.Resource, field string, value interface{}) []*check.Problem {
		labels := value.(map[string]string)
		if len(labels) == 0 {
			return nil
		}
		totalSize := 0
		var probs []*check.Problem
		for k, v := range labels {
			probs = append(probs, checkAnnotation(k, v)...)
			totalSize += len(k) + len(v)
		}
		if totalSize > totalSizeLimit {
			probs = append(probs, &check.Problem{
				Severity:   check.Problem_ERROR,
				Message:    `Maximum size of all annotations is 256k.`,
				Suggestion: fmt.Sprintf(`Reduce size by %d bytes.`, totalSize-totalSizeLimit),
			})
		}

		return probs
	},
}

func checkAnnotation(k string, v string) []*check.Problem {
	var probs []*check.Problem
	if r, _ := utf8.DecodeRuneInString(k); r == utf8.RuneError || !unicode.In(r, unicode.Ll, unicode.Lo) {
		probs = append(probs, &check.Problem{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf(`Key %q has illegal first character %q.`, k, r),
			Suggestion: `Fix key.`,
		})
	} else if ok, r := validKeyRunes(k); !ok {
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
	return probs
}

func validKeyRunes(s string) (bool, rune) {
	for _, r := range s {
		if !unicode.In(r, unicode.Ll, unicode.Lo, unicode.N) && r != '_' && r != '-' {
			return false, r
		}
	}
	return true, 0
}
