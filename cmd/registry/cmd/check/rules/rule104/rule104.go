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

// Api recommended_version must be an ApiVersion that is a child of this Api.
package rule104

import (
	"context"
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 104
var ruleName = lint.NewRuleName(ruleNum, "version-primary-spec-ref")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		primarySpecRef,
	)
}

var primarySpecRef = &lint.ApiVersionRule{
	Name: ruleName,
	OnlyIf: func(a *rpc.ApiVersion) bool {
		return strings.TrimSpace(a.PrimarySpec) != ""
	},
	ApplyToApiVersion: func(ctx context.Context, a *rpc.ApiVersion) []*check.Problem {
		specName, err := names.ParseSpec(a.PrimarySpec)
		if err != nil {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`primary_spec %q is not a valid ApiSpec name.`, a.PrimarySpec),
				Suggestion: fmt.Sprintf(`Parse error: %s`, err),
			}}
		}

		versionName, _ := names.ParseVersion(a.Name) // name assumed to be valid
		if specName.Api() != versionName.Api() {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`primary_spec %q is not an API sibling of this Version.`, a.PrimarySpec),
				Suggestion: `Correct the primary_spec.`,
			}}
		}

		registryClient := lint.RegistryClient(ctx)
		if _, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
			Name: a.PrimarySpec,
		}); err != nil {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`primary_spec %q not found in registry.`, a.PrimarySpec),
				Suggestion: `Correct the primary_spec.`,
			}}
		}

		return nil
	},
}
