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

package rule1003

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/apihub"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
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

func TestStateIsValidLifecycleStage(t *testing.T) {
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "check-test", nil)
	ctx = context.WithValue(ctx, lint.ContextKeyRegistryClient, registryClient)

	// missing apihub-lifecycle artifact
	got := stateIsValidLifecycleStage.ApplyToApiVersion(ctx, &rpc.ApiVersion{})
	expected := []*check.Problem{{
		Severity:   check.Problem_ERROR,
		Message:    `APIVersion state is empty but must match a valid api-lifecycle stage.`,
		Suggestion: `Assign a appropriate value to the APIVersion state.`,
	}}
	if diff := cmp.Diff(got, expected, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
		t.Errorf("unexpected diff: (-want +got):\n%s", diff)
	}

	lc, _ := proto.Marshal(&apihub.Lifecycle{
		Id: "x",
		Stages: []*apihub.Lifecycle_Stage{
			{Id: "good1"},
			{Id: "good2"},
		},
	})
	_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
		ArtifactId: "apihub-lifecycle",
		Parent:     "projects/check-test/locations/global",
		Artifact: &rpc.Artifact{
			MimeType: mime.MimeTypeForMessageType("google.cloud.apigeeregistry.v1.apihub.Lifecycle"),
			Contents: lc,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		state    string
		expected []*check.Problem
	}{
		{"good2", nil},
		{"bad1", []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `APIVersion must match a valid api-lifecycle stage.`,
			Suggestion: `APIVersion state: "bad1", valid stages: good1, good2.`,
		}}},
		{"", []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    `APIVersion state is empty but must match a valid api-lifecycle stage.`,
			Suggestion: `Assign a appropriate value to the APIVersion state.`,
		}}},
	} {
		t.Run(tt.state, func(t *testing.T) {
			v := &rpc.ApiVersion{
				Name:  "projects/check-test/locations/global/apis/myapi/versions/myversion",
				State: tt.state,
			}
			if stateIsValidLifecycleStage.OnlyIf(v) {
				got := stateIsValidLifecycleStage.ApplyToApiVersion(ctx, v)
				if diff := cmp.Diff(got, tt.expected, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
					t.Errorf("unexpected diff: (-want +got):\n%s", diff)
				}
			}
		})
	}
}
