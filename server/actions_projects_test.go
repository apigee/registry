package server

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func seedProjects(ctx context.Context, t *testing.T, s *RegistryServer, projects ...*rpc.Project) {
	t.Helper()

	for _, p := range projects {
		name, err := names.ParseProject(p.Name)
		if err != nil {
			t.Fatalf("Setup/Seeding: ParseProject(%q) returned error: %s", p.Name, err)
		}

		req := &rpc.CreateProjectRequest{
			ProjectId: name.ProjectID,
			Project:   p,
		}

		switch _, err := s.CreateProject(ctx, req); status.Code(err) {
		case codes.OK, codes.AlreadyExists:
			// Project is now ready for use in test.
		default:
			t.Fatalf("Setup/Seeding: CreateProject(%+v) returned error: %s", req, err)
		}
	}
}

func TestCreateProject(t *testing.T) {
	tests := []struct {
		desc      string
		req       *rpc.CreateProjectRequest
		want      *rpc.Project
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			req: &rpc.CreateProjectRequest{
				Project: &rpc.Project{
					DisplayName: "My Project",
					Description: "Project for my APIs",
				},
			},
			want: &rpc.Project{
				DisplayName: "My Project",
				Description: "Project for my APIs",
			},
			// Name field is generated.
			extraOpts: protocmp.IgnoreFields(new(rpc.Project), "name"),
		},
		{
			desc: "custom identifier",
			req: &rpc.CreateProjectRequest{
				ProjectId: "my-project",
				Project:   &rpc.Project{},
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
				test.extraOpts,
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateProject(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if !strings.HasPrefix(created.GetName(), "projects/") {
				t.Errorf("CreateProject(%+v) returned unexpected name %q, expected collection prefix", test.req, created.GetName())
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
	t.Skip("Validation rules are not implemented")

	tests := []struct {
		desc string
		req  *rpc.CreateProjectRequest
		want codes.Code
	}{
		{
			desc: "short custom identifier",
			req: &rpc.CreateProjectRequest{
				ProjectId: "abc",
				Project:   &rpc.Project{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			req: &rpc.CreateProjectRequest{
				ProjectId: "this-identifier-exceeds-the-sixty-three-character-maximum-length",
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
			desc: "customer identifier dots",
			req: &rpc.CreateProjectRequest{
				ProjectId: "dot.identifier",
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
	ctx := context.Background()
	server := defaultTestServer(t)
	seedProjects(ctx, t, server, &rpc.Project{
		Name: "projects/my-project",
	})

	t.Run("case sensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateProjectRequest{
			ProjectId: "my-project",
			Project:   &rpc.Project{},
		}

		if _, err := server.CreateProject(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateProject(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
		}
	})

	t.Skip("Resource names are not yet case insensitive")
	t.Run("case insensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateProjectRequest{
			ProjectId: "My-Project",
			Project:   &rpc.Project{},
		}

		if _, err := server.CreateProject(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateProject(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
		}
	})
}

func TestGetProjectResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.GetProjectRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.GetProjectRequest{
				Name: "projects/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.GetProject(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListProjects(t *testing.T) {
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
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedProjects(ctx, t, server, test.seed...)

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
				// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
				t.Logf("ListProjects(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListProjectsResponseCodes(t *testing.T) {
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
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := []*rpc.Project{
		{Name: "projects/project1"},
		{Name: "projects/project2"},
		{Name: "projects/project3"},
	}
	seedProjects(ctx, t, server, seed...)

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

		if got.GetNextPageToken() != "" {
			// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
			t.Logf("ListProjects(%+v) returned next_page_token, expected no next page", req)
		}

		listed = append(listed, got.Projects...)
	})

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
	ctx := context.Background()
	server := defaultTestServer(t)
	for i := 1; i <= 100; i++ {
		seedProjects(ctx, t, server, &rpc.Project{
			Name: fmt.Sprintf("projects/project%d", i),
		})
	}

	req := &rpc.ListProjectsRequest{
		PageSize: 1,
		Filter:   "name == 'projects/project99'",
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
}

func TestUpdateProject(t *testing.T) {
	t.Skip("Default/empty mask behavior is incorrect and replacement wildcard is not implemented")

	tests := []struct {
		desc string
		seed *rpc.Project
		req  *rpc.UpdateProjectRequest
		want *rpc.Project
	}{
		{
			desc: "default parameters",
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
			seedProjects(ctx, t, server, test.seed)

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
	t.Skip("Update mask validation is not implemented")

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
			seedProjects(ctx, t, server, test.seed)

			if _, err := server.UpdateProject(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Project
		req  *rpc.DeleteProjectRequest
	}{
		{
			desc: "existing project",
			seed: &rpc.Project{
				Name: "projects/my-project",
			},
			req: &rpc.DeleteProjectRequest{
				Name: "projects/my-project",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedProjects(ctx, t, server, test.seed)

			if _, err := server.DeleteProject(ctx, test.req); err != nil {
				t.Fatalf("DeleteProject(%+v) returned error: %s", test.req, err)
			}

			t.Run("GetProject", func(t *testing.T) {
				req := &rpc.GetProjectRequest{
					Name: test.req.GetName(),
				}

				if _, err := server.GetProject(ctx, req); status.Code(err) != codes.NotFound {
					t.Fatalf("GetProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), codes.NotFound, err)
				}
			})
		})
	}
}

func TestDeleteProjectResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.DeleteProjectRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.DeleteProjectRequest{
				Name: "projects/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.DeleteProject(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("DeleteProject(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}
