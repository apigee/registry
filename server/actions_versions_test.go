package server

import (
	"context"
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

var (
	// Basic version view does not include annotations.
	basicVersion = &rpc.ApiVersion{
		Name:        "projects/my-project/apis/my-api/versions/my-version",
		DisplayName: "My Api",
		Description: "Api for my versions",
		State:       "PRODUCTION",
		Labels: map[string]string{
			"label-key": "label-value",
		},
	}
	// Full version view includes annotations.
	fullVersion = &rpc.ApiVersion{
		Name:        "projects/my-project/apis/my-api/versions/my-version",
		DisplayName: "My Api",
		Description: "Api for my versions",
		State:       "PRODUCTION",
		Labels: map[string]string{
			"label-key": "label-value",
		},
		Annotations: map[string]string{
			"annotation-key": "annotation-value",
		},
	}
)

func seedVersions(ctx context.Context, t *testing.T, s *RegistryServer, versions ...*rpc.ApiVersion) {
	t.Helper()

	for _, version := range versions {
		name, err := names.ParseVersion(version.Name)
		if err != nil {
			t.Fatalf("Setup/Seeding: ParseVersion(%q) returned error: %s", version.Name, err)
		}

		parent := name.Api()
		seedApis(ctx, t, s, &rpc.Api{
			Name: parent.String(),
		})

		req := &rpc.CreateApiVersionRequest{
			Parent:       parent.String(),
			ApiVersionId: name.VersionID,
			ApiVersion:   version,
		}

		switch _, err := s.CreateApiVersion(ctx, req); status.Code(err) {
		case codes.OK, codes.AlreadyExists:
			// ApiVersion is now ready for use in test.
		default:
			t.Fatalf("Setup/Seeding: CreateApiVersion(%+v) returned error: %s", req, err)
		}
	}
}

func TestCreateApiVersion(t *testing.T) {
	tests := []struct {
		desc      string
		seed      *rpc.Api
		req       *rpc.CreateApiVersionRequest
		want      *rpc.ApiVersion
		extraOpts cmp.Option
	}{
		{
			desc: "populated resource with default parameters",
			seed: &rpc.Api{
				Name: "projects/my-project/apis/my-api",
			},
			req: &rpc.CreateApiVersionRequest{
				Parent:     "projects/my-project/apis/my-api",
				ApiVersion: fullVersion,
			},
			want: basicVersion,
			// Name field is generated.
			extraOpts: protocmp.IgnoreFields(new(rpc.ApiVersion), "name"),
		},
		{
			desc: "custom identifier",
			seed: &rpc.Api{
				Name: "projects/my-project/apis/my-api",
			},
			req: &rpc.CreateApiVersionRequest{
				Parent:       "projects/my-project/apis/my-api",
				ApiVersionId: "my-version",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: &rpc.ApiVersion{
				Name: "projects/my-project/apis/my-api/versions/my-version",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedApis(ctx, t, server, test.seed)

			created, err := server.CreateApiVersion(ctx, test.req)
			if err != nil {
				t.Fatalf("CreateApiVersion(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiVersion), "create_time", "update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateApiVersion(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if !strings.HasPrefix(created.GetName(), test.req.GetParent()+"/versions/") {
				t.Errorf("CreateApiVersion(%+v) returned unexpected name %q, expected collection prefix", test.req, created.GetName())
			}

			if created.CreateTime == nil || created.UpdateTime == nil {
				t.Errorf("CreateApiVersion(%+v) returned unset create_time (%v) or update_time (%v)", test.req, created.CreateTime, created.UpdateTime)
			} else if !created.CreateTime.AsTime().Equal(created.UpdateTime.AsTime()) {
				t.Errorf("CreateApiVersion(%+v) returned unexpected timestamps: create_time %v != update_time %v", test.req, created.CreateTime, created.UpdateTime)
			}

			t.Run("GetApiVersion", func(t *testing.T) {
				req := &rpc.GetApiVersionRequest{
					Name: created.GetName(),
					View: rpc.View_BASIC,
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
	t.Skip("Validation rules are not implemented")

	tests := []struct {
		desc string
		req  *rpc.CreateApiVersionRequest
		want codes.Code
	}{
		{
			desc: "parent not found",
			req: &rpc.CreateApiVersionRequest{
				Parent:     "projects/my-project/apis/my-api",
				ApiVersion: fullVersion,
			},
			want: codes.NotFound,
		},
		{
			desc: "short custom identifier",
			req: &rpc.CreateApiVersionRequest{
				ApiVersionId: "abc",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			req: &rpc.CreateApiVersionRequest{
				ApiVersionId: "this-identifier-exceeds-the-sixty-three-character-maximum-length",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier underscores",
			req: &rpc.CreateApiVersionRequest{
				ApiVersionId: "underscore_identifier",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier dots",
			req: &rpc.CreateApiVersionRequest{
				ApiVersionId: "dot.identifier",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier uuid format",
			req: &rpc.CreateApiVersionRequest{
				ApiVersionId: "072d2288-c685-42d8-9df0-5edbb2a809ea",
				ApiVersion:   &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.CreateApiVersion(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestCreateApiVersionDuplicates(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seedVersions(ctx, t, server, &rpc.ApiVersion{
		Name: "projects/my-project/apis/my-api/versions/v1",
	})

	t.Run("case sensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateApiVersionRequest{
			Parent:       "projects/my-project/apis/my-api",
			ApiVersionId: "v1",
			ApiVersion:   &rpc.ApiVersion{},
		}

		if _, err := server.CreateApiVersion(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateApiVersion(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
		}
	})

	t.Skip("Resource names are not yet case insensitive")
	t.Run("case insensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateApiVersionRequest{
			Parent:       "projects/my-project/apis/my-api",
			ApiVersionId: "V1",
			ApiVersion:   &rpc.ApiVersion{},
		}

		if _, err := server.CreateApiVersion(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateApiVersion(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
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
			desc: "default view",
			seed: fullVersion,
			req: &rpc.GetApiVersionRequest{
				Name: fullVersion.Name,
			},
			want: basicVersion,
		},
		{
			desc: "basic view",
			seed: fullVersion,
			req: &rpc.GetApiVersionRequest{
				Name: fullVersion.Name,
				View: rpc.View_BASIC,
			},
			want: basicVersion,
		},
		{
			desc: "full view",
			seed: fullVersion,
			req: &rpc.GetApiVersionRequest{
				Name: fullVersion.Name,
				View: rpc.View_FULL,
			},
			want: fullVersion,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedVersions(ctx, t, server, test.seed)

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
		req  *rpc.GetApiVersionRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.GetApiVersionRequest{
				Name: "projects/my-project/apis/my-api/versions/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.GetApiVersion(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListApiVersions(t *testing.T) {
	tests := []struct {
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
				{Name: "projects/my-project/apis/my-api/versions/v1"},
				{Name: "projects/my-project/apis/my-api/versions/v2"},
				{Name: "projects/my-project/apis/my-api/versions/v3"},
				{Name: "projects/my-project/apis/other-api/versions/v1"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/apis/my-api",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/apis/my-api/versions/v1"},
					{Name: "projects/my-project/apis/my-api/versions/v2"},
					{Name: "projects/my-project/apis/my-api/versions/v3"},
				},
			},
		},
		{
			desc: "across all apis in a specific project",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/apis/my-api/versions/v1"},
				{Name: "projects/my-project/apis/other-api/versions/v1"},
				{Name: "projects/other-project/apis/my-api/versions/v1"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/apis/-",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/apis/my-api/versions/v1"},
					{Name: "projects/my-project/apis/other-api/versions/v1"},
				},
			},
		},
		{
			desc: "across all projects and apis",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/apis/my-api/versions/v1"},
				{Name: "projects/other-project/apis/other-api/versions/v1"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/-/apis/-",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/apis/my-api/versions/v1"},
					{Name: "projects/other-project/apis/other-api/versions/v1"},
				},
			},
		},
		{
			desc: "in a specific api across all projects",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/apis/my-api/versions/v1"},
				{Name: "projects/other-project/apis/my-api/versions/v1"},
				{Name: "projects/my-project/apis/other-api/versions/v1"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/-/apis/my-api",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/apis/my-api/versions/v1"},
					{Name: "projects/other-project/apis/my-api/versions/v1"},
				},
			},
		},
		{
			desc: "custom page size",
			seed: []*rpc.ApiVersion{
				{Name: "projects/my-project/apis/my-api/versions/v1"},
				{Name: "projects/my-project/apis/my-api/versions/v2"},
				{Name: "projects/my-project/apis/my-api/versions/v3"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent:   "projects/my-project/apis/my-api",
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
				{Name: "projects/my-project/apis/my-api/versions/v1"},
				{Name: "projects/my-project/apis/my-api/versions/v2"},
				{Name: "projects/my-project/apis/my-api/versions/v3"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/apis/my-api",
				Filter: "name == 'projects/my-project/apis/my-api/versions/v2'",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{Name: "projects/my-project/apis/my-api/versions/v2"},
				},
			},
		},
		{
			desc: "description inequality filtering",
			seed: []*rpc.ApiVersion{
				{
					Name:        "projects/my-project/apis/my-api/versions/v1",
					Description: "First ApiVersion",
				},
				{Name: "projects/my-project/apis/my-api/versions/v2"},
				{Name: "projects/my-project/apis/my-api/versions/v3"},
			},
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/apis/my-api",
				Filter: "description != ''",
			},
			want: &rpc.ListApiVersionsResponse{
				ApiVersions: []*rpc.ApiVersion{
					{
						Name:        "projects/my-project/apis/my-api/versions/v1",
						Description: "First ApiVersion",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedVersions(ctx, t, server, test.seed...)

			got, err := server.ListApiVersions(ctx, test.req)
			if err != nil {
				t.Fatalf("ListApiVersions(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListApiVersionsResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.ApiVersion), "create_time", "update_time"),
				protocmp.SortRepeated(func(a, b *rpc.ApiVersion) bool {
					return a.GetName() < b.GetName()
				}),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListApiVersions(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListApiVersions(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
				t.Logf("ListApiVersions(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListApiVersionsResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.ListApiVersionsRequest
		want codes.Code
	}{
		{
			desc: "parent api not found",
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/apis/my-api",
			},
			want: codes.NotFound,
		},
		{
			desc: "parent project not found",
			req: &rpc.ListApiVersionsRequest{
				Parent: "projects/my-project/apis/-",
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
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

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
		{Name: "projects/my-project/apis/my-api/versions/v1"},
		{Name: "projects/my-project/apis/my-api/versions/v2"},
		{Name: "projects/my-project/apis/my-api/versions/v3"},
	}
	seedVersions(ctx, t, server, seed...)

	listed := make([]*rpc.ApiVersion, 0, 3)

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListApiVersionsRequest{
			Parent:   "projects/my-project/apis/my-api",
			PageSize: 1,
		}

		got, err := server.ListApiVersions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiVersions(%+v) returned error: %s", req, err)
		}

		listed = append(listed, got.ApiVersions...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test intermediate page after failure on first page")
	}

	t.Run("intermediate page", func(t *testing.T) {
		req := &rpc.ListApiVersionsRequest{
			Parent:    "projects/my-project/apis/my-api",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiVersions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiVersions(%+v) returned error: %s", req, err)
		}

		listed = append(listed, got.ApiVersions...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test final page after failure on intermediate page")
	}

	t.Run("final page", func(t *testing.T) {
		req := &rpc.ListApiVersionsRequest{
			Parent:    "projects/my-project/apis/my-api",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListApiVersions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiVersions(%+v) returned error: %s", req, err)
		}

		if got.GetNextPageToken() != "" {
			// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
			t.Logf("ListApiVersions(%+v) returned next_page_token, expected no next page", req)
		}

		listed = append(listed, got.ApiVersions...)
	})

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

func TestUpdateApiVersion(t *testing.T) {
	t.Skip("Default/empty mask behavior is incorrect and replacement wildcard is not implemented")

	tests := []struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.UpdateApiVersionRequest
		want *rpc.ApiVersion
	}{
		{
			desc: "populated resource with default parameters",
			seed: fullVersion,
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: fullVersion.Name,
				},
			},
			want: fullVersion,
		},
		{
			desc: "implicit mask",
			seed: &rpc.ApiVersion{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My ApiVersion",
			},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name:        "projects/my-project/apis/my-api/versions/v1",
					Description: "My Updated ApiVersion",
				},
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My Updated ApiVersion",
			},
		},
		{
			desc: "field specific mask",
			seed: &rpc.ApiVersion{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My ApiVersion",
			},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name:        "projects/my-project/apis/my-api/versions/v1",
					DisplayName: "Ignored",
					Description: "My Updated ApiVersion",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"description"}},
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My Updated ApiVersion",
			},
		},
		{
			desc: "full replacement wildcard mask",
			seed: &rpc.ApiVersion{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				DisplayName: "Version One",
				Description: "My ApiVersion",
			},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name:        "projects/my-project/apis/my-api/versions/v1",
					Description: "My Updated ApiVersion",
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			},
			want: &rpc.ApiVersion{
				Name:        "projects/my-project/apis/my-api/versions/v1",
				DisplayName: "",
				Description: "My Updated ApiVersion",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedVersions(ctx, t, server, test.seed)

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
	t.Skip("Update mask validation is not implemented")

	tests := []struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.UpdateApiVersionRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.ApiVersion{Name: "projects/my-project/apis/my-api/versions/v1"},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "projects/my-project/apis/my-api/versions/doesnt-exist",
				},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.ApiVersion{Name: "projects/my-project/apis/my-api/versions/v1"},
			req:  &rpc.UpdateApiVersionRequest{},
			want: codes.InvalidArgument,
		},
		{
			desc: "missing resource name",
			seed: &rpc.ApiVersion{Name: "projects/my-project/apis/my-api/versions/v1"},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "nonexistent field in mask",
			seed: &rpc.ApiVersion{Name: "projects/my-project/apis/my-api/versions/v1"},
			req: &rpc.UpdateApiVersionRequest{
				ApiVersion: &rpc.ApiVersion{
					Name: "projects/my-project/apis/my-api/versions/v1",
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
			seedVersions(ctx, t, server, test.seed)

			if _, err := server.UpdateApiVersion(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
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
			desc: "existing version",
			seed: &rpc.ApiVersion{
				Name: "projects/my-project/apis/my-api/versions/v1",
			},
			req: &rpc.DeleteApiVersionRequest{
				Name: "projects/my-project/apis/my-api/versions/v1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedVersions(ctx, t, server, test.seed)

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
		req  *rpc.DeleteApiVersionRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.DeleteApiVersionRequest{
				Name: "projects/my-project/apis/my-api/versions/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.DeleteApiVersion(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("DeleteApiVersion(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}
