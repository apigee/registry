// Copyright 2022 Google LLC. All Rights Reserved.
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

	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/gorm/clause"
)

// SaveProject will return codes.NotFound if key not found
func (c *Client) SaveProject(ctx context.Context, v *models.Project, fieldMask *fieldmaskpb.FieldMask) error {
	v.Key = v.Name()
	return c.saveWithMask(ctx, v, fieldMask)
}

// SaveApi will return codes.NotFound if key not found
func (c *Client) SaveApi(ctx context.Context, v *models.Api, fieldMask *fieldmaskpb.FieldMask) error {
	v.Key = v.Name()
	return c.saveWithMask(ctx, v, fieldMask)
}

// SaveVersion will return codes.NotFound if key not found
func (c *Client) SaveVersion(ctx context.Context, v *models.Version, fieldMask *fieldmaskpb.FieldMask) error {
	v.Key = v.Name()
	return c.saveWithMask(ctx, v, fieldMask)
}

// SaveSpecRevision will upsert if key not found
func (c *Client) SaveSpecRevision(ctx context.Context, v *models.Spec) error {
	v.Key = v.RevisionName()
	return c.save(ctx, v)
}

// SaveSpecRevisionContents will create a new blob for the spec
func (c *Client) SaveSpecRevisionContents(ctx context.Context, spec *models.Spec, contents []byte) error {
	v := models.NewBlobForSpec(spec, contents)
	v.Key = spec.RevisionName()
	return c.save(ctx, v)
}

// SaveSpecRevisionTag will upsert if key not found
func (c *Client) SaveSpecRevisionTag(ctx context.Context, v *models.SpecRevisionTag) error {
	v.Key = v.String()
	return c.save(ctx, v)
}

// SaveDeploymentRevision will upsert if key not found
func (c *Client) SaveDeploymentRevision(ctx context.Context, v *models.Deployment) error {
	v.Key = v.RevisionName()
	return c.save(ctx, v)
}

// SaveDeploymentRevisionTag will upsert if key not found
func (c *Client) SaveDeploymentRevisionTag(ctx context.Context, v *models.DeploymentRevisionTag) error {
	v.Key = v.String()
	return c.save(ctx, v)
}

// SaveArtifact will upsert if key not found
func (c *Client) SaveArtifact(ctx context.Context, v *models.Artifact) error {
	v.Key = v.Name()
	return c.save(ctx, v)
}

// SaveSpecRevisionContents will create a new blob for the artifact
func (c *Client) SaveArtifactContents(ctx context.Context, artifact *models.Artifact, contents []byte) error {
	v := models.NewBlobForArtifact(artifact, contents)
	v.Key = artifact.Name()
	return c.save(ctx, v)
}

func (c *Client) save(ctx context.Context, v interface{}) error {
	// Insert or update, see: https://gorm.io/docs/create.html#Upsert-x2F-On-Conflict
	err := c.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(v).Error
	return grpcErrorForDBError(ctx, errors.Wrapf(err, "save %#v", v))
}

func (c *Client) saveWithMask(ctx context.Context, v interface{}, fieldMask *fieldmaskpb.FieldMask) error {
	op := c.db.WithContext(ctx).
		Select(fieldMask.GetPaths()).
		Clauses(clause.Returning{})
	err := op.Save(v).Error
	if err == nil && op.RowsAffected == 0 {
		err = status.Errorf(codes.NotFound, "%s not found in database", v)
	}
	return grpcErrorForDBError(ctx, err)
}
