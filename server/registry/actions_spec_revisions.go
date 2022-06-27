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

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListApiSpecRevisions handles the corresponding API request.
func (s *RegistryServer) ListApiSpecRevisions(ctx context.Context, req *rpc.ListApiSpecRevisionsRequest) (*rpc.ListApiSpecRevisionsResponse, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	if req.GetPageSize() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid page_size %d: must not be negative", req.GetPageSize())
	} else if req.GetPageSize() > 1000 {
		req.PageSize = 1000
	} else if req.GetPageSize() == 0 {
		req.PageSize = 50
	}

	parent, err := names.ParseSpec(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	listing, err := db.ListSpecRevisions(ctx, parent, storage.PageOptions{
		Size:  req.GetPageSize(),
		Token: req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}

	response := &rpc.ListApiSpecRevisionsResponse{
		ApiSpecs:      make([]*rpc.ApiSpec, len(listing.Specs)),
		NextPageToken: listing.Token,
	}

	for i, spec := range listing.Specs {
		response.ApiSpecs[i], err = spec.BasicMessage(spec.RevisionName())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

// DeleteApiSpecRevision handles the corresponding API request.
func (s *RegistryServer) DeleteApiSpecRevision(ctx context.Context, req *rpc.DeleteApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	name, err := names.ParseSpecRevision(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	revision, err := db.GetSpecRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	// Parse the retrieved spec revision name, which has a non-tag revision ID.
	// This is necessary to ensure the actual revision is deleted.
	name, err = names.ParseSpecRevision(revision.RevisionName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := db.DeleteSpecRevision(ctx, name); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_DELETED, name.String())

	// return the latest revision of the current spec
	spec, err := s.getApiSpec(ctx, name.Spec())
	if err != nil {
		// This will fail if we just deleted the only revision of this spec.
		// TODO: prevent this.
		return nil, status.Error(codes.Internal, err.Error())
	}
	return spec, nil
}

// TagApiSpecRevision handles the corresponding API request.
func (s *RegistryServer) TagApiSpecRevision(ctx context.Context, req *rpc.TagApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	if req.GetTag() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tag %q, must not be empty", req.GetTag())
	} else if len(req.GetTag()) > 40 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tag %q, must be 40 characters or less", req.GetTag())
	}

	// Parse the requested spec revision name, which may include a tag name.
	name, err := names.ParseSpecRevision(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	revision, err := db.GetSpecRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	// Parse the retrieved spec revision name, which has a non-tag revision ID.
	// This is necessary to ensure the new tag is associated with a revision ID, not another tag.
	name, err = names.ParseSpecRevision(revision.RevisionName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tag := models.NewSpecRevisionTag(name, req.GetTag())
	if err := db.SaveSpecRevisionTag(ctx, tag); err != nil {
		return nil, err
	}

	message, err := revision.BasicMessage(tag.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_UPDATED, name.String())
	return message, nil
}

// RollbackApiSpec handles the corresponding API request.
func (s *RegistryServer) RollbackApiSpec(ctx context.Context, req *rpc.RollbackApiSpecRequest) (*rpc.ApiSpec, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	if req.GetRevisionId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid revision ID %q, must not be empty", req.GetRevisionId())
	}

	parent, err := names.ParseSpec(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get the target spec revision to use as a base for the new rollback revision.
	name := parent.Revision(req.GetRevisionId())
	target, err := db.GetSpecRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	// Save a new rollback revision based on the target revision.
	rollback := target.NewRevision()
	if err := db.SaveSpecRevision(ctx, rollback); err != nil {
		return nil, err
	}

	blob, err := db.GetSpecRevisionContents(ctx, name)
	if err != nil {
		return nil, err
	}

	// Save a new copy of the target revision blob for the rollback revision.
	blob.RevisionID = name.RevisionID
	if err := db.SaveSpecRevisionContents(ctx, rollback, blob.Contents); err != nil {
		return nil, err
	}

	message, err := rollback.BasicMessage(rollback.RevisionName())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_CREATED, rollback.RevisionName())
	return message, nil
}
