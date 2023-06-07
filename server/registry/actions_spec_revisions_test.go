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
	"strings"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestTagApiSpecRevision(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	if err := seeder.SeedSpecs(ctx, server, &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"}); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	updateReq := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			Contents: specContents,
		},
	}

	revision, err := server.UpdateApiSpec(ctx, updateReq)
	if err != nil {
		t.Fatalf("Setup: UpdateApiSpec(%+v) returned error: %s", updateReq, err)
	}

	req := &rpc.TagApiSpecRevisionRequest{
		Name: fmt.Sprintf("%s@%s", revision.GetName(), revision.GetRevisionId()),
		Tag:  "my-tag",
	}

	got, err := server.TagApiSpecRevision(ctx, req)
	if err != nil {
		t.Fatalf("TagApiSpecRevision(%+v) returned error: %s", req, err)
	}

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(revision, "name"),
	}

	t.Run("response", func(t *testing.T) {
		if !cmp.Equal(revision, got, opts) {
			t.Errorf("TagApiSpecRevision(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(revision, got, opts))
		}

		if want := fmt.Sprintf("%s@my-tag", revision.GetName()); want != got.GetName() {
			t.Errorf("TagApiSpecRevision(%+v) returned unexpected name %q, want %q", req, got.GetName(), want)
		}
	})

	t.Run("GetApiSpec", func(t *testing.T) {
		req := &rpc.GetApiSpecRequest{
			Name: got.GetName(),
		}

		got, err := server.GetApiSpec(ctx, req)
		if err != nil {
			t.Fatalf("GetApiSpec(%+v) returned error: %s", req, err)
		}

		if !cmp.Equal(revision, got, opts) {
			t.Errorf("GetApiSpec(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(revision, got, opts))
		}

		if got.GetName() != req.GetName() {
			t.Errorf("GetApiSpec(%+v) returned unexpected name %q, want %q", req, got.GetName(), req.GetName())
		}
	})

	t.Run("add another tag to a tagged revision", func(t *testing.T) {
		req := &rpc.TagApiSpecRevisionRequest{
			Name: got.GetName(),
			Tag:  "my-second-tag",
		}

		got, err := server.TagApiSpecRevision(ctx, req)
		if err != nil {
			t.Fatalf("TagApiSpecRevision(%+v) returned error: %s", req, err)
		}

		opts := cmp.Options{
			protocmp.Transform(),
			protocmp.IgnoreFields(revision, "name"),
		}

		if !cmp.Equal(revision, got, opts) {
			t.Errorf("TagApiSpecRevision(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(revision, got, opts))
		}

		if want := fmt.Sprintf("%s@my-second-tag", revision.GetName()); want != got.GetName() {
			t.Errorf("TagApiSpecRevision(%+v) returned unexpected name %q, want %q", req, got.GetName(), want)
		}

		t.Run("GetApiSpec", func(t *testing.T) {
			req := &rpc.GetApiSpecRequest{
				Name: got.GetName(),
			}

			got, err := server.GetApiSpec(ctx, req)
			if err != nil {
				t.Fatalf("GetApiSpec(%+v) returned error: %s", req, err)
			}

			if !cmp.Equal(revision, got, opts) {
				t.Errorf("GetApiSpec(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(revision, got, opts))
			}

			if got.GetName() != req.GetName() {
				t.Errorf("GetApiSpec(%+v) returned unexpected name %q, want %q", req, got.GetName(), req.GetName())
			}
		})
	})

	t.Run("DeleteApiSpecRevision", func(t *testing.T) {
		req := &rpc.DeleteApiSpecRevisionRequest{
			Name: got.GetName(),
		}

		if _, err := server.DeleteApiSpecRevision(ctx, req); err != nil {
			t.Fatalf("DeleteApiSpecRevision(%+v) returned error: %s", req, err)
		}

		t.Run("GetApiSpec", func(t *testing.T) {
			req := &rpc.GetApiSpecRequest{
				Name: req.GetName(),
			}

			if _, err := server.GetApiSpec(ctx, req); status.Code(err) != codes.NotFound {
				t.Fatalf("GetApiSpec(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.NotFound, err)
			}
		})
	})
}

func TestTagApiSpecRevisionResponseCodes(t *testing.T) {
	tests := []struct {
		desc string
		tag  string
		want codes.Code
	}{
		{
			desc: "empty tag",
			tag:  "",
			want: codes.InvalidArgument,
		},
		{
			desc: "too long",
			tag:  strings.Repeat("x", 41),
			want: codes.InvalidArgument,
		},
		{
			desc: "contains uppercase leters",
			tag:  "TestTag",
			want: codes.InvalidArgument,
		},
		{
			desc: "single dash",
			tag:  "-",
			want: codes.InvalidArgument,
		},
		{
			desc: "valid one-character tag",
			tag:  "x",
			want: codes.OK,
		},
		{
			desc: "valid tag",
			tag:  "latest",
			want: codes.OK,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)
			if err := seeder.SeedSpecs(ctx, server, &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s"}); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}
			if _, err := server.TagApiSpecRevision(ctx, &rpc.TagApiSpecRevisionRequest{
				Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s",
				Tag:  test.tag,
			}); status.Code(err) != test.want {
				t.Errorf("TagApiSpecRevision(%+v) returned status code %q, want %q: %v", test.tag, status.Code(err), test.want, err)
			}
		})
	}

	t.Run("invalid revision name", func(t *testing.T) {
		ctx := context.Background()
		server := defaultTestServer(t)
		if err := seeder.SeedSpecs(ctx, server, &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s"}); err != nil {
			t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
		}
		if _, err := server.TagApiSpecRevision(ctx, &rpc.TagApiSpecRevisionRequest{
			Name: "invalid",
			Tag:  "test",
		}); status.Code(err) != codes.InvalidArgument {
			t.Errorf("TagApiSpecRevision(%+v) returned status code %q, want %q: %v", "test", status.Code(err), codes.InvalidArgument, err)
		}
	})

	t.Run("missing revision", func(t *testing.T) {
		ctx := context.Background()
		server := defaultTestServer(t)
		if err := seeder.SeedSpecs(ctx, server, &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s"}); err != nil {
			t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
		}
		if _, err := server.TagApiSpecRevision(ctx, &rpc.TagApiSpecRevisionRequest{
			Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s@9999",
			Tag:  "test",
		}); status.Code(err) != codes.NotFound {
			t.Errorf("TagApiSpecRevision(%+v) returned status code %q, want %q: %v", "test", status.Code(err), codes.NotFound, err)
		}
	})
}

func TestRollbackApiSpec(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	if err := seeder.SeedVersions(ctx, server, &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"}); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	createReq := &rpc.CreateApiSpecRequest{
		Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
		ApiSpecId: "my-spec",
		ApiSpec:   &rpc.ApiSpec{},
	}

	firstRevision, err := server.CreateApiSpec(ctx, createReq)
	if err != nil {
		t.Fatalf("Setup: CreateApiSpec(%+v) returned error: %s", createReq, err)
	}

	// Create a new revision so we can roll back from it.
	updateReq := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:     firstRevision.GetName(),
			Contents: specContents,
		},
	}

	secondRevision, err := server.UpdateApiSpec(ctx, updateReq)
	if err != nil {
		t.Fatalf("Setup: UpdateApiSpec(%+v) returned error: %s", updateReq, err)
	}

	if secondRevision.GetRevisionId() == firstRevision.GetRevisionId() {
		t.Fatalf("Setup: UpdateApiSpec(%+v) returned unexpected revision_id %q matching first revision, expected new revision ID", updateReq, secondRevision.GetRevisionId())
	}

	req := &rpc.RollbackApiSpecRequest{
		Name:       secondRevision.GetName(),
		RevisionId: firstRevision.GetRevisionId(),
	}

	rollback, err := server.RollbackApiSpec(ctx, req)
	if err != nil {
		t.Fatalf("RollbackApiSpec(%+v) returned error: %s", req, err)
	}

	want := &rpc.ApiSpec{
		Name:               fmt.Sprintf("%s@%s", firstRevision.GetName(), rollback.GetRevisionId()),
		Hash:               firstRevision.GetHash(),
		SizeBytes:          firstRevision.GetSizeBytes(),
		CreateTime:         firstRevision.GetCreateTime(),
		RevisionCreateTime: firstRevision.GetRevisionCreateTime(),
		RevisionUpdateTime: firstRevision.GetRevisionUpdateTime(),
	}

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(rpc.ApiSpec), "revision_id", "revision_create_time", "revision_update_time"),
	}

	if !cmp.Equal(want, rollback, opts) {
		t.Errorf("RollbackApiSpec(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(want, rollback, opts))
	}

	// Rollback should create a new revision, i.e. it should not reuse an existing revision ID.
	if rollback.GetRevisionId() == firstRevision.GetRevisionId() {
		t.Fatalf("RollbackApiSpec(%+v) returned unexpected revision_id %q matching first revision, expected new revision ID", req, rollback.GetRevisionId())
	} else if rollback.GetRevisionId() == secondRevision.GetRevisionId() {
		t.Fatalf("RollbackApiSpec(%+v) returned unexpected revision_id %q matching second revision, expected new revision ID", req, rollback.GetRevisionId())
	}
}

