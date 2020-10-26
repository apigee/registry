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

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/models"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateProject handles the corresponding API request.
func (s *RegistryServer) CreateProject(ctx context.Context, request *rpc.CreateProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	project, err := models.NewProjectFromProjectID(request.GetProjectId())
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
	err = project.Update(request.GetProject(), nil)
	project.CreateTime = project.UpdateTime
	k, err = client.Put(ctx, k, project)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_CREATED, project.ResourceName())
	return project.Message()
}

// DeleteProject handles the corresponding API request.
func (s *RegistryServer) DeleteProject(ctx context.Context, request *rpc.DeleteProjectRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	// Validate name and create dummy project (we just need the ID fields).
	project, err := models.NewProjectFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete children first and then delete the project.
	err = client.DeleteChildrenOfProject(ctx, project)
	if err != nil {
		return &empty.Empty{}, internalError(err)
	}
	k := client.NewKey(models.ProjectEntityName, request.GetName())
	err = client.Delete(ctx, k)
	s.notify(rpc.Notification_DELETED, request.GetName())
	return &empty.Empty{}, internalError(err)
}

// GetProject handles the corresponding API request.
func (s *RegistryServer) GetProject(ctx context.Context, request *rpc.GetProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	project, err := models.NewProjectFromResourceName(request.GetName())
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
	q := client.NewQuery(models.ProjectEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"project_id", filterArgTypeString},
			{"display_name", filterArgTypeString},
			{"description", filterArgTypeString},
		})
	if err != nil {
		return nil, internalError(err)
	}
	var projectMessages []*rpc.Project
	var project models.Project
	it := client.Run(ctx, q)
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&project); err == nil; _, err = it.Next(&project) {
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id":   project.ProjectID,
				"display_name": project.DisplayName,
				"description":  project.Description,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
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
func (s *RegistryServer) UpdateProject(ctx context.Context, request *rpc.UpdateProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	project, err := models.NewProjectFromResourceName(request.GetProject().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ProjectEntityName, project.ResourceName())
	err = client.Get(ctx, k, project)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = project.Update(request.GetProject(), request.GetUpdateMask())
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, project)
	if err != nil {
		return nil, internalError(err)
	}
	return project.Message()
}
