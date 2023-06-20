// Copyright 2020 Google LLC.
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

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateApiVersion handles the corresponding API request.
func (s *RegistryServer) CreateApiVersion(ctx context.Context, req *rpc.CreateApiVersionRequest) (*rpc.ApiVersion, error) {
	// Parent name must be valid.
	parent, err := names.ParseApi(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Version name must be valid.
	name := parent.Version(req.GetApiVersionId())
	if err := name.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var response *rpc.ApiVersion
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		var err error
		response, err = s.createApiVersion(ctx, db, name, req.GetApiVersion())
		return err
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_CREATED, response.GetName())
	return response, nil
}

func (s *RegistryServer) createApiVersion(ctx context.Context, db *storage.Client, name names.Version, body *rpc.ApiVersion) (*rpc.ApiVersion, error) {
	version, err := models.NewVersion(name, body)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := db.CreateVersion(ctx, version); err != nil {
		return nil, err
	}

	return version.Message()
}

// DeleteApiVersion handles the corresponding API request.
func (s *RegistryServer) DeleteApiVersion(ctx context.Context, req *rpc.DeleteApiVersionRequest) (*emptypb.Empty, error) {
	// Version name must be valid.
	name, err := names.ParseVersion(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		return db.LockVersions(ctx).DeleteVersion(ctx, name, req.GetForce())
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_DELETED, req.GetName())
	return &emptypb.Empty{}, nil
}

// GetApiVersion handles the corresponding API request.
func (s *RegistryServer) GetApiVersion(ctx context.Context, req *rpc.GetApiVersionRequest) (*rpc.ApiVersion, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

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

	listing, err := db.ListVersions(ctx, parent, storage.PageOptions{
		Size:   req.GetPageSize(),
		Filter: req.GetFilter(),
		Order:  req.GetOrderBy(),
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

// UpdateApiVersion handles the corresponding API request.
func (s *RegistryServer) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	// Version body must be valid.
	if req.GetApiVersion() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api_version %+v: body must be provided", req.GetApiVersion())
	}
	// Version name must be valid.
	name, err := names.ParseVersion(req.ApiVersion.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Update mask must be valid.
	if err := models.ValidateMask(req.GetApiVersion(), req.GetUpdateMask()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid update_mask %v: %s", req.GetUpdateMask(), err)
	}
	var response *rpc.ApiVersion
	if err = s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		db.LockVersions(ctx)
		version, err := db.GetVersion(ctx, name)
		if err == nil {
			if err := version.Update(req.GetApiVersion(), models.ExpandMask(req.GetApiVersion(), req.GetUpdateMask())); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			if err := db.SaveVersion(ctx, version); err != nil {
				return err
			}
			response, err = version.Message()
			return err
		} else if status.Code(err) == codes.NotFound && req.GetAllowMissing() {
			response, err = s.createApiVersion(ctx, db, name, req.GetApiVersion())
			return err
		} else {
			return err
		}
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_UPDATED, response.GetName())
	return response, nil
}
