// Copyright 2021 Google LLC.
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

package registry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestCreateApiDeployment(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Api
		req  *rpc.CreateApiDeploymentRequest
		want *rpc.ApiDeployment
	}{
		{
			desc: "fully populated resource",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "my-deployment",
				ApiDeployment: &rpc.ApiDeployment{
					Description: "My Description",
					Labels: map[string]string{
						"label-key": "label-value",
					},
					Annotations: map[string]string{
						"annotation-key": "annotation-value",
					},
				},
			},
			want: &rpc.ApiDeployment{
				Name:        "projects/my-project/locations/global/apis/a/deployments/my-deployment",
				Description: "My Description",
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
		},
		{
			desc: "empty resource",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "my-deployment",
				ApiDeployment:   nil,
			},
			want: &rpc.ApiDeployment{
				Name: "projects/my-project/locations/global/apis/a/deployments/my-deployment",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedApis(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			created, err := server.CreateApiDeployment(ctx, test.req)
			if err != nil {
				t.Fatalf("CreateApiDeployment(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiDeployment), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateApiDeployment(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if created.RevisionId == "" {
				t.Errorf("CreateApiDeployment(%+v) returned unexpected revision_id %q, expected non-empty ID", test.req, created.GetRevisionId())
			}

			if created.CreateTime == nil || created.RevisionCreateTime == nil || created.RevisionUpdateTime == nil {
				t.Errorf("CreateApiDeployment(%+v) returned unset create_time (%v), revision_create_time (%v), or revision_update_time (%v)", test.req, created.CreateTime, created.RevisionCreateTime, created.RevisionUpdateTime)
			} else if !created.CreateTime.AsTime().Equal(created.RevisionCreateTime.AsTime()) {
				t.Errorf("CreateApiDeployment(%+v) returned unexpected timestamps: create_time %v != revision_create_time %v", test.req, created.CreateTime, created.RevisionCreateTime)
			} else if !created.RevisionCreateTime.AsTime().Equal(created.RevisionUpdateTime.AsTime()) {
				t.Errorf("CreateApiDeployment(%+v) returned unexpected timestamps: revision_create_time %v != revision_update_time %v", test.req, created.RevisionCreateTime, created.RevisionUpdateTime)
			}

			t.Run("GetApiDeployment", func(t *testing.T) {
				req := &rpc.GetApiDeploymentRequest{
					Name: created.GetName(),
				}

				got, err := server.GetApiDeployment(ctx, req)
				if err != nil {
					t.Fatalf("GetApiDeployment(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(created, got, opts) {
					t.Errorf("GetApiDeployment(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(created, got, opts))
				}
			})
		})
	}
}

func TestCreateApiDeploymentResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Api
		req  *rpc.CreateApiDeploymentRequest
		want codes.Code
	}{
		{
			desc: "invalid parent name",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent: "invalid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "parent not found",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a2",
				ApiDeploymentId: "valid-id",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "valid-id",
				ApiDeployment:   nil,
			},
			want: codes.OK,
		},
		{
			desc: "missing custom identifier",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "specific revision",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "my-spec@12345678",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "this-identifier-is-invalid-because-it-exceeds-the-eighty-character-maximum-length",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier underscores",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "underscore_identifier",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen prefix",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "-identifier",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen suffix",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "identifier-",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier uuid format",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "072d2288-c685-42d8-9df0-5edbb2a809ea",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier mixed case",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "IDentifier",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedApis(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.CreateApiDeployment(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateApiDeployment(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestCreateApiDeploymentDuplicates(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiDeployment
		req  *rpc.CreateApiDeploymentRequest
		want codes.Code
	}{
		{
			desc: "case sensitive",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/my-deployment"},
			req: &rpc.CreateApiDeploymentRequest{
				Parent:          "projects/my-project/locations/global/apis/a",
				ApiDeploymentId: "my-deployment",
				ApiDeployment:   &rpc.ApiDeployment{},
			},
			want: codes.AlreadyExists,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedDeployments(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.CreateApiDeployment(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateApiDeployment(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestGetApiDeployment(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiDeployment
		req  *rpc.GetApiDeploymentRequest
		want *rpc.ApiDeployment
	}{
		{
			desc: "fully populated resource",
			seed: &rpc.ApiDeployment{
				Name:        "projects/my-project/locations/global/apis/a/deployments/d",
				Description: "My API Deployment",
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
			req: &rpc.GetApiDeploymentRequest{
				Name: "projects/my-project/locations/global/apis/a/deployments/d",
			},
			want: &rpc.ApiDeployment{
				Name:        "projects/my-project/locations/global/apis/a/deployments/d",
				Description: "My API Deployment",
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedDeployments(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.GetApiDeployment(ctx, test.req)
			if err != nil {
				t.Fatalf("GetApiDeployment(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiDeployment), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("GetApiDeployment(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}
		})
	}
}

func TestGetApiDeploymentResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiDeployment
		req  *rpc.GetApiDeploymentRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
			req: &rpc.GetApiDeploymentRequest{
				Name: "projects/my-project/locations/global/apis/a/deployments/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "invalid name",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
			req: &rpc.GetApiDeploymentRequest{
				Name: "invalid",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedDeployments(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.GetApiDeployment(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetApiDeployment(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApiDeployments(t *testing.T) {
	tests := []struct {
		admin     bool
		desc      string
		seed      []*rpc.ApiDeployment
		req       *rpc.ListApiDeploymentsRequest
		want      *rpc.ListApiDeploymentsResponse
		wantToken bool
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			seed: []*rpc.ApiDeployment{
				{Name: "projects/my-project/locations/global/apis/a1/deployments/d1"},
				{Name: "projects/my-project/locations/global/apis/a1/deployments/d2"},
				{Name: "projects/my-project/locations/global/apis/a1/deployments/d3"},
				{Name: "projects/my-project/locations/global/apis/a2/deployments/d1"},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent: "projects/my-project/locations/global/apis/a1",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{Name: "projects/my-project/locations/global/apis/a1/deployments/d1"},
					{Name: "projects/my-project/locations/global/apis/a1/deployments/d2"},
					{Name: "projects/my-project/locations/global/apis/a1/deployments/d3"},
				},
			},
		},
		{
			desc: "with specs containing multiple revisions",
			seed: []*rpc.ApiDeployment{
				{
					Name: "projects/my-project/locations/global/apis/a1/deployments/d1",
				},
				{
					Name:            "projects/my-project/locations/global/apis/a1/deployments/d1",
					ApiSpecRevision: "some-revision",
				},
				{
					Name: "projects/my-project/locations/global/apis/a1/deployments/d2",
				},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent: "projects/my-project/locations/global/apis/a1",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{
						Name:            "projects/my-project/locations/global/apis/a1/deployments/d1",
						ApiSpecRevision: "some-revision",
					},
					{
						Name: "projects/my-project/locations/global/apis/a1/deployments/d2",
					},
				},
			},
		},
		{
			admin: true,
			desc:  "across all apis in a specific project",
			seed: []*rpc.ApiDeployment{
				{Name: "projects/my-project/locations/global/apis/a1/deployments/d"},
				{Name: "projects/my-project/locations/global/apis/a2/deployments/d"},
				{Name: "projects/other-project/locations/global/apis/a1/deployments/d"},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent: "projects/my-project/locations/global/apis/-",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{Name: "projects/my-project/locations/global/apis/a1/deployments/d"},
					{Name: "projects/my-project/locations/global/apis/a2/deployments/d"},
				},
			},
		},
		{
			admin: true,
			desc:  "across all projects and apis",
			seed: []*rpc.ApiDeployment{
				{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
				{Name: "projects/other-project/locations/global/apis/other-api/deployments/d"},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent: "projects/-/locations/global/apis/-",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
					{Name: "projects/other-project/locations/global/apis/other-api/deployments/d"},
				},
			},
		},
		{
			admin: true,
			desc:  "in a specific api across all projects",
			seed: []*rpc.ApiDeployment{
				{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
				{Name: "projects/other-project/locations/global/apis/a/deployments/d"},
				{Name: "projects/my-project/locations/global/apis/other-api/deployments/d"},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent: "projects/-/locations/global/apis/a",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
					{Name: "projects/other-project/locations/global/apis/a/deployments/d"},
				},
			},
		},
		{
			desc: "custom page size",
			seed: []*rpc.ApiDeployment{
				{Name: "projects/my-project/locations/global/apis/a/deployments/d1"},
				{Name: "projects/my-project/locations/global/apis/a/deployments/d2"},
				{Name: "projects/my-project/locations/global/apis/a/deployments/d3"},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent:   "projects/my-project/locations/global/apis/a",
				PageSize: 1,
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{},
				},
			},
			wantToken: true,
			// Ordering is not guaranteed by API, so any resource may be returned.
			extraOpts: protocmp.IgnoreFields(new(rpc.ApiDeployment), "name"),
		},
		{
			desc: "name equality filtering",
			seed: []*rpc.ApiDeployment{
				{Name: "projects/my-project/locations/global/apis/a/deployments/d1"},
				{Name: "projects/my-project/locations/global/apis/a/deployments/d2"},
				{Name: "projects/my-project/locations/global/apis/a/deployments/d3"},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent: "projects/my-project/locations/global/apis/a",
				Filter: "name == 'projects/my-project/locations/global/apis/a/deployments/d2'",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{Name: "projects/my-project/locations/global/apis/a/deployments/d2"},
				},
			},
		},
		{
			desc: "description inequality filtering",
			seed: []*rpc.ApiDeployment{
				{
					Name:        "projects/my-project/locations/global/apis/a/deployments/d1",
					Description: "First ApiDeployment",
				},
				{Name: "projects/my-project/locations/global/apis/a/deployments/d2"},
				{Name: "projects/my-project/locations/global/apis/a/deployments/d3"},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent: "projects/my-project/locations/global/apis/a",
				Filter: "description != ''",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{
						Name:        "projects/my-project/locations/global/apis/a/deployments/d1",
						Description: "First ApiDeployment",
					},
				},
			},
		},
		{
			desc: "ordered by description",
			seed: []*rpc.ApiDeployment{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/deployments/d1",
					Description: "111: this should be returned first",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/deployments/d2",
					Description: "333: this should be returned third",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/deployments/d3",
					Description: "222: this should be returned second",
				},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "description",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/deployments/d1",
						Description: "111: this should be returned first",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/deployments/d3",
						Description: "222: this should be returned second",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/deployments/d2",
						Description: "333: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by description descending",
			seed: []*rpc.ApiDeployment{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/deployments/d1",
					Description: "111: this should be returned third",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/deployments/d2",
					Description: "333: this should be returned first",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/deployments/d3",
					Description: "222: this should be returned second",
				},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "description desc",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/deployments/d2",
						Description: "333: this should be returned first",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/deployments/d3",
						Description: "222: this should be returned second",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/deployments/d1",
						Description: "111: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by description then by name",
			seed: []*rpc.ApiDeployment{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/deployments/d1",
					Description: "222: this should be returned second or third (the name is the tie-breaker)",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/deployments/d3",
					Description: "111: this should be returned first",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/deployments/d2",
					Description: "222: this should be returned second or third (the name is the tie-breaker)",
				},
			},
			req: &rpc.ListApiDeploymentsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "description,name",
			},
			want: &rpc.ListApiDeploymentsResponse{
				ApiDeployments: []*rpc.ApiDeployment{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/deployments/d3",
						Description: "111: this should be returned first",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/deployments/d1",
						Description: "222: this should be returned second or third (the name is the tie-breaker)",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/deployments/d2",
						Description: "222: this should be returned second or third (the name is the tie-breaker)",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if test.admin && adminServiceUnavailable() {
				t.Skip(testRequiresAdminService)
			}
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedDeployments(ctx, server, test.seed...); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.ListApiDeployments(ctx, test.req)
			if err != nil {
				t.Fatalf("ListApiDeployments(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListApiDeploymentsResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.ApiDeployment), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListApiDeployments(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListApiDeployments(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				t.Errorf("ListApiDeployments(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListApiDeploymentsResponseCodes(t *testing.T) {
	tests := []struct {
		admin bool
		desc  string
		seed  *rpc.ApiDeployment
		req   *rpc.ListApiDeploymentsRequest
		want  codes.Code
	}{
		{
			desc: "parent api not found",
			req: &rpc.ListApiDeploymentsRequest{
				Parent: "projects/my-project/locations/global/apis/a",
			},
			want: codes.NotFound,
		},
		{
			admin: true,
			desc:  "parent project not found",
			req: &rpc.ListApiDeploymentsRequest{
				Parent: "projects/my-project/locations/global/apis/-",
			},
			want: codes.NotFound,
		},
		{
			desc: "negative page size",
			req: &rpc.ListApiDeploymentsRequest{
				PageSize: -1,
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid filter",
			req: &rpc.ListApiDeploymentsRequest{
				Filter: "this filter is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid page token",
			req: &rpc.ListApiDeploymentsRequest{
				PageToken: "this token is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering by unknown field",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment"},
			req: &rpc.ListApiDeploymentsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "something",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering by private field",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment"},
			req: &rpc.ListApiDeploymentsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "key",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering direction",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment"},
			req: &rpc.ListApiDeploymentsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "description asc",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering format",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment"},
			req: &rpc.ListApiDeploymentsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "description,",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if test.admin && adminServiceUnavailable() {
				t.Skip(testRequiresAdminService)
			}
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedDeployments(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.ListApiDeployments(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("ListApiDeployments(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApiDeploymentsSequence(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := []*rpc.ApiDeployment{
		{Name: "projects/my-project/locations/global/apis/a/deployments/d1"},
		{Name: "projects/my-project/locations/global/apis/a/deployments/d2"},
		{Name: "projects/my-project/locations/global/apis/a/deployments/d3"},
	}
	if err := seeder.SeedDeployments(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	listed := make([]*rpc.ApiDeployment, 0, 3)

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListApiDeploymentsRequest{
			Parent:   "projects/my-project/locations/global/apis/a",
			PageSize: 1,
		}

		got, err := server.ListApiDeployments(ctx, req)
		if err != nil {
			t.Fatalf("ListApiDeployments(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiDeployments()); count != 1 {
			t.Errorf("ListApiDeployments(%+v) returned %d specs, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListApiDeployments(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.ApiDeployments...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test intermediate page after failure on first page")
	}

	t.Run("intermediate page", func(t *testing.T) {
		req := &rpc.ListApiDeploymentsRequest{
			Parent:    "projects/my-project/locations/global/apis/a",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiDeployments(ctx, req)
		if err != nil {
			t.Fatalf("ListApiDeployments(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiDeployments()); count != 1 {
			t.Errorf("ListApiDeployments(%+v) returned %d specs, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListApiDeployments(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.ApiDeployments...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test final page after failure on intermediate page")
	}

	t.Run("final page", func(t *testing.T) {
		req := &rpc.ListApiDeploymentsRequest{
			Parent:    "projects/my-project/locations/global/apis/a",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiDeployments(ctx, req)
		if err != nil {
			t.Fatalf("ListApiDeployments(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiDeployments()); count != 1 {
			t.Errorf("ListApiDeployments(%+v) returned %d specs, expected exactly one", req, count)
		}

		if got.GetNextPageToken() != "" {
			t.Errorf("ListApiDeployments(%+v) returned next_page_token, expected no next page", req)
		}

		listed = append(listed, got.ApiDeployments...)
	})

	if t.Failed() {
		t.Fatal("Cannot test sequence result after failure on final page")
	}

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(rpc.ApiDeployment), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
		cmpopts.SortSlices(func(a, b *rpc.ApiDeployment) bool {
			return a.GetName() < b.GetName()
		}),
	}

	if !cmp.Equal(seed, listed, opts) {
		t.Errorf("List sequence returned unexpected diff (-want +got):\n%s", cmp.Diff(seed, listed, opts))
	}
}

func TestListApiDeploymentsLargeCollection(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := make([]*rpc.ApiDeployment, 0, 1001)
	for i := 1; i <= cap(seed); i++ {
		seed = append(seed, &rpc.ApiDeployment{
			Name: fmt.Sprintf("projects/my-project/locations/global/apis/a/deployments/d%03d", i),
		})
	}

	if err := seeder.SeedDeployments(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	// This test prevents the list sequence from ending before a known filter match is listed.
	// For simplicity, it does not guarantee the resource is returned on a later page.
	t.Run("filter", func(t *testing.T) {
		req := &rpc.ListApiDeploymentsRequest{
			Parent:   "projects/my-project/locations/global/apis/a",
			PageSize: 1,
			Filter:   "name == 'projects/my-project/locations/global/apis/a/deployments/d099'",
		}

		got, err := server.ListApiDeployments(ctx, req)
		if err != nil {
			t.Fatalf("ListApiDeployments(%+v) returned error: %s", req, err)
		}

		if len(got.GetApiDeployments()) == 1 && got.GetNextPageToken() != "" {
			t.Errorf("ListApiDeployments(%+v) returned a page token when the only matching resource has been listed: %+v", req, got)
		} else if len(got.GetApiDeployments()) == 0 && got.GetNextPageToken() == "" {
			t.Errorf("ListApiDeployments(%+v) returned an empty next page token before listing the only matching resource", req)
		} else if count := len(got.GetApiDeployments()); count > 1 {
			t.Errorf("ListApiDeployments(%+v) returned %d projects, expected at most one: %+v", req, count, got.GetApiDeployments())
		}
	})

	t.Run("max page size", func(t *testing.T) {
		req := &rpc.ListApiDeploymentsRequest{
			Parent:   "projects/my-project/locations/global/apis/a",
			PageSize: 1001,
		}

		got, err := server.ListApiDeployments(ctx, req)
		if err != nil {
			t.Fatalf("ListApiDeployments(%+v) returned error: %s", req, err)
		}

		if len(got.GetApiDeployments()) != 1000 {
			t.Errorf("GetApiDeployments(%+v) should have returned 1000 items, got: %+v", req, len(got.GetApiDeployments()))
		} else if got.GetNextPageToken() == "" {
			t.Errorf("GetApiDeployments(%+v) should return a next page token", req)
		}
	})
}

func TestUpdateApiDeployment(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiDeployment
		req  *rpc.UpdateApiDeploymentRequest
		want *rpc.ApiDeployment
	}{
		{
			desc: "allow missing updates existing resources",
			seed: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My ApiDeployment",
				AccessGuidance: "openapi.json",
			},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name:        "projects/my-project/locations/global/apis/a/deployments/d",
					Description: "My Updated ApiDeployment",
				},
				UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"description"}},
				AllowMissing: true,
			},
			want: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My Updated ApiDeployment",
				AccessGuidance: "openapi.json",
			},
		},
		{
			desc: "allow missing creates missing resources",
			seed: &rpc.ApiDeployment{
				Name: "projects/my-project/locations/global/apis/a/deployments/d-sibling",
			},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name: "projects/my-project/locations/global/apis/a/deployments/d",
				},
				AllowMissing: true,
			},
			want: &rpc.ApiDeployment{
				Name: "projects/my-project/locations/global/apis/a/deployments/d",
			},
		},
		{
			desc: "implicit nil mask",
			seed: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My ApiDeployment",
				AccessGuidance: "openapi.json",
			},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name:        "projects/my-project/locations/global/apis/a/deployments/d",
					Description: "My Updated ApiDeployment",
				},
			},
			want: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My Updated ApiDeployment",
				AccessGuidance: "openapi.json",
			},
		},
		{
			desc: "implicit empty mask",
			seed: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My ApiDeployment",
				AccessGuidance: "openapi.json",
			},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name:        "projects/my-project/locations/global/apis/a/deployments/d",
					Description: "My Updated ApiDeployment",
				},
				UpdateMask: &fieldmaskpb.FieldMask{},
			},
			want: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My Updated ApiDeployment",
				AccessGuidance: "openapi.json",
			},
		},
		{
			desc: "field specific mask",
			seed: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My ApiDeployment",
				AccessGuidance: "openapi.json",
			},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name:           "projects/my-project/locations/global/apis/a/deployments/d",
					Description:    "My Updated ApiDeployment",
					AccessGuidance: "Ignored",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"description"}},
			},
			want: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My Updated ApiDeployment",
				AccessGuidance: "openapi.json",
			},
		},
		{
			desc: "full replacement wildcard mask",
			seed: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My ApiDeployment",
				AccessGuidance: "openapi.json",
			},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name:        "projects/my-project/locations/global/apis/a/deployments/d",
					Description: "My Updated ApiDeployment",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			},
			want: &rpc.ApiDeployment{
				Name:           "projects/my-project/locations/global/apis/a/deployments/d",
				Description:    "My Updated ApiDeployment",
				AccessGuidance: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedDeployments(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			updated, err := server.UpdateApiDeployment(ctx, test.req)
			if err != nil {
				t.Fatalf("UpdateApiDeployment(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiDeployment), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
			}

			if !cmp.Equal(test.want, updated, opts) {
				t.Errorf("UpdateApiDeployment(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, updated, opts))
			}

			t.Run("GetApiDeployment", func(t *testing.T) {
				req := &rpc.GetApiDeploymentRequest{
					Name: updated.GetName(),
				}

				got, err := server.GetApiDeployment(ctx, req)
				if err != nil {
					t.Fatalf("GetApiDeployment(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(updated, got, opts) {
					t.Errorf("GetApiDeployment(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(updated, got, opts))
				}
			})
		})
	}
}

func TestUpdateApiDeploymentResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiDeployment
		req  *rpc.UpdateApiDeploymentRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name: "projects/my-project/locations/global/apis/a/deployments/doesnt-exist",
				},
			},
			want: codes.NotFound,
		},
		{
			desc: "specific revision",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name: "projects/my-project/locations/global/apis/a/versions/v1/deployments/d@12345678",
				},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "missing resource body",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
			req:  &rpc.UpdateApiDeploymentRequest{},
			want: codes.InvalidArgument,
		},
		{
			desc: "missing resource name",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "nonexistent field in mask",
			seed: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name: "projects/my-project/locations/global/apis/a/deployments/d",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"this field does not exist"}},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedDeployments(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.UpdateApiDeployment(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApiDeployment(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestUpdateApiDeploymentSequence(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.UpdateApiDeploymentRequest
		want codes.Code
	}{
		{
			desc: "create using update with allow_missing=false",
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name: "projects/my-project/locations/global/apis/a/deployments/d",
				},
				AllowMissing: false,
			},
			want: codes.NotFound,
		},
		{
			desc: "create using update with allow_missing=true",
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name: "projects/my-project/locations/global/apis/a/deployments/d",
				},
				AllowMissing: true,
			},
			want: codes.OK,
		},
		{
			desc: "update existing resource with allow_missing=true",
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name: "projects/my-project/locations/global/apis/a/deployments/d",
				},
				AllowMissing: true,
			},
			want: codes.OK,
		},
		{
			desc: "update existing resource with allow_missing=false",
			req: &rpc.UpdateApiDeploymentRequest{
				ApiDeployment: &rpc.ApiDeployment{
					Name: "projects/my-project/locations/global/apis/a/deployments/d",
				},
				AllowMissing: false,
			},
			want: codes.OK,
		},
	}
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := &rpc.Api{Name: "projects/my-project/locations/global/apis/a"}
	if err := seeder.SeedApis(ctx, server, seed); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	var createTime time.Time
	var revisionCreateTime time.Time
	var revisionUpdateTime time.Time
	// NOTE: in the following sequence of tests, each test depends on its predecessor.
	// Resources are successively created and updated using the "Update" RPC and the
	// tests verify that CreateTime/UpdateTime fields are modified appropriately.
	for i, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var result *rpc.ApiDeployment
			var err error
			if result, err = server.UpdateApiDeployment(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApiDeployment(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
			if result != nil {
				if i == 1 {
					createTime = result.CreateTime.AsTime()
					revisionCreateTime = result.RevisionCreateTime.AsTime()
					revisionUpdateTime = result.RevisionUpdateTime.AsTime()
				} else {
					if !createTime.Equal(result.CreateTime.AsTime()) {
						t.Errorf("UpdateApiDeployment create time changed after update (%v %v)", createTime, result.CreateTime.AsTime())
					}
					if !revisionCreateTime.Equal(result.RevisionCreateTime.AsTime()) {
						t.Errorf("UpdateApiDeployment revision create time changed after update (%v %v)", revisionCreateTime, result.RevisionCreateTime.AsTime())
					}
					if !revisionUpdateTime.Before(result.RevisionUpdateTime.AsTime()) {
						t.Errorf("UpdateApiDeployment update time did not increase after update (%v %v)", revisionUpdateTime, result.RevisionUpdateTime.AsTime())
					}
					revisionUpdateTime = result.RevisionUpdateTime.AsTime()
				}
			}
		})
	}
}

