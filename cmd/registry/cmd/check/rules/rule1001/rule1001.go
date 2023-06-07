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

// if label name is the same as a Taxonomy name, its value must match
// one of the Taxonomy's Element's ID
package rule1001

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

var ruleNum = 1001
var ruleName = lint.NewRuleName(ruleNum, "taxonomy-labels")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		taxonomyLabels,
	)
}

var taxonomyLabels = &lint.FieldRule{
	Name: ruleName,
	OnlyIf: func(resource lint.Resource, field string) bool {
		return field == "Labels"
	},
	ApplyToField: func(ctx context.Context, resource lint.Resource, field string, value interface{}) []*check.Problem {
		labels := map[string]string{}
		for k, v := range value.(map[string]string) {
			if strings.HasPrefix(k, "apihub-") {
				labels[k] = v
			}
		}
		if len(labels) == 0 {
			return nil
		}

		name, _ := names.Parse(resource.GetName())
		project := name.Project()
		taxonomies := taxonomies(ctx, project)

		var probs []*check.Problem
		for k, v := range labels {
			if !taxonomies.exists(k, v) {
				probs = append(probs, &check.Problem{
					Severity:   check.Problem_ERROR,
					Message:    fmt.Sprintf(`Label value %q not present in Taxonomy %q`, v, k),
					Suggestion: `Adjust label value or Taxonomy elements.`,
				})
			}
		}

		return probs
	},
}

// TODO: cache?
func taxonomies(ctx context.Context, project names.Project) *TaxList {
	client := lint.RegistryClient(ctx)
	ac, err := client.GetArtifactContents(ctx, &rpc.GetArtifactContentsRequest{
		Name: project.Artifact("apihub-taxonomies").String(),
	})
	if err == nil {
		tl := new(apihub.TaxonomyList)
		if err = patch.UnmarshalContents(ac.Data, ac.GetContentType(), tl); err == nil {
			return &TaxList{tl}
		}
	}
	panic(err)
}

type TaxList struct {
	*apihub.TaxonomyList
}

func (tl *TaxList) exists(k, v string) bool {
	for _, t := range tl.GetTaxonomies() {
		if k == t.GetId() {
			for _, e := range t.GetElements() {
				if v == e.Id {
					return true
				}
			}
		}
	}
	return false
}
