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

package rule104

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

func Test_primarySpecRef(t *testing.T) {
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "check-test", []seeder.RegistryResource{
		&rpc.ApiSpec{
			Name: "projects/check-test/locations/global/apis/myapi/versions/myversion/specs/good",
		},
	})
	ctx = context.WithValue(context.Background(), lint.ContextKeyRegistryClient, registryClient)

	for _, tt := range []struct {
		desc     string
		in       string
		expected []*check.Problem
	}{
		{"empty", "", nil},
		{"unable to parse", "bad", []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `primary_spec "bad" is not a valid ApiSpec name.`,
			Suggestion: `Parse error: invalid spec name "bad": must match "^projects/([a-z0-9-.]+)/locations/global/apis/([a-z0-9-.]+)/versions/([a-z0-9-.]+)/specs/([a-z0-9-.]+)$"`,
		}}},
		{"not a sibling", "projects/check-test/locations/global/apis/bad/versions/myversion/specs/bad", []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `primary_spec "projects/check-test/locations/global/apis/bad/versions/myversion/specs/bad" is not an API sibling of this Version.`,
			Suggestion: `Correct the primary_spec.`,
		}}},
		{"missing", "projects/check-test/locations/global/apis/myapi/versions/myversion/specs/missing", []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `primary_spec "projects/check-test/locations/global/apis/myapi/versions/myversion/specs/missing" not found in registry.`,
			Suggestion: `Correct the primary_spec.`,
		}}},
		{"good", "projects/check-test/locations/global/apis/myapi/versions/myversion/specs/good", nil},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			a := &rpc.ApiVersion{
				Name:        "projects/check-test/locations/global/apis/myapi/versions/myversion",
				PrimarySpec: tt.in,
			}

			if primarySpecRef.OnlyIf(a) {
				got := primarySpecRef.ApplyToApiVersion(ctx, a)
				if diff := cmp.Diff(got, tt.expected, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
					t.Errorf("unexpected diff: (-want +got):\n%s", diff)
				}
			}
		})
	}
}
