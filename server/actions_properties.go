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
	"log"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/models"
	storage "github.com/apigee/registry/server/storage"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateProperty handles the corresponding API request.
func (s *RegistryServer) CreateProperty(ctx context.Context, request *rpc.CreatePropertyRequest) (*rpc.Property, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	property, err := models.NewPropertyFromParentAndPropertyID(request.GetParent(), request.GetPropertyId())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// create a blob for the property contents
	blob := models.NewBlobForProperty(property, nil)
	err = property.Update(request.GetProperty(), blob)
	k := client.NewKey(models.PropertyEntityName, property.ResourceName())
	// fail if property already exists
	existingProperty := &models.Property{}
	err = client.Get(ctx, k, existingProperty)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, property.ResourceName()+" already exists")
	}
	property.CreateTime = property.UpdateTime
	k, err = client.Put(ctx, k, property)
	if err != nil {
		return nil, internalError(err)
	}
	if blob != nil {
		switch property.ValueType {
		case models.BytesType, models.AnyType:
			k2 := client.NewKey(models.BlobEntityName, property.ResourceName())
			_, err = client.Put(ctx,
				k2,
				blob)
		default: // do nothing
		}
	}
	if err != nil {
		log.Printf("save blob error %+v", err)
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_CREATED, property.ResourceName())
	return property.Message(nil)
}

// DeleteProperty handles the corresponding API request.
func (s *RegistryServer) DeleteProperty(ctx context.Context, request *rpc.DeletePropertyRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	// Validate name and create dummy property (we just need the ID fields).
	_, err = models.NewPropertyFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete the property.
	k := client.NewKey(models.PropertyEntityName, request.GetName())
	err = client.Delete(ctx, k)
	// Delete any blob associated with the property.
	k2 := client.NewKey(models.BlobEntityName, request.GetName())
	err = client.Delete(ctx, k2)
	s.notify(rpc.Notification_DELETED, request.GetName())
	return &empty.Empty{}, internalError(err)
}

// GetProperty handles the corresponding API request.
func (s *RegistryServer) GetProperty(ctx context.Context, request *rpc.GetPropertyRequest) (*rpc.Property, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	property, err := models.NewPropertyFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.PropertyEntityName, property.ResourceName())
	err = client.Get(ctx, k, property)
	var blob *models.Blob
	if request.GetView() == rpc.View_FULL {
		if property.ValueType == models.BytesType || property.ValueType == models.AnyType {
			blob, _ = fetchBlobForProperty(ctx, client, property)
		}
	}
	if client.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return property.Message(blob)
}

// ListProperties handles the corresponding API request.
func (s *RegistryServer) ListProperties(ctx context.Context, req *rpc.ListPropertiesRequest) (*rpc.ListPropertiesResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	q := client.NewQuery(models.PropertyEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	p, err := models.NewPropertyFromParentAndPropertyID(req.GetParent(), "-")
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if p.ProjectID != "-" {
		q = q.Require("ProjectID", p.ProjectID)
	}
	if p.ApiID != "-" {
		q = q.Require("ApiID", p.ApiID)
	}
	if p.VersionID != "-" {
		q = q.Require("VersionID", p.VersionID)
	}
	if p.SpecID != "-" {
		q = q.Require("SpecID", p.SpecID)
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
	it := client.Run(ctx, q)
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
		var blob *models.Blob
		if req.GetView() == rpc.View_FULL {
			if property.ValueType == models.BytesType || property.ValueType == models.AnyType {
				blob, _ = fetchBlobForProperty(ctx, client, &property)
			}
		}
		propertyMessage, _ := property.Message(blob)
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
	responses.NextPageToken, err = it.GetCursor(len(propertyMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// UpdateProperty handles the corresponding API request.
func (s *RegistryServer) UpdateProperty(ctx context.Context, request *rpc.UpdatePropertyRequest) (*rpc.Property, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	property, err := models.NewPropertyFromResourceName(request.GetProperty().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.PropertyEntityName, request.GetProperty().GetName())
	err = client.Get(ctx, k, property)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	blob := models.NewBlobForProperty(property, nil)
	err = property.Update(request.GetProperty(), blob)
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, property)
	if err != nil {
		return nil, internalError(err)
	}
	if blob != nil {
		switch property.ValueType {
		case models.BytesType, models.AnyType:
			k2 := client.NewKey(models.BlobEntityName, property.ResourceName())
			_, err = client.Put(ctx,
				k2,
				blob)
		default: // do nothing
		}
	}
	s.notify(rpc.Notification_UPDATED, property.ResourceName())
	return property.Message(nil)
}

func contentsForProperty(property *rpc.Property) []byte {
	switch v := property.Value.(type) {
	case *rpc.Property_StringValue:
		return nil
	case *rpc.Property_Int64Value:
		return nil
	case *rpc.Property_DoubleValue:
		return nil
	case *rpc.Property_BoolValue:
		return nil
	case *rpc.Property_BytesValue:
		return v.BytesValue
	case *rpc.Property_MessageValue:
		return v.MessageValue.Value
	default:
		return nil
	}
}

// fetchBlobForProperty gets the blob containing the property contents.
func fetchBlobForProperty(
	ctx context.Context,
	client storage.Client,
	property *models.Property) (*models.Blob, error) {
	blob := &models.Blob{}
	k := client.NewKey(models.BlobEntityName, property.ResourceName())
	err := client.Get(ctx, k, blob)
	return blob, err
}
