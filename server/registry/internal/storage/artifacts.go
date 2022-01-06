// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Artifact 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the artifactific language governing permissions and
// limitations under the License.

package storage

import (
	"context"

	"github.com/apigee/registry/server/registry/internal/storage/gorm"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
)

func (d *Client) ListSpecArtifacts(ctx context.Context, parent names.Spec, opts gorm.PageOptions) (gorm.ArtifactList, error) {
	return d.Client.ListSpecArtifacts(ctx, parent, opts)
}

func (d *Client) ListVersionArtifacts(ctx context.Context, parent names.Version, opts gorm.PageOptions) (gorm.ArtifactList, error) {
	return d.Client.ListVersionArtifacts(ctx, parent, opts)
}

func (d *Client) ListDeploymentArtifacts(ctx context.Context, parent names.Deployment, opts gorm.PageOptions) (gorm.ArtifactList, error) {
	return d.Client.ListDeploymentArtifacts(ctx, parent, opts)
}

func (d *Client) ListApiArtifacts(ctx context.Context, parent names.Api, opts gorm.PageOptions) (gorm.ArtifactList, error) {
	return d.Client.ListApiArtifacts(ctx, parent, opts)
}

func (d *Client) ListProjectArtifacts(ctx context.Context, parent names.Project, opts gorm.PageOptions) (gorm.ArtifactList, error) {
	return d.Client.ListProjectArtifacts(ctx, parent, opts)
}

func (d *Client) SaveArtifact(ctx context.Context, artifact *models.Artifact) error {
	return d.Client.SaveArtifact(ctx, artifact)
}

func (d *Client) SaveArtifactContents(ctx context.Context, artifact *models.Artifact, contents []byte) error {
	return d.Client.SaveArtifactContents(ctx, artifact, contents)
}

func (d *Client) GetArtifact(ctx context.Context, name names.Artifact) (*models.Artifact, error) {
	return d.Client.GetArtifact(ctx, name)
}

func (d *Client) GetArtifactContents(ctx context.Context, name names.Artifact) (*models.Blob, error) {
	return d.Client.GetArtifactContents(ctx, name)
}

func (d *Client) DeleteArtifact(ctx context.Context, name names.Artifact) error {
	return d.Client.DeleteArtifact(ctx, name)
}
