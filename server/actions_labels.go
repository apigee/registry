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
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateLabel handles the corresponding API request.
func (s *RegistryServer) CreateLabel(ctx context.Context, request *rpc.CreateLabelRequest) (*rpc.Label, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer s.releaseStorageClient(client)
	label, err := models.NewLabelFromParentAndLabelID(request.GetParent(), request.GetLabelId())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	err = label.Update(request.GetLabel())
	k := client.NewKey(models.LabelEntityName, label.ResourceName())
	// fail if label already exists
	existingLabel := &models.Label{}
	err = client.Get(ctx, k, existingLabel)
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
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer s.releaseStorageClient(client)
	// Validate name and create dummy label (we just need the ID fields).
	_, err = models.NewLabelFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete the label.
	k := client.NewKey(models.LabelEntityName, request.GetName())
	err = client.Delete(ctx, k)
	s.notify(rpc.Notification_DELETED, request.GetName())
	return &empty.Empty{}, internalError(err)
}

// GetLabel handles the corresponding API request.
func (s *RegistryServer) GetLabel(ctx context.Context, request *rpc.GetLabelRequest) (*rpc.Label, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer s.releaseStorageClient(client)
	label, err := models.NewLabelFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	log.Printf("looking for %s", label.ResourceName())
	k := client.NewKey(models.LabelEntityName, label.ResourceName())
	err = client.Get(ctx, k, label)
	if client.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return label.Message()
}

// ListLabels handles the corresponding API request.
func (s *RegistryServer) ListLabels(ctx context.Context, req *rpc.ListLabelsRequest) (*rpc.ListLabelsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer s.releaseStorageClient(client)
	q := client.NewQuery(models.LabelEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
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
			{"label_id", filterArgTypeString},
		})
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	var labelMessages []*rpc.Label
	var label models.Label
	it := client.Run(ctx, q)
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&label); err == nil; _, err = it.Next(&label) {
		// don't allow wildcarded-names to be empty
		if p.SpecID == "-" && label.SpecID == "" {
			continue
		}
		if p.VersionID == "-" && label.VersionID == "" {
			continue
		}
		if p.ApiID == "-" && label.ApiID == "" {
			continue
		}
		if p.ProjectID == "-" && label.ProjectID == "" {
			continue
		}
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id": label.ProjectID,
				"api_id":     label.ApiID,
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
	responses.NextPageToken, err = it.GetCursor(len(labelMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// UpdateLabel handles the corresponding API request.
func (s *RegistryServer) UpdateLabel(ctx context.Context, request *rpc.UpdateLabelRequest) (*rpc.Label, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer s.releaseStorageClient(client)
	label, err := models.NewLabelFromResourceName(request.GetLabel().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.LabelEntityName, request.GetLabel().GetName())
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
