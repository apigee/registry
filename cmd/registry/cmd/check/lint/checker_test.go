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
	"fmt"
	"strings"
	"testing"

	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestChecker_run(t *testing.T) {
	defaultConfigs := Configs{}

	testRuleName := NewRuleName(111, "test-rule")
	ruleProblems := []*check.Problem{{
		Message:  "rule1_problem",
		Location: "projects/checker-test",
		RuleId:   string(testRuleName),
	}}

	tests := []struct {
		testName string
		configs  Configs
		problems []*check.Problem
	}{
		{"Empty", Configs{}, ruleProblems},
		{
			"NonMatchingFile",
			append(
				defaultConfigs,
				Config{
					IncludedPaths: []string{"nofile"},
				},
			),
			ruleProblems,
		},
		{
			"NonMatchingRule",
			append(
				defaultConfigs,
				Config{
					DisabledRules: []string{"foo::bar"},
				},
			),
			ruleProblems,
		},
		{
			"DisabledRule",
			append(
				defaultConfigs,
				Config{
					DisabledRules: []string{string(testRuleName)},
				},
			),
			[]*check.Problem{},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			rules := NewRuleRegistry()
			err := rules.Register(111, &ProjectRule{
				Name: NewRuleName(111, "test-rule"),
				ApplyToProject: func(ctx context.Context, p *rpc.Project) []*check.Problem {
					if c := RegistryClient(ctx); c == nil {
						t.Errorf("RegistryClient missing in context: %v", ctx)
					}
					return []*check.Problem{{
						Message: "rule1_problem",
					}}
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			registryClient, adminClient := grpctest.SetupRegistry(ctx, t, "checker-test", []seeder.RegistryResource{
				&rpc.Project{
					Name: "projects/checker-test",
				},
			})

			l := New(rules, test.configs)
			root := names.Project{
				ProjectID: "checker-test",
			}
			resp, err := l.Check(ctx, adminClient, registryClient, root, "", 10)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.problems, resp.Problems, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
				t.Errorf("Unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestChecker_panic(t *testing.T) {
	testRuleNum := 111

	tests := []struct {
		testName string
		rule     Rule
	}{
		{
			testName: "Panic",
			rule: &ProjectRule{
				Name: NewRuleName(testRuleNum, "panic"),
				ApplyToProject: func(ctx context.Context, p *rpc.Project) []*check.Problem {
					panic("panic")
				},
			},
		},
		{
			testName: "PanicError",
			rule: &ProjectRule{
				Name: NewRuleName(testRuleNum, "panic"),
				ApplyToProject: func(ctx context.Context, p *rpc.Project) []*check.Problem {
					panic(fmt.Errorf("panic"))
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			rules := NewRuleRegistry()
			err := rules.Register(testRuleNum, test.rule)
			if err != nil {
				t.Fatalf("Failed to create Rules: %q", err)
			}

			ctx := context.Background()
			registryClient, adminClient := grpctest.SetupRegistry(ctx, t, "checker-test", []seeder.RegistryResource{
				&rpc.Project{
					Name: "projects/checker-test",
				},
			})

			l := New(rules, nil)
			root := names.Project{
				ProjectID: "checker-test",
			}
			_, err = l.Check(ctx, adminClient, registryClient, root, "", 10)
			if err == nil || !strings.Contains(err.Error(), "panic") {
				t.Fatalf("Expected error with panic, got: %v", err)
			}
		})
	}
}
