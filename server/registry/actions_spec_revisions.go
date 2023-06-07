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

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage"
	"github.com/apigee/registry/server/registry/internal/storage/models"
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

	parent, err := names.ParseSpecRevision(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	listing, err := db.ListSpecRevisions(ctx, parent, storage.PageOptions{
		Size:   req.GetPageSize(),
		Token:  req.GetPageToken(),
		Filter: req.GetFilter(),
		// order_by is disallowed by AIP-162: https://google.aip.dev/162
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
	// The spec revision name must be valid.
	name, err := names.ParseSpecRevision(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var response *rpc.ApiSpec
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		if err := db.DeleteSpecRevision(ctx, name); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// return the latest revision of the current spec
	response, err = s.getApiSpec(ctx, name.Spec())
	if err != nil {
		// The get will fail if we are deleting the only revision.
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.notify(ctx, rpc.Notification_DELETED, name.String())
	return response, nil
}

// TagApiSpecRevision handles the corresponding API request.
func (s *RegistryServer) TagApiSpecRevision(ctx context.Context, req *rpc.TagApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	// The tag must be nonempty.
	if req.GetTag() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tag %q, must not be empty", req.GetTag())
	}
	// The tag length must be valid.
	if len(req.GetTag()) > 40 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tag %q, must be 40 characters or less", req.GetTag())
	}
	// The tag must match the required format.
	if err := names.ValidateRevisionTag(req.GetTag()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}
	// The requested spec revision name must be valid. It may include a tag name.
	name, err := names.ParseSpecRevision(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var response *rpc.ApiSpec
	var revisionName string
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		// The revision to be tagged must exist.
		revision, err := db.GetSpecRevision(ctx, name)
		if err != nil {
			return err
		}
		// Parse the retrieved spec revision name, which has a non-tag revision ID.
		// This is necessary to ensure the new tag is associated with a revision ID, not another tag.
		name, err = names.ParseSpecRevision(revision.RevisionName())
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		tag := models.NewSpecRevisionTag(name, req.GetTag())
		if err := db.SaveSpecRevisionTag(ctx, tag); err != nil {
			return err
		}
		response, err = revision.BasicMessage(tag.String())
		if err != nil {
			return err
		}
		revisionName = name.String()
		return nil
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_UPDATED, revisionName)
	return response, nil
}

// RollbackApiSpec handles the corresponding API request.
func (s *RegistryServer) RollbackApiSpec(ctx context.Context, req *rpc.RollbackApiSpecRequest) (*rpc.ApiSpec, error) {
	// Revision ID must be nonempty.
	if req.GetRevisionId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid revision ID %q, must not be empty", req.GetRevisionId())
	}
	// Spec name must be valid.
	parent, err := names.ParseSpec(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var response *rpc.ApiSpec
	var revisionName string
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		// Get the target spec revision to use as a base for the new rollback revision.
		name := parent.Revision(req.GetRevisionId())
		target, err := db.GetSpecRevision(ctx, name)
		if err != nil {
			return err
		}
		// Save a new rollback revision based on the target revision.
		rollback := target.NewRevision()
		if err := db.SaveSpecRevision(ctx, rollback); err != nil {
			return err
		}
		blob, err := db.GetSpecRevisionContents(ctx, name)
		if err != nil {
			return err
		}
		// Save a new copy of the target revision blob for the rollback revision.
		blob.RevisionID = name.RevisionID
		if err := db.SaveSpecRevisionContents(ctx, rollback, blob.Contents); err != nil {
			return err
		}
		response, err = rollback.BasicMessage(rollback.RevisionName())
		if err != nil {
			return err
		}
		revisionName = rollback.RevisionName()
		return nil
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_CREATED, revisionName)
	return response, nil
}
