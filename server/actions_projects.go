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
	"github.com/apigee/registry/server/storage"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
)

// CreateProject handles the corresponding API request.
func (s *RegistryServer) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	name := names.Project{}
	if req.GetProjectId() != "" {
		name.ID = req.GetProjectId()
	} else {
		name.ID = names.GenerateID()
	}

	if _, err := getProject(ctx, client, name); err == nil {
		return nil, alreadyExistsError(fmt.Errorf("project %q already exists", name))
	} else if !isNotFound(err) {
		return nil, err
	}

	if err := name.Validate(); err != nil {
		return nil, invalidArgumentError(err)
	}

	project := models.NewProject(name, req.GetProject())
	if err := saveProject(ctx, client, project); err != nil {
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

	name, err := names.ParseProject(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Deletion should only succeed on projects that currently exist.
	if _, err := getProject(ctx, client, name); err != nil {
		return nil, err
	}

	if err := deleteProject(ctx, client, name); err != nil {
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

	name, err := names.ParseProject(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	project, err := getProject(ctx, client, name)
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

	if req.GetPageSize() < 0 {
		return nil, invalidArgumentError(fmt.Errorf("invalid page_size %q: must not be negative", req.GetPageSize()))
	}

	q := client.NewQuery(storage.ProjectEntityName)
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
				"name":         project.Name(),
				"project_id":   project.ProjectID,
				"display_name": project.DisplayName,
				"description":  project.Description,
				"create_time":  project.CreateTime,
				"update_time":  project.UpdateTime,
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
func (s *RegistryServer) UpdateProject(ctx context.Context, req *rpc.UpdateProjectRequest) (*rpc.Project, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	name, err := names.ParseProject(req.GetProject().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	project, err := getProject(ctx, client, name)
	if err != nil {
		return nil, err
	}

	project.Update(req.GetProject(), req.GetUpdateMask())
	if err := saveProject(ctx, client, project); err != nil {
		return nil, err
	}

	message, err := project.Message()
	if err != nil {
		return nil, internalError(err)
	}

	return message, nil
}

func saveProject(ctx context.Context, client storage.Client, project *models.Project) error {
	k := client.NewKey(storage.ProjectEntityName, project.Name())
	if _, err := client.Put(ctx, k, project); err != nil {
		return internalError(err)
	}

	return nil
}

func getProject(ctx context.Context, client storage.Client, name names.Project) (*models.Project, error) {
	project := &models.Project{
		ProjectID: name.ID,
	}

	k := client.NewKey(storage.ProjectEntityName, name.String())
	if err := client.Get(ctx, k, project); client.IsNotFound(err) {
		return nil, notFoundError(fmt.Errorf("project %q not found", name))
	} else if err != nil {
		return nil, internalError(err)
	}

	return project, nil
}

func deleteProject(ctx context.Context, client storage.Client, name names.Project) error {
	if err := client.DeleteChildrenOfProject(ctx, name); err != nil {
		return internalError(err)
	}

	k := client.NewKey(storage.ProjectEntityName, name.String())
	if err := client.Delete(ctx, k); err != nil {
		return internalError(err)
	}

	return nil
}
