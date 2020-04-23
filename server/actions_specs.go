// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"
	"log"

	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FlameServer) CreateSpec(ctx context.Context, request *rpc.CreateSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, err := models.NewSpecFromParentAndSpecID(request.GetParent(), request.GetSpecId())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceName()}
	// fail if spec already exists
	var existingSpec models.Spec
	err = client.Get(ctx, k, &existingSpec)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, spec.ResourceName()+" already exists")
	}
	err = spec.Update(request.GetSpec())
	spec.CreateTime = spec.UpdateTime
	k, err = client.Put(ctx, k, spec)
	if err != nil {
		return nil, err
	}
	return spec.Message(rpc.SpecView_BASIC)
}

func (s *FlameServer) DeleteSpec(ctx context.Context, request *rpc.DeleteSpecRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	// Validate name and create dummy spec (we just need the ID fields).
	_, err = models.NewSpecFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete the spec.
	k := &datastore.Key{Kind: models.SpecEntityName, Name: request.GetName()}
	err = client.Delete(ctx, k)
	return &empty.Empty{}, err
}

func (s *FlameServer) GetSpec(ctx context.Context, request *rpc.GetSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, err := models.NewSpecFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceName()}
	err = client.Get(ctx, k, spec)
	if err == datastore.ErrNoSuchEntity {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return spec.Message(request.GetView())
}

func (s *FlameServer) ListSpecs(ctx context.Context, req *rpc.ListSpecsRequest) (*rpc.ListSpecsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	q := datastore.NewQuery(models.SpecEntityName)
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := models.ParseParentVersion(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if m[1] != "-" {
		q = q.Filter("ProjectID =", m[1])
	}
	if m[2] != "-" {
		q = q.Filter("ProductID =", m[2])
	}
	if m[3] != "-" {
		q = q.Filter("VersionID =", m[3])
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"spec_id", filterArgTypeString},
			{"style", filterArgTypeString},
		})
	if err != nil {
		return nil, internalError(err)
	}
	var specMessages []*rpc.Spec
	var spec models.Spec
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for i, err := it.Next(&spec); err == nil; _, err = it.Next(&spec) {
		log.Printf("%d", i)
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"spec_id": spec.SpecID,
				"style":   spec.Style,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		specMessage, _ := spec.Message(req.GetView())
		specMessages = append(specMessages, specMessage)
		if len(specMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListSpecsResponse{
		Specs: specMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(specMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

func (s *FlameServer) UpdateSpec(ctx context.Context, request *rpc.UpdateSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, err := models.NewSpecFromResourceName(request.GetSpec().GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceName()}
	err = client.Get(ctx, k, spec)
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
	return spec.Message(rpc.SpecView_BASIC)
}
