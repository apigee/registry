// Copyright 2021 Google LLC. All Rights Reserved.
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
// See the License for the artifact language governing permissions and
// limitations under the License.

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
)

var (
	// Example artifact contents for a JSON artifact.
	artifactContents = []byte(`{"contents": "foo"}`)
	// Basic artifact view does not include contents.
	basicArtifact = &rpc.Artifact{
		Name:      "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact",
		MimeType:  "application/json",
		SizeBytes: int32(len(artifactContents)),
		Hash:      sha256hash(artifactContents),
	}
	// Full artifact view includes contents.
	fullArtifact = &rpc.Artifact{
		Name:      "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact",
		MimeType:  "application/json",
		SizeBytes: int32(len(artifactContents)),
		Hash:      sha256hash(artifactContents),
		Contents:  artifactContents,
	}
)

func seedArtifacts(ctx context.Context, t *testing.T, s *RegistryServer, artifacts ...*rpc.Artifact) {
	t.Helper()

	for _, artifact := range artifacts {
		name, err := names.ParseArtifact(artifact.Name)
		if err != nil {
			t.Fatalf("Setup/Seeding: ParseArtifact(%q) returned error: %s", artifact.Name, err)
		}

		parent := strings.TrimSuffix(name.String(), "/artifacts/"+name.ArtifactID())
		if _, err := names.ParseSpec(parent); err == nil {
			seedSpecs(ctx, t, s, &rpc.ApiSpec{
				Name: parent,
			})
		} else if _, err := names.ParseVersion(parent); err == nil {
			seedVersions(ctx, t, s, &rpc.ApiVersion{
				Name: parent,
			})
		} else if _, err := names.ParseApi(parent); err == nil {
			seedApis(ctx, t, s, &rpc.Api{
				Name: parent,
			})
		} else if _, err := names.ParseProject(parent); err == nil {
			seedProjects(ctx, t, s, &rpc.Project{
				Name: parent,
			})
		} else {
			t.Log("Failed to identify parent resource: proceeding without seeding parent")
		}

		req := &rpc.CreateArtifactRequest{
			Parent:     parent,
			ArtifactId: name.ArtifactID(),
			Artifact:   artifact,
		}

		switch _, err := s.CreateArtifact(ctx, req); status.Code(err) {
		case codes.OK, codes.AlreadyExists:
			// Artifact is now ready for use in test.
		default:
			t.Fatalf("Setup/Seeding: CreateArtifact(%+v) returned error: %s", req, err)
		}
	}
}

