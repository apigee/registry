// Copyright 2021 Google LLC.
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

// CreateApiDeployment handles the corresponding API request.
func (s *RegistryServer) CreateApiDeployment(ctx context.Context, req *rpc.CreateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	// Parent name must be valid.
	parent, err := names.ParseApi(req.GetParent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Deployment name must be valid.
	name := parent.Deployment(req.GetApiDeploymentId())
	if err := name.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var response *rpc.ApiDeployment
	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		var err error
		response, err = s.createDeployment(ctx, db, name, req.GetApiDeployment())
		return err
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_CREATED, response.GetName())
	return response, nil
}

func (s *RegistryServer) createDeployment(ctx context.Context, db *storage.Client, name names.Deployment, body *rpc.ApiDeployment) (*rpc.ApiDeployment, error) {
	// The deployment must not already exist.
	if _, err := db.GetDeployment(ctx, name); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "API deployment %q already exists", name)
	} else if !isNotFound(err) {
		return nil, err
	}

	deployment, err := models.NewDeployment(name, body)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := db.CreateDeploymentRevision(ctx, deployment); err != nil {
		return nil, err
	}

	return deployment.BasicMessage(name.String())
}

// DeleteApiDeployment handles the corresponding API request.
func (s *RegistryServer) DeleteApiDeployment(ctx context.Context, req *rpc.DeleteApiDeploymentRequest) (*emptypb.Empty, error) {
	// Deployment name must be valid.
	name, err := names.ParseDeployment(req.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		return db.LockDeployments(ctx).DeleteDeployment(ctx, name, req.GetForce())
	}); err != nil {
		return nil, err
	}
	s.notify(ctx, rpc.Notification_DELETED, req.GetName())
	return &emptypb.Empty{}, nil
}

// GetApiDeployment handles the corresponding API request.
func (s *RegistryServer) GetApiDeployment(ctx context.Context, req *rpc.GetApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	if name, err := names.ParseDeployment(req.GetName()); err == nil {
		return s.getApiDeployment(ctx, name)
	} else if name, err := names.ParseDeploymentRevision(req.GetName()); err == nil {
		return s.getApiDeploymentRevision(ctx, name)
	}

	return nil, status.Errorf(codes.InvalidArgument, "invalid resource name %q, must be an API deployment or revision", req.GetName())
}

func (s *RegistryServer) getApiDeployment(ctx context.Context, name names.Deployment) (*rpc.ApiDeployment, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	deployment, err := db.GetDeployment(ctx, name)
	if err != nil {
		return nil, err
	}

	message, err := deployment.BasicMessage(name.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return message, nil
}

func (s *RegistryServer) getApiDeploymentRevision(ctx context.Context, name names.DeploymentRevision) (*rpc.ApiDeployment, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	revision, err := db.GetDeploymentRevision(ctx, name)
	if err != nil {
		return nil, err
	}

	message, err := revision.BasicMessage(name.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return message, nil
}

// ListApiDeployments handles the corresponding API request.
func (s *RegistryServer) ListApiDeployments(ctx context.Context, req *rpc.ListApiDeploymentsRequest) (*rpc.ListApiDeploymentsResponse, error) {
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

	listing, err := db.ListDeployments(ctx, parent, storage.PageOptions{
		Size:   req.GetPageSize(),
		Filter: req.GetFilter(),
		Order:  req.GetOrderBy(),
		Token:  req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}

	response := &rpc.ListApiDeploymentsResponse{
		ApiDeployments: make([]*rpc.ApiDeployment, len(listing.Deployments)),
		NextPageToken:  listing.Token,
	}

	for i, deployment := range listing.Deployments {
		response.ApiDeployments[i], err = deployment.BasicMessage(deployment.Name())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

// UpdateApiDeployment handles the corresponding API request.
func (s *RegistryServer) UpdateApiDeployment(ctx context.Context, req *rpc.UpdateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	// Deployment body must be valid.
	if req.GetApiDeployment() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid api_deployment %+v: body must be provided", req.GetApiDeployment())
	}
	// Deployment name must be valid.
	name, err := names.ParseDeployment(req.ApiDeployment.GetName())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Update mask must be valid.
	if err := models.ValidateMask(req.GetApiDeployment(), req.GetUpdateMask()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid update_mask %v: %s", req.GetUpdateMask(), err)
	}
	var response *rpc.ApiDeployment
	if err = s.runInTransaction(ctx, func(ctx context.Context, db *storage.Client) error {
		deployment, err := db.GetDeployment(ctx, name)
		if err == nil {
			// Apply the update to the deployment - possibly changing the revision ID.
			maskExpansion := models.ExpandMask(req.GetApiDeployment(), req.GetUpdateMask())
			if err := deployment.Update(req.GetApiDeployment(), maskExpansion); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			// Save the updated/current deployment. This creates a new revision or updates the previous one.
			if err := db.SaveDeploymentRevision(ctx, deployment); err != nil {
				return err
			}
			response, err = deployment.BasicMessage(name.String())
			return err
		} else if status.Code(err) == codes.NotFound && req.GetAllowMissing() {
			response, err = s.createDeployment(ctx, db, name, req.GetApiDeployment())
			if status.Code(err) == codes.AlreadyExists {
				err = status.Error(codes.Aborted, err.Error())
			}
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
