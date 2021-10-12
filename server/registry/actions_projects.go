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
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateProject handles the corresponding API request.
func (s *RegistryServer) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	if req.GetProject() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project %+v: body must be provided", req.GetProject())
	}

	name := names.Project{ProjectID: req.GetProjectId()}
	if _, err := db.GetProject(ctx, name); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "project %q already exists", name)
	} else if !isNotFound(err) {
		return nil, err
	}

	if err := name.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	project := models.NewProject(name, req.GetProject())
	if err := db.SaveProject(ctx, project); err != nil {
		return nil, err
	}

	s.notify(ctx, rpc.Notification_CREATED, name.String())
	return project.Message(), nil
}

// DeleteProject handles the corresponding API request.
func (s *RegistryServer) DeleteProject(ctx context.Context, req *rpc.DeleteProjectRequest) (*emptypb.Empty, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	name, err := names.ParseProject(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Deletion should only succeed on projects that currently exist.
	if _, err := db.GetProject(ctx, name); err != nil {
		return nil, err
	}

	if err := db.DeleteProject(ctx, name); err != nil {
		return nil, err
	}

	s.notify(ctx, rpc.Notification_DELETED, name.String())
	return &emptypb.Empty{}, nil
}

// GetProject handles the corresponding API request.
func (s *RegistryServer) GetProject(ctx context.Context, req *rpc.GetProjectRequest) (*rpc.Project, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	name, err := names.ParseProject(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	project, err := db.GetProject(ctx, name)
	if err != nil {
		return nil, err
	}

	return project.Message(), nil
}

// ListProjects handles the corresponding API request.
func (s *RegistryServer) ListProjects(ctx context.Context, req *rpc.ListProjectsRequest) (*rpc.ListProjectsResponse, error) {
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

	listing, err := db.ListProjects(ctx, storage.PageOptions{
		Size:   req.GetPageSize(),
		Filter: req.GetFilter(),
		Token:  req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}

	response := &rpc.ListProjectsResponse{
		Projects:      make([]*rpc.Project, len(listing.Projects)),
		NextPageToken: listing.Token,
	}

	for i, project := range listing.Projects {
		response.Projects[i] = project.Message()
	}

	return response, nil
}

// UpdateProject handles the corresponding API request.
func (s *RegistryServer) UpdateProject(ctx context.Context, req *rpc.UpdateProjectRequest) (*rpc.Project, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	if req.GetProject() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project %+v: body must be provided", req.GetProject())
	} else if err := models.ValidateMask(req.GetProject(), req.GetUpdateMask()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid update_mask %v: %s", req.GetUpdateMask(), err)
	}

	name, err := names.ParseProject(req.GetProject().GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	project, err := db.GetProject(ctx, name)
	if err != nil {
		return nil, err
	}

	project.Update(req.GetProject(), models.ExpandMask(req.GetProject(), req.GetUpdateMask()))
	if err := db.SaveProject(ctx, project); err != nil {
		return nil, err
	}

	s.notify(ctx, rpc.Notification_UPDATED, name.String())
	return project.Message(), nil
}