func TestRollbackApiSpecRevisionResponseCodes(t *testing.T) {
	tests := []struct {
		desc       string
		name       string
		revisionID string
		want       codes.Code
	}{
		{
			desc:       "empty revisionID",
			name:       "projects/my-project/locations/global/apis/my-api/versions/v1/specs/s",
			revisionID: "",
			want:       codes.InvalidArgument,
		},
		{
			desc:       "invalid name",
			name:       "invalid",
			revisionID: "whatever",
			want:       codes.InvalidArgument,
		},
		{
			desc:       "missing revisionID",
			name:       "projects/my-project/locations/global/apis/my-api/versions/v1/specs/s",
			revisionID: "revision",
			want:       codes.NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			server := defaultTestServer(t)

			req := &rpc.RollbackApiSpecRequest{
				Name:       test.name,
				RevisionId: test.revisionID,
			}

			if _, err := server.RollbackApiSpec(ctx, req); status.Code(err) != test.want {
				t.Errorf("RollbackApiSpec(%+v) returned status code %q, want %q: %v", test.name, status.Code(err), test.want, err)
			}
		})
	}
}

func TestDeleteApiSpecRevision(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	if err := seeder.SeedSpecs(ctx, server, &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"}); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	getReq := &rpc.GetApiSpecRequest{
		Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
	}
	firstRevision, err := server.GetApiSpec(ctx, getReq)
	if err != nil {
		t.Fatalf("Setup: GetApiSpecRequest(%+v) returned error: %s", getReq, err)
	}

	t.Run("invalid name", func(t *testing.T) {
		req := &rpc.DeleteApiSpecRevisionRequest{
			Name: "invalid",
		}

		if _, err := server.DeleteApiSpecRevision(ctx, req); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("DeleteApiSpecRevision(%+v) returned unexpected status code %q, want %q: %v", req, status.Code(err), codes.FailedPrecondition, err)
		}
	})

	t.Run("only remaining revision", func(t *testing.T) {
		req := &rpc.DeleteApiSpecRevisionRequest{
			Name: fmt.Sprintf("projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec@%s", firstRevision.GetRevisionId()),
		}

		if _, err := server.DeleteApiSpecRevision(ctx, req); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("DeleteApiSpecRevision(%+v) returned unexpected status code %q, want %q: %v", req, status.Code(err), codes.FailedPrecondition, err)
		}
	})

	// Create a new revision so we can delete it.
	updateReq := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
			Contents: specContents,
		},
	}

	secondRevision, err := server.UpdateApiSpec(ctx, updateReq)
	if err != nil {
		t.Fatalf("Setup: UpdateApiSpec(%+v) returned error: %s", updateReq, err)
	}

	t.Run("one of multiple existing revisions", func(t *testing.T) {
		req := &rpc.DeleteApiSpecRevisionRequest{
			Name: fmt.Sprintf("projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec@%s", secondRevision.GetRevisionId()),
		}

		if _, err := server.DeleteApiSpecRevision(ctx, req); err != nil {
			t.Fatalf("DeleteApiSpecRevision(%+v) returned error: %s", req, err)
		}

		t.Run("GetApiSpec", func(t *testing.T) {
			req := &rpc.GetApiSpecRequest{
				Name: req.GetName(),
			}

			if _, err := server.GetApiSpec(ctx, req); status.Code(err) != codes.NotFound {
				t.Fatalf("GetApiSpec(%+v) returned status code %q, want %q: %v", req, status.Code(err), codes.NotFound, err)
			}
		})
	})
}

