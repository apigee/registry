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

package rule106

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/connection/grpctest"
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

func TestAddRules(t *testing.T) {
	if err := AddRules(lint.NewRuleRegistry()); err != nil {
		t.Errorf("AddRules got an error: %v", err)
	}
}

func Test_apiSpecRevisionRef(t *testing.T) {
	ctx := context.Background()
	name := "projects/check-test/locations/global/apis/myapi/versions/myversion/specs/good"
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "check-test", []seeder.RegistryResource{
		&rpc.ApiSpec{
			Name: name,
		},
	})
	spec, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	revName := name + "@" + spec.RevisionId

	ctx = context.WithValue(context.Background(), lint.ContextKeyRegistryClient, registryClient)

	for _, tt := range []struct {
		desc     string
		in       string
		expected []*check.Problem
	}{
		{"empty", "", nil},
		{"unable to parse", "bad", []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `api_spec_revision "bad" is not a valid ApiSpecRevision name.`,
			Suggestion: `Parse error: invalid spec revision name "bad": must match "^projects/([a-z0-9-.]+)/locations/global/apis/([a-z0-9-.]+)/versions/([a-z0-9-.]+)/specs/([a-z0-9-.]+)(?:@([a-z0-9-]+))?$"`,
		}}},
		{"not a revision", name, []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `api_spec_revision "projects/check-test/locations/global/apis/myapi/versions/myversion/specs/good" is not a valid ApiSpecRevision name.`,
			Suggestion: `A revision ID is required.`,
		}}},
		{"not a sibling", "projects/check-test/locations/global/apis/bad/versions/myversion/specs/spec@foo", []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `api_spec_revision "projects/check-test/locations/global/apis/bad/versions/myversion/specs/spec@foo" is not an API sibling of this Deployment.`,
			Suggestion: `Correct the api_spec_revision.`,
		}}},
		{"missing", name + "@missing", []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `api_spec_revision "projects/check-test/locations/global/apis/myapi/versions/myversion/specs/good@missing" not found in registry.`,
			Suggestion: `Correct the api_spec_revision.`,
		}}},
		{"good", revName, nil},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			a := &rpc.ApiDeployment{
				Name:            "projects/check-test/locations/global/apis/myapi/deployments/mydeployment",
				ApiSpecRevision: tt.in,
			}
			if apiSpecRevisionRef.OnlyIf(a) {
				got := apiSpecRevisionRef.ApplyToApiDeployment(ctx, a)
				if diff := cmp.Diff(got, tt.expected, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
					t.Errorf("unexpected diff: (-want +got):\n%s", diff)
				}
			}
		})
	}
}
