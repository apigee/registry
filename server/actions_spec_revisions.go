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

package server

import (
	"context"
	"fmt"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/dao"
	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes/empty"
)

// ListApiSpecRevisions handles the corresponding API request.
func (s *RegistryServer) ListApiSpecRevisions(ctx context.Context, req *rpc.ListApiSpecRevisionsRequest) (*rpc.ListApiSpecRevisionsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetPageSize() < 0 {
		return nil, invalidArgumentError(fmt.Errorf("invalid page_size %d: must not be negative", req.GetPageSize()))
	} else if req.GetPageSize() > 1000 {
		req.PageSize = 1000
	} else if req.GetPageSize() == 0 {
		req.PageSize = 50
	}

	parent, err := names.ParseSpec(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	listing, err := db.ListSpecRevisions(ctx, parent, dao.PageOptions{
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
			return nil, internalError(err)
		}
	}

	return response, nil
}

// DeleteApiSpecRevision handles the corresponding API request.
func (s *RegistryServer) DeleteApiSpecRevision(ctx context.Context, req *rpc.DeleteApiSpecRevisionRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	name, err := names.ParseSpecRevision(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	revision, err := db.GetSpecRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	// Parse the retrieved spec revision name, which has a non-tag revision ID.
	// This is necessary to ensure the actual revision is deleted.
	name, err = names.ParseSpecRevision(revision.RevisionName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	if err := db.DeleteSpecRevision(ctx, name); err != nil {
		return nil, internalError(err)
	}

	s.notify(ctx, rpc.Notification_DELETED, name.String())
	return &empty.Empty{}, nil
}

// TagApiSpecRevision handles the corresponding API request.
func (s *RegistryServer) TagApiSpecRevision(ctx context.Context, req *rpc.TagApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetTag() == "" {
		return nil, invalidArgumentError(fmt.Errorf("invalid tag %q, must not be empty", req.GetTag()))
	} else if len(req.GetTag()) > 40 {
		return nil, invalidArgumentError(fmt.Errorf("invalid tag %q, must be 40 characters or less", req.GetTag()))
	}

	// Parse the requested spec revision name, which may include a tag name.
	name, err := names.ParseSpecRevision(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	revision, err := db.GetSpecRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	// Parse the retrieved spec revision name, which has a non-tag revision ID.
	// This is necessary to ensure the new tag is associated with a revision ID, not another tag.
	name, err = names.ParseSpecRevision(revision.RevisionName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	tag := models.NewSpecRevisionTag(name, req.GetTag())
	if err := db.SaveSpecRevisionTag(ctx, tag); err != nil {
		return nil, err
	}

	message, err := revision.BasicMessage(tag.String())
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(ctx, rpc.Notification_UPDATED, name.String())
	return message, nil
}

// RollbackApiSpec handles the corresponding API request.
func (s *RegistryServer) RollbackApiSpec(ctx context.Context, req *rpc.RollbackApiSpecRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetRevisionId() == "" {
		return nil, invalidArgumentError(fmt.Errorf("invalid revision ID %q, must not be empty", req.GetRevisionId()))
	}

	parent, err := names.ParseSpec(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
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
		return nil, internalError(err)
	}

	s.notify(ctx, rpc.Notification_CREATED, rollback.RevisionName())
	return message, nil
}