func TestListApiSpecRevisions(t *testing.T) {
	tests := []struct {
		admin     bool
		desc      string
		seed      []*rpc.ApiSpec
		req       *rpc.ListApiSpecRevisionsRequest
		want      *rpc.ListApiSpecRevisionsResponse
		wantToken bool
	}{
		{
			desc: "single spec all revs",
			seed: []*rpc.ApiSpec{
				{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				},
				{
					Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
					Contents: specContents,
				},
				{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/other-spec",
				},
			},
			req: &rpc.ListApiSpecRevisionsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec@-",
			},
			want: &rpc.ListApiSpecRevisionsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{
						Name:      "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
						Hash:      sha256hash(specContents),
						SizeBytes: int32(len(specContents)),
					},
					{
						Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
					},
				},
			},
		},
		{
			desc: "across multiple specs all revs",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/other-spec"},
			},
			req: &rpc.ListApiSpecRevisionsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/-@-",
			},
			want: &rpc.ListApiSpecRevisionsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/other-spec"},
				},
			},
		},
		{
			desc: "across multiple versions all revs",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/specs/my-spec"},
			},
			req: &rpc.ListApiSpecRevisionsRequest{
				Name: "projects/my-project/locations/global/apis/my-api/versions/-/specs/my-spec@-",
			},
			want: &rpc.ListApiSpecRevisionsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v2/specs/my-spec"},
				},
			},
		},
		{
			desc: "across multiple apis all revs",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/my-project/locations/global/apis/other-api/versions/v1/specs/my-spec"},
			},
			req: &rpc.ListApiSpecRevisionsRequest{
				Name: "projects/my-project/locations/global/apis/-/versions/v1/specs/my-spec@-",
			},
			want: &rpc.ListApiSpecRevisionsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/my-project/locations/global/apis/other-api/versions/v1/specs/my-spec"},
				},
			},
		},
		{
			admin: true,
			desc:  "across multiple projects all revs",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
			},
			req: &rpc.ListApiSpecRevisionsRequest{
				Name: "projects/-/locations/global/apis/my-api/versions/v1/specs/my-spec@-",
			},
			want: &rpc.ListApiSpecRevisionsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
					{Name: "projects/other-project/locations/global/apis/my-api/versions/v1/specs/my-spec"},
				},
			},
		},
		{
			desc: "custom page size, single spec all revs",
			seed: []*rpc.ApiSpec{
				{
					Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
				},
				{
					Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
					Contents: specContents,
				},
			},
			req: &rpc.ListApiSpecRevisionsRequest{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec@-",
				PageSize: 1,
			},
			want: &rpc.ListApiSpecRevisionsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{
						Name:      "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
						Hash:      sha256hash(specContents),
						SizeBytes: int32(len(specContents)),
					},
				},
			},
			wantToken: true,
		},
		{
			desc: "name filtering",
			seed: []*rpc.ApiSpec{
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec1"},
				{
					Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
					Description: "match",
				},
				{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec3"},
			},
			req: &rpc.ListApiSpecRevisionsRequest{
				Name:   "projects/my-project/locations/global/apis/my-api/versions/v1/specs/-@-",
				Filter: "description == 'match'",
			},
			want: &rpc.ListApiSpecRevisionsResponse{
				ApiSpecs: []*rpc.ApiSpec{
					{
						Name:        "projects/my-project/locations/global/apis/my-api/versions/v1/specs/spec2",
						Description: "match",
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
			if err := seeder.SeedSpecs(ctx, server, test.seed...); err != nil {
				t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
			}

			got, err := server.ListApiSpecRevisions(ctx, test.req)
			if err != nil {
				t.Fatalf("ListApiSpecRevisions(%+v) returned error: %s", test.req, err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ListApiSpecRevisionsResponse), "next_page_token"),
				protocmp.IgnoreFields(new(rpc.ApiSpec), "name", "revision_id", "create_time", "revision_create_time", "revision_update_time"),
			}

			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("ListApiSpecRevisions(%+v) returned unexpected diff (-want +got):\n%s", test.req, cmp.Diff(test.want, got, opts))
			}

			if test.wantToken && got.NextPageToken == "" {
				t.Errorf("ListApiSpecRevisions(%+v) returned empty next_page_token, expected non-empty next_page_token", test.req)
			} else if !test.wantToken && got.NextPageToken != "" {
				t.Errorf("ListApiSpecRevisions(%+v) returned non-empty next_page_token, expected empty next_page_token: %s", test.req, got.GetNextPageToken())
			}

			if len(got.ApiSpecs) != len(test.want.ApiSpecs) {
				t.Fatalf("ListApiSpecRevisions(%+v) returned unexpected number of revisions: got %d, want %d", test.req, len(got.ApiSpecs), len(test.want.ApiSpecs))
			}

			for i, got := range got.ApiSpecs {
				if want := test.want.ApiSpecs[i]; !strings.HasPrefix(got.GetName(), want.GetName()) {
					t.Errorf("ListApiSpecRevisions(%+v) returned unexpected revision: got %q, want %q", test.req, got.GetName(), want.GetName())
				}
			}
		})
	}
}

