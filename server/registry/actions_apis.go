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
	"time"

	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateApi handles the corresponding API request.
func (s *RegistryServer) CreateApi(ctx context.Context, req *rpc.CreateApiRequest) (*rpc.Api, error) {
	parent, err := names.ParseProjectWithLocation(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GetApi() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api %+v: body must be provided", req.GetApi())
	}

	return s.createApi(ctx, parent.Api(req.GetApiId()), req.GetApi())
}

func (s *RegistryServer) createApi(ctx context.Context, name names.Api, body *rpc.Api) (*rpc.Api, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	if _, err := db.GetApi(ctx, name); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "API %q already exists", name)
	} else if !isNotFound(err) {
		return nil, err
	}

	if err := name.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Creation should only succeed when the parent exists.
	if _, err := db.GetProject(ctx, name.Project()); err != nil {
		return nil, err
	}

	api, err := models.NewApi(name, body)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := db.SaveApi(ctx, api); err != nil {
		return nil, err
	}

	message, err := api.Message()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_CREATED, name.String())
	return message, nil
}

// DeleteApi handles the corresponding API request.
func (s *RegistryServer) DeleteApi(ctx context.Context, req *rpc.DeleteApiRequest) (*emptypb.Empty, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	name, err := names.ParseApi(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Deletion should only succeed on APIs that currently exist.
	if _, err := db.GetApi(ctx, name); err != nil {
		return nil, err
	}

	if err := db.DeleteApi(ctx, name, req.GetForce()); err != nil {
		return nil, err
	}

	s.notify(ctx, rpc.Notification_DELETED, name.String())
	return &emptypb.Empty{}, nil
}

// GetApi handles the corresponding API request.
func (s *RegistryServer) GetApi(ctx context.Context, req *rpc.GetApiRequest) (*rpc.Api, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

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
	defer db.Close()

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

var updateApiMutex sync.Mutex

// UpdateApi handles the corresponding API request.
func (s *RegistryServer) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	defer db.Close()

	if req.GetApi() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api %v: body must be provided", req.GetApi())
	} else if err := models.ValidateMask(req.GetApi(), req.GetUpdateMask()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid update_mask %v: %s", req.GetUpdateMask(), err)
	}

	name, err := names.ParseApi(req.Api.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GetAllowMissing() {
		before := time.Now()
		// Prevent a race condition that can occur when two updates are made
		// to the same non-existent resource. The db.Get...() call returns
		// NotFound for both updates, and after one creates the resource,
		// the other creation fails. The lock() prevents this by serializing
		// the get and create operations. Future updates could improve this
		// with improvements closer to the database level.
		updateApiMutex.Lock()
		defer updateApiMutex.Unlock()
		log.Debugf(ctx, "Acquired lock after blocking for %v", time.Since(before))
	}

	api, err := db.GetApi(ctx, name)
	if req.GetAllowMissing() && isNotFound(err) {
		return s.createApi(ctx, name, req.GetApi())
	} else if err != nil {
		return nil, err
	}

	if err := api.Update(req.GetApi(), models.ExpandMask(req.GetApi(), req.GetUpdateMask())); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := db.SaveApi(ctx, api); err != nil {
		return nil, err
	}

	message, err := api.Message()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	s.notify(ctx, rpc.Notification_UPDATED, name.String())
	return message, nil
}
