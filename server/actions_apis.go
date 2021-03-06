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

// CreateApi handles the corresponding API request.
func (s *RegistryServer) CreateApi(ctx context.Context, req *rpc.CreateApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetApiId() == "" {
		req.ApiId = names.GenerateID()
	}

	api, err := models.NewApiFromParentAndApiID(req.GetParent(), req.GetApiId())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ApiEntityName, api.ResourceName())
	// fail if api already exists
	existingApi := &models.Api{}
	err = client.Get(ctx, k, existingApi)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, api.ResourceName()+" already exists")
	}
	err = api.Update(req.GetApi(), nil)
	api.CreateTime = api.UpdateTime
	k, err = client.Put(ctx, k, api)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_CREATED, api.ResourceName())
	return api.Message(rpc.View_BASIC)
}

// DeleteApi handles the corresponding API request.
func (s *RegistryServer) DeleteApi(ctx context.Context, req *rpc.DeleteApiRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	// Validate name and create dummy api (we just need the ID fields).
	api, err := models.NewApiFromResourceName(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	k := client.NewKey(models.ApiEntityName, req.GetName())
	if err := client.Get(ctx, k, &models.Api{}); client.IsNotFound(err) {
		return nil, notFoundError(err)
	} else if err != nil {
		return nil, internalError(err)
	}

	if err := client.DeleteChildrenOfApi(ctx, api); err != nil {
		return nil, internalError(err)
	}

	if err := client.Delete(ctx, k); err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_DELETED, req.GetName())
	return &empty.Empty{}, nil
}

// GetApi handles the corresponding API request.
func (s *RegistryServer) GetApi(ctx context.Context, req *rpc.GetApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	api, err := models.NewApiFromResourceName(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ApiEntityName, api.ResourceName())
	err = client.Get(ctx, k, api)
	if client.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return api.Message(req.GetView())
}

// ListApis handles the corresponding API request.
func (s *RegistryServer) ListApis(ctx context.Context, req *rpc.ListApisRequest) (*rpc.ListApisResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetPageSize() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid page_size: must not be negative")
	}

	q := client.NewQuery(models.ApiEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	m, err := names.ParseProject(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if m[1] != "-" {
		q = q.Require("ProjectID", m[1])
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"name", filterArgTypeString},
			{"project_id", filterArgTypeString},
			{"api_id", filterArgTypeString},
			{"display_name", filterArgTypeString},
			{"description", filterArgTypeString},
			{"create_time", filterArgTypeTimestamp},
			{"update_time", filterArgTypeTimestamp},
			{"availability", filterArgTypeString},
			{"recommended_version", filterArgTypeString},
			{"labels", filterArgTypeStringMap},
		})
	if err != nil {
		return nil, internalError(err)
	}
	var apiMessages []*rpc.Api
	var api models.Api
	it := client.Run(ctx, q)
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&api); err == nil; _, err = it.Next(&api) {
		if prg != nil {
			filterInputs := map[string]interface{}{
				"name":                api.ResourceName(),
				"project_id":          api.ProjectID,
				"api_id":              api.ApiID,
				"display_name":        api.DisplayName,
				"description":         api.Description,
				"create_time":         api.CreateTime,
				"update_time":         api.UpdateTime,
				"availability":        api.Availability,
				"recommended_version": api.RecommendedVersion,
			}
			filterInputs["labels"], err = api.LabelsMap()
			if err != nil {
				return nil, internalError(err)
			}
			out, _, err := prg.Eval(filterInputs)
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if v, ok := out.Value().(bool); !ok {
				return nil, invalidArgumentError(fmt.Errorf("expression does not evaluate to a boolean (instead yielding %T)", out.Value()))
			} else if !v {
				continue
			}
		}
		apiMessage, _ := api.Message(req.GetView())
		apiMessages = append(apiMessages, apiMessage)
		if len(apiMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListApisResponse{
		Apis: apiMessages,
	}

	responses.NextPageToken, err = it.GetCursor(len(apiMessages))
	if err != nil {
		return nil, internalError(err)
	}

	return responses, nil
}

// UpdateApi handles the corresponding API request.
func (s *RegistryServer) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	api, err := models.NewApiFromResourceName(req.GetApi().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ApiEntityName, api.ResourceName())
	err = client.Get(ctx, k, api)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = api.Update(req.GetApi(), req.GetUpdateMask())
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, api)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_UPDATED, api.ResourceName())
	return api.Message(rpc.View_BASIC)
}
