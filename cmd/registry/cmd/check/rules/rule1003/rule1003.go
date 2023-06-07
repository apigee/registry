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

// APIVersion state must match a valid lifecycle stage
package rule1003

import (
	"context"
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/application/apihub"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 1003
var ruleName = lint.NewRuleName(ruleNum, "version-state-lifecycle-stage")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		stateIsValidLifecycleStage,
	)
}

var stateIsValidLifecycleStage = &lint.ApiVersionRule{
	Name: ruleName,
	OnlyIf: func(a *rpc.ApiVersion) bool {
		return true
	},
	ApplyToApiVersion: func(ctx context.Context, a *rpc.ApiVersion) []*check.Problem {
		client := lint.RegistryClient(ctx)
		if a.State == "" {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    `APIVersion state is empty but must match a valid api-lifecycle stage.`,
				Suggestion: `Assign a appropriate value to the APIVersion state.`,
			}}
		}
		name, _ := names.Parse(a.GetName())
		project := name.Project()
		var lc *apihub.Lifecycle
		ac, err := client.GetArtifactContents(ctx, &rpc.GetArtifactContentsRequest{
			Name: project.Artifact("apihub-lifecycle").String(),
		})
		if err == nil {
			lc = new(apihub.Lifecycle)
			err = patch.UnmarshalContents(ac.Data, ac.ContentType, lc)
		}
		if err != nil {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    `Failed to check APIVersion state against api-lifecycle stage.`,
				Suggestion: fmt.Sprintf(`Unable to access api-lifecycle artifact: %v`, err),
			}}
		}
		ids := []string{}
		for _, s := range lc.Stages {
			if a.State == s.Id {
				return nil
			}
			ids = append(ids, s.Id)
		}
		fmt.Println(strings.Join(ids, ", "))
		return []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `APIVersion must match a valid api-lifecycle stage.`,
			Suggestion: fmt.Sprintf(`APIVersion state: %q, valid stages: %s.`, a.State, strings.Join(ids, ", ")),
		}}
	},
}
