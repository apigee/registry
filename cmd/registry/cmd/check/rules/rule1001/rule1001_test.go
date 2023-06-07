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

package rule1001

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

func TestTaxonomyList(t *testing.T) {
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "check-test", nil)
	ctx = context.WithValue(ctx, lint.ContextKeyRegistryClient, registryClient)

	tl, _ := proto.Marshal(&apihub.TaxonomyList{
		Taxonomies: []*apihub.TaxonomyList_Taxonomy{
			{
				Id: "apihub-list1",
				Elements: []*apihub.TaxonomyList_Taxonomy_Element{
					{Id: "good1"},
					{Id: "good2"},
				},
			},
			{
				Id: "apihub-list2",
				Elements: []*apihub.TaxonomyList_Taxonomy_Element{
					{Id: "good3"},
					{Id: "good4"},
				},
			},
		},
	})
	_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
		ArtifactId: "apihub-taxonomies",
		Parent:     "projects/check-test/locations/global",
		Artifact: &rpc.Artifact{
			MimeType: mime.MimeTypeForMessageType("google.cloud.apigeeregistry.v1.apihub.TaxonomyList"),
			Contents: tl,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name     string
		labels   map[string]string
		expected []*check.Problem
	}{
		{"good", map[string]string{
			"apihub-list1": "good1",
			"apihub-list2": "good3",
		}, nil},
		{"bad1", map[string]string{
			"apihub-list1": "good3",
			"apihub-list2": "good3",
		}, []*check.Problem{{
			Message:    `Label value "good3" not present in Taxonomy "apihub-list1"`,
			Suggestion: `Adjust label value or Taxonomy elements.`,
			Severity:   check.Problem_ERROR,
		}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := &rpc.Project{
				Name: "projects/check-test",
			}
			if taxonomyLabels.OnlyIf(p, "Labels") {
				got := taxonomyLabels.ApplyToField(ctx, p, "Labels", tt.labels)
				if diff := cmp.Diff(got, tt.expected, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
					t.Errorf("unexpected diff: (-want +got):\n%s", diff)
				}
			}
		})
	}
}
