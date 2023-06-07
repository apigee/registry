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

package rule109

import (
	"context"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/cmd/registry/cmd/check/rules/util"
	"github.com/apigee/registry/pkg/application/check"
)

var ruleNum = 109
var fieldName = "display_name"
var ruleName = lint.NewRuleName(ruleNum, "display-name-format")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		displayName,
	)
}

// max length of 64, all characters must use UTF-8 encoding
var displayName = &lint.FieldRule{
	Name: ruleName,
	OnlyIf: func(resource lint.Resource, field string) bool {
		return field == fieldName
	},
	ApplyToField: func(ctx context.Context, resource lint.Resource, field string, value interface{}) []*check.Problem {
		return util.CheckUTF(fieldName, value, 64)
	},
}
