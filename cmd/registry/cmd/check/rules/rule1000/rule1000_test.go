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

package rule1000

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/cmd/registry/types"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
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

func TestRequiredArtifacts(t *testing.T) {
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "check-test", []seeder.RegistryResource{
		&rpc.Artifact{
			Name: "projects/check-test/locations/global/artifacts/good1",
		},
		&rpc.Artifact{
			Name: "projects/check-test/locations/global/artifacts/good2",
		},
	})
	ctx = context.WithValue(ctx, lint.ContextKeyRegistryClient, registryClient)

	bad1 := &rpc.Problem{
		Message:    `artifact "projects/check-test/locations/global/artifacts/bad1" not found in registry.`,
		Suggestion: `Initialize API Hub.`,
		Severity:   rpc.Problem_ERROR,
	}
	bad2 := &rpc.Problem{
		Message:    `artifact "projects/check-test/locations/global/artifacts/bad2" not found in registry.`,
		Suggestion: `Initialize API Hub.`,
		Severity:   rpc.Problem_ERROR,
	}

	for _, tt := range []struct {
		name     string
		required []string
		expected []*rpc.Problem
	}{
		{"good one", []string{"good1"}, nil},
		{"good two", []string{"good1", "good2"}, nil},
		{"bad one", []string{"bad1", "good2"}, []*rpc.Problem{bad1}},
		{"bad two", []string{"bad1", "bad2"}, []*rpc.Problem{bad1, bad2}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oldIds := requiredIDs
			t.Cleanup(func() { requiredIDs = oldIds })
			requiredIDs = tt.required
			p := &rpc.Project{
				Name: "projects/check-test",
			}
			if requiredArtifacts.OnlyIf(p) {
				got := requiredArtifacts.ApplyToProject(ctx, p)
				if diff := cmp.Diff(got, tt.expected, cmpopts.IgnoreUnexported(rpc.Problem{})); diff != "" {
					t.Errorf("unexpected diff: (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestRequiredTaxonomies(t *testing.T) {
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "check-test", nil)
	ctx = context.WithValue(ctx, lint.ContextKeyRegistryClient, registryClient)

	tl, _ := proto.Marshal(&rpc.TaxonomyList{
		Taxonomies: []*rpc.TaxonomyList_Taxonomy{
			{Id: "good1"},
			{Id: "good2"},
		},
	})
	_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
		ArtifactId: "apihub-taxonomies",
		Parent:     "projects/check-test/locations/global",
		Artifact: &rpc.Artifact{
			MimeType: types.MimeTypeForMessageType("google.cloud.apigeeregistry.v1.apihub.TaxonomyList"),
			Contents: tl,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name     string
		required []string
		expected []*rpc.Problem
	}{
		{"good", []string{"good1"}, nil},
		{"good2", []string{"good1", "good2"}, nil},
		{"bad", []string{"bad1"}, []*rpc.Problem{{
			Message:    `TaxonomyList "projects/check-test/locations/global/artifacts/apihub-taxonomies" must include items: bad1.`,
			Suggestion: `Initialize API Hub.`,
			Severity:   rpc.Problem_ERROR,
		}}},
		{"bad2", []string{"bad1", "bad2"}, []*rpc.Problem{{
			Message:    `TaxonomyList "projects/check-test/locations/global/artifacts/apihub-taxonomies" must include items: bad1, bad2.`,
			Suggestion: `Initialize API Hub.`,
			Severity:   rpc.Problem_ERROR,
		}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oldIds := requiredIDs
			t.Cleanup(func() { requiredIDs = oldIds })
			requiredIDs = []string{"apihub-taxonomies"}
			requiredTaxonomies = tt.required
			p := &rpc.Project{
				Name: "projects/check-test",
			}
			if requiredArtifacts.OnlyIf(p) {
				got := requiredArtifacts.ApplyToProject(ctx, p)
				if diff := cmp.Diff(got, tt.expected, cmpopts.IgnoreUnexported(rpc.Problem{})); diff != "" {
					t.Errorf("unexpected diff: (-want +got):\n%s", diff)
				}
			}
		})
	}
}