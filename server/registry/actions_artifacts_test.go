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
// See the License for the artifact language governing permissions and
// limitations under the License.

package registry

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

var (
	// Example artifact contents for a JSON artifact.
	artifactContents = []byte(`{"contents": "foo"}`)
)

func TestCreateArtifact(t *testing.T) {
	tests := []struct {
		desc string
		seed seeder.RegistryResource
		req  *rpc.CreateArtifactRequest
		want *rpc.Artifact
	}{
		{
			desc: "create project artifact",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "my-artifact",
				Artifact: &rpc.Artifact{
					MimeType:  "application/json",
					SizeBytes: int32(len(artifactContents)),
					Hash:      sha256hash(artifactContents),
					Contents:  artifactContents,
					Labels: map[string]string{
						"label-key": "label-value",
					},
					Annotations: map[string]string{
						"annotation-key": "annotation-value",
					},
				},
			},
			want: &rpc.Artifact{
				Name:      "projects/my-project/locations/global/artifacts/my-artifact",
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContents)),
				Hash:      sha256hash(artifactContents),
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
		},
		{
			desc: "create api artifact",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global/apis/my-api",
				ArtifactId: "my-artifact",
				Artifact: &rpc.Artifact{
					MimeType:  "application/json",
					SizeBytes: int32(len(artifactContents)),
					Hash:      sha256hash(artifactContents),
					Contents:  artifactContents,
					Labels: map[string]string{
						"label-key": "label-value",
					},
					Annotations: map[string]string{
						"annotation-key": "annotation-value",
					},
				},
			},
			want: &rpc.Artifact{
				Name:      "projects/my-project/locations/global/apis/my-api/artifacts/my-artifact",
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContents)),
				Hash:      sha256hash(artifactContents),
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
		},
		{
			desc: "create version artifact",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global/apis/my-api/versions/my-version",
				ArtifactId: "my-artifact",
				Artifact: &rpc.Artifact{
					MimeType:  "application/json",
					SizeBytes: int32(len(artifactContents)),
					Hash:      sha256hash(artifactContents),
					Contents:  artifactContents,
					Labels: map[string]string{
						"label-key": "label-value",
					},
					Annotations: map[string]string{
						"annotation-key": "annotation-value",
					},
				},
			},
			want: &rpc.Artifact{
				Name:      "projects/my-project/locations/global/apis/my-api/versions/my-version/artifacts/my-artifact",
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContents)),
				Hash:      sha256hash(artifactContents),
				Labels: map[string]string{
					"label-key": "label-value",
				},
				Annotations: map[string]string{
					"annotation-key": "annotation-value",
				},
			},
		},
		{
			desc: "create empty artifact",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "my-artifact",
				Artifact:   nil,
			},
			want: &rpc.Artifact{
				Name: "projects/my-project/locations/global/artifacts/my-artifact",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedRegistry(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			created, err := server.CreateArtifact(ctx, test.req)
			if err != nil {
				t.Fatalf("CreateArtifact(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
			}

			if !cmp.Equal(test.want, created, opts) {
				t.Errorf("CreateArtifact(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, created, opts))
			}

			if created.CreateTime == nil && created.UpdateTime == nil {
				t.Errorf("CreateArtifact(%+v) returned unset create_time (%v) and update_time (%v)", test.req, created.CreateTime, created.UpdateTime)
			} else if !created.CreateTime.AsTime().Equal(created.UpdateTime.AsTime()) {
				t.Errorf("CreateArtifact(%+v) returned unexpected timestamps: create_time %v != update_time %v", test.req, created.CreateTime, created.UpdateTime)
			}

			t.Run("GetArtifact", func(t *testing.T) {
				req := &rpc.GetArtifactRequest{
					Name: created.GetName(),
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
		admin bool
		desc  string
		seed  seeder.RegistryResource
		req   *rpc.CreateArtifactRequest
		want  codes.Code
	}{
		{
			admin: true,
			desc:  "parent project not found",
			seed:  &rpc.Project{Name: "projects/other-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "valid-id",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.NotFound,
		},
		{
			desc: "parent api not found",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global/apis/a",
				ArtifactId: "valid-id",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.NotFound,
		},
		{
			desc: "parent version not found",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global/apis/a/versions/v",
				ArtifactId: "valid-id",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.NotFound,
		},
		{
			desc: "parent spec not found",
			seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/a/versions/v"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global/apis/a/versions/v/specs/s",
				ArtifactId: "valid-id",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.NotFound,
		},
		{
			desc: "parent deployment not found",
			seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global/apis/a/deployments/d",
				ArtifactId: "valid-id",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "valid-id",
				Artifact:   nil,
			},
			want: codes.OK,
		},
		{
			desc: "missing custom identifier",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "long custom identifier",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "this-identifier-is-invalid-because-it-exceeds-the-eighty-character-maximum-length",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier underscores",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "underscore_identifier",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen prefix",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "-identifier",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier hyphen suffix",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "identifier-",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "customer identifier uuid format",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "072d2288-c685-42d8-9df0-5edbb2a809ea",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "custom identifier mixed case",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "IDentifier",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid parent",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent: "invalid",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid contents",
			seed: &rpc.Project{Name: "projects/my-project"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "identifier",
				Artifact: &rpc.Artifact{
					MimeType: "something+gzip",
					Contents: []byte("invalid"),
				},
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
			if err := seeder.SeedRegistry(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.CreateArtifact(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateArtifact(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestCreateArtifactDuplicates(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.CreateArtifactRequest
		want codes.Code
	}{
		{
			desc: "case sensitive",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/my-artifact"},
			req: &rpc.CreateArtifactRequest{
				Parent:     "projects/my-project/locations/global",
				ArtifactId: "my-artifact",
				Artifact:   &rpc.Artifact{},
			},
			want: codes.AlreadyExists,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedArtifacts(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.CreateArtifact(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("CreateArtifact(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

// See also TestSpecRevisionArtifacts and TestDeploymentRevisionArtifacts
func TestGetArtifact(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.GetArtifactRequest
		want *rpc.Artifact
	}{
		{
			desc: "fully populated resource",
			seed: &rpc.Artifact{
				Name:      "projects/my-project/locations/global/artifacts/my-artifact",
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContents)),
				Hash:      sha256hash(artifactContents),
				Contents:  artifactContents,
			},
			req: &rpc.GetArtifactRequest{
				Name: "projects/my-project/locations/global/artifacts/my-artifact",
			},
			want: &rpc.Artifact{
				Name:      "projects/my-project/locations/global/artifacts/my-artifact",
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContents)),
				Hash:      sha256hash(artifactContents),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedArtifacts(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

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
		seed *rpc.Artifact
		req  *rpc.GetArtifactRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/my-artifact"},
			req: &rpc.GetArtifactRequest{
				Name: "projects/my-project/locations/global/artifacts/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "invalid name",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/my-artifact"},
			req: &rpc.GetArtifactRequest{
				Name: "invalid",
			},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedArtifacts(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.GetArtifact(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetArtifact(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

// See also TestSpecRevisionArtifacts and TestDeploymentRevisionArtifacts
func TestGetArtifactContents(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.GetArtifactContentsRequest
		want codes.Code
	}{
		{
			desc: "resource not found",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/my-artifact"},
			req: &rpc.GetArtifactContentsRequest{
				Name: "projects/my-project/locations/global/artifacts/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "inappropriate contents suffix in resource name",
			seed: &rpc.Artifact{
				Name:     "projects/my-project/locations/global/artifacts/my-artifact",
				Contents: []byte{},
			},
			req: &rpc.GetArtifactContentsRequest{
				Name: "projects/my-project/locations/global/artifacts/my-artifact/contents",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "gzip mimetype with empty contents",
			seed: &rpc.Artifact{
				Name:     "projects/my-project/locations/global/artifacts/my-artifact",
				MimeType: "application/x.openapi+gzip;version=3.0.0",
				Contents: []byte{},
			},
			req: &rpc.GetArtifactContentsRequest{
				Name: "projects/my-project/locations/global/artifacts/my-artifact",
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

			if _, err := server.GetArtifactContents(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("GetArtifactContents(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestListArtifacts(t *testing.T) {
	tests := []struct {
		admin     bool
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
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact1"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact3"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/artifacts/artifact1"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact1"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact2"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact3"},
				},
			},
		},
		{
			admin: true,
			desc:  "across all version in a artifact project and api",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/artifacts/my-artifact"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/-",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/artifacts/my-artifact"},
				},
			},
		},
		{
			admin: true,
			desc:  "across all apis and version in a artifact project",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v2/artifacts/my-artifact"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/-/versions/-",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/my-project/locations/global/apis/other-api/versions/v2/artifacts/my-artifact"},
				},
			},
		},
		{
			admin: true,
			desc:  "across all projects, apis, and version",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/other-project/locations/global/apis/other-api/versions/v2/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/-/locations/global/apis/-/versions/-",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/other-project/locations/global/apis/other-api/versions/v2/artifacts/my-artifact"},
				},
			},
		},
		{
			admin: true,
			desc:  "in a artifact api and parent across all projects",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/-/locations/global/apis/my-api/versions/v1",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
				},
			},
		},
		{
			admin: true,
			desc:  "in a artifact parent across all projects and apis",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/other-project/locations/global/apis/other-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/-/locations/global/apis/-/versions/v1",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/other-project/locations/global/apis/other-api/versions/v1/artifacts/my-artifact"},
				},
			},
		},
		{
			admin: true,
			desc:  "in all version of a artifact api across all projects",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v2/artifacts/my-artifact"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v1/artifacts/my-artifact"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/-/locations/global/apis/my-api/versions/-",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
					{Name: "projects/other-project/locations/global/apis/my-api/versions/v2/artifacts/my-artifact"},
				},
			},
		},
		{
			desc: "custom page size",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact1"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact3"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent:   "projects/my-project/locations/global/apis/my-api/versions/v1",
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
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact1"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact3"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1",
				Filter: "name == 'projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact2'",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact2"},
				},
			},
		},
		{
			desc: "description inequality filtering",
			seed: []*rpc.Artifact{
				{
					Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact1",
					MimeType: "application/json",
				},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact2"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact3"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1",
				Filter: "mime_type != ''",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{
						Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact1",
						MimeType: "application/json",
					},
				},
			},
		},
		{
			admin: true,
			desc:  "artifacts owned by a project",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/artifacts/artifact1"},
				{Name: "projects/my-project/locations/global/artifacts/artifact2"},
				{Name: "projects/my-project/locations/global/artifacts/artifact3"},
				{Name: "projects/another-project/locations/global/artifacts/artifact4"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/artifacts/artifact1"},
					{Name: "projects/my-project/locations/global/artifacts/artifact2"},
					{Name: "projects/my-project/locations/global/artifacts/artifact3"},
				},
			},
		},
		{
			desc: "artifacts owned by an api",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/apis/a1/artifacts/artifact1"},
				{Name: "projects/my-project/locations/global/apis/a1/artifacts/artifact2"},
				{Name: "projects/my-project/locations/global/apis/a1/artifacts/artifact3"},
				{Name: "projects/my-project/locations/global/apis/a2/artifacts/artifact4"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/a1",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/a1/artifacts/artifact1"},
					{Name: "projects/my-project/locations/global/apis/a1/artifacts/artifact2"},
					{Name: "projects/my-project/locations/global/apis/a1/artifacts/artifact3"},
				},
			},
		},
		{
			desc: "artifacts owned by a version",
			seed: []*rpc.Artifact{
				{Name: "projects/my-project/locations/global/apis/a1/versions/v1/artifacts/artifact1"},
				{Name: "projects/my-project/locations/global/apis/a1/versions/v1/artifacts/artifact2"},
				{Name: "projects/my-project/locations/global/apis/a1/versions/v1/artifacts/artifact3"},
				{Name: "projects/my-project/locations/global/apis/a1/versions/v2/artifacts/artifact4"},
			},
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/a1/versions/v1",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{Name: "projects/my-project/locations/global/apis/a1/versions/v1/artifacts/artifact1"},
					{Name: "projects/my-project/locations/global/apis/a1/versions/v1/artifacts/artifact2"},
					{Name: "projects/my-project/locations/global/apis/a1/versions/v1/artifacts/artifact3"},
				},
			},
		},
		{
			desc: "ordered by mime_type",
			seed: []*rpc.Artifact{
				{
					Name:     "projects/my-project/locations/global/artifacts/artifact1",
					MimeType: "111: this should be returned first",
				},
				{
					Name:     "projects/my-project/locations/global/artifacts/artifact2",
					MimeType: "333: this should be returned third",
				},
				{
					Name:     "projects/my-project/locations/global/artifacts/artifact3",
					MimeType: "222: this should be returned second",
				},
			},
			req: &rpc.ListArtifactsRequest{
				Parent:  "projects/my-project/locations/global",
				OrderBy: "mime_type",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{
						Name:     "projects/my-project/locations/global/artifacts/artifact1",
						MimeType: "111: this should be returned first",
					},
					{
						Name:     "projects/my-project/locations/global/artifacts/artifact3",
						MimeType: "222: this should be returned second",
					},
					{
						Name:     "projects/my-project/locations/global/artifacts/artifact2",
						MimeType: "333: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by mime_type descending",
			seed: []*rpc.Artifact{
				{
					Name:     "projects/my-project/locations/global/artifacts/artifact1",
					MimeType: "111: this should be returned third",
				},
				{
					Name:     "projects/my-project/locations/global/artifacts/artifact2",
					MimeType: "333: this should be returned first",
				},
				{
					Name:     "projects/my-project/locations/global/artifacts/artifact3",
					MimeType: "222: this should be returned second",
				},
			},
			req: &rpc.ListArtifactsRequest{
				Parent:  "projects/my-project/locations/global",
				OrderBy: "mime_type desc",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{
						Name:     "projects/my-project/locations/global/artifacts/artifact2",
						MimeType: "333: this should be returned first",
					},
					{
						Name:     "projects/my-project/locations/global/artifacts/artifact3",
						MimeType: "222: this should be returned second",
					},
					{
						Name:     "projects/my-project/locations/global/artifacts/artifact1",
						MimeType: "111: this should be returned third",
					},
				},
			},
		},
		{
			desc: "ordered by mime_type then by name",
			seed: []*rpc.Artifact{
				{
					Name:     "projects/my-project/locations/global/artifacts/artifact1",
					MimeType: "222: this should be returned second or third (the name is the tie-breaker)",
				},
				{
					Name:     "projects/my-project/locations/global/artifacts/artifact3",
					MimeType: "111: this should be returned first",
				},
				{
					Name:     "projects/my-project/locations/global/artifacts/artifact2",
					MimeType: "222: this should be returned second or third (the name is the tie-breaker)",
				},
			},
			req: &rpc.ListArtifactsRequest{
				Parent:  "projects/my-project/locations/global",
				OrderBy: "mime_type,name",
			},
			want: &rpc.ListArtifactsResponse{
				Artifacts: []*rpc.Artifact{
					{
						Name:     "projects/my-project/locations/global/artifacts/artifact3",
						MimeType: "111: this should be returned first",
					},
					{
						Name:     "projects/my-project/locations/global/artifacts/artifact1",
						MimeType: "222: this should be returned second or third (the name is the tie-breaker)",
					},
					{
						Name:     "projects/my-project/locations/global/artifacts/artifact2",
						MimeType: "222: this should be returned second or third (the name is the tie-breaker)",
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
			if err := seeder.SeedArtifacts(ctx, server, test.seed...); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.ListArtifacts(ctx, test.req)
			if err != nil {
				t.Fatalf("ListArtifacts(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListArtifactsResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
				test.extraOpts,
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListArtifacts(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListArtifacts(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				t.Errorf("ListArtifacts(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}
		})
	}
}

func TestListArtifactsResponseCodes(t *testing.T) {
	tests := []struct {
		admin bool
		desc  string
		seed  *rpc.Artifact
		req   *rpc.ListArtifactsRequest
		want  codes.Code
	}{
		{
			desc: "parent spec not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/s",
			},
			want: codes.NotFound,
		},
		{
			desc: "parent spec rev not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/s@123",
			},
			want: codes.NotFound,
		},
		{
			desc: "parent deployment not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/deployments/d",
			},
			want: codes.NotFound,
		},
		{
			desc: "parent deployment rev not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/deployments/d@123",
			},
			want: codes.NotFound,
		},
		{
			desc: "parent version not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/v1",
			},
			want: codes.NotFound,
		},
		{
			desc: "parent api not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/my-api/versions/-",
			},
			want: codes.NotFound,
		},
		{
			admin: true,
			desc:  "parent project not found",
			req: &rpc.ListArtifactsRequest{
				Parent: "projects/my-project/locations/global/apis/-/versions/-",
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
		{
			desc: "invalid ordering by unknown field",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/my-artifact"},
			req: &rpc.ListArtifactsRequest{
				Parent:  "projects/my-project/locations/global",
				OrderBy: "something",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering by private field",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/my-artifact"},
			req: &rpc.ListArtifactsRequest{
				Parent:  "projects/my-project/locations/global",
				OrderBy: "key",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering direction",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/my-artifact"},
			req: &rpc.ListArtifactsRequest{
				Parent:  "projects/my-project/locations/global",
				OrderBy: "description asc",
			},
			want: codes.InvalidArgument,
		},
		{
			desc: "invalid ordering format",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/my-artifact"},
			req: &rpc.ListArtifactsRequest{
				Parent:  "projects/my-project/locations/global",
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
			if err := seeder.SeedArtifacts(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

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
		{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact1"},
		{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact2"},
		{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/artifact3"},
	}
	if err := seeder.SeedArtifacts(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	listed := make([]*rpc.Artifact, 0, 3)

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListArtifactsRequest{
			Parent:   "projects/my-project/locations/global/apis/my-api/versions/v1",
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
			Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
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
			Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
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
			t.Errorf("ListArtifacts(%+v) returned next_page_token, expected no next page", req)
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

func TestListArtifactsLargeCollection(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := make([]*rpc.Artifact, 0, 1001)
	for i := 1; i <= cap(seed); i++ {
		seed = append(seed, &rpc.Artifact{
			Name: fmt.Sprintf("projects/my-project/locations/global/artifacts/a%03d", i),
		})
	}

	if err := seeder.SeedArtifacts(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	// This test prevents the list sequence from ending before a known filter match is listed.
	// For simplicity, it does not guarantee the resource is returned on a later page.
	t.Run("filter", func(t *testing.T) {
		req := &rpc.ListArtifactsRequest{
			Parent:   "projects/my-project/locations/global",
			PageSize: 1,
			Filter:   "name == 'projects/my-project/locations/global/artifacts/a099'",
		}

		got, err := server.ListArtifacts(ctx, req)
		if err != nil {
			t.Fatalf("ListArtifacts(%+v) returned error: %s", req, err)
		}

		if len(got.GetArtifacts()) == 1 && got.GetNextPageToken() != "" {
			t.Errorf("ListArtifacts(%+v) returned a page token when the only matching resource has been listed: %+v", req, got)
		} else if len(got.GetArtifacts()) == 0 && got.GetNextPageToken() == "" {
			t.Errorf("ListArtifacts(%+v) returned an empty next page token before listing the only matching resource", req)
		} else if count := len(got.GetArtifacts()); count > 1 {
			t.Errorf("ListArtifacts(%+v) returned %d projects, expected at most one: %+v", req, count, got.GetArtifacts())
		}
	})

	t.Run("max page size", func(t *testing.T) {
		req := &rpc.ListArtifactsRequest{
			Parent:   "projects/my-project/locations/global",
			PageSize: 1001,
		}

		got, err := server.ListArtifacts(ctx, req)
		if err != nil {
			t.Fatalf("ListArtifacts(%+v) returned error: %s", req, err)
		}

		if len(got.GetArtifacts()) != 1000 {
			t.Errorf("ListArtifacts(%+v) should have returned 1000 items, got: %+v", req, len(got.GetArtifacts()))
		} else if got.GetNextPageToken() == "" {
			t.Errorf("ListArtifacts(%+v) should return a next page token", req)
		}
	})
}

func TestReplaceArtifact(t *testing.T) {
	tests := []struct {
		desc string
		seed *rpc.Artifact
		req  *rpc.ReplaceArtifactRequest
		want *rpc.Artifact
	}{
		{
			desc: "fully populated resource",
			seed: &rpc.Artifact{
				Name: "projects/my-project/locations/global/artifacts/my-artifact",
			},
			req: &rpc.ReplaceArtifactRequest{
				Artifact: &rpc.Artifact{
					Name:      "projects/my-project/locations/global/artifacts/my-artifact",
					MimeType:  "application/json",
					SizeBytes: int32(len(artifactContents)),
					Hash:      sha256hash(artifactContents),
					Contents:  artifactContents,
				},
			},
			want: &rpc.Artifact{
				Name:      "projects/my-project/locations/global/artifacts/my-artifact",
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContents)),
				Hash:      sha256hash(artifactContents),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedArtifacts(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

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
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
			req: &rpc.ReplaceArtifactRequest{
				Artifact: &rpc.Artifact{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/doesnt-exist",
				},
			},
			want: codes.NotFound,
		},
		{
			desc: "missing resource body",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
			req:  &rpc.ReplaceArtifactRequest{},
			want: codes.InvalidArgument,
		},
		{
			desc: "missing resource name",
			seed: &rpc.Artifact{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact"},
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
			if err := seeder.SeedArtifacts(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			if _, err := server.ReplaceArtifact(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("ReplaceArtifact(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}

func TestReplaceArtifactSequence(t *testing.T) {
	tests := []struct {
		desc string
		req  *rpc.ReplaceArtifactRequest
		want codes.Code
	}{
		{
			desc: "first replacement",
			req: &rpc.ReplaceArtifactRequest{
				Artifact: &rpc.Artifact{
					Name: "projects/my-project/locations/global/artifacts/a",
				},
			},
			want: codes.OK,
		},
		{
			desc: "second replacement",
			req: &rpc.ReplaceArtifactRequest{
				Artifact: &rpc.Artifact{
					Name: "projects/my-project/locations/global/artifacts/a",
				},
			},
			want: codes.OK,
		},
	}
	ctx := context.Background()
	server := defaultTestServer(t)
	seed := &rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/a"}
	if err := seeder.SeedArtifacts(ctx, server, seed); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	a, err := server.GetArtifact(ctx, &rpc.GetArtifactRequest{
		Name: "projects/my-project/locations/global/artifacts/a",
	})
	if err != nil {
		t.Fatalf("Setup/Seeding: Failed to get seeded artifact: %s", err)
	}
	createTime := a.CreateTime.AsTime()
	updateTime := a.UpdateTime.AsTime()
	// NOTE: in the following sequence of tests, each test depends on its predecessor.
	// Resources are successively updated using the "Replace" RPC and the
	// tests verify that CreateTime/UpdateTime fields are modified appropriately.
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var result *rpc.Artifact
			var err error
			if result, err = server.ReplaceArtifact(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("UpdateApi(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
			if result != nil {
				if !createTime.Equal(result.CreateTime.AsTime()) {
					t.Errorf("ReplaceArtifact create time changed after replace (%v %v)", createTime, result.CreateTime.AsTime())
				}
				if !updateTime.Before(result.UpdateTime.AsTime()) {
					t.Errorf("ReplaceArtifact update time did not increase after replace (%v %v)", updateTime, result.UpdateTime.AsTime())
				}
				updateTime = result.UpdateTime.AsTime()
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
			desc: "existing resource",
			seed: &rpc.Artifact{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact",
			},
			req: &rpc.DeleteArtifactRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/my-artifact",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedArtifacts(ctx, server, test.seed); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

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
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/artifacts/doesnt-exist",
			},
			want: codes.NotFound,
		},
		{
			desc: "invalid name",
			req: &rpc.DeleteArtifactRequest{
				Name: "invalid",
			},
			want: codes.InvalidArgument,
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

func TestSpecRevisionArtifacts(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	if err := seeder.SeedVersions(ctx, server,
		&rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version"}); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	createSpec := &rpc.CreateApiSpecRequest{
		Parent:    "projects/my-project/locations/global/apis/my-api/versions/my-version",
		ApiSpecId: "my-spec",
		ApiSpec: &rpc.ApiSpec{
			Description: "Empty First Revision",
		},
	}
	spec1r1, err := server.CreateApiSpec(ctx, createSpec)
	if err != nil {
		t.Fatalf("Setup: CreateApiSpec(%+v) returned error: %s", createSpec, err)
	}

	updateSpec := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:        "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec",
			Contents:    specContents,
			Description: "Second revision",
		},
	}
	spec1r2, err := server.UpdateApiSpec(ctx, updateSpec)
	if err != nil {
		t.Fatalf("Setup: UpdateApiSpec(%+v) returned error: %s", updateSpec, err)
	}

	createSpec2 := &rpc.CreateApiSpecRequest{
		Parent:    createSpec.Parent,
		ApiSpecId: "my-spec2",
		ApiSpec: &rpc.ApiSpec{
			Description: "Empty First Revision",
		},
	}
	spec2r1, err := server.CreateApiSpec(ctx, createSpec2)
	if err != nil {
		t.Fatalf("Setup: CreateApiSpec(%+v) returned error: %s", createSpec2, err)
	}

	artifactContentsR1 := []byte(`{"contents": "r1"}`)
	artifactContentsR2 := []byte(`{"contents": "r2"}`)

	createArtifact := &rpc.CreateArtifactRequest{
		Parent:     spec2r1.Name,
		ArtifactId: "my-artifact-r1",
		Artifact: &rpc.Artifact{
			MimeType:  "application/json",
			SizeBytes: int32(len(artifactContentsR1)),
			Hash:      sha256hash(artifactContentsR1),
			Contents:  artifactContentsR1,
		},
	}
	_, err = server.CreateArtifact(ctx, createArtifact)
	if err != nil {
		t.Fatalf("CreateArtifact(%+v) returned error: %s", createArtifact, err)
	}

	t.Run("create under default revision", func(t *testing.T) {
		createArtifact := &rpc.CreateArtifactRequest{
			Parent:     "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec",
			ArtifactId: "my-artifact-r2",
			Artifact: &rpc.Artifact{
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContentsR2)),
				Hash:      sha256hash(artifactContentsR2),
				Contents:  artifactContentsR2,
			},
		}
		artifact, err := server.CreateArtifact(ctx, createArtifact)
		if err != nil {
			t.Fatalf("CreateArtifact(%+v) returned error: %s", createArtifact, err)
		}

		want := &rpc.Artifact{
			Name:      fmt.Sprintf(createArtifact.Parent+"@%s/artifacts/my-artifact-r2", spec1r2.RevisionId),
			MimeType:  "application/json",
			SizeBytes: int32(len(artifactContentsR2)),
			Hash:      sha256hash(artifactContentsR2),
		}

		opts := cmp.Options{
			protocmp.Transform(),
			protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
		}

		if !cmp.Equal(want, artifact, opts) {
			t.Errorf("CreateArtifact(%+v) returned unexpected diff (-want +got):\n%s", createArtifact, cmp.Diff(want, artifact, opts))
		}
	})

	t.Run("create under a specified revision", func(t *testing.T) {
		createArtifact := &rpc.CreateArtifactRequest{
			Parent:     "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId(),
			ArtifactId: "my-artifact-r1",
			Artifact: &rpc.Artifact{
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContentsR1)),
				Hash:      sha256hash(artifactContentsR1),
				Contents:  artifactContentsR1,
			},
		}
		artifact, err := server.CreateArtifact(ctx, createArtifact)
		if err != nil {
			t.Fatalf("CreateArtifact(%+v) returned error: %s", createArtifact, err)
		}

		want := &rpc.Artifact{
			Name:      createArtifact.Parent + "/artifacts/my-artifact-r1",
			MimeType:  "application/json",
			SizeBytes: int32(len(artifactContentsR1)),
			Hash:      sha256hash(artifactContentsR1),
		}

		opts := cmp.Options{
			protocmp.Transform(),
			protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
		}

		if !cmp.Equal(want, artifact, opts) {
			t.Errorf("CreateArtifact(%+v) returned unexpected diff (-want +got):\n%s", createArtifact, cmp.Diff(want, artifact, opts))
		}
	})

	t.Run("get artifact", func(t *testing.T) {
		tests := []struct {
			desc string
			req  *rpc.GetArtifactRequest
			want *rpc.Artifact
		}{
			{
				desc: "specified revision",
				req: &rpc.GetArtifactRequest{
					Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId() + "/artifacts/my-artifact-r1",
				},
				want: &rpc.Artifact{
					Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId() + "/artifacts/my-artifact-r1",
				},
			},
			{
				desc: "latest revision",
				req: &rpc.GetArtifactRequest{
					Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2",
				},
				want: &rpc.Artifact{
					Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2",
				},
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				got, err := server.GetArtifact(ctx, test.req)
				if err != nil {
					t.Fatalf("GetArtifact(%+v) returned error: %s", test.req, err)
				}

				opts := cmp.Options{
					protocmp.Transform(),
					protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time", "hash", "mime_type", "size_bytes"),
				}

				if !cmp.Equal(test.want, got, opts) {
					t.Errorf("GetArtifact(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
				}
			})
		}
	})

	t.Run("list artifacts across", func(t *testing.T) {
		tests := []struct {
			admin bool
			desc  string
			req   *rpc.ListArtifactsRequest
			want  *rpc.ListArtifactsResponse
		}{
			{
				desc: "specified spec revision",
				req: &rpc.ListArtifactsRequest{
					Parent: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId(),
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
			{
				desc: "latest spec revision",
				req: &rpc.ListArtifactsRequest{
					Parent: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
					},
				},
			},
			{
				desc: "all spec revisions",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
			{
				desc: "latest revisions of all specs",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec2@" + spec2r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
					},
				},
			},
			{
				desc: "all revisions of all specs",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/-@-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec2@" + spec2r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
			{
				desc: "all revisions in all versions",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/my-project/locations/global/apis/my-api/versions/-/specs/-@-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec2@" + spec2r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
			{
				desc: "all revisions in all apis",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/my-project/locations/global/apis/-/versions/-/specs/-@-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec2@" + spec2r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
			{
				admin: true,
				desc:  "all revisions in all projects",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/-/locations/global/apis/-/versions/-/specs/-@-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec2@" + spec2r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
						{Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				if test.admin && adminServiceUnavailable() {
					t.Skip(testRequiresAdminService)
				}
				got, err := server.ListArtifacts(ctx, test.req)
				if err != nil {
					t.Fatalf("ListArtifacts(%+v) returned error: %s", test.req, err)
				}

				opts := cmp.Options{
					protocmp.Transform(),
					protocmp.IgnoreFields(new(rpc.ListArtifactsResponse), "next_page_token"),
					protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time", "hash", "mime_type", "size_bytes"),
				}

				if !cmp.Equal(test.want, got, opts) {
					t.Errorf("ListArtifacts(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
				}
			})
		}
	})

	t.Run("get contents", func(t *testing.T) {
		tests := []struct {
			desc string
			req  *rpc.GetArtifactContentsRequest
			want []byte
		}{
			{
				desc: "specified revision",
				req: &rpc.GetArtifactContentsRequest{
					Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r1.GetRevisionId() + "/artifacts/my-artifact-r1",
				},
				want: artifactContentsR1,
			},
			{
				desc: "latest revision",
				req: &rpc.GetArtifactContentsRequest{
					Name: "projects/my-project/locations/global/apis/my-api/versions/my-version/specs/my-spec@" + spec1r2.GetRevisionId() + "/artifacts/my-artifact-r2",
				},
				want: artifactContentsR2,
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				got, err := server.GetArtifactContents(ctx, test.req)
				if err != nil {
					t.Fatalf("GetArtifact(%+v) returned error: %s", test.req, err)
				}

				if !cmp.Equal(test.want, got.Data, nil) {
					t.Errorf("GetArtifactContents(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got.Data, nil))
				}
			})
		}
	})
}

func TestDeploymentRevisionArtifacts(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	if err := seeder.SeedApis(ctx, server,
		&rpc.Api{Name: "projects/my-project/locations/global/apis/my-api"}); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	createDeployment := &rpc.CreateApiDeploymentRequest{
		Parent:          "projects/my-project/locations/global/apis/my-api",
		ApiDeploymentId: "my-deployment",
		ApiDeployment: &rpc.ApiDeployment{
			Description: "Empty First Revision",
		},
	}
	deployment1r1, err := server.CreateApiDeployment(ctx, createDeployment)
	if err != nil {
		t.Fatalf("Setup: CreateApiDeployment(%+v) returned error: %s", createDeployment, err)
	}

	updateDeployment := &rpc.UpdateApiDeploymentRequest{
		ApiDeployment: &rpc.ApiDeployment{
			Name:        "projects/my-project/locations/global/apis/my-api/deployments/my-deployment",
			Description: "Second revision",
			EndpointUri: "updated",
		},
	}
	deployment1r2, err := server.UpdateApiDeployment(ctx, updateDeployment)
	if err != nil {
		t.Fatalf("Setup: UpdateApiDeployment(%+v) returned error: %s", updateDeployment, err)
	}

	createDeployment2 := &rpc.CreateApiDeploymentRequest{
		Parent:          "projects/my-project/locations/global/apis/my-api",
		ApiDeploymentId: "my-deployment2",
		ApiDeployment: &rpc.ApiDeployment{
			Description: "Empty First Revision",
		},
	}
	deployment2r1, err := server.CreateApiDeployment(ctx, createDeployment2)
	if err != nil {
		t.Fatalf("Setup: CreateApiDeployment(%+v) returned error: %s", createDeployment, err)
	}

	artifactContentsR1 := []byte(`{"contents": "r1"}`)
	artifactContentsR2 := []byte(`{"contents": "r2"}`)

	createArtifact := &rpc.CreateArtifactRequest{
		Parent:     deployment2r1.Name,
		ArtifactId: "my-artifact-r1",
		Artifact: &rpc.Artifact{
			MimeType:  "application/json",
			SizeBytes: int32(len(artifactContentsR1)),
			Hash:      sha256hash(artifactContentsR1),
			Contents:  artifactContentsR1,
		},
	}
	_, err = server.CreateArtifact(ctx, createArtifact)
	if err != nil {
		t.Fatalf("CreateArtifact(%+v) returned error: %s", createArtifact, err)
	}

	t.Run("create under default revision", func(t *testing.T) {
		createArtifact := &rpc.CreateArtifactRequest{
			Parent:     "projects/my-project/locations/global/apis/my-api/deployments/my-deployment",
			ArtifactId: "my-artifact-r2",
			Artifact: &rpc.Artifact{
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContentsR2)),
				Hash:      sha256hash(artifactContentsR2),
				Contents:  artifactContentsR2,
			},
		}
		artifact, err := server.CreateArtifact(ctx, createArtifact)
		if err != nil {
			t.Fatalf("CreateArtifact(%+v) returned error: %s", createArtifact, err)
		}

		want := &rpc.Artifact{
			Name:      fmt.Sprintf(createArtifact.Parent+"@%s/artifacts/my-artifact-r2", deployment1r2.RevisionId),
			MimeType:  "application/json",
			SizeBytes: int32(len(artifactContentsR2)),
			Hash:      sha256hash(artifactContentsR2),
		}

		opts := cmp.Options{
			protocmp.Transform(),
			protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
		}

		if !cmp.Equal(want, artifact, opts) {
			t.Errorf("CreateArtifact(%+v) returned unexpected diff (-want +got):\n%s", createArtifact, cmp.Diff(want, artifact, opts))
		}
	})

	t.Run("create under a specified revision", func(t *testing.T) {
		createArtifact := &rpc.CreateArtifactRequest{
			Parent:     "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId(),
			ArtifactId: "my-artifact-r1",
			Artifact: &rpc.Artifact{
				MimeType:  "application/json",
				SizeBytes: int32(len(artifactContentsR1)),
				Hash:      sha256hash(artifactContentsR1),
				Contents:  artifactContentsR1,
			},
		}
		artifact, err := server.CreateArtifact(ctx, createArtifact)
		if err != nil {
			t.Fatalf("CreateArtifact(%+v) returned error: %s", createArtifact, err)
		}

		want := &rpc.Artifact{
			Name:      createArtifact.Parent + "/artifacts/my-artifact-r1",
			MimeType:  "application/json",
			SizeBytes: int32(len(artifactContentsR1)),
			Hash:      sha256hash(artifactContentsR1),
		}

		opts := cmp.Options{
			protocmp.Transform(),
			protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time"),
		}

		if !cmp.Equal(want, artifact, opts) {
			t.Errorf("CreateArtifact(%+v) returned unexpected diff (-want +got):\n%s", createArtifact, cmp.Diff(want, artifact, opts))
		}
	})

	t.Run("get artifact", func(t *testing.T) {
		tests := []struct {
			desc string
			req  *rpc.GetArtifactRequest
			want *rpc.Artifact
		}{
			{
				desc: "specified revision",
				req: &rpc.GetArtifactRequest{
					Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId() + "/artifacts/my-artifact-r1",
				},
				want: &rpc.Artifact{
					Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId() + "/artifacts/my-artifact-r1",
				},
			},
			{
				desc: "latest revision",
				req: &rpc.GetArtifactRequest{
					Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r2.GetRevisionId() + "/artifacts/my-artifact-r2",
				},
				want: &rpc.Artifact{
					Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r2.GetRevisionId() + "/artifacts/my-artifact-r2",
				},
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				got, err := server.GetArtifact(ctx, test.req)
				if err != nil {
					t.Fatalf("GetArtifact(%+v) returned error: %s", test.req, err)
				}

				opts := cmp.Options{
					protocmp.Transform(),
					protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time", "hash", "mime_type", "size_bytes"),
				}

				if !cmp.Equal(test.want, got, opts) {
					t.Errorf("GetArtifact(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
				}
			})
		}
	})

	t.Run("list artifacts across", func(t *testing.T) {
		tests := []struct {
			admin bool
			desc  string
			req   *rpc.ListArtifactsRequest
			want  *rpc.ListArtifactsResponse
		}{
			{
				desc: "specified deployment revision",
				req: &rpc.ListArtifactsRequest{
					Parent: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId(),
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
			{
				desc: "latest deployment revision",
				req: &rpc.ListArtifactsRequest{
					Parent: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
					},
				},
			},
			{
				desc: "all deployment revisions",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
			{
				desc: "latest revisions of all deployments",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/my-project/locations/global/apis/my-api/deployments/-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment2@" + deployment2r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
					},
				},
			},
			{
				desc: "all revisions of all deployments",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/my-project/locations/global/apis/my-api/deployments/-@-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment2@" + deployment2r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
			{
				desc: "all revisions in all apis",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/my-project/locations/global/apis/-/deployments/-@-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment2@" + deployment2r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
			{
				admin: true,
				desc:  "all revisions in all projects",
				req: &rpc.ListArtifactsRequest{
					Parent:  "projects/-/locations/global/apis/-/deployments/-@-",
					OrderBy: "create_time",
				},
				want: &rpc.ListArtifactsResponse{
					Artifacts: []*rpc.Artifact{
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment2@" + deployment2r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r2.GetRevisionId() + "/artifacts/my-artifact-r2"},
						{Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId() + "/artifacts/my-artifact-r1"},
					},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				if test.admin && adminServiceUnavailable() {
					t.Skip(testRequiresAdminService)
				}
				got, err := server.ListArtifacts(ctx, test.req)
				if err != nil {
					t.Fatalf("ListArtifacts(%+v) returned error: %s", test.req, err)
				}

				opts := cmp.Options{
					protocmp.Transform(),
					protocmp.IgnoreFields(new(rpc.ListArtifactsResponse), "next_page_token"),
					protocmp.IgnoreFields(new(rpc.Artifact), "create_time", "update_time", "hash", "mime_type", "size_bytes"),
				}

				if !cmp.Equal(test.want, got, opts) {
					t.Errorf("ListArtifacts(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
				}
			})
		}
	})

	t.Run("get contents", func(t *testing.T) {
		tests := []struct {
			desc string
			req  *rpc.GetArtifactContentsRequest
			want []byte
		}{
			{
				desc: "specified revision",
				req: &rpc.GetArtifactContentsRequest{
					Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r1.GetRevisionId() + "/artifacts/my-artifact-r1",
				},
				want: artifactContentsR1,
			},
			{
				desc: "latest revision",
				req: &rpc.GetArtifactContentsRequest{
					Name: "projects/my-project/locations/global/apis/my-api/deployments/my-deployment@" + deployment1r2.GetRevisionId() + "/artifacts/my-artifact-r2",
				},
				want: artifactContentsR2,
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				got, err := server.GetArtifactContents(ctx, test.req)
				if err != nil {
					t.Fatalf("GetArtifact(%+v) returned error: %s", test.req, err)
				}

				if !cmp.Equal(test.want, got.Data, nil) {
					t.Errorf("GetArtifactContents(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got.Data, nil))
				}
			})
		}
	})
}
