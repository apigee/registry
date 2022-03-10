// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var sortActions = cmpopts.SortSlices(func(a, b *Action) bool { return a.Command < b.Command })
var styleguide = &rpc.StyleGuide{
	Id:        "registry-styleguide",
	MimeTypes: []string{gzipOpenAPIv3},
	Guidelines: []*rpc.Guideline{
		{
			Id: "Operation",
			Rules: []*rpc.Rule{
				{
					Id:             "OperationIdValidInURL",
					Linter:         "spectral",
					LinterRulename: "operation-operationId-valid-in-url",
					Severity:       rpc.Rule_WARNING,
				},
			},
			Status: rpc.Guideline_ACTIVE,
		},
	},
}

// Tests for artifacts as resources and specs as dependencies
func TestArtifacts(t *testing.T) {
	tests := []struct {
		desc  string
		setup func(context.Context, connection.Client, connection.AdminClient)
		want  []*Action
	}{
		{
			desc: "single spec",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
			},
			want: []*Action{
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
				},
			},
		},
		{
			desc: "multiple specs",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
			},
			want: []*Action{
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
				},
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic",
				},
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
				},
			},
		},
		{
			desc: "partially existing artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
			},
			want: []*Action{
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
				},
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
				},
			},
		},
		{
			desc: "outdated artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
				// Update spec 1.0.1 to make the artifact outdated
				updateSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml")
			},
			want: []*Action{
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic",
				},
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
				},
			},
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			manifest := &rpc.Manifest{
				Id: "controller-test",
				GeneratedResources: []*rpc.GeneratedResource{
					{
						Pattern: "apis/-/versions/-/specs/-/artifacts/lint-gnostic",
						Dependencies: []*rpc.Dependency{
							{
								Pattern: "$resource.spec",
								Filter:  "mime_type.contains('openapi')",
							},
						},
						Action: "registry compute lint $resource.spec --linter gnostic",
					},
				},
			}
			actions := ProcessManifest(ctx, registryClient, projectID, manifest)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}

			deleteProject(ctx, adminClient, t, "controller-test")
		})
	}
}

// Tests for aggregated artifacts at api level and specs as resources
func TestAggregateArtifacts(t *testing.T) {
	tests := []struct {
		desc  string
		setup func(context.Context, connection.Client, connection.AdminClient)
		want  []*Action
	}{
		{
			desc: "create artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "test-api-1")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

				// Test API 2
				createApi(ctx, client, t, "projects/controller-test/locations/global", "test-api-2")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
			},
			want: []*Action{
				{
					Command:           "registry compute vocabulary projects/controller-test/locations/global/apis/test-api-1",
					GeneratedResource: "projects/controller-test/locations/global/apis/test-api-1/artifacts/vocabulary",
				},
				{
					Command:           "registry compute vocabulary projects/controller-test/locations/global/apis/test-api-2",
					GeneratedResource: "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary",
				},
			},
		},
		{
			desc: "outdated arttifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "test-api-1")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-1/artifacts/vocabulary")

				// Test API 2
				createApi(ctx, client, t, "projects/controller-test/locations/global", "test-api-2")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary")
				// Update underlying spec to make artifact outdated
				updateSpec(ctx, client, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.1/specs/openapi.yaml")
			},
			want: []*Action{
				{
					Command:           "registry compute vocabulary projects/controller-test/locations/global/apis/test-api-2",
					GeneratedResource: "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary",
				},
			},
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			manifest := &rpc.Manifest{

				Id: "controller-test",
				GeneratedResources: []*rpc.GeneratedResource{
					{
						Pattern: "apis/-/artifacts/vocabulary",
						Dependencies: []*rpc.Dependency{
							{
								Pattern: "$resource.api/versions/-/specs/-",
							},
						},
						Action: "registry compute vocabulary $resource.api",
					},
				},
			}
			actions := ProcessManifest(ctx, registryClient, projectID, manifest)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}

			deleteProject(ctx, adminClient, t, "controller-test")
		})
	}

}

// Tests for derived artifacts with artifacts as dependencies
func TestDerivedArtifacts(t *testing.T) {
	tests := []struct {
		desc  string
		setup func(context.Context, connection.Client, connection.AdminClient)
		want  []*Action
	}{
		{
			desc: "create artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity")
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity")
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity")
			},
			want: []*Action{
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/summary",
				},
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/summary",
				},
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/summary",
				},
			},
		},
		{
			desc: "missing artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity")
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity")
			},
			want: []*Action{
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/summary",
				},
			},
		},
		{
			desc: "outdated artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")

				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/summary")
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/summary")
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/summary")

				// Make some artifacts outdated from the above setup
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity")
			},
			want: []*Action{
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/summary",
				},
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/summary",
				},
			},
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			manifest := &rpc.Manifest{
				Id: "controller-test",
				GeneratedResources: []*rpc.GeneratedResource{
					{
						Pattern: "apis/-/versions/-/specs/-/artifacts/summary",
						Dependencies: []*rpc.Dependency{
							{
								Pattern: "$resource.spec/artifacts/lint-gnostic",
							},
							{
								Pattern: "$resource.spec/artifacts/complexity",
							},
						},
						Action: "registry compute summary $resource.spec/artifacts/lint-gnostic $resource.spec/artifacts/complexity",
					},
				},
			}
			actions := ProcessManifest(ctx, registryClient, projectID, manifest)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}

			deleteProject(ctx, adminClient, t, "controller-test")
		})
	}

}

