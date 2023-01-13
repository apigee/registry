// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lint

import (
	"errors"
	"fmt"
)

// RuleRegistry is a registry for registering and looking up rules.
type RuleRegistry map[RuleName]Rule

var (
	errInvalidRuleName    = errors.New("not a valid rule name")
	errInvalidRuleGroup   = errors.New("invalid rule group")
	errDuplicatedRuleName = errors.New("duplicate rule name")
)

// NewRuleRegistry creates a new rule registry.
func NewRuleRegistry() RuleRegistry {
	return make(RuleRegistry)
}

// Register registers the list of rules of the same ruleNum.
// Return an error if any of the rules is found duplicate in the registry.
func (r RuleRegistry) Register(ruleNum int, rules ...Rule) error {
	rulePrefix := getRuleGroup(ruleNum, ruleGroup) + nameSeparator + fmt.Sprintf("%04d", ruleNum)
	for _, rl := range rules {
		if !rl.GetName().IsValid() {
			return errInvalidRuleName
		}

		if !rl.GetName().HasPrefix(rulePrefix) {
			return errInvalidRuleGroup
		}

		if _, found := r[rl.GetName()]; found {
			return errDuplicatedRuleName
		}

		r[rl.GetName()] = rl
	}
	return nil
}
