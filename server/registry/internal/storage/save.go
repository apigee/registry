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
	"gorm.io/gorm/clause"
)

func (c *Client) SaveProject(ctx context.Context, v *models.Project) error {
	v.Key = v.Name()
	return c.save(ctx, v)
}

func (c *Client) SaveApi(ctx context.Context, v *models.Api) error {
	v.Key = v.Name()
	return c.save(ctx, v)
}

func (c *Client) SaveVersion(ctx context.Context, v *models.Version) error {
	v.Key = v.Name()
	return c.save(ctx, v)
}

func (c *Client) SaveSpecRevision(ctx context.Context, v *models.Spec) error {
	v.Key = v.RevisionName()
	return c.save(ctx, v)
}

func (c *Client) SaveSpecRevisionContents(ctx context.Context, spec *models.Spec, contents []byte) error {
	v := models.NewBlobForSpec(spec, contents)
	v.Key = spec.RevisionName()
	return c.save(ctx, v)
}

func (c *Client) SaveSpecRevisionTag(ctx context.Context, v *models.SpecRevisionTag) error {
	v.Key = v.String()
	return c.save(ctx, v)
}

func (c *Client) SaveDeploymentRevision(ctx context.Context, v *models.Deployment) error {
	v.Key = v.RevisionName()
	return c.save(ctx, v)
}

func (c *Client) SaveDeploymentRevisionTag(ctx context.Context, v *models.DeploymentRevisionTag) error {
	v.Key = v.String()
	return c.save(ctx, v)
}

func (c *Client) SaveArtifact(ctx context.Context, v *models.Artifact) error {
	v.Key = v.Name()
	return c.save(ctx, v)
}

func (c *Client) SaveArtifactContents(ctx context.Context, artifact *models.Artifact, contents []byte) error {
	v := models.NewBlobForArtifact(artifact, contents)
	v.Key = artifact.Name()
	return c.save(ctx, v)
}

func (c *Client) save(ctx context.Context, v interface{}) error {
	// update if exists or create as needed
	err := c.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(v).Error
	return grpcErrorForDBError(ctx, err)
}