func TestListApiSpecRevisionsSequence(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	if err := seeder.SeedVersions(ctx, server, &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"}); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	createReq := &rpc.CreateApiSpecRequest{
		Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
		ApiSpecId: "my-spec",
		ApiSpec:   &rpc.ApiSpec{},
	}

	firstRevision, err := server.CreateApiSpec(ctx, createReq)
	if err != nil {
		t.Fatalf("Setup: CreateApiSpec(%+v) returned error: %s", createReq, err)
	}

	firstWant := &rpc.ApiSpec{
		Name:               fmt.Sprintf("%s@%s", firstRevision.GetName(), firstRevision.GetRevisionId()),
		CreateTime:         firstRevision.GetCreateTime(),
		RevisionCreateTime: firstRevision.GetRevisionCreateTime(),
		RevisionUpdateTime: firstRevision.GetRevisionUpdateTime(),
		RevisionId:         firstRevision.GetRevisionId(),
	}

	updateReq := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:      firstRevision.GetName(),
			Contents:  specContents,
			Hash:      sha256hash(specContents),
			SizeBytes: int32(len(specContents)),
		},
	}

	secondRevision, err := server.UpdateApiSpec(ctx, updateReq)
	if err != nil {
		t.Fatalf("Setup: UpdateApiSpec(%+v) returned error: %s", updateReq, err)
	}

	secondWant := &rpc.ApiSpec{
		Name:               fmt.Sprintf("%s@%s", secondRevision.GetName(), secondRevision.GetRevisionId()),
		Hash:               secondRevision.GetHash(),
		SizeBytes:          secondRevision.GetSizeBytes(),
		CreateTime:         secondRevision.GetCreateTime(),
		RevisionCreateTime: secondRevision.GetRevisionCreateTime(),
		RevisionUpdateTime: secondRevision.GetRevisionUpdateTime(),
		RevisionId:         secondRevision.GetRevisionId(),
	}

	opts := cmp.Options{
		protocmp.Transform(),
	}

	var nextToken string
	t.Run("first page", func(t *testing.T) {
		req := &rpc.ListApiSpecRevisionsRequest{
			Name:     firstRevision.GetName() + "@-",
			PageSize: 1,
		}

		got, err := server.ListApiSpecRevisions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecRevisions(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiSpecs()); count != 1 {
			t.Errorf("ListApiSpecRevisions(%+v) returned %d specs, expected exactly one", req, count)
		}

		// Check that the most recent revision is returned.
		want := []*rpc.ApiSpec{secondWant}
		if !cmp.Equal(want, got.GetApiSpecs(), opts) {
			t.Errorf("List sequence returned unexpected diff (-want +got):\n%s", cmp.Diff(want, got.GetApiSpecs(), opts))
		}

		if got.GetNextPageToken() == "" {
			t.Errorf("ListApiSpecRevisions(%+v) returned empty next_page_token, expected another page", req)
		}

		nextToken = got.GetNextPageToken()
	})

	if t.Failed() {
		t.Fatal("Cannot test final page after failure on first page")
	}

	t.Run("final page", func(t *testing.T) {
		req := &rpc.ListApiSpecRevisionsRequest{
			Name:      firstRevision.GetName() + "@-",
			PageToken: nextToken,
		}

		got, err := server.ListApiSpecRevisions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecRevisions(%+v) returned error: %s", req, err)
		}

		if count := len(got.GetApiSpecs()); count != 1 {
			t.Errorf("ListApiSpecRevisions(%+v) returned %d specs, expected exactly one", req, count)
		}

		// Check that the original revision is returned.
		want := []*rpc.ApiSpec{firstWant}
		if !cmp.Equal(want, got.GetApiSpecs(), opts) {
			t.Errorf("List sequence returned unexpected diff (-want +got):\n%s", cmp.Diff(want, got.GetApiSpecs(), opts))
		}

		if got.GetNextPageToken() != "" {
			t.Errorf("ListApiSpecRevisions(%+v) returned next_page_token, expected no next page", req)
		}
	})
}

