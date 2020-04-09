// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"

	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) CreateProperty(ctx context.Context, request *rpc.CreatePropertyRequest) (*rpc.Property, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	propertyID := request.GetPropertyId()
	if propertyID == "" {
		propertyID = newRandomID()
	}
	property, err := models.NewPropertyFromParentAndPropertyID(request.GetParent(), propertyID)
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := &datastore.Key{Kind: models.PropertyEntityName, Name: property.ResourceName()}
	// fail if property already exists
	var existingProperty models.Property
	err = client.Get(ctx, k, &existingProperty)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, property.ResourceName()+" already exists")
	}
	err = property.Update(request.GetProperty())
	property.CreateTime = property.UpdateTime
	k, err = client.Put(ctx, k, property)
	if err != nil {
		return nil, internalError(err)
	}
	return property.Message()
}

func (s *server) DeleteProperty(ctx context.Context, request *rpc.DeletePropertyRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	// Validate name and create dummy property (we just need the ID fields).
	_, err = models.NewPropertyFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete the property.
	k := &datastore.Key{Kind: models.PropertyEntityName, Name: request.GetName()}
	err = client.Delete(ctx, k)
	return &empty.Empty{}, internalError(err)
}

func (s *server) GetProperty(ctx context.Context, request *rpc.GetPropertyRequest) (*rpc.Property, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	property, err := models.NewPropertyFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := &datastore.Key{Kind: models.PropertyEntityName, Name: property.ResourceName()}
	err = client.Get(ctx, k, property)
	if err == datastore.ErrNoSuchEntity {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return property.Message()
}

func (s *server) ListProperties(ctx context.Context, req *rpc.ListPropertiesRequest) (*rpc.ListPropertiesResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	q := datastore.NewQuery(models.PropertyEntityName)
	q = queryApplyPageSize(q, req.GetPageSize())
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := models.ParseParentProject(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	q = q.Filter("ProjectID =", m[1])
	var propertyMessages []*rpc.Property
	var property models.Property
	it := client.Run(ctx, q.Distinct())
	_, err = it.Next(&property)
	for err == nil {
		propertyMessage, _ := property.Message()
		propertyMessages = append(propertyMessages, propertyMessage)
		_, err = it.Next(&property)
	}
	if err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListPropertiesResponse{
		Properties: propertyMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(propertyMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

func (s *server) UpdateProperty(ctx context.Context, request *rpc.UpdatePropertyRequest) (*rpc.Property, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	property, err := models.NewPropertyFromResourceName(request.GetProperty().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := &datastore.Key{Kind: models.PropertyEntityName, Name: property.ResourceName()}
	err = client.Get(ctx, k, &property)
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
	return property.Message()
}
