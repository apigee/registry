// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"
	"errors"
	"log"

	"apigov.dev/registry/models"
	rpc "apigov.dev/registry/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateSpec handles the corresponding API request.
func (s *RegistryServer) CreateSpec(ctx context.Context, request *rpc.CreateSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, err := models.NewSpecFromParentAndSpecID(request.GetParent(), request.GetSpecId())
	if err != nil {
		return nil, err
	}
	// fail if spec already exists
	q := datastore.NewQuery(models.SpecEntityName)
	q = q.Filter("ProjectID =", spec.ProjectID)
	q = q.Filter("ProductID =", spec.ProductID)
	q = q.Filter("VersionID =", spec.VersionID)
	q = q.Filter("SpecID =", spec.SpecID)
	it := client.Run(ctx, q.Distinct())
	var existingSpec models.Spec
	existingKey, err := it.Next(&existingSpec)
	if existingKey != nil {
		return nil, status.Error(codes.AlreadyExists, spec.ResourceName()+" already exists")
	}
	// save the spec under its full resource@revision name
	err = spec.Update(request.GetSpec())
	spec.CreateTime = spec.UpdateTime
	k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceNameWithRevision()}
	k, err = client.Put(ctx, k, spec)
	if err != nil {
		return nil, err
	}
	return spec.Message(rpc.SpecView_BASIC, false)
}

// DeleteSpec handles the corresponding API request.
func (s *RegistryServer) DeleteSpec(ctx context.Context, request *rpc.DeleteSpecRequest) (*empty.Empty, error) {
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

// GetSpec handles the corresponding API request.
func (s *RegistryServer) GetSpec(ctx context.Context, request *rpc.GetSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, request.GetName())
	if err != nil {
		return nil, err
		// return nil, status.Error(codes.NotFound, "not found")
	}
	return spec.Message(request.GetView(), userSpecifiedRevision)
}

// ListSpecs handles the corresponding API request.
func (s *RegistryServer) ListSpecs(ctx context.Context, req *rpc.ListSpecsRequest) (*rpc.ListSpecsResponse, error) {
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
			{"project_id", filterArgTypeString},
			{"product_id", filterArgTypeString},
			{"version_id", filterArgTypeString},
			{"spec_id", filterArgTypeString},
			{"filename", filterArgTypeString},
			{"description", filterArgTypeString},
			{"style", filterArgTypeString},
		})
	if err != nil {
		return nil, internalError(err)
	}
	var specMessages []*rpc.Spec
	var spec models.Spec
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for _, err := it.Next(&spec); err == nil; _, err = it.Next(&spec) {
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id":  spec.ProjectID,
				"product_id":  spec.ProductID,
				"version_id":  spec.VersionID,
				"spec_id":     spec.SpecID,
				"filename":    spec.FileName,
				"description": spec.Description,
				"style":       spec.Style,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		specMessage, _ := spec.Message(req.GetView(), false)
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

// UpdateSpec handles the corresponding API request.
func (s *RegistryServer) UpdateSpec(ctx context.Context, request *rpc.UpdateSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, request.GetSpec().GetName())
	if err != nil {
		return nil, err
	}
	if userSpecifiedRevision {
		return nil, invalidArgumentError(errors.New("updates to specific revisions are unsupported"))
	}
	err = spec.Update(request.GetSpec())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceNameWithRevision()}
	k, err = client.Put(ctx, k, spec)
	if err != nil {
		return nil, err
	}
	return spec.Message(rpc.SpecView_BASIC, false)
}

// ListSpecRevisions handles the corresponding API request.
func (s *RegistryServer) ListSpecRevisions(ctx context.Context, req *rpc.ListSpecRevisionsRequest) (*rpc.ListSpecRevisionsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	targetSpec, err := models.NewSpecFromResourceName(req.GetName())
	if err != nil {
		return nil, err
	}
	q := datastore.NewQuery(models.SpecEntityName)
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	q = q.Filter("ProjectID =", targetSpec.ProjectID)
	q = q.Filter("ProductID =", targetSpec.ProductID)
	q = q.Filter("VersionID =", targetSpec.VersionID)
	q = q.Filter("SpecID =", targetSpec.SpecID)
	q = q.Order("-CreateTime")

	var specMessages []*rpc.Spec
	var spec models.Spec
	log.Printf("%+v", q)
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for _, err := it.Next(&spec); err == nil; _, err = it.Next(&spec) {
		specMessage, _ := spec.Message(rpc.SpecView_BASIC, true)
		specMessages = append(specMessages, specMessage)
		log.Printf("%+v", specMessage)
		if len(specMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListSpecRevisionsResponse{
		Specs: specMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(specMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// fetchSpec gets the stored model of a Spec.
func fetchSpec(
	ctx context.Context,
	client *datastore.Client,
	name string,
) (*models.Spec, bool, error) {
	spec, err := models.NewSpecFromResourceName(name)
	if err != nil {
		return nil, false, err
	}
	// if there's no revision, get the spec with the most recent revision
	if spec.RevisionID == "" {
		q := datastore.NewQuery(models.SpecEntityName)
		q = q.Filter("ProjectID =", spec.ProjectID)
		q = q.Filter("ProductID =", spec.ProductID)
		q = q.Filter("VersionID =", spec.VersionID)
		q = q.Filter("SpecID =", spec.SpecID)
		q = q.Order("-CreateTime")
		log.Printf("query %+v", q)
		it := client.Run(ctx, q.Distinct())
		_, err = it.Next(spec)
		if err != nil {
			return nil, false, err
		}
		return spec, false, nil
	}
	// if the revision reference is a tag, resolve the tag
	// ... todo ...
	// if we know the revision, get the spec by revision
	k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceNameWithRevision()}
	err = client.Get(ctx, k, spec)
	if err == datastore.ErrNoSuchEntity {
		return nil, true, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, true, internalError(err)
	}
	return spec, true, nil
}