func TestListApiSpecRevisionsLargeCollection(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	if err := seeder.SeedSpecs(ctx, server, &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/s"}); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	for i := 0; i <= 1001; i++ {
		contents := []byte(fmt.Sprintf(`{"openapi": "3.0.0", "info": {"title": "My API", "version": "%d"}, "paths": {}}`, i))
		updateReq := &rpc.UpdateApiSpecRequest{
			ApiSpec: &rpc.ApiSpec{
				Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/s",
				Contents: contents,
			},
		}

		_, err := server.UpdateApiSpec(ctx, updateReq)
		if err != nil {
			t.Fatalf("Setup: UpdateApiSpec(%+v) returned error: %s", updateReq, err)
		}
	}

	t.Run("max page size", func(t *testing.T) {
		req := &rpc.ListApiSpecRevisionsRequest{
			Name:     "projects/my-project/locations/global/apis/my-api/versions/v1/specs/s",
			PageSize: 1001,
		}

		got, err := server.ListApiSpecRevisions(ctx, req)
		if err != nil {
			t.Fatalf("ListApiSpecRevisions(%+v) returned error: %s", req, err)
		}

		if len(got.GetApiSpecs()) != 1000 {
			t.Errorf("GetApiSpecs(%+v) should have returned 1000 items, got: %+v", req, len(got.GetApiSpecs()))
		} else if got.GetNextPageToken() == "" {
			t.Errorf("GetApiSpecs(%+v) should return a next page token", req)
		}
	})
}

