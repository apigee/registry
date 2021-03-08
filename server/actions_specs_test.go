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

func seedSpecs(ctx context.Context, t *testing.T, s *RegistryServer, specs ...*rpc.ApiSpec) {
	t.Helper()

	for _, spec := range specs {
		m, err := names.ParseSpec(spec.Name)
		if err != nil {
			t.Fatalf("Setup/Seeding: ParseSpec(%q) returned error: %s", spec.Name, err)
		}

		parent := fmt.Sprintf("projects/%s/apis/%s/versions/%s", m[1], m[2], m[3])
		seedVersions(ctx, t, s, &rpc.ApiVersion{
			Name: parent,
		})

		req := &rpc.CreateApiSpecRequest{
			Parent:    parent,
			ApiSpecId: m[4],
			ApiSpec:   spec,
		}

		switch _, err := s.CreateApiSpec(ctx, req); status.Code(err) {
		case codes.OK, codes.AlreadyExists:
			// ApiSpec is now ready for use in test.
		default:
			t.Fatalf("Setup/Seeding: CreateApiSpec(%+v) returned error: %s", req, err)
		}
	}
}

func TestCreateApiSpec(t *testing.T) {
	tests := []struct {
		desc      string
		req       *rpc.CreateApiSpecRequest
		want      *rpc.ApiSpec
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			req: &rpc.CreateApiSpecRequest{
				Parent: "projects/my-project/apis/my-api/versions/v1",
				ApiSpec: &rpc.ApiSpec{
					Description: "ApiSpec for my versions",
				},
			},
			want: &rpc.ApiSpec{
				Description: "ApiSpec for my versions",
			},
			// Name field is generated.
			extraOpts: protocmp.IgnoreFields(new(rpc.ApiSpec), "name"),
		},
		{
			desc: "custom identifier",
			req: &rpc.CreateApiSpecRequest{
				Parent:    "projects/my-project/apis/my-api/versions/v1",
				ApiSpecId: "my-version",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: &rpc.ApiSpec{
				Name: "projects/my-project/apis/my-api/versions/v1/specs/my-version",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			created, err := server.CreateApiSpec(ctx, test.req)
			if err != nil {
				t.Fatalf("CreateApiSpec(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiSpec), "create_time", "revision_create_time", "revision_update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateApiSpec(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if !strings.HasPrefix(created.GetName(), test.req.GetParent()+"/specs/") {
				t.Errorf("CreateApiSpec(%+v) returned unexpected name %q, expected collection prefix", test.req, created.GetName())
			}

			if created.CreateTime == nil || created.RevisionCreateTime == nil || created.RevisionUpdateTime == nil {
				t.Errorf("CreateApiSpec(%+v) returned unset create_time (%v), revision_create_time (%v), or revision_update_time (%v)", test.req, created.CreateTime, created.RevisionCreateTime, created.RevisionUpdateTime)
			} else if !created.CreateTime.AsTime().Equal(created.RevisionCreateTime.AsTime()) {
				t.Errorf("CreateApiSpec(%+v) returned unexpected timestamps: create_time %v != revision_create_time %v", test.req, created.CreateTime, created.RevisionCreateTime)
			} else if !created.RevisionCreateTime.AsTime().Equal(created.RevisionUpdateTime.AsTime()) {
				t.Errorf("CreateApiSpec(%+v) returned unexpected timestamps: revision_create_time %v != revision_update_time %v", test.req, created.RevisionCreateTime, created.RevisionUpdateTime)
			}

			t.Run("GetApiSpec", func(t *testing.T) {
				req := &rpc.GetApiSpecRequest{
					Name: created.GetName(),
				}

				got, err := server.GetApiSpec(ctx, req)
				if err != nil {
					t.Fatalf("GetApiSpec(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(created, got, opts) {
					t.Errorf("GetApiSpec(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(created, got, opts))
				}
			})
		})
	}
}

func TestCreateApiSpecResponseCodes(t *testing.T) {
	t.Skip("Validation rules are not implemented")

	tests := []struct {
		desc string
		req  *rpc.CreateApiSpecRequest
		want codes.Code
	}{
		{
			desc: "short custom identifier",
			req: &rpc.CreateApiSpecRequest{
				ApiSpecId: "abc",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			req: &rpc.CreateApiSpecRequest{
				ApiSpecId: "this-identifier-exceeds-the-sixty-three-character-maximum-length",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier underscores",
			req: &rpc.CreateApiSpecRequest{
				ApiSpecId: "underscore_identifier",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier dots",
			req: &rpc.CreateApiSpecRequest{
				ApiSpecId: "dot.identifier",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier uuid format",
			req: &rpc.CreateApiSpecRequest{
				ApiSpecId: "072d2288-c685-42d8-9df0-5edbb2a809ea",
				ApiSpec:   &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.CreateApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestCreateApiSpecDuplicates(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seedSpecs(ctx, t, server, &rpc.ApiSpec{
		Name: "projects/my-project/apis/my-api/versions/v1/specs/my-spec",
	})

	t.Run("case sensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateApiSpecRequest{
			Parent:    "projects/my-project/apis/my-api/versions/v1",
			ApiSpecId: "my-spec",
			ApiSpec:   &rpc.ApiSpec{},
		}

		if _, err := server.CreateApiSpec(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateApiSpec(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
		}
	})

	t.Skip("Resource names are not yet case insensitive")
	t.Run("case insensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateApiSpecRequest{
			Parent:    "projects/my-project/apis/my-api/versions/v1",
			ApiSpecId: "My-Spec",
			ApiSpec:   &rpc.ApiSpec{},
		}

		if _, err := server.CreateApiSpec(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateApiSpec(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
		}
	})
}

func TestGetApiSpecResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.GetApiSpecRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.GetApiSpecRequest{
				Name: "projects/my-project/apis/my-api/versions/v1/specs/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.GetApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApiSpecs(t *testing.T) {
	tests := []struct {
		desc      string
		seed      []*rpc.ApiSpec
		req       *rpc.ListApiSpecsRequest
		want      *rpc.ListApiSpecsResponse
		wantToken bool
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1"},
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec2"},
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec3"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/apis/my-api/versions/v1",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1"},
					{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec2"},
					{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec3"},
				},
			},
		},
		{
			desc: "custom page size",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1"},
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec2"},
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec3"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent:   "projects/my-project/apis/my-api/versions/v1",
				PageSize: 1,
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{},
				},
			},
			wantToken: true,
			// Ordering is not guaranteed by API, so any resource may be returned.
			extraOpts: protocmp.IgnoreFields(new(rpc.ApiSpec), "name"),
		},
		{
			desc: "name equality filtering",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1"},
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec2"},
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec3"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/apis/my-api/versions/v1",
				Filter: "name == 'projects/my-project/apis/my-api/versions/v1/specs/spec2'",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec2"},
				},
			},
		},
		{
			desc: "description inequality filtering",
			seed: []*rpc.ApiSpec{
				{
					Name:        "projects/my-project/apis/my-api/versions/v1/specs/spec1",
					Description: "First ApiSpec",
				},
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec2"},
				{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec3"},
			},
			req: &rpc.ListApiSpecsRequest{
				Parent: "projects/my-project/apis/my-api/versions/v1",
				Filter: "description != ''",
			},
			want: &rpc.ListApiSpecsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{
						Name:        "projects/my-project/apis/my-api/versions/v1/specs/spec1",
						Description: "First ApiSpec",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedSpecs(ctx, t, server, test.seed...)

			got, err := server.ListApiSpecs(ctx, test.req)
			if err != nil {
				t.Fatalf("ListApiSpecs(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListApiSpecsResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.ApiSpec), "create_time", "revision_create_time", "revision_update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListApiSpecs(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListApiSpecs(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
				t.Logf("ListApiSpecs(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListApiSpecsResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.ListApiSpecsRequest
		want codes.Code
	}{
		{
			desc: "negative page size",
			req: &rpc.ListApiSpecsRequest{
				PageSize: -1,
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid filter",
			req: &rpc.ListApiSpecsRequest{
				Filter: "this filter is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid page token",
			req: &rpc.ListApiSpecsRequest{
				PageToken: "this token is not valid",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.ListApiSpecs(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("ListApiSpecs(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApiSpecsSequence(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := []*rpc.ApiSpec{
		{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1"},
		{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec2"},
		{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec3"},
	}
	seedSpecs(ctx, t, server, seed...)

	listed := make([]*rpc.ApiSpec, 0, 3)

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListApiSpecsRequest{
			Parent:   "projects/my-project/apis/my-api/versions/v1",
			PageSize: 1,
		}

		got, err := server.ListApiSpecs(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecs(%+v) returned error: %s", req, err)
		}

		listed = append(listed, got.ApiSpecs...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test intermediate page after failure on first page")
	}

	t.Run("intermediate page", func(t *testing.T) {
		req := &rpc.ListApiSpecsRequest{
			Parent:    "projects/my-project/apis/my-api/versions/v1",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiSpecs(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecs(%+v) returned error: %s", req, err)
		}

		listed = append(listed, got.ApiSpecs...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test final page after failure on intermediate page")
	}

	t.Run("final page", func(t *testing.T) {
		req := &rpc.ListApiSpecsRequest{
			Parent:    "projects/my-project/apis/my-api/versions/v1",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiSpecs(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecs(%+v) returned error: %s", req, err)
		}

		if got.GetNextPageToken() != "" {
			// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
			t.Logf("ListApiSpecs(%+v) returned next_page_token, expected no next page", req)
		}

		listed = append(listed, got.ApiSpecs...)
	})

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(rpc.ApiSpec), "create_time", "revision_create_time", "revision_update_time"),
		cmpopts.SortSlices(func(a, b *rpc.ApiSpec) bool {
			return a.GetName() < b.GetName()
		}),
	}

	if !cmp.Equal(seed, listed, opts) {
		t.Errorf("List sequence returned unexpected diff (-want +got):\n%s", cmp.Diff(seed, listed, opts))
	}
}

func TestUpdateApiSpec(t *testing.T) {
	t.Skip("Default/empty mask behavior is incorrect and replacement wildcard is not implemented")

	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.UpdateApiSpecRequest
		want *rpc.ApiSpec
	}{
		{
			desc: "default parameters",
			seed: &rpc.ApiSpec{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				Description: "ApiSpec for my APIs",
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1",
				},
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				Description: "ApiSpec for my APIs",
			},
		},
		{
			desc: "field specific mask",
			seed: &rpc.ApiSpec{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				Description: "ApiSpec for my APIs",
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name:        "projects/my-project/apis/my-api/versions/v1",
					Description: "Ignored",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"display_name"}},
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				Description: "ApiSpec for my APIs",
			},
		},
		{
			desc: "full replacement wildcard mask",
			seed: &rpc.ApiSpec{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				Description: "ApiSpec for my APIs",
			},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			},
			want: &rpc.ApiSpec{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				Description: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedSpecs(ctx, t, server, test.seed)

			updated, err := server.UpdateApiSpec(ctx, test.req)
			if err != nil {
				t.Fatalf("UpdateApiSpec(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiSpec), "create_time", "revision_create_time", "revision_update_time"),
			}

			if !cmp.Equal(test.want, updated, opts) {
				t.Errorf("UpdateApiSpec(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, updated, opts))
			}

			t.Run("GetApiSpec", func(t *testing.T) {
				req := &rpc.GetApiSpecRequest{
					Name: updated.GetName(),
				}

				got, err := server.GetApiSpec(ctx, req)
				if err != nil {
					t.Fatalf("GetApiSpec(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(updated, got, opts) {
					t.Errorf("GetApiSpec(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(updated, got, opts))
				}
			})
		})
	}
}

func TestUpdateApiSpecsResponseCodes(t *testing.T) {
	t.Skip("Update mask validation is not implemented")

	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.UpdateApiSpecRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.ApiSpec{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1"},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/apis/my-api/versions/v1/specs/doesnt-exist",
				},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource name",
			seed: &rpc.ApiSpec{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1"},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "nonexistent field in mask",
			seed: &rpc.ApiSpec{Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1"},
			req: &rpc.UpdateApiSpecRequest{
				ApiSpec: &rpc.ApiSpec{
					Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1",
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
			seedSpecs(ctx, t, server, test.seed)

			if _, err := server.UpdateApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestDeleteApiSpec(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.ApiSpec
		req  *rpc.DeleteApiSpecRequest
	}{
		{
			desc: "existing version",
			seed: &rpc.ApiSpec{
				Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1",
			},
			req: &rpc.DeleteApiSpecRequest{
				Name: "projects/my-project/apis/my-api/versions/v1/specs/spec1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedSpecs(ctx, t, server, test.seed)

			if _, err := server.DeleteApiSpec(ctx, test.req); err != nil {
				t.Fatalf("DeleteApiSpec(%+v) returned error: %s", test.req, err)
			}

			t.Run("GetApiSpec", func(t *testing.T) {
				req := &rpc.GetApiSpecRequest{
					Name: test.req.GetName(),
				}

				if _, err := server.GetApiSpec(ctx, req); status.Code(err) != codes.NotFound {
					t.Fatalf("GetApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), codes.NotFound, err)
				}
			})
		})
	}
}

func TestDeleteApiSpecResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.DeleteApiSpecRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.DeleteApiSpecRequest{
				Name: "projects/my-project/apis/my-api/versions/v1/specs/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.DeleteApiSpec(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("DeleteApiSpec(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}
