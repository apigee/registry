// Copyright 2022 Google LLC
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

package lint

import (
	"context"
	"reflect"
	"testing"

	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/rpc"
)

func TestProjectRule(t *testing.T) {
	resource := &rpc.Project{
		Name: "resource/myresource",
	}
	for _, test := range makeRuleTests(resource) {
		t.Run(test.testName, func(t *testing.T) {
			rule := &ProjectRule{
				Name: RuleName("test"),
				OnlyIf: func(p *rpc.Project) bool {
					return p.GetName() == resource.Name
				},
				ApplyToProject: func(ctx context.Context, p *rpc.Project) []*check.Problem {
					return test.problems
				},
			}

			test.runRule(rule, resource, t)
		})
	}
}

func TestArtifactRule(t *testing.T) {
	resource := &rpc.Artifact{
		Name: "resource/myresource",
	}
	for _, test := range makeRuleTests(resource) {
		t.Run(test.testName, func(t *testing.T) {
			rule := &ArtifactRule{
				Name: RuleName("test"),
				OnlyIf: func(p *rpc.Artifact) bool {
					return p.GetName() == resource.Name
				},
				ApplyToArtifact: func(ctx context.Context, p *rpc.Artifact) []*check.Problem {
					return test.problems
				},
			}

			test.runRule(rule, resource, t)
		})
	}
}

func TestApiRule(t *testing.T) {
	resource := &rpc.Api{
		Name: "resource/myresource",
	}
	for _, test := range makeRuleTests(resource) {
		t.Run(test.testName, func(t *testing.T) {
			rule := &ApiRule{
				Name: RuleName("test"),
				OnlyIf: func(p *rpc.Api) bool {
					return p.GetName() == resource.Name
				},
				ApplyToApi: func(ctx context.Context, p *rpc.Api) []*check.Problem {
					return test.problems
				},
			}

			test.runRule(rule, resource, t)
		})
	}
}

func TestApiDeploymentRule(t *testing.T) {
	resource := &rpc.ApiDeployment{
		Name: "resource/myresource",
	}
	for _, test := range makeRuleTests(resource) {
		t.Run(test.testName, func(t *testing.T) {
			rule := &ApiDeploymentRule{
				Name: RuleName("test"),
				OnlyIf: func(p *rpc.ApiDeployment) bool {
					return p.GetName() == resource.Name
				},
				ApplyToApiDeployment: func(ctx context.Context, p *rpc.ApiDeployment) []*check.Problem {
					return test.problems
				},
			}

			test.runRule(rule, resource, t)
		})
	}
}

func TestApiVersion(t *testing.T) {
	resource := &rpc.ApiVersion{
		Name: "resource/myresource",
	}
	for _, test := range makeRuleTests(resource) {
		t.Run(test.testName, func(t *testing.T) {
			rule := &ApiVersionRule{
				Name: RuleName("test"),
				OnlyIf: func(p *rpc.ApiVersion) bool {
					return p.GetName() == resource.Name
				},
				ApplyToApiVersion: func(ctx context.Context, p *rpc.ApiVersion) []*check.Problem {
					return test.problems
				},
			}

			test.runRule(rule, resource, t)
		})
	}
}

func TestApiSpec(t *testing.T) {
	resource := &rpc.ApiSpec{
		Name: "resource/myresource",
	}
	for _, test := range makeRuleTests(resource) {
		t.Run(test.testName, func(t *testing.T) {
			rule := &ApiSpecRule{
				Name: RuleName("test"),
				OnlyIf: func(p *rpc.ApiSpec) bool {
					return p.GetName() == resource.Name
				},
				ApplyToApiSpec: func(ctx context.Context, p *rpc.ApiSpec) []*check.Problem {
					return test.problems
				},
			}

			test.runRule(rule, resource, t)
		})
	}
}

func TestFieldRule(t *testing.T) {
	project := &rpc.Project{
		Name: "projects/myproject",
	}

	// Iterate over the tests and run them.
	for _, test := range makeRuleTests(project) {
		t.Run(test.testName, func(t *testing.T) {
			rule := &FieldRule{
				Name: RuleName("test"),
				OnlyIf: func(resource Resource, name string) bool {
					return name == "Name"
				},
				ApplyToField: func(ctx context.Context, resource Resource, name string, value interface{}) []*check.Problem {
					return test.problems
				},
			}

			test.runRule(rule, project, t)
		})
	}
}

type ruleTest struct {
	testName string
	problems []*check.Problem
}

// runRule runs a rule within a test environment.
func (test *ruleTest) runRule(rule Rule, r Resource, t *testing.T) {
	t.Helper()

	// Establish that the metadata methods work.
	if got, want := string(rule.GetName()), string(RuleName("test")); got != want {
		t.Errorf("Got %q for GetName(), expected %q", got, want)
	}

	// Run the rule's lint function on the file descriptor
	if got, want := rule.Apply(context.Background(), r), test.problems; !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v problems; expected %v.", got, want)
	}
}

// makeRuleTests generates boilerplate tests that are consistent for
// each type of rule.
func makeRuleTests(r Resource) []ruleTest {
	return []ruleTest{
		{"NoProblems", nil},
		{"OneProblem", []*check.Problem{{
			Message:  "There was a problem.",
			Location: r.GetName(),
		}}},
		{"TwoProblems", []*check.Problem{
			{
				Message:  "This was the first problem.",
				Location: r.GetName(),
			},
			{
				Message:  "This was the second problem.",
				Location: r.GetName(),
			},
		}},
	}
}
