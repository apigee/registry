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
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateVersion handles the corresponding API request.
func (s *RegistryServer) CreateVersion(ctx context.Context, request *rpc.CreateVersionRequest) (*rpc.Version, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer s.releaseStorageClient(client)
	version, err := models.NewVersionFromParentAndVersionID(request.GetParent(), request.GetVersionId())
	if err != nil {
		return nil, err
	}
	k := client.NewKey(models.VersionEntityName, version.ResourceName())
	// fail if version already exists
	existingVersion := &models.Version{}
	err = client.Get(ctx, k, existingVersion)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, version.ResourceName()+" already exists")
	}
	err = version.Update(request.GetVersion())
	version.CreateTime = version.UpdateTime
	k, err = client.Put(ctx, k, version)
	if err != nil {
		return nil, err
	}
	s.notify(rpc.Notification_CREATED, version.ResourceName())
	return version.Message()
}

// DeleteVersion handles the corresponding API request.
func (s *RegistryServer) DeleteVersion(ctx context.Context, request *rpc.DeleteVersionRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer s.releaseStorageClient(client)
	// Validate name and create dummy version (we just need the ID fields).
	version, err := models.NewVersionFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete children first and then delete the version.
	client.DeleteChildrenOfVersion(ctx, version)
	k := client.NewKey(models.VersionEntityName, request.GetName())
	err = client.Delete(ctx, k)
	s.notify(rpc.Notification_DELETED, request.GetName())
	return &empty.Empty{}, err
}

// GetVersion handles the corresponding API request.
func (s *RegistryServer) GetVersion(ctx context.Context, request *rpc.GetVersionRequest) (*rpc.Version, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer s.releaseStorageClient(client)
	version, err := models.NewVersionFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := client.NewKey(models.VersionEntityName, version.ResourceName())
	err = client.Get(ctx, k, version)
	if client.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return version.Message()
}

// ListVersions handles the corresponding API request.
func (s *RegistryServer) ListVersions(ctx context.Context, req *rpc.ListVersionsRequest) (*rpc.ListVersionsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer s.releaseStorageClient(client)
	q := client.NewQuery(models.VersionEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := names.ParseParentApi(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if m[1] != "-" {
		q = q.Require("ProjectID", m[1])
	}
	if m[2] != "-" {
		q = q.Require("ApiID", m[2])
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"project_id", filterArgTypeString},
			{"api_id", filterArgTypeString},
			{"version_id", filterArgTypeString},
			{"display_name", filterArgTypeString},
			{"description", filterArgTypeString},
			{"state", filterArgTypeString},
		})
	if err != nil {
		return nil, internalError(err)
	}
	var versionMessages []*rpc.Version
	var version models.Version
	it := client.Run(ctx, q)
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&version); err == nil; _, err = it.Next(&version) {
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id":   version.ProjectID,
				"api_id":       version.ApiID,
				"version_id":   version.VersionID,
				"display_name": version.DisplayName,
				"description":  version.Description,
				"state":        version.State,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		versionMessage, _ := version.Message()
		versionMessages = append(versionMessages, versionMessage)
		if len(versionMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListVersionsResponse{
		Versions: versionMessages,
	}
	responses.NextPageToken, err = it.GetCursor(len(versionMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// UpdateVersion handles the corresponding API request.
func (s *RegistryServer) UpdateVersion(ctx context.Context, request *rpc.UpdateVersionRequest) (*rpc.Version, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, err
	}
	defer s.releaseStorageClient(client)
	version, err := models.NewVersionFromResourceName(request.GetVersion().GetName())
	if err != nil {
		return nil, err
	}
	k := client.NewKey(models.VersionEntityName, version.ResourceName())
	err = client.Get(ctx, k, version)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = version.Update(request.GetVersion())
	if err != nil {
		return nil, err
	}
	k, err = client.Put(ctx, k, version)
	if err != nil {
		return nil, err
	}
	s.notify(rpc.Notification_UPDATED, version.ResourceName())
	return version.Message()
}
