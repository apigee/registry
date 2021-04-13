// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Spec 2.0 (the "License");
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

package dao

import (
	"context"

	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/apigee/registry/server/storage"
	"github.com/apigee/registry/server/storage/filtering"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SpecList contains a page of spec resources.
type SpecList struct {
	Specs []models.Spec
	Token string
}

var specFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "api_id", Type: filtering.String},
	{Name: "version_id", Type: filtering.String},
	{Name: "spec_id", Type: filtering.String},
	{Name: "filename", Type: filtering.String},
	{Name: "description", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "revision_create_time", Type: filtering.Timestamp},
	{Name: "revision_update_time", Type: filtering.Timestamp},
	{Name: "mime_type", Type: filtering.String},
	{Name: "size_bytes", Type: filtering.Int},
	{Name: "source_uri", Type: filtering.String},
	{Name: "labels", Type: filtering.StringMap},
}

func (d *DAO) ListSpecs(ctx context.Context, parent names.Version, opts PageOptions) (SpecList, error) {
	q := d.NewQuery(storage.SpecEntityName)
	q = q.Require("Currency", models.IsCurrent)
	q, err := q.ApplyCursor(opts.Token)
	if err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if id := parent.ProjectID; id != "-" {
		q = q.Require("ProjectID", id)
	}
	if id := parent.ApiID; id != "-" {
		q = q.Require("ApiID", id)
	}
	if id := parent.VersionID; id != "-" {
		q = q.Require("VersionID", id)
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" {
		if _, err := d.GetVersion(ctx, parent); err != nil {
			return SpecList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID == "-" {
		if _, err := d.GetApi(ctx, parent.Api()); err != nil {
			return SpecList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.VersionID == "-" {
		if _, err := d.GetProject(ctx, parent.Project()); err != nil {
			return SpecList{}, err
		}
	}

	filter, err := filtering.NewFilter(opts.Filter, specFields)
	if err != nil {
		return SpecList{}, err
	}

	it := d.Run(ctx, q)
	response := SpecList{
		Specs: make([]models.Spec, 0, opts.Size),
	}

	spec := new(models.Spec)
	for _, err = it.Next(spec); err == nil; _, err = it.Next(spec) {
		specMap, err := specMap(*spec)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(specMap)
		if err != nil {
			return response, err
		} else if !match {
			continue
		}

		response.Specs = append(response.Specs, *spec)
		if len(response.Specs) == int(opts.Size) {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return response, status.Error(codes.Internal, err.Error())
	}

	if err == nil {
		response.Token, err = it.GetCursor()
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

func specMap(spec models.Spec) (map[string]interface{}, error) {
	labels, err := spec.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":                 spec.Name(),
		"project_id":           spec.ProjectID,
		"api_id":               spec.ApiID,
		"version_id":           spec.VersionID,
		"spec_id":              spec.SpecID,
		"filename":             spec.FileName,
		"description":          spec.Description,
		"create_time":          spec.CreateTime,
		"revision_create_time": spec.RevisionCreateTime,
		"revision_update_time": spec.RevisionUpdateTime,
		"mime_type":            spec.MimeType,
		"size_bytes":           spec.SizeInBytes,
		"source_uri":           spec.SourceURI,
		"labels":               labels,
	}, nil
}

func (d *DAO) GetSpec(ctx context.Context, name names.Spec) (*models.Spec, error) {
	q := d.NewQuery(storage.SpecEntityName)
	q = q.Require("ProjectID", name.ProjectID)
	q = q.Require("ApiID", name.ApiID)
	q = q.Require("VersionID", name.VersionID)
	q = q.Require("SpecID", name.SpecID)
	q = q.Require("Currency", models.IsCurrent)

	it := d.Run(ctx, q)
	spec := &models.Spec{}
	if _, err := it.Next(spec); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return spec, nil
}

func (d *DAO) SaveSpec(ctx context.Context, spec *models.Spec) error {
	k := d.NewKey(storage.SpecEntityName, spec.Name())
	if _, err := d.Put(ctx, k, spec); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *DAO) DeleteSpec(ctx context.Context, name names.Spec) error {
	if err := d.DeleteChildrenOfSpec(ctx, name); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	k := d.NewKey(storage.SpecEntityName, name.String())
	if err := d.Delete(ctx, k); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	q := d.NewQuery(storage.SpecEntityName)
	q = q.Require("ProjectID", name.ProjectID)
	q = q.Require("ApiID", name.ApiID)
	q = q.Require("VersionID", name.VersionID)
	q = q.Require("SpecID", name.SpecID)
	if err := d.DeleteAllMatches(ctx, q); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}
