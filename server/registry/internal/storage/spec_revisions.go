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
)

func (d *Client) ListSpecRevisions(ctx context.Context, parent names.Spec, opts gorm.PageOptions) (gorm.SpecList, error) {
	return d.Client.ListSpecRevisions(ctx, parent, opts)
}

func (d *Client) SaveSpecRevision(ctx context.Context, revision *models.Spec) error {
	return d.Client.SaveSpecRevision(ctx, revision)
}

func (d *Client) SaveSpecRevisionContents(ctx context.Context, spec *models.Spec, contents []byte) error {
	return d.Client.SaveSpecRevisionContents(ctx, spec, contents)
}

func (d *Client) GetSpecRevision(ctx context.Context, name names.SpecRevision) (*models.Spec, error) {
	return d.Client.GetSpecRevision(ctx, name)
}

func (d *Client) GetSpecRevisionContents(ctx context.Context, name names.SpecRevision) (*models.Blob, error) {
	return d.Client.GetSpecRevisionContents(ctx, name)
}

func (d *Client) DeleteSpecRevision(ctx context.Context, name names.SpecRevision) error {
	return d.Client.DeleteSpecRevision(ctx, name)
}

func (d *Client) SaveSpecRevisionTag(ctx context.Context, tag *models.SpecRevisionTag) error {
	return d.Client.SaveSpecRevisionTag(ctx, tag)
}
