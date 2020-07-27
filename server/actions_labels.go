// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"
	"log"

	"apigov.dev/registry/rpc"
	"apigov.dev/registry/server/models"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateLabel handles the corresponding API request.
func (s *RegistryServer) CreateLabel(ctx context.Context, request *rpc.CreateLabelRequest) (*rpc.Label, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
	label, err := models.NewLabelFromParentAndLabelID(request.GetParent(), request.GetLabelId())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	err = label.Update(request.GetLabel())
	k := &datastore.Key{Kind: models.LabelEntityName, Name: label.ResourceName()}
	// fail if label already exists
	var existingLabel models.Label
	err = client.Get(ctx, k, &existingLabel)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, label.ResourceName()+" already exists")
	}
	label.CreateTime = label.UpdateTime
	k, err = client.Put(ctx, k, label)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_CREATED, label.ResourceName())
	return label.Message()
}

// DeleteLabel handles the corresponding API request.
func (s *RegistryServer) DeleteLabel(ctx context.Context, request *rpc.DeleteLabelRequest) (*empty.Empty, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
	// Validate name and create dummy label (we just need the ID fields).
	_, err = models.NewLabelFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete the label.
	k := &datastore.Key{Kind: models.LabelEntityName, Name: request.GetName()}
	err = client.Delete(ctx, k)
	s.notify(rpc.Notification_DELETED, request.GetName())
	return &empty.Empty{}, internalError(err)
}

// GetLabel handles the corresponding API request.
func (s *RegistryServer) GetLabel(ctx context.Context, request *rpc.GetLabelRequest) (*rpc.Label, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
	label, err := models.NewLabelFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	log.Printf("looking for %s", label.ResourceName())
	k := &datastore.Key{Kind: models.LabelEntityName, Name: label.ResourceName()}
	err = client.Get(ctx, k, label)
	if err == datastore.ErrNoSuchEntity {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return label.Message()
}

// ListLabels handles the corresponding API request.
func (s *RegistryServer) ListLabels(ctx context.Context, req *rpc.ListLabelsRequest) (*rpc.ListLabelsResponse, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
	q := datastore.NewQuery(models.LabelEntityName)
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	p, err := models.NewLabelFromParentAndLabelID(req.GetParent(), "-")
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if p.ProjectID != "-" {
		q = q.Filter("ProjectID =", p.ProjectID)
	}
	if p.ProductID != "-" {
		q = q.Filter("ProductID =", p.ProductID)
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
			{"product_id", filterArgTypeString},
			{"version_id", filterArgTypeString},
			{"spec_id", filterArgTypeString},
			{"label_id", filterArgTypeString},
		})
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	var labelMessages []*rpc.Label
	var label models.Label
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&label); err == nil; _, err = it.Next(&label) {
		// don't allow wildcarded-names to be empty
		if p.SpecID == "-" && label.SpecID == "" {
			continue
		}
		if p.VersionID == "-" && label.VersionID == "" {
			continue
		}
		if p.ProductID == "-" && label.ProductID == "" {
			continue
		}
		if p.ProjectID == "-" && label.ProjectID == "" {
			continue
		}
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id": label.ProjectID,
				"product_id": label.ProductID,
				"version_id": label.VersionID,
				"spec_id":    label.SpecID,
				"label_id":   label.LabelID,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		labelMessage, _ := label.Message()
		labelMessages = append(labelMessages, labelMessage)
		if len(labelMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListLabelsResponse{
		Labels: labelMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(labelMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// UpdateLabel handles the corresponding API request.
func (s *RegistryServer) UpdateLabel(ctx context.Context, request *rpc.UpdateLabelRequest) (*rpc.Label, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
	label, err := models.NewLabelFromResourceName(request.GetLabel().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := &datastore.Key{Kind: models.LabelEntityName, Name: request.GetLabel().GetName()}
	err = client.Get(ctx, k, label)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = label.Update(request.GetLabel())
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, label)
	if err != nil {
		return nil, internalError(err)
	}
	s.notify(rpc.Notification_UPDATED, label.ResourceName())
	return label.Message()
}