func TestUpdateApiSpecRevisions(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	if err := seeder.SeedVersions(ctx, server, &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/my-api/versions/v1"}); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	createReq := &rpc.CreateApiSpecRequest{
		Parent:    "projects/my-project/locations/global/apis/my-api/versions/v1",
		ApiSpecId: "my-spec",
		ApiSpec: &rpc.ApiSpec{
			Description: "Empty First Revision",
		},
	}

	created, err := server.CreateApiSpec(ctx, createReq)
	if err != nil {
		t.Fatalf("Setup: CreateApiSpec(%+v) returned error: %s", createReq, err)
	}

	opts := cmp.Options{
		protocmp.Transform(),
		protocmp.IgnoreFields(new(rpc.ApiSpec), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
	}

	t.Run("modify revision without content changes", func(t *testing.T) {
		req := &rpc.UpdateApiSpecRequest{
			ApiSpec: &rpc.ApiSpec{
				Name: created.GetName(),
			},
		}

		got, err := server.UpdateApiSpec(ctx, req)
		if err != nil {
			t.Fatalf("UpdateApiSpec(%+v) returned error: %s", req, err)
		}

		if got.GetRevisionId() != created.GetRevisionId() {
			t.Errorf("UpdateApiSpec(%+v) returned unexpected revision_id %q, expected no change (%q)", req, got.GetRevisionId(), created.GetRevisionId())
		}

		if ct, ut := got.GetRevisionCreateTime().AsTime(), got.GetRevisionUpdateTime().AsTime(); !ct.Before(ut) {
			t.Errorf("UpdateApiSpec(%+v) returned unexpected timestamps, expected revision_update_time %v > revision_create_time %v", req, ut, ct)
		}
	})

	t.Run("modify revision with content changes", func(t *testing.T) {
		req := &rpc.UpdateApiSpecRequest{
			ApiSpec: &rpc.ApiSpec{
				Name:     created.GetName(),
				Contents: specContents,
			},
		}
		want := proto.Clone(created).(*rpc.ApiSpec)
		want.SizeBytes = int32(len(req.ApiSpec.GetContents()))
		want.Hash = sha256hash(req.ApiSpec.GetContents())

		got, err := server.UpdateApiSpec(ctx, req)
		if err != nil {
			t.Fatalf("UpdateApiSpec(%+v) returned error: %s", req, err)
		}

		if !cmp.Equal(want, got, opts) {
			t.Errorf("UpdateApiSpec(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(want, got, opts))
		}

		if got.GetRevisionId() == created.GetRevisionId() {
			t.Errorf("UpdateApiSpec(%+v) returned unexpected revision_id %q, expected new revision", req, got.GetRevisionId())
		}

		if ct, ut := got.GetCreateTime().AsTime(), got.GetRevisionUpdateTime().AsTime(); !ct.Before(ut) {
			t.Errorf("UpdateApiSpec(%+v) returned unexpected timestamps, expected revision_update_time %v > create_time %v", req, ut, ct)
		}
	})

	t.Run("modify specific revision", func(t *testing.T) {
		req := &rpc.UpdateApiSpecRequest{
			ApiSpec: &rpc.ApiSpec{
				Name: fmt.Sprintf("%s@%s", created.GetName(), created.GetRevisionId()),
			},
		}

		if _, err := server.UpdateApiSpec(ctx, req); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("UpdateApiSpec(%+v) returned unexpected status code %q, want %q: %v", req, status.Code(err), codes.InvalidArgument, err)
		}
	})
}
