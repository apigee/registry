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

// CreateProject handles the corresponding API request.
func (s *RegistryServer) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	name := names.Project{}
	if req.GetProjectId() != "" {
		name.ProjectID = req.GetProjectId()
	} else {
		name.ProjectID = names.GenerateID()
	}

	if _, err := db.GetProject(ctx, name); err == nil {
		return nil, alreadyExistsError(fmt.Errorf("project %q already exists", name))
	} else if !isNotFound(err) {
		return nil, err
	}

	if err := name.Validate(); err != nil {
		return nil, invalidArgumentError(err)
	}

	project := models.NewProject(name, req.GetProject())
	if err := db.SaveProject(ctx, project); err != nil {
		return nil, err
	}

	message, err := project.Message()
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_CREATED, name.String())
	return message, nil
}

// DeleteProject handles the corresponding API request.
func (s *RegistryServer) DeleteProject(ctx context.Context, req *rpc.DeleteProjectRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	name, err := names.ParseProject(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Deletion should only succeed on projects that currently exist.
	if _, err := db.GetProject(ctx, name); err != nil {
		return nil, err
	}

	if err := db.DeleteProject(ctx, name); err != nil {
		return nil, err
	}

	s.notify(rpc.Notification_DELETED, name.String())
	return &empty.Empty{}, nil
}

// GetProject handles the corresponding API request.
func (s *RegistryServer) GetProject(ctx context.Context, req *rpc.GetProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	name, err := names.ParseProject(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	project, err := db.GetProject(ctx, name)
	if err != nil {
		return nil, err
	}

	message, err := project.Message()
	if err != nil {
		return nil, internalError(err)
	}

	return message, nil
}

// ListProjects handles the corresponding API request.
func (s *RegistryServer) ListProjects(ctx context.Context, req *rpc.ListProjectsRequest) (*rpc.ListProjectsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetPageSize() < 0 {
		return nil, invalidArgumentError(fmt.Errorf("invalid page_size %q: must not be negative", req.GetPageSize()))
	} else if req.GetPageSize() > 1000 {
		req.PageSize = 1000
	}

	listing, err := db.ListProjects(ctx, dao.PageOptions{
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
		response.Projects[i], err = project.Message()
		if err != nil {
			return nil, internalError(err)
		}
	}

	return response, nil
}

// UpdateProject handles the corresponding API request.
func (s *RegistryServer) UpdateProject(ctx context.Context, req *rpc.UpdateProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetProject() == nil {
		return nil, invalidArgumentError(fmt.Errorf("invalid project %+v: body must be provided", req.GetProject()))
	}

	name, err := names.ParseProject(req.GetProject().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	project, err := db.GetProject(ctx, name)
	if err != nil {
		return nil, err
	}

	project.Update(req.GetProject(), req.GetUpdateMask())
	if err := db.SaveProject(ctx, project); err != nil {
		return nil, err
	}

	message, err := project.Message()
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_UPDATED, name.String())
	return message, nil
}
