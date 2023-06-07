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

// ApiDeployment api_spec_revision must be an SpecRevision and sibling under the parent Api.
package rule106

import (
	"context"
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 106
var ruleName = lint.NewRuleName(ruleNum, "deployment-api-spec-revision-ref")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		apiSpecRevisionRef,
	)
}

var apiSpecRevisionRef = &lint.ApiDeploymentRule{
	Name: ruleName,
	OnlyIf: func(a *rpc.ApiDeployment) bool {
		return strings.TrimSpace(a.ApiSpecRevision) != ""
	},
	ApplyToApiDeployment: func(ctx context.Context, a *rpc.ApiDeployment) []*check.Problem {
		specName, err := names.ParseSpecRevision(a.ApiSpecRevision)
		if err != nil {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`api_spec_revision %q is not a valid ApiSpecRevision name.`, a.ApiSpecRevision),
				Suggestion: fmt.Sprintf(`Parse error: %s`, err),
			}}
		}
		if specName.RevisionID == "" {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`api_spec_revision %q is not a valid ApiSpecRevision name.`, a.ApiSpecRevision),
				Suggestion: `A revision ID is required.`,
			}}
		}

		deploymentName, _ := names.ParseDeployment(a.Name) // name assumed to be valid
		if specName.Api() != deploymentName.Api() {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`api_spec_revision %q is not an API sibling of this Deployment.`, a.ApiSpecRevision),
				Suggestion: `Correct the api_spec_revision.`,
			}}
		}

		registryClient := lint.RegistryClient(ctx)
		if _, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
			Name: a.ApiSpecRevision,
		}); err != nil {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`api_spec_revision %q not found in registry.`, a.ApiSpecRevision),
				Suggestion: `Correct the api_spec_revision.`,
			}}
		}

		return nil
	},
}
