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

func seedApis(ctx context.Context, t *testing.T, s *RegistryServer, apis ...*rpc.Api) {
	t.Helper()

	for _, p := range apis {
		m, err := names.ParseApi(p.Name)
		if err != nil {
			t.Fatalf("Setup/Seeding: ParseApi(%q) returned error: %s", p.Name, err)
		}

		parent := fmt.Sprintf("projects/%s", m[1])
		seedProjects(ctx, t, s, &rpc.Project{
			Name: parent,
		})

		req := &rpc.CreateApiRequest{
			Parent: parent,
			ApiId:  m[2],
			Api:    p,
		}

		switch _, err := s.CreateApi(ctx, req); status.Code(err) {
		case codes.OK, codes.AlreadyExists:
			// Api is now ready for use in test.
		default:
			t.Fatalf("Setup/Seeding: CreateApi(%+v) returned error: %s", req, err)
		}
	}
}

func TestCreateApi(t *testing.T) {
	tests := []struct {
		desc      string
		req       *rpc.CreateApiRequest
		want      *rpc.Api
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			req: &rpc.CreateApiRequest{
				Parent: "projects/p",
				Api: &rpc.Api{
					DisplayName: "My Api",
					Description: "Api for my versions",
				},
			},
			want: &rpc.Api{
				DisplayName: "My Api",
				Description: "Api for my versions",
			},
			// Name field is generated.
			extraOpts: protocmp.IgnoreFields(new(rpc.Api), "name"),
		},
		{
			desc: "custom identifier",
			req: &rpc.CreateApiRequest{
				Parent: "projects/p",
				ApiId:  "my-api",
				Api:    &rpc.Api{},
			},
			want: &rpc.Api{
				Name: "projects/p/apis/my-api",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			created, err := server.CreateApi(ctx, test.req)
			if err != nil {
				t.Fatalf("CreateApi(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Api), "create_time", "update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateApi(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if !strings.HasPrefix(created.GetName(), test.req.GetParent()+"/apis/") {
				t.Errorf("CreateApi(%+v) returned unexpected name %q, expected collection prefix", test.req, created.GetName())
			}

			if created.CreateTime == nil || created.UpdateTime == nil {
				t.Errorf("CreateApi(%+v) returned unset create_time (%v) or update_time (%v)", test.req, created.CreateTime, created.UpdateTime)
			} else if !created.CreateTime.AsTime().Equal(created.UpdateTime.AsTime()) {
				t.Errorf("CreateApi(%+v) returned unexpected timestamps: create_time %v != update_time %v", test.req, created.CreateTime, created.UpdateTime)
			}

			t.Run("GetApi", func(t *testing.T) {
				req := &rpc.GetApiRequest{
					Name: created.GetName(),
				}

				got, err := server.GetApi(ctx, req)
				if err != nil {
					t.Fatalf("GetApi(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(created, got, opts) {
					t.Errorf("GetApi(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(created, got, opts))
				}
			})
		})
	}
}

func TestCreateApiResponseCodes(t *testing.T) {
	t.Skip("Validation rules are not implemented")

	tests := []struct {
		desc string
		req  *rpc.CreateApiRequest
		want codes.Code
	}{
		{
			desc: "short custom identifier",
			req: &rpc.CreateApiRequest{
				ApiId: "abc",
				Api:   &rpc.Api{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			req: &rpc.CreateApiRequest{
				ApiId: "this-identifier-exceeds-the-sixty-three-character-maximum-length",
				Api:   &rpc.Api{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier underscores",
			req: &rpc.CreateApiRequest{
				ApiId: "underscore_identifier",
				Api:   &rpc.Api{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier dots",
			req: &rpc.CreateApiRequest{
				ApiId: "dot.identifier",
				Api:   &rpc.Api{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier uuid format",
			req: &rpc.CreateApiRequest{
				ApiId: "072d2288-c685-42d8-9df0-5edbb2a809ea",
				Api:   &rpc.Api{},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.CreateApi(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateApi(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestCreateApiDuplicates(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seedApis(ctx, t, server, &rpc.Api{
		Name: "projects/p/apis/a1",
	})

	t.Run("case sensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateApiRequest{
			Parent: "projects/p",
			ApiId:  "a1",
			Api:    &rpc.Api{},
		}

		if _, err := server.CreateApi(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateApi(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
		}
	})

	t.Skip("Resource names are not yet case insensitive")
	t.Run("case insensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateApiRequest{
			Parent: "projects/p",
			ApiId:  "A1",
			Api:    &rpc.Api{},
		}

		if _, err := server.CreateApi(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateApi(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
		}
	})
}

func TestGetApiResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.GetApiRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.GetApiRequest{
				Name: "projects/p/apis/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.GetApi(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetApi(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApis(t *testing.T) {
	tests := []struct {
		desc      string
		seed      []*rpc.Api
		req       *rpc.ListApisRequest
		want      *rpc.ListApisResponse
		wantToken bool
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			seed: []*rpc.Api{
				{Name: "projects/p/apis/a1"},
				{Name: "projects/p/apis/a2"},
				{Name: "projects/p/apis/a3"},
			},
			req: &rpc.ListApisRequest{
				Parent: "projects/p",
			},
			want: &rpc.ListApisResponse{
				Apis: []*rpc.Api{
					{Name: "projects/p/apis/a1"},
					{Name: "projects/p/apis/a2"},
					{Name: "projects/p/apis/a3"},
				},
			},
		},
		{
			desc: "custom page size",
			seed: []*rpc.Api{
				{Name: "projects/p/apis/a1"},
				{Name: "projects/p/apis/a2"},
				{Name: "projects/p/apis/a3"},
			},
			req: &rpc.ListApisRequest{
				Parent:   "projects/p",
				PageSize: 1,
			},
			want: &rpc.ListApisResponse{
				Apis: []*rpc.Api{
					{},
				},
			},
			wantToken: true,
			// Ordering is not guaranteed by API, so any resource may be returned.
			extraOpts: protocmp.IgnoreFields(new(rpc.Api), "name"),
		},
		{
			desc: "name equality filtering",
			seed: []*rpc.Api{
				{Name: "projects/p/apis/a1"},
				{Name: "projects/p/apis/a2"},
				{Name: "projects/p/apis/a3"},
			},
			req: &rpc.ListApisRequest{
				Parent: "projects/p",
				Filter: "name == 'projects/p/apis/a2'",
			},
			want: &rpc.ListApisResponse{
				Apis: []*rpc.Api{
					{Name: "projects/p/apis/a2"},
				},
			},
		},
		{
			desc: "description inequality filtering",
			seed: []*rpc.Api{
				{
					Name:        "projects/p/apis/a1",
					Description: "First Api",
				},
				{Name: "projects/p/apis/a2"},
				{Name: "projects/p/apis/a3"},
			},
			req: &rpc.ListApisRequest{
				Parent: "projects/p",
				Filter: "description != ''",
			},
			want: &rpc.ListApisResponse{
				Apis: []*rpc.Api{
					{
						Name:        "projects/p/apis/a1",
						Description: "First Api",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedApis(ctx, t, server, test.seed...)

			got, err := server.ListApis(ctx, test.req)
			if err != nil {
				t.Fatalf("ListApis(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListApisResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.Api), "create_time", "update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListApis(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListApis(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
				t.Logf("ListApis(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListApisResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.ListApisRequest
		want codes.Code
	}{
		{
			desc: "negative page size",
			req: &rpc.ListApisRequest{
				PageSize: -1,
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid filter",
			req: &rpc.ListApisRequest{
				Filter: "this filter is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid page token",
			req: &rpc.ListApisRequest{
				PageToken: "this token is not valid",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.ListApis(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("ListApis(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApisSequence(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := []*rpc.Api{
		{Name: "projects/p/apis/a1"},
		{Name: "projects/p/apis/a2"},
		{Name: "projects/p/apis/a3"},
	}
	seedApis(ctx, t, server, seed...)

	listed := make([]*rpc.Api, 0, 3)

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListApisRequest{
			Parent:   "projects/p",
			PageSize: 1,
		}

		got, err := server.ListApis(ctx, req)
		if err != nil {
			t.Fatalf("ListApis(%+v) returned error: %s", req, err)
		}

		listed = append(listed, got.Apis...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test intermediate page after failure on first page")
	}

	t.Run("intermediate page", func(t *testing.T) {
		req := &rpc.ListApisRequest{
			Parent:    "projects/p",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApis(ctx, req)
		if err != nil {
			t.Fatalf("ListApis(%+v) returned error: %s", req, err)
		}

		listed = append(listed, got.Apis...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test final page after failure on intermediate page")
	}

	t.Run("final page", func(t *testing.T) {
		req := &rpc.ListApisRequest{
			Parent:    "projects/p",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApis(ctx, req)
		if err != nil {
			t.Fatalf("ListApis(%+v) returned error: %s", req, err)
		}

		if got.GetNextPageToken() != "" {
			// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
			t.Logf("ListApis(%+v) returned next_page_token, expected no next page", req)
		}

		listed = append(listed, got.Apis...)
	})

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(rpc.Api), "create_time", "update_time"),
		cmpopts.SortSlices(func(a, b *rpc.Api) bool {
			return a.GetName() < b.GetName()
		}),
	}

	if !cmp.Equal(seed, listed, opts) {
		t.Errorf("List sequence returned unexpected diff (-want +got):\n%s", cmp.Diff(seed, listed, opts))
	}
}

// This test prevents the list sequence from ending before a known filter match is listed.
// For simplicity, it does not guarantee the resource is returned on a later page.
func TestListApisLargeCollectionFiltering(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	for i := 1; i <= 100; i++ {
		seedApis(ctx, t, server, &rpc.Api{
			Name: fmt.Sprintf("projects/p/apis/a%d", i),
		})
	}

	req := &rpc.ListApisRequest{
		Parent:   "projects/p",
		PageSize: 1,
		Filter:   "name == 'projects/p/apis/a99'",
	}

	got, err := server.ListApis(ctx, req)
	if err != nil {
		t.Fatalf("ListApis(%+v) returned error: %s", req, err)
	}

	if len(got.GetApis()) == 1 && got.GetNextPageToken() != "" {
		t.Errorf("ListApis(%+v) returned a page token when the only matching resource has been listed: %+v", req, got)
	} else if len(got.GetApis()) == 0 && got.GetNextPageToken() == "" {
		t.Errorf("ListApis(%+v) returned an empty next page token before listing the only matching resource", req)
	} else if count := len(got.GetApis()); count > 1 {
		t.Errorf("ListApis(%+v) returned %d projects, expected at most one: %+v", req, count, got.GetApis())
	}
}

func TestUpdateApi(t *testing.T) {
	t.Skip("Default/empty mask behavior is incorrect and replacement wildcard is not implemented")

	tests := []struct {
		desc string
		seed *rpc.Api
		req  *rpc.UpdateApiRequest
		want *rpc.Api
	}{
		{
			desc: "default parameters",
			seed: &rpc.Api{
				Name:        "projects/p/apis/p",
				DisplayName: "My Api",
				Description: "Api for my APIs",
			},
			req: &rpc.UpdateApiRequest{
				Api: &rpc.Api{
					Name:        "projects/p/apis/p",
					DisplayName: "My Updated Api",
				},
			},
			want: &rpc.Api{
				Name:        "projects/p/apis/p",
				DisplayName: "My Updated Api",
				Description: "Api for my APIs",
			},
		},
		{
			desc: "field specific mask",
			seed: &rpc.Api{
				Name:        "projects/p/apis/p",
				DisplayName: "My Api",
				Description: "Api for my APIs",
			},
			req: &rpc.UpdateApiRequest{
				Api: &rpc.Api{
					Name:        "projects/p/apis/p",
					DisplayName: "My Updated Api",
					Description: "Ignored",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
			},
			want: &rpc.Api{
				Name:        "projects/p/apis/p",
				DisplayName: "My Updated Api",
				Description: "Api for my APIs",
			},
		},
		{
			desc: "full replacement wildcard mask",
			seed: &rpc.Api{
				Name:        "projects/p/apis/p",
				DisplayName: "My Api",
				Description: "Api for my APIs",
			},
			req: &rpc.UpdateApiRequest{
				Api: &rpc.Api{
					Name:        "projects/p/apis/p",
					DisplayName: "My Updated Api",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			},
			want: &rpc.Api{
				Name:        "projects/p/apis/p",
				DisplayName: "My Updated Api",
				Description: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedApis(ctx, t, server, test.seed)

			updated, err := server.UpdateApi(ctx, test.req)
			if err != nil {
				t.Fatalf("UpdateApi(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Api), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, updated, opts) {
				t.Errorf("UpdateApi(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, updated, opts))
			}

			t.Run("GetApi", func(t *testing.T) {
				req := &rpc.GetApiRequest{
					Name: updated.GetName(),
				}

				got, err := server.GetApi(ctx, req)
				if err != nil {
					t.Fatalf("GetApi(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(updated, got, opts) {
					t.Errorf("GetApi(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(updated, got, opts))
				}
			})
		})
	}
}

func TestUpdateApisResponseCodes(t *testing.T) {
	t.Skip("Update mask validation is not implemented")

	tests := []struct {
		desc string
		seed *rpc.Api
		req  *rpc.UpdateApiRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.Api{Name: "projects/p/apis/p"},
			req: &rpc.UpdateApiRequest{
				Api: &rpc.Api{
					Name: "projects/p/apis/doesnt-exist",
				},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource name",
			seed: &rpc.Api{Name: "projects/p/apis/p"},
			req: &rpc.UpdateApiRequest{
				Api: &rpc.Api{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "nonexistent field in mask",
			seed: &rpc.Api{Name: "projects/p/apis/p"},
			req: &rpc.UpdateApiRequest{
				Api: &rpc.Api{
					Name: "projects/p/apis/p",
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
			seedApis(ctx, t, server, test.seed)

			if _, err := server.UpdateApi(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApi(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestDeleteApi(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Api
		req  *rpc.DeleteApiRequest
	}{
		{
			desc: "existing api",
			seed: &rpc.Api{
				Name: "projects/p/apis/p",
			},
			req: &rpc.DeleteApiRequest{
				Name: "projects/p/apis/p",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedApis(ctx, t, server, test.seed)

			if _, err := server.DeleteApi(ctx, test.req); err != nil {
				t.Fatalf("DeleteApi(%+v) returned error: %s", test.req, err)
			}

			t.Run("GetApi", func(t *testing.T) {
				req := &rpc.GetApiRequest{
					Name: test.req.GetName(),
				}

				if _, err := server.GetApi(ctx, req); status.Code(err) != codes.NotFound {
					t.Fatalf("GetApi(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), codes.NotFound, err)
				}
			})
		})
	}
}

func TestDeleteApiResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.DeleteApiRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.DeleteApiRequest{
				Name: "projects/p/apis/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.DeleteApi(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("DeleteApi(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}