func TestDeleteApiDeployment(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiDeployment
		req  *rpc.DeleteApiDeploymentRequest
	}{
		{
			desc: "existing resource",
			seed: &rpc.ApiDeployment{
				Name: "projects/my-project/locations/global/apis/a/deployments/d",
			},
			req: &rpc.DeleteApiDeploymentRequest{
				Name: "projects/my-project/locations/global/apis/a/deployments/d",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedDeployments(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.DeleteApiDeployment(ctx, test.req); err != nil {
				t.Fatalf("DeleteApiDeployment(%+v) returned error: %s", test.req, err)
			}

			t.Run("GetApiDeployment", func(t *testing.T) {
				req := &rpc.GetApiDeploymentRequest{
					Name: test.req.GetName(),
				}

				if _, err := server.GetApiDeployment(ctx, req); status.Code(err) != codes.NotFound {
					t.Fatalf("GetApiDeployment(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), codes.NotFound, err)
				}
			})
		})
	}
}

func TestDeleteApiDeploymentResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.DeleteApiDeploymentRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.DeleteApiDeploymentRequest{
				Name: "projects/my-project/locations/global/apis/a/deployments/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "specific revision",
			req: &rpc.DeleteApiDeploymentRequest{
				Name: "projects/my-project/locations/global/apis/a/deployments/d@12345678",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "resource has children",
			seed: &rpc.Artifact{
				Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment/artifacts/my-artifact",
			},
			req: &rpc.DeleteApiDeploymentRequest{
				Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment",
			},
			want: codes.FailedPrecondition,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedArtifacts(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.DeleteApiDeployment(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("DeleteApiDeployment(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestDeleteApiDeploymentCascading(t *testing.T) {
	var (
		ctx      = context.Background()
		server   = defaultTestServer(t)
		artifact = &rpc.Artifact{
			Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment/artifacts/my-artifact",
		}
	)

	if err := seeder.SeedArtifacts(ctx, server, artifact); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	req := &rpc.DeleteApiDeploymentRequest{
		Name:  "projects/my-project/locations/global/apis/my-api/deployments/my-deployment",
		Force: true,
	}

	if _, err := server.DeleteApiDeployment(ctx, req); err != nil {
		t.Errorf("DeleteApiDeployment(%+v) returned error: %s", req, err)
	}

	if _, err := server.GetApiDeployment(ctx, &rpc.GetApiDeploymentRequest{Name: req.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetApiDeployment(%q) returned status code %q, want %q: %s", req.GetName(), status.Code(err), codes.NotFound, err)
	}

	if _, err := server.GetArtifact(ctx, &rpc.GetArtifactRequest{Name: artifact.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetArtifact(%q) returned status code %q, want %q: %s", artifact.GetName(), status.Code(err), codes.NotFound, err)
	}
}
