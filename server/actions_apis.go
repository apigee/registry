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
	"strings"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateApi handles the corresponding API request.
func (s *RegistryServer) CreateApi(ctx context.Context, request *rpc.CreateApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	api, err := models.NewApiFromParentAndApiID(request.GetParent(), request.GetApiId())
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
	err = api.Update(request.GetApi(), nil)
	api.CreateTime = api.UpdateTime
	k, err = client.Put(ctx, k, api)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_CREATED, api.ResourceName())
	return api.Message()
}

// DeleteApi handles the corresponding API request.
func (s *RegistryServer) DeleteApi(ctx context.Context, request *rpc.DeleteApiRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	// Validate name and create dummy api (we just need the ID fields).
	api, err := models.NewApiFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete children first and then delete the api.
	client.DeleteChildrenOfApi(ctx, api)
	k := client.NewKey(models.ApiEntityName, request.GetName())
	err = client.Delete(ctx, k)
	s.notify(rpc.Notification_DELETED, request.GetName())
	return &empty.Empty{}, internalError(err)
}

// GetApi handles the corresponding API request.
func (s *RegistryServer) GetApi(ctx context.Context, request *rpc.GetApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	api, err := models.NewApiFromResourceName(request.GetName())
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
	return api.Message()
}

// ListApis handles the corresponding API request.
func (s *RegistryServer) ListApis(ctx context.Context, req *rpc.ListApisRequest) (*rpc.ListApisResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	q := client.NewQuery(models.ApiEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := names.ParseParentProject(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if m[1] != "-" {
		q = q.Require("ProjectID", m[1])
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"project_id", filterArgTypeString},
			{"api_id", filterArgTypeString},
			{"display_name", filterArgTypeString},
			{"description", filterArgTypeString},
			{"create_time", filterArgTypeTimestamp},
			{"update_time", filterArgTypeTimestamp},
			{"availability", filterArgTypeString},
			{"recommended_version", filterArgTypeString},
			{"owner", filterArgTypeString},
			{"labels", filterArgTypeStringArray},
		})
	if err != nil {
		return nil, internalError(err)
	}
	// If the filter contains the "in labels" string,
	// include labels associated with each item.
	hasLabels := strings.Contains(req.GetFilter(), "in labels")
	var apiMessages []*rpc.Api
	var api models.Api
	it := client.Run(ctx, q)
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&api); err == nil; _, err = it.Next(&api) {
		labels := make([]string, 0)
		if hasLabels {
			// Proof-of-concept only. This is extremely slow!
			q2 := client.NewQuery(models.LabelEntityName)
			q2 = q2.Require("ProjectID", api.ProjectID)
			q2 = q2.Require("ApiID", api.ApiID)
			var label models.Label
			it2 := client.Run(ctx, q2)
			for _, err = it2.Next(&label); err == nil; _, err = it.Next(&label) {
				labels = append(labels, label.LabelID)
			}
		}
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id":          api.ProjectID,
				"api_id":              api.ApiID,
				"display_name":        api.DisplayName,
				"description":         api.Description,
				"create_time":         api.CreateTime,
				"update_time":         api.UpdateTime,
				"availability":        api.Availability,
				"recommended_version": api.RecommendedVersion,
				"owner":               api.Owner,
				"labels":              labels,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		apiMessage, _ := api.Message()
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
func (s *RegistryServer) UpdateApi(ctx context.Context, request *rpc.UpdateApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	api, err := models.NewApiFromResourceName(request.GetApi().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ApiEntityName, api.ResourceName())
	err = client.Get(ctx, k, api)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = api.Update(request.GetApi(), request.GetUpdateMask())
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, api)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_UPDATED, api.ResourceName())
	return api.Message()
}
