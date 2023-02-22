// Copyright 2023 Google LLC. All Rights Reserved.
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

package rule102

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/artifacts"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestAddRules(t *testing.T) {
	if err := AddRules(lint.NewRuleRegistry()); err != nil {
		t.Errorf("AddRules got an error: %v", err)
	}
}

func Test_recommendedDeploymentRef(t *testing.T) {
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "check-test", []seeder.RegistryResource{
		&rpc.ApiDeployment{
			Name: "projects/check-test/locations/global/apis/myapi/deployments/good",
		},
	})
	ctx = context.WithValue(context.Background(), lint.ContextKeyRegistryClient, registryClient)

	for _, tt := range []struct {
		desc     string
		in       string
		expected []*artifacts.Problem
	}{
		{"empty", "", nil},
		{"unable to parse", "bad", []*artifacts.Problem{{
			Severity:   artifacts.Problem_ERROR,
			Message:    `recommended_deployment "bad" is not a valid ApiDeployment name.`,
			Suggestion: `Parse error: invalid deployment name "bad": must match "^projects/([A-Za-z0-9-.]+)/locations/global/apis/([A-Za-z0-9-.]+)/deployments/([A-Za-z0-9-.]+)$"`,
		}}},
		{"not a child", "projects/check-test/locations/global/apis/bad/deployments/bad", []*artifacts.Problem{{
			Severity:   artifacts.Problem_ERROR,
			Message:    `recommended_deployment "projects/check-test/locations/global/apis/bad/deployments/bad" is not a child of this Api.`,
			Suggestion: `Correct the recommended_deployment.`,
		}}},
		{"missing", "projects/check-test/locations/global/apis/myapi/deployments/missing", []*artifacts.Problem{{
			Severity:   artifacts.Problem_ERROR,
			Message:    `recommended_deployment "projects/check-test/locations/global/apis/myapi/deployments/missing" not found in registry.`,
			Suggestion: `Correct the recommended_deployment.`,
		}}},
		{"good", "projects/check-test/locations/global/apis/myapi/deployments/good", nil},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			a := &rpc.Api{
				Name:                  "projects/check-test/locations/global/apis/myapi",
				RecommendedDeployment: tt.in,
			}

			if recommendedDeploymentRef.OnlyIf(a) {
				got := recommendedDeploymentRef.ApplyToApi(ctx, a)
				if diff := cmp.Diff(got, tt.expected, cmpopts.IgnoreUnexported(artifacts.Problem{})); diff != "" {
					t.Errorf("unexpected diff: (-want +got):\n%s", diff)
				}
			}
		})
	}
}
