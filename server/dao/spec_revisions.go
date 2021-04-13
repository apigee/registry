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
// See the License for the revisionific language governing permissions and
// limitations under the License.

package dao

import (
	"context"

	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/apigee/registry/server/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *DAO) ListSpecRevisions(ctx context.Context, parent names.Spec, opts PageOptions) (SpecList, error) {
	q := d.NewQuery(storage.SpecEntityName)
	q = q.Require("ProjectID", parent.ProjectID)
	q = q.Require("ApiID", parent.ApiID)
	q = q.Require("VersionID", parent.VersionID)
	q = q.Require("SpecID", parent.SpecID)
	q = q.Order("-RevisionCreateTime")
	q, err := q.ApplyCursor(opts.Token)
	if err != nil {
		return SpecList{}, status.Error(codes.Internal, err.Error())
	}

	it := d.Run(ctx, q)
	response := SpecList{
		Specs: make([]models.Spec, 0, opts.Size),
	}

	revision := new(models.Spec)
	for _, err = it.Next(revision); err == nil; _, err = it.Next(revision) {
		response.Specs = append(response.Specs, *revision)
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

func (d *DAO) SaveSpecRevision(ctx context.Context, revision *models.Spec) error {
	k := d.NewKey(storage.SpecEntityName, revision.RevisionName())
	if _, err := d.Put(ctx, k, revision); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *DAO) SaveSpecRevisionContents(ctx context.Context, spec *models.Spec, contents []byte) error {
	blob := models.NewBlobForSpec(spec, contents)
	k := d.NewKey(models.BlobEntityName, spec.RevisionName())
	if _, err := d.Put(ctx, k, blob); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *DAO) GetSpecRevision(ctx context.Context, name names.SpecRevision) (*models.Spec, error) {
	name, err := d.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return nil, err
	}

	spec := new(models.Spec)
	k := d.NewKey(storage.SpecEntityName, name.String())
	if err := d.Get(ctx, k, spec); d.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "spec revision %q not found", name)
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return spec, nil
}

func (d *DAO) GetSpecRevisionContents(ctx context.Context, name names.SpecRevision) (*models.Blob, error) {
	name, err := d.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return nil, err
	}

	blob := new(models.Blob)
	k := d.NewKey(models.BlobEntityName, name.String())
	if err := d.Get(ctx, k, blob); d.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "spec revision contents %q not found", name)
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return blob, nil
}

func (d *DAO) DeleteSpecRevision(ctx context.Context, name names.SpecRevision) error {
	name, err := d.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return err
	}

	k := d.NewKey(models.BlobEntityName, name.String())
	if err := d.Delete(ctx, k); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	k = d.NewKey(storage.SpecEntityName, name.String())
	if err := d.Delete(ctx, k); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *DAO) SaveSpecRevisionTag(ctx context.Context, tag *models.SpecRevisionTag) error {
	k := d.NewKey(storage.SpecRevisionTagEntityName, tag.String())
	if _, err := d.Put(ctx, k, tag); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *DAO) unwrapSpecRevisionTag(ctx context.Context, name names.SpecRevision) (names.SpecRevision, error) {
	tag := new(models.SpecRevisionTag)
	if err := d.Get(ctx, d.NewKey(storage.SpecRevisionTagEntityName, name.String()), tag); d.IsNotFound(err) {
		return name, nil
	} else if err != nil {
		return names.SpecRevision{}, status.Error(codes.Internal, err.Error())
	}

	return name.Spec().Revision(tag.RevisionID), nil
}
