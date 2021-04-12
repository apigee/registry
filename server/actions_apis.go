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

// CreateApi handles the corresponding API request.
func (s *RegistryServer) CreateApi(ctx context.Context, req *rpc.CreateApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	parent, err := names.ParseProject(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Creation should only succeed when the parent exists.
	if _, err := db.GetProject(ctx, parent); err != nil {
		return nil, err
	}

	name := parent.Api(req.GetApiId())
	if name.ApiID == "" {
		name.ApiID = names.GenerateID()
	}

	if _, err := db.GetApi(ctx, name); err == nil {
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

	if err := db.SaveApi(ctx, api); err != nil {
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
	db := dao.NewDAO(client)

	name, err := names.ParseApi(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Deletion should only succeed on APIs that currently exist.
	if _, err := db.GetApi(ctx, name); err != nil {
		return nil, err
	}

	if err := db.DeleteApi(ctx, name); err != nil {
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
	db := dao.NewDAO(client)

	name, err := names.ParseApi(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	api, err := db.GetApi(ctx, name)
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
	db := dao.NewDAO(client)

	if req.GetPageSize() < 0 {
		return nil, invalidArgumentError(fmt.Errorf("invalid page_size %d: must not be negative", req.GetPageSize()))
	} else if req.GetPageSize() > 1000 {
		req.PageSize = 1000
	} else if req.GetPageSize() == 0 {
		req.PageSize = 50
	}

	parent, err := names.ParseProject(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	listing, err := db.ListApis(ctx, parent, dao.PageOptions{
		Size:   req.GetPageSize(),
		Filter: req.GetFilter(),
		Token:  req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}

	response := &rpc.ListApisResponse{
		Apis:          make([]*rpc.Api, len(listing.Apis)),
		NextPageToken: listing.Token,
	}

	for i, api := range listing.Apis {
		response.Apis[i], err = api.Message(req.GetView())
		if err != nil {
			return nil, internalError(err)
		}
	}

	return response, nil
}

// UpdateApi handles the corresponding API request.
func (s *RegistryServer) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetApi() == nil {
		return nil, invalidArgumentError(fmt.Errorf("invalid api %v: body must be provided", req.GetApi()))
	}

	name, err := names.ParseApi(req.Api.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	api, err := db.GetApi(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := api.Update(req.GetApi(), req.GetUpdateMask()); err != nil {
		return nil, internalError(err)
	}

	if err := db.SaveApi(ctx, api); err != nil {
		return nil, err
	}

	message, err := api.Message(rpc.View_BASIC)
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_UPDATED, name.String())
	return message, nil
}
