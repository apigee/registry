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
	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateProject handles the corresponding API request.
func (s *RegistryServer) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetProjectId() == "" {
		req.ProjectId = names.GenerateID()
	}

	project, err := models.NewProjectFromProjectID(req.GetProjectId())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ProjectEntityName, project.ResourceName())
	// fail if project already exists
	existingProject := &models.Project{}
	err = client.Get(ctx, k, existingProject)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, project.ResourceName()+" already exists")
	}
	err = project.Update(req.GetProject(), nil)
	project.CreateTime = project.UpdateTime
	k, err = client.Put(ctx, k, project)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_CREATED, project.ResourceName())
	return project.Message()
}

// DeleteProject handles the corresponding API request.
func (s *RegistryServer) DeleteProject(ctx context.Context, req *rpc.DeleteProjectRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	// Validate name and create dummy project (we just need the ID field).
	project, err := models.NewProjectFromResourceName(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	k := client.NewKey(models.ProjectEntityName, project.ResourceName())
	if err := client.Get(ctx, k, &models.Project{}); client.IsNotFound(err) {
		return nil, notFoundError(err)
	} else if err != nil {
		return nil, internalError(err)
	}

	if err := client.DeleteChildrenOfProject(ctx, project); err != nil {
		return nil, internalError(err)
	}

	if err := client.Delete(ctx, k); err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_DELETED, req.GetName())
	return &empty.Empty{}, nil
}

// GetProject handles the corresponding API request.
func (s *RegistryServer) GetProject(ctx context.Context, req *rpc.GetProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	project, err := models.NewProjectFromResourceName(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ProjectEntityName, project.ResourceName())
	err = client.Get(ctx, k, project)
	if client.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return project.Message()
}

// ListProjects handles the corresponding API request.
func (s *RegistryServer) ListProjects(ctx context.Context, req *rpc.ListProjectsRequest) (*rpc.ListProjectsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetPageSize() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid page_size: must not be negative")
	}

	q := client.NewQuery(models.ProjectEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"name", filterArgTypeString},
			{"project_id", filterArgTypeString},
			{"display_name", filterArgTypeString},
			{"description", filterArgTypeString},
			{"create_time", filterArgTypeTimestamp},
			{"update_time", filterArgTypeTimestamp},
		})
	if err != nil {
		return nil, err
	}
	var projectMessages []*rpc.Project
	var project models.Project
	it := client.Run(ctx, q)
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&project); err == nil; _, err = it.Next(&project) {
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"name":         project.ResourceName(),
				"project_id":   project.ProjectID,
				"display_name": project.DisplayName,
				"description":  project.Description,
				"create_time":  project.CreateTime,
				"update_time":  project.UpdateTime,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if v, ok := out.Value().(bool); !ok {
				return nil, invalidArgumentError(fmt.Errorf("expression does not evaluate to a boolean (instead yielding %T)", out.Value()))
			} else if !v {
				continue
			}
		}
		projectMessage, _ := project.Message()
		projectMessages = append(projectMessages, projectMessage)
		if len(projectMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListProjectsResponse{
		Projects: projectMessages,
	}

	responses.NextPageToken, err = it.GetCursor(len(projectMessages))
	if err != nil {
		return nil, internalError(err)
	}

	return responses, nil
}

// UpdateProject handles the corresponding API request.
func (s *RegistryServer) UpdateProject(ctx context.Context, req *rpc.UpdateProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	project, err := models.NewProjectFromResourceName(req.GetProject().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ProjectEntityName, project.ResourceName())
	err = client.Get(ctx, k, project)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = project.Update(req.GetProject(), req.GetUpdateMask())
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, project)
	if err != nil {
		return nil, internalError(err)
	}
	return project.Message()
}
