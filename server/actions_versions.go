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
	"fmt"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/apigee/registry/server/storage"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
)

// CreateApiVersion handles the corresponding API request.
func (s *RegistryServer) CreateApiVersion(ctx context.Context, req *rpc.CreateApiVersionRequest) (*rpc.ApiVersion, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	parent, err := names.ParseApi(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Creation should only succeed when the parent exists.
	if _, err := getApi(ctx, client, parent); err != nil {
		return nil, err
	}

	name := parent.Version(req.GetApiVersionId())
	if name.VersionID == "" {
		name.VersionID = names.GenerateID()
	}

	if _, err := getVersion(ctx, client, name); err == nil {
		return nil, alreadyExistsError(fmt.Errorf("API version %q already exists", name))
	} else if !isNotFound(err) {
		return nil, err
	}

	if err := name.Validate(); err != nil {
		return nil, invalidArgumentError(err)
	}

	version, err := models.NewVersion(name, req.GetApiVersion())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	if err := saveVersion(ctx, client, version); err != nil {
		return nil, err
	}

	message, err := version.Message(rpc.View_BASIC)
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_CREATED, name.String())
	return message, nil
}

// DeleteApiVersion handles the corresponding API request.
func (s *RegistryServer) DeleteApiVersion(ctx context.Context, req *rpc.DeleteApiVersionRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	name, err := names.ParseVersion(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// Deletion should only succeed on API versions that currently exist.
	if _, err := getVersion(ctx, client, name); err != nil {
		return nil, err
	}

	if err := deleteVersion(ctx, client, name); err != nil {
		return nil, err
	}

	s.notify(rpc.Notification_DELETED, name.String())
	return &empty.Empty{}, nil
}

// GetApiVersion handles the corresponding API request.
func (s *RegistryServer) GetApiVersion(ctx context.Context, req *rpc.GetApiVersionRequest) (*rpc.ApiVersion, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	name, err := names.ParseVersion(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	version, err := getVersion(ctx, client, name)
	if err != nil {
		return nil, err
	}

	message, err := version.Message(req.GetView())
	if err != nil {
		return nil, internalError(err)
	}

	return message, nil
}

// ListApiVersions handles the corresponding API request.
func (s *RegistryServer) ListApiVersions(ctx context.Context, req *rpc.ListApiVersionsRequest) (*rpc.ListApiVersionsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetPageSize() < 0 {
		return nil, invalidArgumentError(fmt.Errorf("invalid page_size %q: must not be negative", req.GetPageSize()))
	}

	q := client.NewQuery(storage.VersionEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	parent, err := names.ParseApi(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if parent.ProjectID != "-" {
		q = q.Require("ProjectID", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		q = q.Require("ApiID", parent.ApiID)
	}
	if parent.ProjectID != "-" && parent.ApiID != "-" {
		if _, err := getApi(ctx, client, parent); err != nil {
			return nil, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" {
		if _, err := getProject(ctx, client, parent.Project()); err != nil {
			return nil, err
		}
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"name", filterArgTypeString},
			{"project_id", filterArgTypeString},
			{"api_id", filterArgTypeString},
			{"version_id", filterArgTypeString},
			{"display_name", filterArgTypeString},
			{"description", filterArgTypeString},
			{"create_time", filterArgTypeTimestamp},
			{"update_time", filterArgTypeTimestamp},
			{"state", filterArgTypeString},
			{"labels", filterArgTypeStringMap},
		})
	if err != nil {
		return nil, internalError(err)
	}
	var versionMessages []*rpc.ApiVersion
	var version models.Version
	it := client.Run(ctx, q)
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&version); err == nil; _, err = it.Next(&version) {
		if prg != nil {
			filterInputs := map[string]interface{}{
				"name":         version.Name(),
				"project_id":   version.ProjectID,
				"api_id":       version.ApiID,
				"version_id":   version.VersionID,
				"display_name": version.DisplayName,
				"description":  version.Description,
				"create_time":  version.CreateTime,
				"update_time":  version.UpdateTime,
				"state":        version.State,
			}
			filterInputs["labels"], err = version.LabelsMap()
			if err != nil {
				return nil, internalError(err)
			}
			out, _, err := prg.Eval(filterInputs)
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		versionMessage, _ := version.Message(req.GetView())
		versionMessages = append(versionMessages, versionMessage)
		if len(versionMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListApiVersionsResponse{
		ApiVersions: versionMessages,
	}
	responses.NextPageToken, err = it.GetCursor(len(versionMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// UpdateApiVersion handles the corresponding API request.
func (s *RegistryServer) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetApiVersion() == nil {
		return nil, invalidArgumentError(fmt.Errorf("invalid api_version %+v: body must be provided", req.GetApiVersion()))
	}

	name, err := names.ParseVersion(req.ApiVersion.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	version, err := getVersion(ctx, client, name)
	if err != nil {
		return nil, err
	}

	if err := version.Update(req.GetApiVersion(), req.GetUpdateMask()); err != nil {
		return nil, internalError(err)
	}

	if err := saveVersion(ctx, client, version); err != nil {
		return nil, internalError(err)
	}

	message, err := version.Message(rpc.View_BASIC)
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_UPDATED, name.String())
	return message, nil
}

func saveVersion(ctx context.Context, client storage.Client, version *models.Version) error {
	k := client.NewKey(storage.VersionEntityName, version.Name())
	if _, err := client.Put(ctx, k, version); err != nil {
		return internalError(err)
	}

	return nil
}

func getVersion(ctx context.Context, client storage.Client, name names.Version) (*models.Version, error) {
	version := new(models.Version)
	k := client.NewKey(storage.VersionEntityName, name.String())
	if err := client.Get(ctx, k, version); client.IsNotFound(err) {
		return nil, notFoundError(fmt.Errorf("api version %q not found", name))
	} else if err != nil {
		return nil, internalError(err)
	}

	return version, nil
}

func deleteVersion(ctx context.Context, client storage.Client, name names.Version) error {
	if err := client.DeleteChildrenOfVersion(ctx, name); err != nil {
		return internalError(err)
	}

	k := client.NewKey(storage.VersionEntityName, name.String())
	if err := client.Delete(ctx, k); err != nil {
		return internalError(err)
	}

	return nil
}
