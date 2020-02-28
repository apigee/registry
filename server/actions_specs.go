// Copyright 2020 Google Inc. All Rights Reserved.

package server

import (
	"context"
	"time"

	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const specEntityName = "Spec"

func (s *server) CreateSpec(ctx context.Context, request *rpc.CreateSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	spec, err := models.NewSpecFromMessage(request.Spec)
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: specEntityName, Name: spec.ResourceName()}
	spec.CreateTime = time.Now()
	spec.UpdateTime = spec.CreateTime
	k, err = client.Put(ctx, k, spec)
	if err != nil {
		return nil, err
	}
	return spec.Message()
}

func (s *server) DeleteSpec(ctx context.Context, request *rpc.DeleteSpecRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	// validate name
	_, err = models.NewSpecFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: specEntityName, Name: request.GetName()}
	// TODO: delete children
	err = client.Delete(ctx, k)
	return &empty.Empty{}, err
}

func (s *server) GetSpec(ctx context.Context, request *rpc.GetSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	spec, err := models.NewSpecFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: specEntityName, Name: spec.ResourceName()}
	err = client.Get(ctx, k, &spec)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return spec.Message()
}

func (s *server) ListSpecs(ctx context.Context, req *rpc.ListSpecsRequest) (*rpc.ListSpecsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	q := datastore.NewQuery(specEntityName)
	var specs []*models.Spec
	_, err = client.GetAll(ctx, q, &specs)
	var specMessages []*rpc.Spec
	for _, spec := range specs {
		specMessage, _ := spec.Message()
		specMessages = append(specMessages, specMessage)
	}
	responses := &rpc.ListSpecsResponse{
		Specs: specMessages,
	}
	return responses, nil
}

func (s *server) UpdateSpec(ctx context.Context, request *rpc.UpdateSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	spec, err := models.NewSpecFromResourceName(request.GetSpec().GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: specEntityName, Name: spec.ResourceName()}
	err = client.Get(ctx, k, &spec)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = spec.Update(request.GetSpec())
	if err != nil {
		return nil, err
	}
	k, err = client.Put(ctx, k, spec)
	if err != nil {
		return nil, err
	}
	return spec.Message()
}
