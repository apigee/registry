// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"

	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
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
	q = queryApplyPageSize(q, req.GetPageSize())
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := models.ParseParentVersion(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	q = q.Filter("ProjectID =", m[1])
	if m[2] != "-" {
		q = q.Filter("ProductID =", m[2])
	}
	if m[3] != "-" {
		q = q.Filter("VersionID =", m[3])
	}

	var prg cel.Program
	filter := req.GetFilter()
	if filter != "" {
		d := cel.Declarations(decls.NewIdent("style", decls.String, nil))
		env, err := cel.NewEnv(d)
		if err != nil {
			return nil, invalidArgumentError(err)
		}
		ast, iss := env.Compile(filter)
		// Check iss for compilation errors.
		if iss.Err() != nil {
			return nil, invalidArgumentError(iss.Err())
		}
		prg, err = env.Program(ast)
		if err != nil {
			return nil, invalidArgumentError(err)
		}
	}

	var specMessages []*rpc.Spec
	var spec models.Spec
	it := client.Run(ctx, q.Distinct())
	_, err = it.Next(&spec)
	for err == nil {
		specMessage, _ := spec.Message(req.GetView())
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"style": specMessage.GetStyle(),
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if out.Value().(bool) {
				specMessages = append(specMessages, specMessage)
			}
		} else {
			specMessages = append(specMessages, specMessage)
		}
		_, err = it.Next(&spec)
	}
	if err != iterator.Done {
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
	return spec.Message(rpc.SpecView_BASIC)
}
