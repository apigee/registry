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

// ListApiDeploymentRevisions handles the corresponding API request.
func (s *RegistryServer) ListApiDeploymentRevisions(ctx context.Context, req *rpc.ListApiDeploymentRevisionsRequest) (*rpc.ListApiDeploymentRevisionsResponse, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	if req.GetPageSize() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid page_size %d: must not be negative", req.GetPageSize())
	} else if req.GetPageSize() > 1000 {
		req.PageSize = 1000
	} else if req.GetPageSize() == 0 {
		req.PageSize = 50
	}

	parent, err := names.ParseDeployment(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := db.ListDeploymentRevisions(ctx, parent, storage.PageOptions{
		Size:  req.GetPageSize(),
		Token: req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteApiDeploymentRevision handles the corresponding API request.
func (s *RegistryServer) DeleteApiDeploymentRevision(ctx context.Context, req *rpc.DeleteApiDeploymentRevisionRequest) (*rpc.ApiDeployment, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	name, err := names.ParseDeploymentRevision(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	revision, err := db.GetDeploymentRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	// Parse the retrieved deployment revision name, which has a non-tag revision ID.
	// This is necessary to ensure the actual revision is deleted.
	name, err = names.ParseDeploymentRevision(revision.RevisionName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := db.DeleteDeploymentRevision(ctx, name); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_DELETED, name.String())

	// return the latest revision of the current deployment
	deployment, err := s.getApiDeployment(ctx, name.Deployment())
	if err != nil {
		// This will fail if we just deleted the only revision of this deployment.
		// TODO: prevent this.
		return nil, status.Error(codes.Internal, err.Error())
	}
	return deployment, nil
}

// TagApiDeploymentRevision handles the corresponding API request.
func (s *RegistryServer) TagApiDeploymentRevision(ctx context.Context, req *rpc.TagApiDeploymentRevisionRequest) (*rpc.ApiDeployment, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	if req.GetTag() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tag %q, must not be empty", req.GetTag())
	} else if len(req.GetTag()) > 40 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tag %q, must be 40 characters or less", req.GetTag())
	}

	// Parse the requested deployment revision name, which may include a tag name.
	name, err := names.ParseDeploymentRevision(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	revision, err := db.GetDeploymentRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	// Parse the retrieved deployment revision name, which has a non-tag revision ID.
	// This is necessary to ensure the new tag is associated with a revision ID, not another tag.
	name, err = names.ParseDeploymentRevision(revision.RevisionName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tag := models.NewDeploymentRevisionTag(name, req.GetTag())
	if err := db.SaveDeploymentRevisionTag(ctx, tag); err != nil {
		return nil, err
	}

	tags, err := deploymentRevisionTags(ctx, db, name)
	if err != nil {
		return nil, err
	}

	message, err := revision.BasicMessage(tag.String(), tags)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_UPDATED, name.String())
	return message, nil
}

// RollbackApiDeployment handles the corresponding API request.
func (s *RegistryServer) RollbackApiDeployment(ctx context.Context, req *rpc.RollbackApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	if req.GetRevisionId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid revision ID %q, must not be empty", req.GetRevisionId())
	}

	parent, err := names.ParseDeployment(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get the target deployment revision to use as a base for the new rollback revision.
	name := parent.Revision(req.GetRevisionId())
	target, err := db.GetDeploymentRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	// Save a new rollback revision based on the target revision.
	rollback := target.NewRevision()
	if err := db.SaveDeploymentRevision(ctx, rollback); err != nil {
		return nil, err
	}

	message, err := rollback.BasicMessage(rollback.RevisionName(), []string{})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_CREATED, rollback.RevisionName())
	return message, nil
}
