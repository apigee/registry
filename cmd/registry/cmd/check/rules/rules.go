// Copyright 2023 Google LLC
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

// Package rules contains implementations of checker rules.
//
// Rules are sorted into groups. Every rule represented in code here must
// be documented in English in a corresponding registry rule. Conversely,
// anything mandated in such a rule should have a rule here if it is feasible
// to enforce in code (sometimes it is infeasible, however).
//
// A rule is technically anything with a `GetName()` and `Apply(Resource)`
// but most rule authors will want to use the rule structs provided in the
// lint package (`lint.ProjectRule`, lint.FieldRule`, and so on). These run
// against each applicable registry resource type or field. They also have
// an `OnlyIf` property that can be used to run against a subset of resources.
//
// Once a rule is written, it needs to be registered. This involves adding
// the rule to the `AddRules` method for the appropriate group.
// If this is the first rule for a new package, then the `rules.go` init() function
// must also be updated to run the `Add` function for the new package.
package rules

import (
	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule100"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule1000"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule1001"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule1002"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule1003"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule101"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule102"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule103"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule104"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule105"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule106"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule107"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule108"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule109"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule110"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule111"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule112"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/rule113"
)

type addRulesFuncType func(lint.RuleRegistry) error

var addRulesFuncs = []addRulesFuncType{
	rule100.AddRules,
	rule101.AddRules,
	rule102.AddRules,
	rule103.AddRules,
	rule104.AddRules,
	rule105.AddRules,
	rule106.AddRules,
	rule107.AddRules,
	rule108.AddRules,
	rule109.AddRules,
	rule110.AddRules,
	rule111.AddRules,
	rule112.AddRules,
	rule113.AddRules,
	rule1000.AddRules,
	rule1001.AddRules,
	rule1002.AddRules,
	rule1003.AddRules,
}

// Add all rules to the given registry.
func Add(r lint.RuleRegistry) error {
	return addRules(r, addRulesFuncs)
}

func addRules(r lint.RuleRegistry, addRulesFuncs []addRulesFuncType) error {
	for _, addRules := range addRulesFuncs {
		if err := addRules(r); err != nil {
			return err
		}
	}
	return nil
}
