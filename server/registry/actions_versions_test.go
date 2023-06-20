// Copyright 2020 Google LLC.
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

func TestCreateApiVersion(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Api
		req  *rpc.CreateApiVersionRequest
		want *rpc.ApiVersion
	}{
		{
			desc: "fully populated resource",
			seed: &rpc.Api{
				Name: "projects/my-project/locations/global/apis/my-api",
			},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "v1",
				ApiVersion: &rpc.ApiVersion{
					DisplayName: "My Display Name",
					Description: "My Description",
					State:       "My State",
					PrimarySpec: "specs/my spec",
					Labels: map[string]string{
						"label-key": "label-value",
					},
					Annotations: map[string]string{
						"annotation-key": "annotation-value",
					},
				},
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
				DisplayName: "My Display Name",
				Description: "My Description",
				State:       "My State",
				PrimarySpec: "specs/my spec",
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
			seed: &rpc.Api{
				Name: "projects/my-project/locations/global/apis/my-api",
			},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "v1",
				ApiVersion:   nil,
			},
			want: &rpc.ApiVersion{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1",
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

			created, err := server.CreateApiVersion(ctx, test.req)
			if err != nil {
				t.Fatalf("CreateApiVersion(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiVersion), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateApiVersion(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if created.CreateTime == nil || created.UpdateTime == nil {
				t.Errorf("CreateApiVersion(%+v) returned unset create_time (%v) or update_time (%v)", test.req, created.CreateTime, created.UpdateTime)
			} else if !created.CreateTime.AsTime().Equal(created.UpdateTime.AsTime()) {
				t.Errorf("CreateApiVersion(%+v) returned unexpected timestamps: create_time %v != update_time %v", test.req, created.CreateTime, created.UpdateTime)
			}

			t.Run("GetApiVersion", func(t *testing.T) {
				req := &rpc.GetApiVersionRequest{
					Name: created.GetName(),
				}

				got, err := server.GetApiVersion(ctx, req)
				if err != nil {
					t.Fatalf("GetApiVersion(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(created, got, opts) {
					t.Errorf("GetApiVersion(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(created, got, opts))
				}
			})
		})
	}
}

func TestCreateApiVersionResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Api
		req  *rpc.CreateApiVersionRequest
		want codes.Code
	}{
		{
			desc: "invalid parent",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent: "invalid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "parent not found",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/other-api",
				ApiVersionId: "valid-id",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "valid-id",
				ApiVersion:   nil,
			},
			want: codes.OK,
		},
		{
			desc: "missing custom identifier",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "this-identifier-is-invalid-because-it-exceeds-the-eighty-character-maximum-length",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier underscores",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "underscore_identifier",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen prefix",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "-identifier",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen suffix",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "identifier-",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier uuid format",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "072d2288-c685-42d8-9df0-5edbb2a809ea",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier mixed case",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/locations/global/apis/my-api",
				ApiVersionId: "IDentifier",
				ApiVersion:   &rpc.ApiVersion{},
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

			if _, err := server.CreateApiVersion(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestCreateApiVersionDuplicates(t *testing.T) {
	test := struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.CreateApiVersionRequest
		want codes.Code
	}{
		desc: "case sensitive",
		seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
		req: &rpc.CreateApiVersionRequest{
			Parent:       "projects/my-project/locations/global/apis/my-api",
			ApiVersionId: "v1",
			ApiVersion:   &rpc.ApiVersion{},
		},
		want: codes.AlreadyExists,
	}
	t.Run(test.desc, func(t *testing.T) {
		ctx := context.Background()
		server := defaultTestServer(t)
		if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
			t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
		}

		if _, err := server.CreateApiVersion(ctx, test.req); status.Code(err) != test.want {
			t.Errorf("CreateApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
		}
	})
}

func TestGetApiVersion(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.GetApiVersionRequest
		want *rpc.ApiVersion
	}{
		{
			desc: "fully populated resource",
			seed: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/my-version",
				DisplayName: "My Display Name",
				Description: "My Description",
				State:       "My State",
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
			req: &rpc.GetApiVersionRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/my-version",
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/my-version",
				DisplayName: "My Display Name",
				Description: "My Description",
				State:       "My State",
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
			if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.GetApiVersion(ctx, test.req)
			if err != nil {
				t.Fatalf("GetApiVersion(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiVersion), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("GetApiVersion(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}
		})
	}
}

func TestGetApiVersionResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.GetApiVersionRequest
		want codes.Code
	}{
		{
			desc: "invalid name",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.GetApiVersionRequest{
				Name: "invalid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "resource not found",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.GetApiVersionRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.GetApiVersion(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApiVersions(t *testing.T) {
	tests := []struct {
		admin     bool
		desc      string
		seed      []*rpc.ApiVersion
		req       *rpc.ListApiVersionsRequest
		want      *rpc.ListApiVersionsResponse
		wantToken bool
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v3"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v1"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v2"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v3"},
				},
			},
		},
		{
			admin: true,
			desc:  "across all apis in a specific project",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v1"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v1"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/locations/global/apis/-",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
					{Name: "projects/my-project/locations/global/apis/other-api/versions/v1"},
				},
			},
		},
		{
			admin: true,
			desc:  "across all projects and apis",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
				{Name: "projects/other-project/locations/global/apis/other-api/versions/v1"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/-/locations/global/apis/-",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
					{Name: "projects/other-project/locations/global/apis/other-api/versions/v1"},
				},
			},
		},
		{
			admin: true,
			desc:  "in a specific api across all projects",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v1"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v1"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/-/locations/global/apis/my-api",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
					{Name: "projects/other-project/locations/global/apis/my-api/versions/v1"},
				},
			},
		},
		{
			desc: "custom page size",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v3"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent:   "projects/my-project/locations/global/apis/my-api",
				PageSize: 1,
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{},
				},
			},
			wantToken: true,
			// Ordering is not guaranteed by API, so any resource may be returned.
			extraOpts: protocmp.IgnoreFields(new(rpc.ApiVersion), "name"),
		},
		{
			desc: "name equality filtering",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v3"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api",
				Filter: "name == 'projects/my-project/locations/global/apis/my-api/versions/v2'",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v2"},
				},
			},
		},
		{
			desc: "description inequality filtering",
			seed: []*rpc.ApiVersion{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
					Description: "First ApiVersion",
				},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v3"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api",
				Filter: "description != ''",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
						Description: "First ApiVersion",
					},
				},
			},
		},
		{
			desc: "ordered by description",
			seed: []*rpc.ApiVersion{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
					Description: "111: this should be returned first",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v2",
					Description: "333: this should be returned third",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v3",
					Description: "222: this should be returned second",
				},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "description",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
						Description: "111: this should be returned first",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v3",
						Description: "222: this should be returned second",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v2",
						Description: "333: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by description descending",
			seed: []*rpc.ApiVersion{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
					Description: "111: this should be returned third",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v2",
					Description: "333: this should be returned first",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v3",
					Description: "222: this should be returned second",
				},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "description desc",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v2",
						Description: "333: this should be returned first",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v3",
						Description: "222: this should be returned second",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
						Description: "111: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by description then by name",
			seed: []*rpc.ApiVersion{
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
					Description: "222: this should be returned second or third (the name is the tie-breaker)",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v3",
					Description: "111: this should be returned first",
				},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v2",
					Description: "222: this should be returned second or third (the name is the tie-breaker)",
				},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "description,name",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v3",
						Description: "111: this should be returned first",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
						Description: "222: this should be returned second or third (the name is the tie-breaker)",
					},
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v2",
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
			if err := seeder.SeedVersions(ctx, server, test.seed...); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.ListApiVersions(ctx, test.req)
			if err != nil {
				t.Fatalf("ListApiVersions(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListApiVersionsResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.ApiVersion), "create_time", "update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListApiVersions(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListApiVersions(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				t.Errorf("ListApiVersions(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListApiVersionsResponseCodes(t *testing.T) {
	tests := []struct {
		admin bool
		desc  string
		seed  *rpc.ApiVersion
		req   *rpc.ListApiVersionsRequest
		want  codes.Code
	}{
		{
			desc: "invalid parent",
			req: &rpc.ListApiVersionsRequest{
				Parent: "invalid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "parent api not found",
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api",
			},
			want: codes.NotFound,
		},
		{
			admin: true,
			desc:  "parent project not found",
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/locations/global/apis/-",
			},
			want: codes.NotFound,
		},
		{
			desc: "negative page size",
			req: &rpc.ListApiVersionsRequest{
				PageSize: -1,
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid filter",
			req: &rpc.ListApiVersionsRequest{
				Filter: "this filter is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid page token",
			req: &rpc.ListApiVersionsRequest{
				PageToken: "this token is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering by unknown field",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.ListApiVersionsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "something",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering by private field",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.ListApiVersionsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "key",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering direction",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.ListApiVersionsRequest{
				Parent:  "projects/my-project/locations/global/apis/my-api",
				OrderBy: "description asc",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering format",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.ListApiVersionsRequest{
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
			if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.ListApiVersions(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("ListApiVersions(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApiVersionsSequence(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := []*rpc.ApiVersion{
		{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
		{Name: "projects/my-project/locations/global/apis/my-api/versions/v2"},
		{Name: "projects/my-project/locations/global/apis/my-api/versions/v3"},
	}
	if err := seeder.SeedVersions(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	listed := make([]*rpc.ApiVersion, 0, 3)

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListApiVersionsRequest{
			Parent:   "projects/my-project/locations/global/apis/my-api",
			PageSize: 1,
		}

		got, err := server.ListApiVersions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiVersions(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiVersions()); count != 1 {
			t.Errorf("ListApiVersions(%+v) returned %d versions, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListApiVersions(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.ApiVersions...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test intermediate page after failure on first page")
	}

	t.Run("intermediate page", func(t *testing.T) {
		req := &rpc.ListApiVersionsRequest{
			Parent:    "projects/my-project/locations/global/apis/my-api",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiVersions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiVersions(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiVersions()); count != 1 {
			t.Errorf("ListApiVersions(%+v) returned %d versions, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListApiVersions(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.ApiVersions...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test final page after failure on intermediate page")
	}

	t.Run("final page", func(t *testing.T) {
		req := &rpc.ListApiVersionsRequest{
			Parent:    "projects/my-project/locations/global/apis/my-api",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiVersions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiVersions(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiVersions()); count != 1 {
			t.Errorf("ListApiVersions(%+v) returned %d versions, expected exactly one", req, count)
		}

		if got.GetNextPageToken() != "" {
			t.Errorf("ListApiVersions(%+v) returned next_page_token, expected no next page", req)
		}

		listed = append(listed, got.ApiVersions...)
	})

	if t.Failed() {
		t.Fatal("Cannot test sequence result after failure on final page")
	}

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(rpc.ApiVersion), "create_time", "update_time"),
		cmpopts.SortSlices(func(a, b *rpc.ApiVersion) bool {
			return a.GetName() < b.GetName()
		}),
	}

	if !cmp.Equal(seed, listed, opts) {
		t.Errorf("List sequence returned unexpected diff (-want +got):\n%s", cmp.Diff(seed, listed, opts))
	}
}

func TestListApiVersionsLargeCollection(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := make([]*rpc.ApiVersion, 0, 1001)
	for i := 1; i <= cap(seed); i++ {
		seed = append(seed, &rpc.ApiVersion{
			Name: fmt.Sprintf("projects/my-project/locations/global/apis/my-api/versions/v%03d", i),
		})
	}

	if err := seeder.SeedVersions(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	// This test prevents the list sequence from ending before a known filter match is listed.
	// For simplicity, it does not guarantee the resource is returned on a later page.
	t.Run("filter", func(t *testing.T) {
		req := &rpc.ListApiVersionsRequest{
			Parent:   "projects/my-project/locations/global/apis/my-api",
			PageSize: 1,
			Filter:   "name == 'projects/my-project/locations/global/apis/my-api/versions/v099'",
		}

		got, err := server.ListApiVersions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiVersions(%+v) returned error: %s", req, err)
		}

		if len(got.GetApiVersions()) == 1 && got.GetNextPageToken() != "" {
			t.Errorf("ListApiVersions(%+v) returned a page token when the only matching resource has been listed: %+v", req, got)
		} else if len(got.GetApiVersions()) == 0 && got.GetNextPageToken() == "" {
			t.Errorf("ListApiVersions(%+v) returned an empty next page token before listing the only matching resource", req)
		} else if count := len(got.GetApiVersions()); count > 1 {
			t.Errorf("ListApiVersions(%+v) returned %d projects, expected at most one: %+v", req, count, got.GetApiVersions())
		}
	})

	t.Run("max page size", func(t *testing.T) {
		req := &rpc.ListApiVersionsRequest{
			Parent:   "projects/my-project/locations/global/apis/my-api",
			PageSize: 1001,
		}

		got, err := server.ListApiVersions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiVersions(%+v) returned error: %s", req, err)
		}

		if len(got.GetApiVersions()) != 1000 {
			t.Errorf("GetApiVersions(%+v) should have returned 1000 items, got: %+v", req, len(got.GetApiVersions()))
		} else if got.GetNextPageToken() == "" {
			t.Errorf("GetApiVersions(%+v) should return a next page token", req)
		}
	})
}

func TestUpdateApiVersion(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.UpdateApiVersionRequest
		want *rpc.ApiVersion
	}{
		{
			desc: "allow missing updates existing resources",
			seed: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/a/versions/v",
				Description: "My ApiVersion",
			},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name:        "projects/my-project/locations/global/apis/a/versions/v",
					Description: "My Updated ApiVersion",
				},
				UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"description"}},
				AllowMissing: true,
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/a/versions/v",
				Description: "My Updated ApiVersion",
			},
		},
		{
			desc: "allow missing creates missing resources",
			seed: &rpc.ApiVersion{
				Name: "projects/my-project/locations/global/apis/a/versions/v-sibling",
			},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "projects/my-project/locations/global/apis/a/versions/v",
				},
				AllowMissing: true,
			},
			want: &rpc.ApiVersion{
				Name: "projects/my-project/locations/global/apis/a/versions/v",
			},
		},
		{
			desc: "implicit nil mask",
			seed: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My ApiVersion",
			},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
					Description: "My Updated ApiVersion",
				},
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My Updated ApiVersion",
			},
		},
		{
			desc: "implicit empty mask",
			seed: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My ApiVersion",
			},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
					Description: "My Updated ApiVersion",
				},
				UpdateMask: &fieldmaskpb.FieldMask{},
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My Updated ApiVersion",
			},
		},
		{
			desc: "field specific mask",
			seed: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My ApiVersion",
			},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
					DisplayName: "Ignored",
					Description: "My Updated ApiVersion",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"description"}},
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My Updated ApiVersion",
			},
		},
		{
			desc: "full replacement wildcard mask",
			seed: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My ApiVersion",
			},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
					Description: "My Updated ApiVersion",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/locations/global/apis/my-api/versions/v1",
				DisplayName: "",
				Description: "My Updated ApiVersion",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			updated, err := server.UpdateApiVersion(ctx, test.req)
			if err != nil {
				t.Fatalf("UpdateApiVersion(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiVersion), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, updated, opts) {
				t.Errorf("UpdateApiVersion(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, updated, opts))
			}

			t.Run("GetApiVersion", func(t *testing.T) {
				req := &rpc.GetApiVersionRequest{
					Name: updated.GetName(),
				}

				got, err := server.GetApiVersion(ctx, req)
				if err != nil {
					t.Fatalf("GetApiVersion(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(updated, got, opts) {
					t.Errorf("GetApiVersion(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(updated, got, opts))
				}
			})
		})
	}
}

func TestUpdateApiVersionResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.UpdateApiVersionRequest
		want codes.Code
	}{
		{
			desc: "invalid name",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "invalid",
				},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "resource not found",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "projects/my-project/locations/global/apis/my-api/versions/doesnt-exist",
				},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req:  &rpc.UpdateApiVersionRequest{},
			want: codes.InvalidArgument,
		},
		{
			desc: "missing resource name",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "nonexistent field in mask",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1",
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
			if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.UpdateApiVersion(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestUpdateApiVersionSequence(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.UpdateApiVersionRequest
		want codes.Code
	}{
		{
			desc: "create using update with allow_missing=false",
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "projects/my-project/locations/global/apis/a/versions/v",
				},
				AllowMissing: false,
			},
			want: codes.NotFound,
		},
		{
			desc: "create using update with allow_missing=true",
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "projects/my-project/locations/global/apis/a/versions/v",
				},
				AllowMissing: true,
			},
			want: codes.OK,
		},
		{
			desc: "update existing resource with allow_missing=true",
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "projects/my-project/locations/global/apis/a/versions/v",
				},
				AllowMissing: true,
			},
			want: codes.OK,
		},
		{
			desc: "update existing resource with allow_missing=false",
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "projects/my-project/locations/global/apis/a/versions/v",
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
	var updateTime time.Time
	// NOTE: in the following sequence of tests, each test depends on its predecessor.
	// Resources are successively created and updated using the "Update" RPC and the
	// tests verify that CreateTime/UpdateTime fields are modified appropriately.
	for i, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var result *rpc.ApiVersion
			var err error
			if result, err = server.UpdateApiVersion(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
			if result != nil {
				if i == 1 {
					createTime = result.CreateTime.AsTime()
					updateTime = result.UpdateTime.AsTime()
				} else {
					if !createTime.Equal(result.CreateTime.AsTime()) {
						t.Errorf("UpdateApiVersion create time changed after update (%v %v)", createTime, result.CreateTime.AsTime())
					}
					if !updateTime.Before(result.UpdateTime.AsTime()) {
						t.Errorf("UpdateApiVersion update time did not increase after update (%v %v)", updateTime, result.UpdateTime.AsTime())
					}
					updateTime = result.UpdateTime.AsTime()
				}
			}
		})
	}
}

