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
// See the License for the revisionific language governing permissions and
// limitations under the License.

package storage

import (
	"context"

	"github.com/apigee/registry/server/registry/internal/storage/gorm"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *Client) ListSpecRevisions(ctx context.Context, parent names.Spec, opts PageOptions) (SpecList, error) {
	q := d.NewQuery(gorm.SpecEntityName)
	q = q.Require("ProjectID", parent.ProjectID)
	q = q.Require("ApiID", parent.ApiID)
	q = q.Require("VersionID", parent.VersionID)
	q = q.Require("SpecID", parent.SpecID)
	q = q.Descending("RevisionCreateTime")

	token, err := decodeToken(opts.Token)
	if err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	}

	q = q.ApplyOffset(token.Offset)

	it := d.Run(ctx, q)
	response := SpecList{
		Specs: make([]models.Spec, 0, opts.Size),
	}

	revision := new(models.Spec)
	for _, err = it.Next(revision); err == nil; _, err = it.Next(revision) {
		token.Offset++

		response.Specs = append(response.Specs, *revision)
		if len(response.Specs) == int(opts.Size) {
			break
		}
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

func (d *Client) SaveSpecRevision(ctx context.Context, revision *models.Spec) error {
	return d.Client.SaveSpecRevision(ctx, revision)
}

func (d *Client) SaveSpecRevisionContents(ctx context.Context, spec *models.Spec, contents []byte) error {
	return d.Client.SaveSpecRevisionContents(ctx, spec, contents)
}

func (d *Client) GetSpecRevision(ctx context.Context, name names.SpecRevision) (*models.Spec, error) {
	name, err := d.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return nil, err
	}

	return d.Client.GetSpecRevision(ctx, name)
}

func (d *Client) GetSpecRevisionContents(ctx context.Context, name names.SpecRevision) (*models.Blob, error) {
	name, err := d.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return nil, err
	}

	return d.Client.GetSpecRevisionContents(ctx, name)
}

func (d *Client) DeleteSpecRevision(ctx context.Context, name names.SpecRevision) error {
	name, err := d.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return err
	}

	return d.Client.DeleteSpecRevision(ctx, name)
}

func (d *Client) SaveSpecRevisionTag(ctx context.Context, tag *models.SpecRevisionTag) error {
	return d.Client.SaveSpecRevisionTag(ctx, tag)
}

func (d *Client) unwrapSpecRevisionTag(ctx context.Context, name names.SpecRevision) (names.SpecRevision, error) {
	tag := new(models.SpecRevisionTag)
	if err := d.Get(ctx, d.NewKey(gorm.SpecRevisionTagEntityName, name.String()), tag); d.IsNotFound(err) {
		return name, nil
	} else if err != nil {
		return names.SpecRevision{}, status.Error(codes.Internal, err.Error())
	}

	return name.Spec().Revision(tag.RevisionID), nil
}
