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

// CreateApi handles the corresponding API request.
func (s *RegistryServer) CreateApi(ctx context.Context, req *rpc.CreateApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	parent, err := names.ParseProject(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Creation should only succeed when the parent exists.
	if _, err := getProject(ctx, client, parent); err != nil {
		return nil, err
	}

	name := parent.Api(req.GetApiId())
	if name.ApiID == "" {
		name.ApiID = names.GenerateID()
	}

	if _, err := getApi(ctx, client, name); err == nil {
		return nil, alreadyExistsError(fmt.Errorf("API %q already exists", name))
	} else if !isNotFound(err) {
		return nil, err
	}

	if err := name.Validate(); err != nil {
		return nil, invalidArgumentError(err)
	}

	api, err := models.NewApi(name, req.GetApi())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	if err := saveApi(ctx, client, api); err != nil {
		return nil, err
	}

	message, err := api.Message(rpc.View_BASIC)
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_CREATED, name.String())
	return message, nil
}

// DeleteApi handles the corresponding API request.
func (s *RegistryServer) DeleteApi(ctx context.Context, req *rpc.DeleteApiRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	name, err := names.ParseApi(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Deletion should only succeed on APIs that currently exist.
	if _, err := getApi(ctx, client, name); err != nil {
		return nil, err
	}

	if err := deleteApi(ctx, client, name); err != nil {
		return nil, err
	}

	s.notify(rpc.Notification_DELETED, name.String())
	return &empty.Empty{}, nil
}

// GetApi handles the corresponding API request.
func (s *RegistryServer) GetApi(ctx context.Context, req *rpc.GetApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	name, err := names.ParseApi(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	api, err := getApi(ctx, client, name)
	if err != nil {
		return nil, err
	}

	message, err := api.Message(req.GetView())
	if err != nil {
		return nil, internalError(err)
	}

	return message, nil
}

// ListApis handles the corresponding API request.
func (s *RegistryServer) ListApis(ctx context.Context, req *rpc.ListApisRequest) (*rpc.ListApisResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetPageSize() < 0 {
		return nil, invalidArgumentError(fmt.Errorf("invalid page_size %q: must not be negative", req.GetPageSize()))
	}

	q := client.NewQuery(storage.ApiEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	name, err := names.ParseProject(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if name.ProjectID != "-" {
		q = q.Require("ProjectID", name.ProjectID)
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
			if !out.Value().(bool) {
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

	name, err := names.ParseApi(req.Api.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	api, err := getApi(ctx, client, name)
	if err != nil {
		return nil, err
	}

	if err := api.Update(req.GetApi(), req.GetUpdateMask()); err != nil {
		return nil, internalError(err)
	}

	if err := saveApi(ctx, client, api); err != nil {
		return nil, err
	}

	message, err := api.Message(rpc.View_BASIC)
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_UPDATED, name.String())
	return message, nil
}

func saveApi(ctx context.Context, client storage.Client, api *models.Api) error {
	k := client.NewKey(storage.ApiEntityName, api.ResourceName())
	if _, err := client.Put(ctx, k, api); err != nil {
		return internalError(err)
	}

	return nil
}

func getApi(ctx context.Context, client storage.Client, name names.Api) (*models.Api, error) {
	api := &models.Api{
		ApiID: name.ApiID,
	}

	k := client.NewKey(storage.ApiEntityName, name.String())
	if err := client.Get(ctx, k, api); client.IsNotFound(err) {
		return nil, notFoundError(fmt.Errorf("api %q not found", name))
	} else if err != nil {
		return nil, internalError(err)
	}

	return api, nil
}

func deleteApi(ctx context.Context, client storage.Client, name names.Api) error {
	if err := client.DeleteChildrenOfApi(ctx, name); err != nil {
		return internalError(err)
	}

	k := client.NewKey(storage.ApiEntityName, name.String())
	if err := client.Delete(ctx, k); err != nil {
		return internalError(err)
	}

	return nil
}