func TestDeleteApiVersion(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.DeleteApiVersionRequest
	}{
		{
			desc: "existing resource",
			seed: &rpc.ApiVersion{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1",
			},
			req: &rpc.DeleteApiVersionRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.DeleteApiVersion(ctx, test.req); err != nil {
				t.Fatalf("DeleteApiVersion(%+v) returned error: %s", test.req, err)
			}

			t.Run("GetApiVersion", func(t *testing.T) {
				req := &rpc.GetApiVersionRequest{
					Name: test.req.GetName(),
				}

				if _, err := server.GetApiVersion(ctx, req); status.Code(err) != codes.NotFound {
					t.Fatalf("GetApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), codes.NotFound, err)
				}
			})
		})
	}
}

func TestDeleteApiVersionResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.DeleteApiVersionRequest
		want codes.Code
	}{
		{
			desc: "invalid name",
			req: &rpc.DeleteApiVersionRequest{
				Name: "invalid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "resource not found",
			req: &rpc.DeleteApiVersionRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "resource has children",
			seed: &rpc.Artifact{
				Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/artifacts/my-artifact",
			},
			req: &rpc.DeleteApiVersionRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/my-version",
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

			if _, err := server.DeleteApiVersion(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("DeleteApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestDeleteApiVersionCascading(t *testing.T) {
	var (
		ctx    = context.Background()
		server = defaultTestServer(t)
		spec   = &rpc.ApiSpec{
			Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec",
		}
		artifact = &rpc.Artifact{
			Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/artifacts/my-artifact",
		}
	)

	if err := seeder.SeedRegistry(ctx, server, spec, artifact); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	req := &rpc.DeleteApiVersionRequest{
		Name:  "projects/my-project/locations/global/apis/my-api/versions/my-version",
		Force: true,
	}

	if _, err := server.DeleteApiVersion(ctx, req); err != nil {
		t.Errorf("DeleteApiVersion(%+v) returned error: %s", req, err)
	}

	if _, err := server.GetApiVersion(ctx, &rpc.GetApiVersionRequest{Name: req.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetApiVersion(%q) returned status code %q, want %q: %s", req.GetName(), status.Code(err), codes.NotFound, err)
	}

	if _, err := server.GetApiSpec(ctx, &rpc.GetApiSpecRequest{Name: spec.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetApiSpec(%q) returned status code %q, want %q: %s", spec.GetName(), status.Code(err), codes.NotFound, err)
	}

	if _, err := server.GetArtifact(ctx, &rpc.GetArtifactRequest{Name: artifact.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetArtifact(%q) returned status code %q, want %q: %s", artifact.GetName(), status.Code(err), codes.NotFound, err)
	}
}
