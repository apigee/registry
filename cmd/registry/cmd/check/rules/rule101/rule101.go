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
package rule101

import (
	"context"
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 101
var ruleName = lint.NewRuleName(ruleNum, "api-recommended-version-ref")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		recommendedVersionRef,
	)
}

var recommendedVersionRef = &lint.ApiRule{
	Name: ruleName,
	OnlyIf: func(a *rpc.Api) bool {
		return strings.TrimSpace(a.RecommendedVersion) != ""
	},
	ApplyToApi: func(ctx context.Context, a *rpc.Api) []*check.Problem {
		versionName, err := names.ParseVersion(a.RecommendedVersion)
		if err != nil {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`recommended_version %q is not a valid ApiVersion name.`, a.RecommendedVersion),
				Suggestion: fmt.Sprintf(`Parse error: %s`, err),
			}}
		}

		apiName, _ := names.ParseApi(a.Name) // name assumed to be valid
		if versionName.Api() != apiName {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`recommended_version %q is not a child of this Api.`, a.RecommendedVersion),
				Suggestion: `Correct the recommended_version.`,
			}}
		}

		registryClient := lint.RegistryClient(ctx)
		if _, err := registryClient.GetApiVersion(ctx, &rpc.GetApiVersionRequest{
			Name: a.RecommendedVersion,
		}); err != nil {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`recommended_version %q not found in registry.`, a.RecommendedVersion),
				Suggestion: `Correct the recommended_version.`,
			}}
		}

		return nil
	},
}
