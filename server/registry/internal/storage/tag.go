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

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (c *Client) unwrapSpecRevisionTag(ctx context.Context, name names.SpecRevision) (names.SpecRevision, error) {
	v := new(models.SpecRevisionTag)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return name, nil
	} else if err != nil {
		return names.SpecRevision{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
	}

	return name.Spec().Revision(v.RevisionID), nil
}

func (c *Client) unwrapDeploymentRevisionTag(ctx context.Context, name names.DeploymentRevision) (names.DeploymentRevision, error) {
	v := new(models.DeploymentRevisionTag)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return name, nil
	} else if err != nil {
		return names.DeploymentRevision{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
	}

	return name.Deployment().Revision(v.RevisionID), nil
}
