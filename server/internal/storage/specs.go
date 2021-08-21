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

package storage

import (
	"context"

	"github.com/apigee/registry/server/internal/storage/filtering"
	"github.com/apigee/registry/server/internal/storage/gorm"
	"github.com/apigee/registry/server/internal/storage/models"
	"github.com/apigee/registry/server/names"
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

func (d *Client) ListSpecs(ctx context.Context, parent names.Version, opts PageOptions) (SpecList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
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

	it := d.GetRecentSpecRevisions(ctx, token.Offset, parent.ProjectID, parent.ApiID, parent.VersionID)
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
			token.Offset++
			continue
		} else if len(response.Specs) == int(opts.Size) {
			break
		}

		response.Specs = append(response.Specs, *spec)
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
		"revision_id":          spec.RevisionID,
		"create_time":          spec.CreateTime,
		"revision_create_time": spec.RevisionCreateTime,
		"revision_update_time": spec.RevisionUpdateTime,
		"mime_type":            spec.MimeType,
		"size_bytes":           spec.SizeInBytes,
		"hash":                 spec.Hash,
		"source_uri":           spec.SourceURI,
		"labels":               labels,
	}, nil
}

func (d *Client) GetSpec(ctx context.Context, name names.Spec) (*models.Spec, error) {
	normal := name.Normal()
	q := d.NewQuery(gorm.SpecEntityName)
	q = q.Require("ProjectID", normal.ProjectID)
	q = q.Require("ApiID", normal.ApiID)
	q = q.Require("VersionID", normal.VersionID)
	q = q.Require("SpecID", normal.SpecID)
	q = q.Descending("RevisionCreateTime")

	it := d.Run(ctx, q)
	spec := &models.Spec{}
	if _, err := it.Next(spec); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return spec, nil
}

func (d *Client) DeleteSpec(ctx context.Context, name names.Spec) error {
	for _, entityName := range []string{
		gorm.SpecEntityName,
		gorm.SpecRevisionTagEntityName,
		gorm.ArtifactEntityName,
		gorm.BlobEntityName,
	} {
		q := d.NewQuery(entityName)
		q = q.Require("ProjectID", name.ProjectID)
		q = q.Require("ApiID", name.ApiID)
		q = q.Require("VersionID", name.VersionID)
		q = q.Require("SpecID", name.SpecID)
		if err := d.Delete(ctx, q); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

func (d *Client) GetSpecTags(ctx context.Context, name names.Spec) ([]*models.SpecRevisionTag, error) {
	q := d.NewQuery(gorm.SpecRevisionTagEntityName)
	q = q.Require("ProjectID", name.ProjectID)
	q = q.Require("ApiID", name.ApiID)
	q = q.Require("VersionID", name.VersionID)
	if name.SpecID != "-" {
		q = q.Require("SpecID", name.SpecID)
	}

	var (
		tags = make([]*models.SpecRevisionTag, 0)
		tag  = new(models.SpecRevisionTag)
		it   = d.Run(ctx, q)
		err  error
	)

	for _, err = it.Next(tag); err == nil; _, err = it.Next(tag) {
		tags = append(tags, tag)
	}
	if err != nil && err != iterator.Done {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return tags, nil
}
