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

func TestCreateProject(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		req  *rpc.CreateProjectRequest
		want *rpc.Project
	}{
		{
			desc: "fully populated resource",
			req: &rpc.CreateProjectRequest{
				ProjectId: "my-project",
				Project: &rpc.Project{
					DisplayName: "My Display Name",
					Description: "My Description",
				},
			},
			want: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Display Name",
				Description: "My Description",
			},
		},
		{
			desc: "empty resource",
			req: &rpc.CreateProjectRequest{
				ProjectId: "my-project",
				Project:   nil,
			},
			want: &rpc.Project{
				Name: "projects/my-project",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			created, err := server.CreateProject(ctx, test.req)
			if err != nil {
				t.Fatalf("CreateProject(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Project), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateProject(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if created.CreateTime == nil || created.UpdateTime == nil {
				t.Errorf("CreateProject(%+v) returned unset create_time (%v) or update_time (%v)", test.req, created.CreateTime, created.UpdateTime)
			} else if !created.CreateTime.AsTime().Equal(created.UpdateTime.AsTime()) {
				t.Errorf("CreateProject(%+v) returned unexpected timestamps: create_time %v != update_time %v", test.req, created.CreateTime, created.UpdateTime)
			}

			t.Run("GetProject", func(t *testing.T) {
				req := &rpc.GetProjectRequest{
					Name: created.GetName(),
				}

				got, err := server.GetProject(ctx, req)
				if err != nil {
					t.Fatalf("GetProject(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(created, got, opts) {
					t.Errorf("GetProject(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(created, got, opts))
				}
			})
		})
	}
}

func TestCreateProjectResponseCodes(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		req  *rpc.CreateProjectRequest
		want codes.Code
	}{
		{
			desc: "missing resource body",
			req: &rpc.CreateProjectRequest{
				ProjectId: "valid-id",
				Project:   nil,
			},
			want: codes.OK,
		},
		{
			desc: "missing custom identifier",
			req: &rpc.CreateProjectRequest{
				ProjectId: "",
				Project:   &rpc.Project{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			req: &rpc.CreateProjectRequest{
				ProjectId: "this-identifier-is-invalid-because-it-exceeds-the-eighty-character-maximum-length",
				Project:   &rpc.Project{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier underscores",
			req: &rpc.CreateProjectRequest{
				ProjectId: "underscore_identifier",
				Project:   &rpc.Project{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen prefix",
			req: &rpc.CreateProjectRequest{
				ProjectId: "-identifier",
				Project:   &rpc.Project{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen suffix",
			req: &rpc.CreateProjectRequest{
				ProjectId: "identifier-",
				Project:   &rpc.Project{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier uuid format",
			req: &rpc.CreateProjectRequest{
				ProjectId: "072d2288-c685-42d8-9df0-5edbb2a809ea",
				Project:   &rpc.Project{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier mixed case",
			req: &rpc.CreateProjectRequest{
				ProjectId: "IDentifier",
				Project:   &rpc.Project{},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.CreateProject(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestCreateProjectDuplicates(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	test := struct {
		desc string
		seed *rpc.Project
		req  *rpc.CreateProjectRequest
		want codes.Code
	}{
		desc: "case sensitive",
		seed: &rpc.Project{Name: "projects/my-project"},
		req: &rpc.CreateProjectRequest{
			ProjectId: "my-project",
			Project:   &rpc.Project{},
		},
		want: codes.AlreadyExists,
	}
	t.Run(test.desc, func(t *testing.T) {
		ctx := context.Background()
		server := defaultTestServer(t)
		if err := seeder.SeedProjects(ctx, server, test.seed); err != nil {
			t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
		}

		if _, err := server.CreateProject(ctx, test.req); status.Code(err) != test.want {
			t.Errorf("CreateProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
		}
	})
}

func TestGetProject(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		seed *rpc.Project
		req  *rpc.GetProjectRequest
		want *rpc.Project
	}{
		{
			desc: "fully populated resource",
			seed: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Display Name",
				Description: "My Description",
			},
			req: &rpc.GetProjectRequest{
				Name: "projects/my-project",
			},
			want: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Display Name",
				Description: "My Description",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedProjects(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.GetProject(ctx, test.req)
			if err != nil {
				t.Fatalf("GetProject(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Project), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("GetProject(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}
		})
	}
}

func TestGetProjectResponseCodes(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		seed *rpc.Project
		req  *rpc.GetProjectRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.GetProjectRequest{
				Name: "projects/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "invalid name",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.GetProjectRequest{
				Name: "invalid",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedProjects(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.GetProject(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListProjects(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc      string
		seed      []*rpc.Project
		req       *rpc.ListProjectsRequest
		want      *rpc.ListProjectsResponse
		wantToken bool
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			seed: []*rpc.Project{
				{Name: "projects/project1"},
				{Name: "projects/project2"},
				{Name: "projects/project3"},
			},
			req: &rpc.ListProjectsRequest{},
			want: &rpc.ListProjectsResponse{
				Projects: []*rpc.Project{
					{Name: "projects/project1"},
					{Name: "projects/project2"},
					{Name: "projects/project3"},
				},
			},
		},
		{
			desc: "custom page size",
			seed: []*rpc.Project{
				{Name: "projects/project1"},
				{Name: "projects/project2"},
				{Name: "projects/project3"},
			},
			req: &rpc.ListProjectsRequest{
				PageSize: 1,
			},
			want: &rpc.ListProjectsResponse{
				Projects: []*rpc.Project{
					{},
				},
			},
			wantToken: true,
			// Ordering is not guaranteed by API, so any resource may be returned.
			extraOpts: protocmp.IgnoreFields(new(rpc.Project), "name"),
		},
		{
			desc: "name equality filtering",
			seed: []*rpc.Project{
				{Name: "projects/project1"},
				{Name: "projects/project2"},
				{Name: "projects/project3"},
			},
			req: &rpc.ListProjectsRequest{
				Filter: "name == 'projects/project2'",
			},
			want: &rpc.ListProjectsResponse{
				Projects: []*rpc.Project{
					{Name: "projects/project2"},
				},
			},
		},
		{
			desc: "description inequality filtering",
			seed: []*rpc.Project{
				{
					Name:        "projects/project1",
					Description: "First Project",
				},
				{Name: "projects/project2"},
				{Name: "projects/project3"},
			},
			req: &rpc.ListProjectsRequest{
				Filter: "description != ''",
			},
			want: &rpc.ListProjectsResponse{
				Projects: []*rpc.Project{
					{
						Name:        "projects/project1",
						Description: "First Project",
					},
				},
			},
		},
		{
			desc: "ordered by description",
			seed: []*rpc.Project{
				{
					Name:        "projects/project1",
					Description: "111: this should be returned first",
				},
				{
					Name:        "projects/project2",
					Description: "333: this should be returned third",
				},
				{
					Name:        "projects/project3",
					Description: "222: this should be returned second",
				},
			},
			req: &rpc.ListProjectsRequest{
				OrderBy: "description",
			},
			want: &rpc.ListProjectsResponse{
				Projects: []*rpc.Project{
					{
						Name:        "projects/project1",
						Description: "111: this should be returned first",
					},
					{
						Name:        "projects/project3",
						Description: "222: this should be returned second",
					},
					{
						Name:        "projects/project2",
						Description: "333: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by description descending",
			seed: []*rpc.Project{
				{
					Name:        "projects/project1",
					Description: "111: this should be returned third",
				},
				{
					Name:        "projects/project2",
					Description: "333: this should be returned first",
				},
				{
					Name:        "projects/project3",
					Description: "222: this should be returned second",
				},
			},
			req: &rpc.ListProjectsRequest{
				OrderBy: "description desc",
			},
			want: &rpc.ListProjectsResponse{
				Projects: []*rpc.Project{
					{
						Name:        "projects/project2",
						Description: "333: this should be returned first",
					},
					{
						Name:        "projects/project3",
						Description: "222: this should be returned second",
					},
					{
						Name:        "projects/project1",
						Description: "111: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by description then by name",
			seed: []*rpc.Project{
				{
					Name:        "projects/project1",
					Description: "222: this should be returned second or third (the name is the tie-breaker)",
				},
				{
					Name:        "projects/project3",
					Description: "111: this should be returned first",
				},
				{
					Name:        "projects/project2",
					Description: "222: this should be returned second or third (the name is the tie-breaker)",
				},
			},
			req: &rpc.ListProjectsRequest{
				OrderBy: "description,name",
			},
			want: &rpc.ListProjectsResponse{
				Projects: []*rpc.Project{
					{
						Name:        "projects/project3",
						Description: "111: this should be returned first",
					},
					{
						Name:        "projects/project1",
						Description: "222: this should be returned second or third (the name is the tie-breaker)",
					},
					{
						Name:        "projects/project2",
						Description: "222: this should be returned second or third (the name is the tie-breaker)",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedProjects(ctx, server, test.seed...); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.ListProjects(ctx, test.req)
			if err != nil {
				t.Fatalf("ListProjects(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListProjectsResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.Project), "create_time", "update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListProjects(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListProjects(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				t.Errorf("ListProjects(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListProjectsResponseCodes(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		req  *rpc.ListProjectsRequest
		want codes.Code
	}{
		{
			desc: "negative page size",
			req: &rpc.ListProjectsRequest{
				PageSize: -1,
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid filter",
			req: &rpc.ListProjectsRequest{
				Filter: "this filter is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid page token",
			req: &rpc.ListProjectsRequest{
				PageToken: "this token is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering by unknown field",
			req: &rpc.ListProjectsRequest{
				OrderBy: "something",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering by private field",
			req: &rpc.ListProjectsRequest{
				OrderBy: "key",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering direction",
			req: &rpc.ListProjectsRequest{
				OrderBy: "description asc",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering format",
			req: &rpc.ListProjectsRequest{
				OrderBy: "description,",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.ListProjects(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("ListProjects(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListProjectsSequence(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := []*rpc.Project{
		{Name: "projects/project1"},
		{Name: "projects/project2"},
		{Name: "projects/project3"},
	}

	if err := seeder.SeedProjects(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	listed := make([]*rpc.Project, 0, 3)

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListProjectsRequest{
			PageSize: 1,
		}

		got, err := server.ListProjects(ctx, req)
		if err != nil {
			t.Fatalf("ListProjects(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetProjects()); count != 1 {
			t.Errorf("ListProjects(%+v) returned %d projects, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListProjects(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.Projects...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test intermediate page after failure on first page")
	}

	t.Run("intermediate page", func(t *testing.T) {
		req := &rpc.ListProjectsRequest{
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListProjects(ctx, req)
		if err != nil {
			t.Fatalf("ListProjects(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetProjects()); count != 1 {
			t.Errorf("ListProjects(%+v) returned %d projects, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListProjects(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.Projects...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test final page after failure on intermediate page")
	}

	t.Run("final page", func(t *testing.T) {
		req := &rpc.ListProjectsRequest{
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListProjects(ctx, req)
		if err != nil {
			t.Fatalf("ListProjects(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetProjects()); count != 1 {
			t.Errorf("ListProjects(%+v) returned %d projects, expected exactly one", req, count)
		}

		if got.GetNextPageToken() != "" {
			t.Errorf("ListProjects(%+v) returned next_page_token, expected no next page", req)
		}

		listed = append(listed, got.Projects...)
	})

	if t.Failed() {
		t.Fatal("Cannot test sequence result after failure on final page")
	}

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(rpc.Project), "create_time", "update_time"),
		cmpopts.SortSlices(func(a, b *rpc.Project) bool {
			return a.GetName() < b.GetName()
		}),
	}

	if !cmp.Equal(seed, listed, opts) {
		t.Errorf("List sequence returned unexpected diff (-want +got):\n%s", cmp.Diff(seed, listed, opts))
	}
}

// This test prevents the list sequence from ending before a known filter match is listed.
// For simplicity, it does not guarantee the resource is returned on a later page.
func TestListProjectsLargeCollectionFiltering(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := make([]*rpc.Project, 0, 1001)
	for i := 1; i <= cap(seed); i++ {
		seed = append(seed, &rpc.Project{
			Name: fmt.Sprintf("projects/project%03d", i),
		})
	}

	if err := seeder.SeedProjects(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	t.Run("filter", func(t *testing.T) {
		req := &rpc.ListProjectsRequest{
			PageSize: 1,
			Filter:   "name == 'projects/project099'",
		}

		got, err := server.ListProjects(ctx, req)
		if err != nil {
			t.Fatalf("ListProjects(%+v) returned error: %s", req, err)
		}

		if len(got.GetProjects()) == 1 && got.GetNextPageToken() != "" {
			t.Errorf("ListProjects(%+v) returned a page token when the only matching resource has been listed: %+v", req, got)
		} else if len(got.GetProjects()) == 0 && got.GetNextPageToken() == "" {
			t.Errorf("ListProjects(%+v) returned an empty next page token before listing the only matching resource", req)
		} else if count := len(got.GetProjects()); count > 1 {
			t.Errorf("ListProjects(%+v) returned %d projects, expected at most one: %+v", req, count, got.GetProjects())
		}
	})

	t.Run("max page size", func(t *testing.T) {
		req := &rpc.ListProjectsRequest{
			PageSize: 1001,
		}

		got, err := server.ListProjects(ctx, req)
		if err != nil {
			t.Fatalf("ListApiDeployments(%+v) returned error: %s", req, err)
		}

		if len(got.GetProjects()) != 1000 {
			t.Errorf("GetProjects(%+v) should have returned 1000 items, got: %+v", req, len(got.GetProjects()))
		} else if got.GetNextPageToken() == "" {
			t.Errorf("GetProjects(%+v) should return a next page token", req)
		}
	})
}

func TestUpdateProject(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		seed *rpc.Project
		req  *rpc.UpdateProjectRequest
		want *rpc.Project
	}{
		{
			desc: "allow missing updates existing resources",
			seed: &rpc.Project{
				Name:        "projects/my-project",
				Description: "My Project",
			},
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name:        "projects/my-project",
					Description: "My Updated Project",
				},
				UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"description"}},
				AllowMissing: true,
			},
			want: &rpc.Project{
				Name:        "projects/my-project",
				Description: "My Updated Project",
			},
		},
		{
			desc: "allow missing creates missing resources",
			seed: &rpc.Project{
				Name: "projects/my-project-sibling",
			},
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name: "projects/my-project",
				},
				AllowMissing: true,
			},
			want: &rpc.Project{
				Name: "projects/my-project",
			},
		},
		{
			desc: "implicit nil mask",
			seed: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Project",
				Description: "Project for my APIs",
			},
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name:        "projects/my-project",
					DisplayName: "My Updated Project",
				},
			},
			want: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Updated Project",
				Description: "Project for my APIs",
			},
		},
		{
			desc: "implicit empty mask",
			seed: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Project",
				Description: "Project for my APIs",
			},
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name:        "projects/my-project",
					DisplayName: "My Updated Project",
				},
				UpdateMask: &fieldmaskpb.FieldMask{},
			},
			want: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Updated Project",
				Description: "Project for my APIs",
			},
		},
		{
			desc: "field specific mask",
			seed: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Project",
				Description: "Project for my APIs",
			},
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name:        "projects/my-project",
					DisplayName: "My Updated Project",
					Description: "Ignored",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
			},
			want: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Updated Project",
				Description: "Project for my APIs",
			},
		},
		{
			desc: "full replacement wildcard mask",
			seed: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Project",
				Description: "Project for my APIs",
			},
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name:        "projects/my-project",
					DisplayName: "My Updated Project",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			},
			want: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "My Updated Project",
				Description: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedProjects(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			updated, err := server.UpdateProject(ctx, test.req)
			if err != nil {
				t.Fatalf("UpdateProject(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Project), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, updated, opts) {
				t.Errorf("UpdateProject(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, updated, opts))
			}

			t.Run("GetProject", func(t *testing.T) {
				req := &rpc.GetProjectRequest{
					Name: updated.GetName(),
				}

				got, err := server.GetProject(ctx, req)
				if err != nil {
					t.Fatalf("GetProject(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(updated, got, opts) {
					t.Errorf("GetProject(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(updated, got, opts))
				}
			})
		})
	}
}

func TestUpdateProjectResponseCodes(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		seed *rpc.Project
		req  *rpc.UpdateProjectRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name: "projects/doesnt-exist",
				},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.Project{Name: "projects/my-project"},
			req:  &rpc.UpdateProjectRequest{},
			want: codes.InvalidArgument,
		},
		{
			desc: "missing resource name",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "nonexistent field in mask",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name: "projects/my-project",
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
			if err := seeder.SeedProjects(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.UpdateProject(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestUpdateProjectSequence(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		req  *rpc.UpdateProjectRequest
		want codes.Code
	}{
		{
			desc: "create using update with allow_missing=false",
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name: "projects/my-project",
				},
				AllowMissing: false,
			},
			want: codes.NotFound,
		},
		{
			desc: "create using update with allow_missing=true",
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name: "projects/my-project",
				},
				AllowMissing: true,
			},
			want: codes.OK,
		},
		{
			desc: "update existing resource with allow_missing=true",
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name: "projects/my-project",
				},
				AllowMissing: true,
			},
			want: codes.OK,
		},
		{
			desc: "update existing resource with allow_missing=false",
			req: &rpc.UpdateProjectRequest{
				Project: &rpc.Project{
					Name: "projects/my-project",
				},
				AllowMissing: false,
			},
			want: codes.OK,
		},
	}
	ctx := context.Background()
	server := defaultTestServer(t)
	var createTime time.Time
	var updateTime time.Time
	// NOTE: in the following sequence of tests, each test depends on its predecessor.
	// Resources are successively created and updated using the "Update" RPC and the
	// tests verify that CreateTime/UpdateTime fields are modified appropriately
	for i, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var result *rpc.Project
			var err error
			if result, err = server.UpdateProject(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
			if result != nil {
				if i == 1 {
					createTime = result.CreateTime.AsTime()
					updateTime = result.UpdateTime.AsTime()
				} else {
					if !createTime.Equal(result.CreateTime.AsTime()) {
						t.Errorf("UpdateProject create time changed after update (%v %v)", createTime, result.CreateTime.AsTime())
					}
					if !updateTime.Before(result.UpdateTime.AsTime()) {
						t.Errorf("UpdateProject update time did not increase after update (%v %v)", updateTime, result.UpdateTime.AsTime())
					}
					updateTime = result.UpdateTime.AsTime()
				}
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		seed *rpc.Project
		req  *rpc.DeleteProjectRequest
	}{
		{
			desc: "existing resource",
			seed: &rpc.Project{
				Name: "projects/my-project",
			},
			req: &rpc.DeleteProjectRequest{
				Name:  "projects/my-project",
				Force: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			artifact := &rpc.Artifact{
				Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec/artifacts/my-artifact",
			}
			if err := seeder.SeedArtifacts(ctx, server, artifact); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.DeleteProject(ctx, test.req); err != nil {
				t.Fatalf("DeleteProject(%+v) returned error: %s", test.req, err)
			}

			if _, err := server.GetProject(ctx, &rpc.GetProjectRequest{Name: test.req.GetName()}); status.Code(err) != codes.NotFound {
				t.Errorf("GetProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), codes.NotFound, err)
			}

			if _, err := server.GetArtifact(ctx, &rpc.GetArtifactRequest{Name: artifact.GetName()}); status.Code(err) != codes.NotFound {
				t.Errorf("GetArtifact(%q) returned status code %q, want %q: %s", artifact.GetName(), status.Code(err), codes.NotFound, err)
			}
		})
	}
}

func TestDeleteProjectResponseCodes(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.DeleteProjectRequest
		want codes.Code
	}{
		{
			desc: "invalid name",
			req: &rpc.DeleteProjectRequest{
				Name: "invalid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "resource not found",
			req: &rpc.DeleteProjectRequest{
				Name: "projects/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "resource has children",
			seed: &rpc.Artifact{
				Name: "projects/my-project/locations/global/artifacts/my-artifact",
			},
			req: &rpc.DeleteProjectRequest{
				Name: "projects/my-project",
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

			if _, err := server.DeleteProject(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("DeleteProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestDeleteProjectCascading(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	var (
		ctx    = context.Background()
		server = defaultTestServer(t)
		api    = &rpc.Api{
			Name: "projects/my-project/locations/global/apis/my-api",
		}
		artifact = &rpc.Artifact{
			Name: "projects/my-project/locations/global/artifacts/my-artifact",
		}
	)

	if err := seeder.SeedRegistry(ctx, server, api, artifact); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	req := &rpc.DeleteProjectRequest{
		Name:  "projects/my-project",
		Force: true,
	}

	if _, err := server.DeleteProject(ctx, req); err != nil {
		t.Errorf("DeleteProject(%+v) returned error: %s", req, err)
	}

	if _, err := server.GetProject(ctx, &rpc.GetProjectRequest{Name: req.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetProject(%q) returned status code %q, want %q: %s", req.GetName(), status.Code(err), codes.NotFound, err)
	}

	if _, err := server.GetApi(ctx, &rpc.GetApiRequest{Name: api.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetApi(%q) returned status code %q, want %q: %s", api.GetName(), status.Code(err), codes.NotFound, err)
	}

	if _, err := server.GetArtifact(ctx, &rpc.GetArtifactRequest{Name: artifact.GetName()}); status.Code(err) != codes.NotFound {
		t.Errorf("GetArtifact(%q) returned status code %q, want %q: %s", artifact.GetName(), status.Code(err), codes.NotFound, err)
	}
}
