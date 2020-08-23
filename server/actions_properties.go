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

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/models"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateProperty handles the corresponding API request.
func (s *RegistryServer) CreateProperty(ctx context.Context, request *rpc.CreatePropertyRequest) (*rpc.Property, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseStorageClient(client)
	property, err := models.NewPropertyFromParentAndPropertyID(request.GetParent(), request.GetPropertyId())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	err = property.Update(request.GetProperty())
	k := client.NewKey(models.PropertyEntityName, property.ResourceName())
	// fail if property already exists
	var existingProperty models.Property
	err = client.Get(ctx, k, &existingProperty)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, property.ResourceName()+" already exists")
	}
	property.CreateTime = property.UpdateTime
	k, err = client.Put(ctx, k, property)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_CREATED, property.ResourceName())
	return property.Message()
}

// DeleteProperty handles the corresponding API request.
func (s *RegistryServer) DeleteProperty(ctx context.Context, request *rpc.DeletePropertyRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseStorageClient(client)
	// Validate name and create dummy property (we just need the ID fields).
	_, err = models.NewPropertyFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete the property.
	k := client.NewKey(models.PropertyEntityName, request.GetName())
	err = client.Delete(ctx, k)
	s.notify(rpc.Notification_DELETED, request.GetName())
	return &empty.Empty{}, internalError(err)
}

// GetProperty handles the corresponding API request.
func (s *RegistryServer) GetProperty(ctx context.Context, request *rpc.GetPropertyRequest) (*rpc.Property, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseStorageClient(client)
	property, err := models.NewPropertyFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.PropertyEntityName, property.ResourceName())
	err = client.Get(ctx, k, property)
	if err == client.ErrNotFound() {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return property.Message()
}

// ListProperties handles the corresponding API request.
func (s *RegistryServer) ListProperties(ctx context.Context, req *rpc.ListPropertiesRequest) (*rpc.ListPropertiesResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseStorageClient(client)
	q := client.NewQuery(models.PropertyEntityName)
	q, err = client.QueryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	p, err := models.NewPropertyFromParentAndPropertyID(req.GetParent(), "-")
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if p.ProjectID != "-" {
		q = q.Filter("ProjectID =", p.ProjectID)
	}
	if p.ApiID != "-" {
		q = q.Filter("ApiID =", p.ApiID)
	}
	if p.VersionID != "-" {
		q = q.Filter("VersionID =", p.VersionID)
	}
	if p.SpecID != "-" {
		q = q.Filter("SpecID =", p.SpecID)
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"project_id", filterArgTypeString},
			{"api_id", filterArgTypeString},
			{"version_id", filterArgTypeString},
			{"spec_id", filterArgTypeString},
			{"property_id", filterArgTypeString},
		})
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	var propertyMessages []*rpc.Property
	var property models.Property
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&property); err == nil; _, err = it.Next(&property) {
		// don't allow wildcarded-names to be empty
		if p.SpecID == "-" && property.SpecID == "" {
			continue
		}
		if p.VersionID == "-" && property.VersionID == "" {
			continue
		}
		if p.ApiID == "-" && property.ApiID == "" {
			continue
		}
		if p.ProjectID == "-" && property.ProjectID == "" {
			continue
		}
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id":  property.ProjectID,
				"api_id":      property.ApiID,
				"version_id":  property.VersionID,
				"spec_id":     property.SpecID,
				"property_id": property.PropertyID,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		propertyMessage, _ := property.Message()
		propertyMessages = append(propertyMessages, propertyMessage)
		if len(propertyMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListPropertiesResponse{
		Properties: propertyMessages,
	}
	responses.NextPageToken, err = client.IteratorGetCursor(it, len(propertyMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// UpdateProperty handles the corresponding API request.
func (s *RegistryServer) UpdateProperty(ctx context.Context, request *rpc.UpdatePropertyRequest) (*rpc.Property, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseStorageClient(client)
	property, err := models.NewPropertyFromResourceName(request.GetProperty().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.PropertyEntityName, request.GetProperty().GetName())
	err = client.Get(ctx, k, property)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = property.Update(request.GetProperty())
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, property)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_UPDATED, property.ResourceName())
	return property.Message()
}
