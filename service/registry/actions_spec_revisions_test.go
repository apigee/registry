// Copyright 2020 Google LLC. All Rights Reserved.
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

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/service/registry/internal/test/seeder"
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
		protocmp.IgnoreFields(revision, "name", "revision_tags"),
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
			protocmp.IgnoreFields(revision, "name", "revision_tags"),
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

func TestDeleteApiSpecRevision(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)
	if err := seeder.SeedSpecs(ctx, server, &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec"}); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	t.Run("only remaining revision", func(t *testing.T) {
		t.Skip("not yet supported")

		req := &rpc.DeleteApiSpecRevisionRequest{
			Name: "projects/my-project/locations/global/apis/my-api/versions/v1/specs/my-spec",
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
			Name:     firstRevision.GetName(),
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
			Name:      firstRevision.GetName(),
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
			t.Errorf("TagApiSpecRevision(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(want, got, opts))
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
