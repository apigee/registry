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
	"strings"

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (c *Client) GetProject(ctx context.Context, name names.Project) (*models.Project, error) {
	v := new(models.Project)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
	}

	return v, nil
}

func (c *Client) GetApi(ctx context.Context, name names.Api) (*models.Api, error) {
	v := new(models.Api)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
	}

	return v, nil
}

func (c *Client) GetVersion(ctx context.Context, name names.Version) (*models.Version, error) {
	v := new(models.Version)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
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
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
	}

	return v, nil
}

func (c *Client) GetSpecRevision(ctx context.Context, name names.SpecRevision) (*models.Spec, error) {
	name, err := c.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return nil, err
	}

	if name.RevisionID == "" { // get latest revision
		return c.GetSpec(ctx, name.Spec())
	}

	v := new(models.Spec)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
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
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
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
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
	}

	return v, nil
}

func (c *Client) GetDeploymentRevision(ctx context.Context, name names.DeploymentRevision) (*models.Deployment, error) {
	name, err := c.unwrapDeploymentRevisionTag(ctx, name)
	if err != nil {
		return nil, err
	}

	if name.RevisionID == "" { // get latest revision
		return c.GetDeployment(ctx, name.Deployment())
	}
	v := new(models.Deployment)
	if err := c.db.WithContext(ctx).Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
	}

	return v, nil
}

func (c *Client) GetArtifact(ctx context.Context, name names.Artifact, forUpdate bool) (*models.Artifact, error) {
	v := new(models.Artifact)
	op := c.db.WithContext(ctx)
	if forUpdate {
		op.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	// if fully specified, just grab via key
	if !((name.SpecID() != "" || name.DeploymentID() != "") && name.RevisionID() == "") {
		if err := op.Take(v, "key = ?", name.String()).Error; err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
		} else if err != nil {
			return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
		}

		return v, nil
	}

	// otherwise, retrieve by latest revision
	op = op.Model(&models.Artifact{}).
		Where("artifacts.project_id = ?", name.ProjectID()).
		Where("artifacts.api_id = ?", name.ApiID()).
		Where("artifacts.version_id = ?", name.VersionID()).
		Where("artifacts.artifact_id = ?", name.ArtifactID())

	if name.SpecID() != "" {
		op = op.Where("artifacts.spec_id = ?", name.SpecID()).
			Joins(`join (?) latest
				ON artifacts.project_id = latest.project_id
				AND artifacts.api_id = latest.api_id
				AND artifacts.version_id = latest.version_id
				AND artifacts.spec_id = latest.spec_id
				AND artifacts.revision_id = latest.revision_id`, c.latestSpecRevisionsQuery(ctx))
	} else if name.DeploymentID() != "" {
		op = op.Where("artifacts.deployment_id = ?", name.DeploymentID()).
			Joins(`join (?) latest
				ON artifacts.project_id = latest.project_id
				AND artifacts.api_id = latest.api_id
				AND artifacts.deployment_id = latest.deployment_id
				AND artifacts.revision_id = latest.revision_id`, c.latestDeploymentRevisionsQuery(ctx))
	}

	if err := op.Take(v).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
	}

	return v, nil
}

func (c *Client) GetArtifactContents(ctx context.Context, name names.Artifact) (*models.Blob, error) {
	v := new(models.Blob)
	op := c.db.WithContext(ctx)

	op = op.Model(&models.Blob{}).
		Where("blobs.project_id = ?", name.ProjectID()).
		Where("blobs.api_id = ?", name.ApiID()).
		Where("blobs.version_id = ?", name.VersionID()).
		Where("blobs.spec_id = ?", name.SpecID()).
		Where("blobs.artifact_id = ?", strings.ToLower(name.ArtifactID()))

	if id := name.RevisionID(); id != "" { // select specific spec revision
		op = op.Where("blobs.revision_id = ?", id)
	} else {
		if name.SpecID() != "" {
			op = op.Joins(`join (?) latest
				ON blobs.project_id = latest.project_id
				AND blobs.api_id = latest.api_id
				AND blobs.version_id = latest.version_id
				AND blobs.spec_id = latest.spec_id
				AND blobs.revision_id = latest.revision_id`, c.latestSpecRevisionsQuery(ctx))
		} else if name.DeploymentID() != "" {
			op = op.Joins(`join (?) latest
				ON blobs.project_id = latest.project_id
				AND blobs.api_id = latest.api_id
				AND blobs.deployment_id = latest.deployment_id
				AND blobs.revision_id = latest.revision_id`, c.latestDeploymentRevisionsQuery(ctx))
		}
	}

	if err := op.Take(v).Error; err == gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.NotFound, "%q not found in database", name)
	} else if err != nil {
		return nil, grpcErrorForDBError(ctx, errors.Wrapf(err, "get %s", name))
	}

	return v, nil
}
