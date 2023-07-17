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

package rule1000

import (
	"context"
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/application/apihub"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 1000
var ruleName = lint.NewRuleName(ruleNum, "required-artifacts")
var requiredIDs = []string{"apihub-lifecycle", "apihub-taxonomies"}
var requiredTaxonomies = []string{"apihub-target-users", "apihub-style", "apihub-team", "apihub-business-unit", "apihub-gateway"}

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		requiredArtifacts,
	)
}

var requiredArtifacts = &lint.ProjectRule{
	Name: ruleName,
	OnlyIf: func(a *rpc.Project) bool {
		return true
	},
	ApplyToProject: func(ctx context.Context, a *rpc.Project) []*check.Problem {
		client := lint.RegistryClient(ctx)

		var filter string
		found := make(map[string]bool, len(requiredIDs))
		for _, n := range requiredIDs {
			if len(filter) > 0 {
				filter += " || "
			}
			filter += fmt.Sprintf("artifact_id == '%s'", n)
		}
		projectName, _ := names.ParseProject(a.GetName())
		var probs []*check.Problem
		if err := visitor.ListArtifacts(ctx, client, projectName.Artifact("-"), 0, filter, true, func(ctx context.Context, a *rpc.Artifact) error {
			found[a.GetName()] = true

			name, _ := names.ParseArtifact(a.GetName())
			if name.ArtifactID() == "apihub-taxonomies" {
				probs = append(probs, checkTaxonomies(a)...)
			}

			return nil
		}); err != nil {
			panic(err)
		}

		for _, id := range requiredIDs {
			fn := projectName.Artifact(id).String()
			if !found[fn] {
				probs = append(probs, &check.Problem{
					Severity:   check.Problem_ERROR,
					Message:    fmt.Sprintf(`artifact %q not found in registry.`, fn),
					Suggestion: `Initialize API Hub.`,
				})
			}
		}

		return probs
	},
}

func checkTaxonomies(a *rpc.Artifact) []*check.Problem {
	message, err := mime.MessageForMimeType(a.GetMimeType())
	if err == nil {
		err = patch.UnmarshalContents(a.GetContents(), a.GetMimeType(), message)
	}
	if err != nil {
		return []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf(`Unable to verify %s`, a.GetName()),
			Suggestion: fmt.Sprintf(`Error: %s`, err),
		}}
	}
	tl := message.(*apihub.TaxonomyList)
	tm := make(map[string]bool)
	for _, t := range tl.Taxonomies {
		tm[t.Id] = true
	}
	var missing []string
	for _, t := range requiredTaxonomies {
		if !tm[t] {
			missing = append(missing, t)
		}
	}
	if len(missing) > 0 {
		return []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf(`TaxonomyList %q must include items: %s.`, a.GetName(), strings.Join(missing, `, `)),
			Suggestion: `Initialize API Hub.`,
		}}
	}
	return nil
}
