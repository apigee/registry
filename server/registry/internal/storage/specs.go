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

	"github.com/apigee/registry/server/registry/internal/storage/gorm"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
)

func (d *Client) ListSpecs(ctx context.Context, parent names.Version, opts gorm.PageOptions) (gorm.SpecList, error) {
	return d.Client.ListSpecs(ctx, parent, opts)
}

func (d *Client) GetSpec(ctx context.Context, name names.Spec) (*models.Spec, error) {
	return d.Client.GetSpec(ctx, name)
}

func (d *Client) DeleteSpec(ctx context.Context, name names.Spec) error {
	return d.Client.DeleteSpec(ctx, name)
}

func (d *Client) GetSpecTags(ctx context.Context, name names.Spec) ([]*models.SpecRevisionTag, error) {
	return d.Client.GetSpecTags(ctx, name)
}