// Tests for receipt artifacts as generated resource
func TestReceiptArtifacts(t *testing.T) {
	tests := []struct {
		desc  string
		setup func(context.Context, connection.Client, connection.AdminClient)
		want  []*Action
	}{
		{
			desc: "create artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")

				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
			},
			want: []*Action{
				{
					Command:           "command projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/custom-artifact",
					RequiresReceipt:   true,
				},
				{
					Command:           "command projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/custom-artifact",
					RequiresReceipt:   true,
				},
				{
					Command:           "command projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/custom-artifact",
					RequiresReceipt:   true,
				},
			},
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			manifest := &rpc.Manifest{
				Id: "controller-test",
				GeneratedResources: []*rpc.GeneratedResource{
					{
						Pattern: "apis/-/versions/-/specs/-/artifacts/custom-artifact",
						Receipt: true,
						Dependencies: []*rpc.Dependency{
							{
								Pattern: "$resource.spec",
							},
						},
						Action: "command $resource.spec",
					},
				},
			}
			actions := ProcessManifest(ctx, registryClient, projectID, manifest)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}

			deleteProject(ctx, adminClient, t, "controller-test")
		})
	}

}

// Tests for receipt aggregate artifacts as generated resource
func TestReceiptAggArtifacts(t *testing.T) {
	tests := []struct {
		desc  string
		setup func(context.Context, connection.Client, connection.AdminClient)
		want  []*Action
	}{
		{
			desc: "create artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")

				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
			},
			want: []*Action{
				{
					Command:           "registry compute search-index projects/controller-test/locations/global/apis/-/versions/-/specs/-",
					GeneratedResource: "projects/controller-test/locations/global/artifacts/search-index",
					RequiresReceipt:   true,
				},
			},
		},
		{
			desc: "updated artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")

				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				// Create target artifact
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/artifacts/search-index")

				// Add a new spec to make the artifact outdated
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
			},
			want: []*Action{
				{
					Command:           "registry compute search-index projects/controller-test/locations/global/apis/-/versions/-/specs/-",
					GeneratedResource: "projects/controller-test/locations/global/artifacts/search-index",
					RequiresReceipt:   true,
				},
			},
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			manifest := &rpc.Manifest{
				Id: "controller-test",
				GeneratedResources: []*rpc.GeneratedResource{
					{
						Pattern: "artifacts/search-index",
						Receipt: true,
						Dependencies: []*rpc.Dependency{
							{
								Pattern: "apis/-/versions/-/specs/-",
							},
						},
						Action: "registry compute search-index projects/controller-test/locations/global/apis/-/versions/-/specs/-",
					},
				},
			}
			actions := ProcessManifest(ctx, registryClient, projectID, manifest)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}

			deleteProject(ctx, adminClient, t, "controller-test")
		})
	}

}

// Tests for manifest with multiple entity references
func TestMultipleEntitiesArtifacts(t *testing.T) {
	tests := []struct {
		desc  string
		setup func(context.Context, connection.Client, connection.AdminClient)
		want  []*Action
	}{
		{
			desc: "create artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")

				uploadStyleguide(ctx, client, t, "controller-test", styleguide)

				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")
				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
			},
			want: []*Action{
				{
					Command:           "registry compute conformance projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-registry-styleguide",
					RequiresReceipt:   true,
				},
				{
					Command:           "registry compute conformance projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/conformance-registry-styleguide",
					RequiresReceipt:   true,
				},
				{
					Command:           "registry compute conformance projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/conformance-registry-styleguide",
					RequiresReceipt:   true,
				},
			},
		},
		{
			desc: "outdated artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")

				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-registry-styleguide")
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/conformance-registry-styleguide")
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
				createUpdateArtifact(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/conformance-registry-styleguide")

				//Update styleguide definition to make sure conformance artifacts are outdated
				uploadStyleguide(ctx, client, t, "controller-test", styleguide)
			},
			want: []*Action{
				{
					Command:           "registry compute conformance projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-registry-styleguide",
					RequiresReceipt:   true,
				},
				{
					Command:           "registry compute conformance projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/conformance-registry-styleguide",
					RequiresReceipt:   true,
				},
				{
					Command:           "registry compute conformance projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/conformance-registry-styleguide",
					RequiresReceipt:   true,
				},
			},
		},
		{
			desc: "missing dependencies",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "controller-test")
				createProject(ctx, adminClient, t, "controller-test")
				createApi(ctx, client, t, "projects/controller-test/locations/global", "petstore")

				uploadStyleguide(ctx, client, t, "controller-test", styleguide)

				// Version 1.0.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// Version 1.0.1
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
				// Version 1.1.0
				createVersion(ctx, client, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
				createSpec(ctx, client, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

			},
			want: []*Action{
				{
					Command:           "registry compute conformance projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-registry-styleguide",
					RequiresReceipt:   true,
				},
				{
					Command:           "registry compute conformance projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/conformance-registry-styleguide",
					RequiresReceipt:   true,
				},
			},
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			manifest := &rpc.Manifest{
				Id: "controller-test",
				GeneratedResources: []*rpc.GeneratedResource{
					{
						Pattern: "apis/-/versions/-/specs/-/artifacts/conformance-registry-styleguide",
						Receipt: true,
						Dependencies: []*rpc.Dependency{
							{
								Pattern: "$resource.spec",
							},
							{
								Pattern: "artifacts/registry-styleguide",
							},
						},
						Action: "registry compute conformance $resource.spec",
					},
				},
			}
			actions := ProcessManifest(ctx, registryClient, projectID, manifest)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}

			deleteProject(ctx, adminClient, t, "controller-test")
		})
	}

}
