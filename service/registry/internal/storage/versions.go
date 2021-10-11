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

package storage

import (
	"context"

	"github.com/apigee/registry/service/registry/internal/storage/filtering"
	"github.com/apigee/registry/service/registry/internal/storage/gorm"
	"github.com/apigee/registry/service/registry/internal/storage/models"
	"github.com/apigee/registry/service/registry/names"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VersionList contains a page of version resources.
type VersionList struct {
	Versions []models.Version
	Token    string
}

var versionFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "api_id", Type: filtering.String},
	{Name: "version_id", Type: filtering.String},
	{Name: "display_name", Type: filtering.String},
	{Name: "description", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "update_time", Type: filtering.Timestamp},
	{Name: "state", Type: filtering.String},
	{Name: "labels", Type: filtering.StringMap},
}

func (d *Client) ListVersions(ctx context.Context, parent names.Api, opts PageOptions) (VersionList, error) {
	q := d.NewQuery(gorm.VersionEntityName)

	token, err := decodeToken(opts.Token)
	if err != nil {
		return VersionList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return VersionList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	q = q.ApplyOffset(token.Offset)

	if parent.ProjectID != "-" {
		q = q.Require("ProjectID", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		q = q.Require("ApiID", parent.ApiID)
	}
	if parent.ProjectID != "-" && parent.ApiID != "-" {
		if _, err := d.GetApi(ctx, parent); err != nil {
			return VersionList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" {
		if _, err := d.GetProject(ctx, parent.Project()); err != nil {
			return VersionList{}, err
		}
	}

	filter, err := filtering.NewFilter(opts.Filter, versionFields)
	if err != nil {
		return VersionList{}, err
	}

	it := d.Run(ctx, q)
	response := VersionList{
		Versions: make([]models.Version, 0, opts.Size),
	}

	version := new(models.Version)
	for _, err = it.Next(version); err == nil; _, err = it.Next(version) {
		versionMap, err := versionMap(*version)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(versionMap)
		if err != nil {
			return response, err
		} else if !match {
			token.Offset++
			continue
		} else if len(response.Versions) == int(opts.Size) {
			break
		}

		response.Versions = append(response.Versions, *version)
		token.Offset++
	}
	if err != nil && err != iterator.Done {
		return response, status.Error(codes.Internal, err.Error())
	}

	if err == nil {
		response.Token, err = encodeToken(token)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

func versionMap(version models.Version) (map[string]interface{}, error) {
	labels, err := version.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":         version.Name(),
		"project_id":   version.ProjectID,
		"version_id":   version.VersionID,
		"display_name": version.DisplayName,
		"description":  version.Description,
		"create_time":  version.CreateTime,
		"update_time":  version.UpdateTime,
		"state":        version.State,
		"labels":       labels,
	}, nil
}

func (d *Client) GetVersion(ctx context.Context, name names.Version) (*models.Version, error) {
	version := new(models.Version)
	k := d.NewKey(gorm.VersionEntityName, name.String())
	if err := d.Get(ctx, k, version); d.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "api version %q not found in database", name)
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return version, nil
}

func (d *Client) SaveVersion(ctx context.Context, version *models.Version) error {
	k := d.NewKey(gorm.VersionEntityName, version.Name())
	if _, err := d.Put(ctx, k, version); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *Client) DeleteVersion(ctx context.Context, name names.Version) error {
	for _, entityName := range []string{
		gorm.VersionEntityName,
		gorm.SpecEntityName,
		gorm.SpecRevisionTagEntityName,
		gorm.ArtifactEntityName,
		gorm.BlobEntityName,
	} {
		q := d.NewQuery(entityName)
		q = q.Require("ProjectID", name.ProjectID)
		q = q.Require("ApiID", name.ApiID)
		q = q.Require("VersionID", name.VersionID)
		if err := d.Delete(ctx, q); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}
