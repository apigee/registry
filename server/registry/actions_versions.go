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
	"sync"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage/gorm"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateApiVersion handles the corresponding API request.
func (s *RegistryServer) CreateApiVersion(ctx context.Context, req *rpc.CreateApiVersionRequest) (*rpc.ApiVersion, error) {
	parent, err := names.ParseApi(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GetApiVersion() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api_version %+v: body must be provided", req.GetApiVersion())
	}

	return s.createApiVersion(ctx, parent.Version(req.GetApiVersionId()), req.GetApiVersion())
}

func (s *RegistryServer) createApiVersion(ctx context.Context, name names.Version, body *rpc.ApiVersion) (*rpc.ApiVersion, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	if _, err := db.GetVersion(ctx, name); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "API version %q already exists", name)
	} else if !isNotFound(err) {
		return nil, err
	}

	if err := name.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Creation should only succeed when the parent exists.
	if _, err := db.GetApi(ctx, name.Api()); err != nil {
		return nil, err
	}

	version, err := models.NewVersion(name, body)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := db.SaveVersion(ctx, version); err != nil {
		return nil, err
	}

	message, err := version.Message()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_CREATED, name.String())
	return message, nil
}

// DeleteApiVersion handles the corresponding API request.
func (s *RegistryServer) DeleteApiVersion(ctx context.Context, req *rpc.DeleteApiVersionRequest) (*emptypb.Empty, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	name, err := names.ParseVersion(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Deletion should only succeed on API versions that currently exist.
	if _, err := db.GetVersion(ctx, name); err != nil {
		return nil, err
	}

	if err := db.DeleteVersion(ctx, name); err != nil {
		return nil, err
	}

	s.notify(ctx, rpc.Notification_DELETED, name.String())
	return &emptypb.Empty{}, nil
}

// GetApiVersion handles the corresponding API request.
func (s *RegistryServer) GetApiVersion(ctx context.Context, req *rpc.GetApiVersionRequest) (*rpc.ApiVersion, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	name, err := names.ParseVersion(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	version, err := db.GetVersion(ctx, name)
	if err != nil {
		return nil, err
	}

	message, err := version.Message()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return message, nil
}

// ListApiVersions handles the corresponding API request.
func (s *RegistryServer) ListApiVersions(ctx context.Context, req *rpc.ListApiVersionsRequest) (*rpc.ListApiVersionsResponse, error) {
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

	parent, err := names.ParseApi(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	listing, err := db.ListVersions(ctx, parent, gorm.PageOptions{
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
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

var updateVersionMutex sync.Mutex

// UpdateApiVersion handles the corresponding API request.
func (s *RegistryServer) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	if req.GetApiVersion() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api_version %+v: body must be provided", req.GetApiVersion())
	} else if err := models.ValidateMask(req.GetApiVersion(), req.GetUpdateMask()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid update_mask %v: %s", req.GetUpdateMask(), err)
	}

	name, err := names.ParseVersion(req.ApiVersion.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GetAllowMissing() {
		// Prevent a race condition that can occur when two updates are made
		// to the same non-existent resource. The db.Get...() call returns
		// NotFound for both updates, and after one creates the resource,
		// the other creation fails. The lock() prevents this by serializing
		// the get and create operations. Future updates could improve this
		// with improvements closer to the database level.
		updateVersionMutex.Lock()
		defer updateVersionMutex.Unlock()
	}

	version, err := db.GetVersion(ctx, name)
	if req.GetAllowMissing() && isNotFound(err) {
		return s.createApiVersion(ctx, name, req.GetApiVersion())
	} else if err != nil {
		return nil, err
	}

	if err := version.Update(req.GetApiVersion(), models.ExpandMask(req.GetApiVersion(), req.GetUpdateMask())); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := db.SaveVersion(ctx, version); err != nil {
		return nil, err
	}

	message, err := version.Message()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_UPDATED, name.String())
	return message, nil
}