func TestCreateArtifact(t *testing.T) {
	tests := []struct {
		desc      string
		seed      *rpc.ApiVersion
		req       *rpc.CreateArtifactRequest
		want      *rpc.Artifact
		extraOpts cmp.Option
	}{
		{
			desc: "populated resource with default parameters",
			seed: &rpc.ApiVersion{Name: "projects/my-project/apis/my-api/versions/v1"},
			req: &rpc.CreateArtifactRequest{
				Parent:   "projects/my-project/apis/my-api/versions/v1",
				Artifact: fullArtifact,
			},
			want: basicArtifact,
			// Name field is generated.
			extraOpts: protocmp.IgnoreFields(new(rpc.Artifact), "name"),
		},
		{
			desc: "custom identifier",
			seed: &rpc.ApiVersion{Name: "projects/my-project/apis/my-api/versions/v1"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/apis/my-api/versions/v1",
				ArtifactId: "my-artifact",
				Artifact:   &rpc.Artifact{},
			},
			want: &rpc.Artifact{
				Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedVersions(ctx, t, server, test.seed)

			created, err := server.CreateArtifact(ctx, test.req)
			if err != nil {
				t.Fatalf("CreateArtifact(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateArtifact(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if !strings.HasPrefix(created.GetName(), test.req.GetParent()+"/artifacts/") {
				t.Errorf("CreateArtifact(%+v) returned unexpected name %q, expected collection prefix", test.req, created.GetName())
			}

			if created.CreateTime == nil && created.UpdateTime == nil {
				t.Errorf("CreateArtifact(%+v) returned unset create_time (%v) and update_time (%v)", test.req, created.CreateTime, created.UpdateTime)
			} else if !created.CreateTime.AsTime().Equal(created.UpdateTime.AsTime()) {
				t.Errorf("CreateArtifact(%+v) returned unexpected timestamps: create_time %v != update_time %v", test.req, created.CreateTime, created.UpdateTime)
			}

			t.Run("GetArtifact", func(t *testing.T) {
				req := &rpc.GetArtifactRequest{
					Name: created.GetName(),
					View: rpc.View_BASIC,
				}

				got, err := server.GetArtifact(ctx, req)
				if err != nil {
					t.Fatalf("GetArtifact(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(created, got, opts) {
					t.Errorf("GetArtifact(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(created, got, opts))
				}
			})
		})
	}
}

func TestCreateArtifactResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Project
		req  *rpc.CreateArtifactRequest
		want codes.Code
	}{
		{
			desc: "parent not found",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:   "projects/other-project",
				Artifact: fullArtifact,
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:   "projects/my-project",
				Artifact: nil,
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project",
				ArtifactId: "this-identifier-exceeds-the-sixty-three-character-maximum-length",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier underscores",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project",
				ArtifactId: "underscore_identifier",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen prefix",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project",
				ArtifactId: "-identifier",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen suffix",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project",
				ArtifactId: "identifier-",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier uuid format",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project",
				ArtifactId: "072d2288-c685-42d8-9df0-5edbb2a809ea",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedProjects(ctx, t, server, test.seed)

			if _, err := server.CreateArtifact(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateArtifact(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestCreateArtifactDuplicates(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seedArtifacts(ctx, t, server, &rpc.Artifact{
		Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact",
	})

	t.Run("case sensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateArtifactRequest{
			Parent:     "projects/my-project/apis/my-api/versions/v1",
			ArtifactId: "my-artifact",
			Artifact:   &rpc.Artifact{},
		}

		if _, err := server.CreateArtifact(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateArtifact(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
		}
	})

	t.Run("case insensitive duplicate", func(t *testing.T) {
		req := &rpc.CreateArtifactRequest{
			Parent:     "projects/my-project/apis/my-api/versions/v1",
			ArtifactId: "My-Artifact",
			Artifact:   &rpc.Artifact{},
		}

		if _, err := server.CreateArtifact(ctx, req); status.Code(err) != codes.AlreadyExists {
			t.Errorf("CreateArtifact(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.AlreadyExists, err)
		}
	})
}

func TestGetArtifact(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.GetArtifactRequest
		want *rpc.Artifact
	}{
		{
			desc: "default view",
			seed: fullArtifact,
			req: &rpc.GetArtifactRequest{
				Name: fullArtifact.Name,
			},
			want: basicArtifact,
		},
		{
			desc: "basic view",
			seed: fullArtifact,
			req: &rpc.GetArtifactRequest{
				Name: fullArtifact.Name,
				View: rpc.View_BASIC,
			},
			want: basicArtifact,
		},
		{
			desc: "full view",
			seed: fullArtifact,
			req: &rpc.GetArtifactRequest{
				Name: fullArtifact.Name,
				View: rpc.View_FULL,
			},
			want: fullArtifact,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedArtifacts(ctx, t, server, test.seed)

			got, err := server.GetArtifact(ctx, test.req)
			if err != nil {
				t.Fatalf("GetArtifact(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("GetArtifact(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}
		})
	}
}

func TestGetArtifactResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.GetArtifactRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.GetArtifactRequest{
				Name: "projects/my-project/apis/my-api/versions/v1/artifacts/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.GetArtifact(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetArtifact(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListArtifacts(t *testing.T) {
	tests := []struct {
		desc      string
		seed      []*rpc.Artifact
		req       *rpc.ListArtifactsRequest
		want      *rpc.ListArtifactsResponse
		wantToken bool
		extraOpts cmp.Option
	}{
		{
			desc: "default parameters",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact1"},
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact2"},
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact3"},
				{Name: "projects/my-project/apis/my-api/versions/v2/artifacts/artifact1"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/apis/my-api/versions/v1",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact1"},
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact2"},
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact3"},
				},
			},
		},
		{
			desc: "across all version in a artifact project and api",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/apis/my-api/versions/v2/artifacts/my-artifact"},
				{Name: "projects/other-project/apis/my-api/versions/v1/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/apis/my-api/versions/-",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/my-project/apis/my-api/versions/v2/artifacts/my-artifact"},
				},
			},
		},
		{
			desc: "across all apis and version in a artifact project",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/apis/other-api/versions/v2/artifacts/my-artifact"},
				{Name: "projects/other-project/apis/my-api/versions/v1/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/apis/-/versions/-",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/my-project/apis/other-api/versions/v2/artifacts/my-artifact"},
				},
			},
		},
		{
			desc: "across all projects, apis, and version",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/other-project/apis/other-api/versions/v2/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/-/apis/-/versions/-",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/other-project/apis/other-api/versions/v2/artifacts/my-artifact"},
				},
			},
		},
		{
			desc: "in a artifact api and parent across all projects",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/other-project/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/apis/other-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/apis/my-api/versions/v2/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/-/apis/my-api/versions/v1",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/other-project/apis/my-api/versions/v1/artifacts/my-artifact"},
				},
			},
		},
		{
			desc: "in a artifact parent across all projects and apis",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/other-project/apis/other-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/apis/my-api/versions/v2/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/-/apis/-/versions/v1",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/other-project/apis/other-api/versions/v1/artifacts/my-artifact"},
				},
			},
		},
		{
			desc: "in all version of a artifact api across all projects",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/other-project/apis/my-api/versions/v2/artifacts/my-artifact"},
				{Name: "projects/my-project/apis/other-api/versions/v1/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/-/apis/my-api/versions/-",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/other-project/apis/my-api/versions/v2/artifacts/my-artifact"},
				},
			},
		},
		{
			desc: "custom page size",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact1"},
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact2"},
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact3"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent:   "projects/my-project/apis/my-api/versions/v1",
				PageSize: 1,
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{},
				},
			},
			wantToken: true,
			// Ordering is not guaranteed by API, so any resource may be returned.
			extraOpts: protocmp.IgnoreFields(new(rpc.Artifact), "name"),
		},
		{
			desc: "name equality filtering",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact1"},
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact2"},
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact3"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/apis/my-api/versions/v1",
				Filter: "name == 'projects/my-project/apis/my-api/versions/v1/artifacts/artifact2'",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact2"},
				},
			},
		},
		{
			desc: "description inequality filtering",
			seed: []*rpc.Artifact{
				{
					Name:     "projects/my-project/apis/my-api/versions/v1/artifacts/artifact1",
					MimeType: "application/json",
				},
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact2"},
				{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact3"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/apis/my-api/versions/v1",
				Filter: "mime_type != ''",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{
						Name:     "projects/my-project/apis/my-api/versions/v1/artifacts/artifact1",
						MimeType: "application/json",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedArtifacts(ctx, t, server, test.seed...)

			got, err := server.ListArtifacts(ctx, test.req)
			if err != nil {
				t.Fatalf("ListArtifacts(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListArtifactsResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
				protocmp.SortRepeated(func(a, b *rpc.Artifact) bool {
					return a.GetName() < b.GetName()
				}),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListArtifacts(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListArtifacts(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
				t.Logf("ListArtifacts(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListArtifactsResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.ListArtifactsRequest
		want codes.Code
	}{
		{
			desc: "parent parent not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/apis/my-api/versions/v1",
			},
			want: codes.NotFound,
		},
		{
			desc: "parent api not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/apis/my-api/versions/-",
			},
			want: codes.NotFound,
		},
		{
			desc: "parent project not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/apis/-/versions/-",
			},
			want: codes.NotFound,
		},
		{
			desc: "negative page size",
			req: &rpc.ListArtifactsRequest{
				PageSize: -1,
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid filter",
			req: &rpc.ListArtifactsRequest{
				Filter: "this filter is not valid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid page token",
			req: &rpc.ListArtifactsRequest{
				PageToken: "this token is not valid",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.ListArtifacts(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("ListArtifacts(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListArtifactsSequence(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := []*rpc.Artifact{
		{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact1"},
		{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact2"},
		{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/artifact3"},
	}
	seedArtifacts(ctx, t, server, seed...)

	listed := make([]*rpc.Artifact, 0, 3)

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListArtifactsRequest{
			Parent:   "projects/my-project/apis/my-api/versions/v1",
			PageSize: 1,
		}

		got, err := server.ListArtifacts(ctx, req)
		if err != nil {
			t.Fatalf("ListArtifacts(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetArtifacts()); count != 1 {
			t.Errorf("ListArtifacts(%+v) returned %d artifacts, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListArtifacts(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.Artifacts...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test intermediate page after failure on first page")
	}

	t.Run("intermediate page", func(t *testing.T) {
		req := &rpc.ListArtifactsRequest{
			Parent:    "projects/my-project/apis/my-api/versions/v1",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListArtifacts(ctx, req)
		if err != nil {
			t.Fatalf("ListArtifacts(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetArtifacts()); count != 1 {
			t.Errorf("ListArtifacts(%+v) returned %d artifacts, expected exactly one", req, count)
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListArtifacts(%+v) returned empty next_page_token, expected another page", req)
		}

		listed = append(listed, got.Artifacts...)
		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test final page after failure on intermediate page")
	}

	t.Run("final page", func(t *testing.T) {
		req := &rpc.ListArtifactsRequest{
			Parent:    "projects/my-project/apis/my-api/versions/v1",
			PageSize:  1,
			PageToken: nextToken,
		}

		got, err := server.ListArtifacts(ctx, req)
		if err != nil {
			t.Fatalf("ListArtifacts(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetArtifacts()); count != 1 {
			t.Errorf("ListArtifacts(%+v) returned %d artifacts, expected exactly one", req, count)
		}

		if got.GetNextPageToken() != "" {
			// TODO: This should be changed to a test error when possible. See: https://github.com/apigee/registry/issues/68
			t.Logf("ListArtifacts(%+v) returned next_page_token, expected no next page", req)
		}

		listed = append(listed, got.Artifacts...)
	})

	if t.Failed() {
		t.Fatal("Cannot test sequence result after failure on final page")
	}

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
		cmpopts.SortSlices(func(a, b *rpc.Artifact) bool {
			return a.GetName() < b.GetName()
		}),
	}

	if !cmp.Equal(seed, listed, opts) {
		t.Errorf("List sequence returned unexpected diff (-want +got):\n%s", cmp.Diff(seed, listed, opts))
	}
}

func TestReplaceArtifact(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.ReplaceArtifactRequest
		want *rpc.Artifact
	}{
		{
			desc: "populated resource",
			seed: &rpc.Artifact{
				Name: fullArtifact.Name,
			},
			req: &rpc.ReplaceArtifactRequest{
				Artifact: fullArtifact,
			},
			want: basicArtifact,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedArtifacts(ctx, t, server, test.seed)

			updated, err := server.ReplaceArtifact(ctx, test.req)
			if err != nil {
				t.Fatalf("ReplaceArtifact(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, updated, opts) {
				t.Errorf("ReplaceArtifact(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, updated, opts))
			}

			t.Run("GetArtifact", func(t *testing.T) {
				req := &rpc.GetArtifactRequest{
					Name: updated.GetName(),
				}

				got, err := server.GetArtifact(ctx, req)
				if err != nil {
					t.Fatalf("GetArtifact(%+v) returned error: %s", req, err)
				}

				opts := protocmp.Transform()
				if !cmp.Equal(updated, got, opts) {
					t.Errorf("GetArtifact(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(updated, got, opts))
				}
			})
		})
	}
}

func TestReplaceArtifactResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.ReplaceArtifactRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.Artifact{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
			req: &rpc.ReplaceArtifactRequest{
				Artifact: &rpc.Artifact{
					Name: "projects/my-project/apis/my-api/versions/v1/artifacts/doesnt-exist",
				},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.Artifact{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
			req:  &rpc.ReplaceArtifactRequest{},
			want: codes.InvalidArgument,
		},
		{
			desc: "missing resource name",
			seed: &rpc.Artifact{Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact"},
			req: &rpc.ReplaceArtifactRequest{
				Artifact: &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedArtifacts(ctx, t, server, test.seed)

			if _, err := server.ReplaceArtifact(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("ReplaceArtifact(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestDeleteArtifact(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.DeleteArtifactRequest
	}{
		{
			desc: "existing parent",
			seed: &rpc.Artifact{
				Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact",
			},
			req: &rpc.DeleteArtifactRequest{
				Name: "projects/my-project/apis/my-api/versions/v1/artifacts/my-artifact",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			seedArtifacts(ctx, t, server, test.seed)

			if _, err := server.DeleteArtifact(ctx, test.req); err != nil {
				t.Fatalf("DeleteArtifact(%+v) returned error: %s", test.req, err)
			}

			t.Run("GetArtifact", func(t *testing.T) {
				req := &rpc.GetArtifactRequest{
					Name: test.req.GetName(),
				}

				if _, err := server.GetArtifact(ctx, req); status.Code(err) != codes.NotFound {
					t.Fatalf("GetArtifact(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), codes.NotFound, err)
				}
			})
		})
	}
}

func TestDeleteArtifactResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.DeleteArtifactRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			req: &rpc.DeleteArtifactRequest{
				Name: "projects/my-project/apis/my-api/versions/v1/artifacts/doesnt-exist",
			},
			want: codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			if _, err := server.DeleteArtifact(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("DeleteArtifact(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}
