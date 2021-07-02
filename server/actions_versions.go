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

// CreateApiVersion handles the corresponding API request.
func (s *RegistryServer) CreateApiVersion(ctx context.Context, req *rpc.CreateApiVersionRequest) (*rpc.ApiVersion, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetApiVersion() == nil {
		return nil, invalidArgumentError(fmt.Errorf("invalid api_version %+v: body must be provided", req.GetApiVersion()))
	}

	parent, err := names.ParseApi(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Creation should only succeed when the parent exists.
	if _, err := db.GetApi(ctx, parent); err != nil {
		return nil, err
	}

	name := parent.Version(req.GetApiVersionId())
	if _, err := db.GetVersion(ctx, name); err == nil {
		return nil, alreadyExistsError(fmt.Errorf("API version %q already exists", name))
	} else if !isNotFound(err) {
		return nil, err
	}

	if err := name.Validate(); err != nil {
		return nil, invalidArgumentError(err)
	}

	version, err := models.NewVersion(name, req.GetApiVersion())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	if err := db.SaveVersion(ctx, version); err != nil {
		return nil, err
	}

	message, err := version.Message()
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(ctx, rpc.Notification_CREATED, name.String())
	return message, nil
}

// DeleteApiVersion handles the corresponding API request.
func (s *RegistryServer) DeleteApiVersion(ctx context.Context, req *rpc.DeleteApiVersionRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	name, err := names.ParseVersion(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Deletion should only succeed on API versions that currently exist.
	if _, err := db.GetVersion(ctx, name); err != nil {
		return nil, err
	}

	if err := db.DeleteVersion(ctx, name); err != nil {
		return nil, err
	}

	s.notify(ctx, rpc.Notification_DELETED, name.String())
	return &empty.Empty{}, nil
}

// GetApiVersion handles the corresponding API request.
func (s *RegistryServer) GetApiVersion(ctx context.Context, req *rpc.GetApiVersionRequest) (*rpc.ApiVersion, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	name, err := names.ParseVersion(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	version, err := db.GetVersion(ctx, name)
	if err != nil {
		return nil, err
	}

	message, err := version.Message()
	if err != nil {
		return nil, internalError(err)
	}

	return message, nil
}

// ListApiVersions handles the corresponding API request.
func (s *RegistryServer) ListApiVersions(ctx context.Context, req *rpc.ListApiVersionsRequest) (*rpc.ListApiVersionsResponse, error) {
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

	parent, err := names.ParseApi(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	listing, err := db.ListVersions(ctx, parent, dao.PageOptions{
		Size:   req.GetPageSize(),
		Filter: req.GetFilter(),
		Token:  req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}

	response := &rpc.ListApiVersionsResponse{
		ApiVersions:   make([]*rpc.ApiVersion, len(listing.Versions)),
		NextPageToken: listing.Token,
	}

	for i, version := range listing.Versions {
		response.ApiVersions[i], err = version.Message()
		if err != nil {
			return nil, internalError(err)
		}
	}

	return response, nil
}

// UpdateApiVersion handles the corresponding API request.
func (s *RegistryServer) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	if req.GetApiVersion() == nil {
		return nil, invalidArgumentError(fmt.Errorf("invalid api_version %+v: body must be provided", req.GetApiVersion()))
	} else if err := models.ValidateMask(req.GetApiVersion(), req.GetUpdateMask()); err != nil {
		return nil, invalidArgumentError(fmt.Errorf("invalid update_mask %v: %s", req.GetUpdateMask(), err))
	}

	name, err := names.ParseVersion(req.ApiVersion.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	version, err := db.GetVersion(ctx, name)
	if err != nil {
		return nil, err
	}

	if err := version.Update(req.GetApiVersion(), models.ExpandMask(req.GetApiVersion(), req.GetUpdateMask())); err != nil {
		return nil, internalError(err)
	}

	if err := db.SaveVersion(ctx, version); err != nil {
		return nil, err
	}

	message, err := version.Message()
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(ctx, rpc.Notification_UPDATED, name.String())
	return message, nil
}
