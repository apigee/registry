// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		https://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/names"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestChecker_run(t *testing.T) {
	defaultConfigs := Configs{}

	testRuleName := NewRuleName(111, "test-rule")
	ruleProblems := []Problem{{
		Message:  "rule1_problem",
		Location: "projects/test",
		RuleID:   testRuleName,
	}}

	tests := []struct {
		testName string
		configs  Configs
		problems []Problem
	}{
		{"Empty", Configs{}, []Problem{}},
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
			[]Problem{},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			rules := NewRuleRegistry()
			err := rules.Register(111, &ProjectRule{
				Name: NewRuleName(111, "test-rule"),
				ApplyToProject: func(p *rpc.Project) []Problem {
					return test.problems
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { client.Close() })
			admin, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { admin.Close() })

			err = admin.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/test",
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Error deleting test project: %+v", err)
			}
			// Create the test project.
			_, err = admin.CreateProject(ctx, &rpc.CreateProjectRequest{
				ProjectId: "test",
				Project: &rpc.Project{
					DisplayName: "Test",
					Description: "A test catalog",
				},
			})
			if err != nil {
				t.Fatalf("Error creating project %s", err)
			}

			l := New(rules, test.configs)
			root := names.Project{
				ProjectID: "test",
			}
			resp, err := l.Check(ctx, admin, client, root, "", 10)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.problems, resp.Problems, cmpopts.IgnoreUnexported(Problem{})); diff != "" {
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
				ApplyToProject: func(p *rpc.Project) []Problem {
					panic("panic")
				},
			},
		},
		{
			testName: "PanicError",
			rule: &ProjectRule{
				Name: NewRuleName(testRuleNum, "panic"),
				ApplyToProject: func(p *rpc.Project) []Problem {
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
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { client.Close() })
			admin, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { admin.Close() })

			err = admin.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/test",
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Error deleting test project: %+v", err)
			}
			// Create the test project.
			_, err = admin.CreateProject(ctx, &rpc.CreateProjectRequest{
				ProjectId: "test",
				Project: &rpc.Project{
					DisplayName: "Test",
					Description: "A test catalog",
				},
			})
			if err != nil {
				t.Fatalf("Error creating project %s", err)
			}

			l := New(rules, nil)
			root := names.Project{
				ProjectID: "test",
			}
			_, err = l.Check(ctx, admin, client, root, "", 10)
			if err == nil || !strings.Contains(err.Error(), "panic") {
				t.Fatalf("Expected error with panic, got: %v", err)
			}
		})
	}
}
