// Copyright 2022 Google LLC.
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
)

func (c *Client) CreateProject(ctx context.Context, v *models.Project) error {
	v.Key = v.Name()
	return c.create(ctx, v)
}

func (c *Client) CreateApi(ctx context.Context, v *models.Api) error {
	v.Key = v.Name()
	return c.create(ctx, v)
}

func (c *Client) CreateVersion(ctx context.Context, v *models.Version) error {
	v.Key = v.Name()
	return c.create(ctx, v)
}

func (c *Client) CreateArtifact(ctx context.Context, v *models.Artifact) error {
	v.Key = v.Name()
	return c.create(ctx, v)
}

func (c *Client) CreateDeploymentRevision(ctx context.Context, v *models.Deployment) error {
	v.Key = v.RevisionName()
	return c.create(ctx, v)
}

func (c *Client) CreateSpecRevision(ctx context.Context, v *models.Spec) error {
	v.Key = v.RevisionName()
	return c.create(ctx, v)
}

func (c *Client) create(ctx context.Context, v interface{}) error {
	return grpcErrorForDBError(ctx,
		errors.Wrapf(c.db.WithContext(ctx).Create(v).Error, "create %#v", v))
}
