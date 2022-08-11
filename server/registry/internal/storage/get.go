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
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (c *Client) GetProject(ctx context.Context, name names.Project) (*models.Project, error) {
	v := new(models.Project)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}

func (c *Client) GetApi(ctx context.Context, name names.Api) (*models.Api, error) {
	v := new(models.Api)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}

func (c *Client) GetVersion(ctx context.Context, name names.Version) (*models.Version, error) {
	v := new(models.Version)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}

func (c *Client) GetSpec(ctx context.Context, name names.Spec) (*models.Spec, error) {
	name = name.Normal()
	op := c.db.WithContext(ctx).
		Where("project_id = ?", name.ProjectID).
		Where("api_id = ?", name.ApiID).
		Where("version_id = ?", name.VersionID).
		Where("spec_id = ?", name.SpecID).
		Order("revision_create_time desc")

	v := new(models.Spec)
	if err := op.First(v).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}

func (c *Client) GetSpecRevision(ctx context.Context, name names.SpecRevision) (*models.Spec, error) {
	name, err := c.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return nil, err
	}

	v := new(models.Spec)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}

func (c *Client) GetSpecRevisionContents(ctx context.Context, name names.SpecRevision) (*models.Blob, error) {
	name, err := c.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return nil, err
	}

	v := new(models.Blob)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}

func (c *Client) GetDeployment(ctx context.Context, name names.Deployment) (*models.Deployment, error) {
	name = name.Normal()
	op := c.db.WithContext(ctx).
		Where("project_id = ?", name.ProjectID).
		Where("api_id = ?", name.ApiID).
		Where("deployment_id = ?", name.DeploymentID).
		Order("revision_create_time desc")

	v := new(models.Deployment)
	if err := op.First(v).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}

func (c *Client) GetDeploymentRevision(ctx context.Context, name names.DeploymentRevision) (*models.Deployment, error) {
	name, err := c.unwrapDeploymentRevisionTag(ctx, name)
	if err != nil {
		return nil, err
	}

	v := new(models.Deployment)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}

func (c *Client) GetArtifact(ctx context.Context, name names.Artifact) (*models.Artifact, error) {
	v := new(models.Artifact)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}

func (c *Client) GetArtifactContents(ctx context.Context, name names.Artifact) (*models.Blob, error) {
	v := new(models.Blob)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(err)
	}

	return v, nil
}
