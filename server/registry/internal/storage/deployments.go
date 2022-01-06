// Copyright 2021 Google LLC. All Rights Reserved.
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

func (d *Client) ListDeployments(ctx context.Context, parent names.Api, opts gorm.PageOptions) (gorm.DeploymentList, error) {
	return d.Client.ListDeployments(ctx, parent, opts)
}

func (d *Client) GetDeployment(ctx context.Context, name names.Deployment) (*models.Deployment, error) {
	return d.Client.GetDeployment(ctx, name)
}

func (d *Client) DeleteDeployment(ctx context.Context, name names.Deployment) error {
	return d.Client.DeleteDeployment(ctx, name)
}

func (d *Client) GetDeploymentTags(ctx context.Context, name names.Deployment) ([]*models.DeploymentRevisionTag, error) {
	return d.Client.GetDeploymentTags(ctx, name)
}
