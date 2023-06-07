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

// Api availability field is free-form, but we expect single words that describe availability.
// https://github.com/apigee/registry/blob/main/google/cloud/apigeeregistry/v1/registry_models.proto#L51
package rule100

import (
	"context"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 100
var ruleName = lint.NewRuleName(ruleNum, "api-availability-single-word")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		availabilitySingleWord,
	)
}

var availabilitySingleWord = &lint.ApiRule{
	Name: ruleName,
	OnlyIf: func(a *rpc.Api) bool {
		return true
	},
	ApplyToApi: func(ctx context.Context, a *rpc.Api) []*check.Problem {
		if arr := strings.SplitN(a.Availability, " ", 2); len(arr) > 1 {
			return []*check.Problem{{
				Severity:   check.Problem_INFO,
				Message:    `Availability is free-form, but we expect single words that describe availability.`,
				Suggestion: `Use single words like: "NONE", "TESTING", "PREVIEW", "GENERAL", "DEPRECATED", "SHUTDOWN"`,
			}}
		}
		return nil
	},
}
