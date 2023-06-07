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

// Api state field is free-form, but we expect single words that describe state.
// https://github.com/apigee/registry/blob/main/google/cloud/apigeeregistry/v1/registry_models.proto#L113
package rule103

import (
	"context"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 103
var ruleName = lint.NewRuleName(ruleNum, "version-state-single-word")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		stateSingleWord,
	)
}

var stateSingleWord = &lint.ApiVersionRule{
	Name: ruleName,
	OnlyIf: func(a *rpc.ApiVersion) bool {
		return true
	},
	ApplyToApiVersion: func(ctx context.Context, a *rpc.ApiVersion) []*check.Problem {
		if arr := strings.SplitN(a.State, " ", 2); len(arr) > 1 {
			return []*check.Problem{{
				Severity:   check.Problem_INFO,
				Message:    `State is free-form, but we expect single words that describe API maturity.`,
				Suggestion: `Use single words like: "CONCEPT", "DESIGN", "DEVELOPMENT", "STAGING", "PRODUCTION", "DEPRECATED", "RETIRED"`,
			}}
		}
		return nil
	},
}
