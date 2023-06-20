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

// CreateApi handles the corresponding API request.
func (s *RegistryServer) CreateApi(ctx context.Context, req *rpc.CreateApiRequest) (*rpc.Api, error) {
	// Parent name must be valid.
	parent, err := names.ParseProjectWithLocation(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	name := parent.Api(req.GetApiId())
	if err := name.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var response *rpc.Api
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		var err error
		response, err = s.createApi(ctx, db, name, req.GetApi())
		return err
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_CREATED, response.GetName())
	return response, nil
}

func (s *RegistryServer) createApi(ctx context.Context, db *storage.Client, name names.Api, body *rpc.Api) (*rpc.Api, error) {
	api, err := models.NewApi(name, body)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := db.CreateApi(ctx, api); err != nil {
		return nil, err
	}

	return api.Message()
}

// DeleteApi handles the corresponding API request.
func (s *RegistryServer) DeleteApi(ctx context.Context, req *rpc.DeleteApiRequest) (*emptypb.Empty, error) {
	// API name must be valid.
	name, err := names.ParseApi(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		return db.LockApis(ctx).DeleteApi(ctx, name, req.GetForce())
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_DELETED, req.GetName())
	return &emptypb.Empty{}, nil
}

// GetApi handles the corresponding API request.
func (s *RegistryServer) GetApi(ctx context.Context, req *rpc.GetApiRequest) (*rpc.Api, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	name, err := names.ParseApi(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	api, err := db.GetApi(ctx, name)
	if err != nil {
		return nil, err
	}

	message, err := api.Message()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return message, nil
}

// ListApis handles the corresponding API request.
func (s *RegistryServer) ListApis(ctx context.Context, req *rpc.ListApisRequest) (*rpc.ListApisResponse, error) {
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

	parent, err := names.ParseProjectWithLocation(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	listing, err := db.ListApis(ctx, parent, storage.PageOptions{
		Size:   req.GetPageSize(),
		Filter: req.GetFilter(),
		Order:  req.GetOrderBy(),
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
		response.Apis[i], err = api.Message()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

// UpdateApi handles the corresponding API request.
func (s *RegistryServer) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	// API body must be nonempty.
	if req.GetApi() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api %v: body must be provided", req.GetApi())
	}
	// API name must be valid.
	name, err := names.ParseApi(req.Api.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Update mask must be valid.
	if err := models.ValidateMask(req.GetApi(), req.GetUpdateMask()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid update_mask %v: %s", req.GetUpdateMask(), err)
	}
	var response *rpc.Api
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		db.LockApis(ctx)
		api, err := db.GetApi(ctx, name)
		if err == nil {
			if err := api.Update(req.GetApi(), models.ExpandMask(req.GetApi(), req.GetUpdateMask())); err != nil {
				return err
			}
			if err := db.SaveApi(ctx, api); err != nil {
				return err
			}
			response, err = api.Message()
			return err
		} else if status.Code(err) == codes.NotFound && req.GetAllowMissing() {
			response, err = s.createApi(ctx, db, name, req.GetApi())
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
