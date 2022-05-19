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

func (c *Client) DeleteProject(ctx context.Context, name names.Project, cascade bool) error {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		for _, model := range []interface{}{
			models.Project{},
			models.Api{},
			models.Deployment{},
			models.DeploymentRevisionTag{},
			models.Version{},
			models.Spec{},
			models.SpecRevisionTag{},
			models.Blob{},
			models.Artifact{},
		} {
			op := tx.Where("project_id = ?", name.ProjectID)
			if err := op.Delete(model).Error; err != nil {
				return err
			}

			count += op.RowsAffected
		}

		if count > 1 && !cascade {
			return status.Errorf(codes.FailedPrecondition, "cannot delete child resources in non-cascading mode")
		}

		return nil
	})

	switch status.Code(err) {
	case codes.OK:
		return nil
	case codes.FailedPrecondition:
		return err
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (c *Client) DeleteApi(ctx context.Context, name names.Api, cascade bool) error {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		for _, model := range []interface{}{
			models.Api{},
			models.Deployment{},
			models.DeploymentRevisionTag{},
			models.Version{},
			models.Spec{},
			models.SpecRevisionTag{},
			models.Blob{},
			models.Artifact{},
		} {
			op := tx.Where("project_id = ?", name.ProjectID).
				Where("api_id = ?", name.ApiID)
			if err := op.Delete(model).Error; err != nil {
				return err
			}

			count += op.RowsAffected
		}

		if count > 1 && !cascade {
			return status.Errorf(codes.FailedPrecondition, "cannot delete child resources in non-cascading mode")
		}

		return nil
	})

	switch status.Code(err) {
	case codes.OK:
		return nil
	case codes.FailedPrecondition:
		return err
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (c *Client) DeleteVersion(ctx context.Context, name names.Version, cascade bool) error {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		for _, model := range []interface{}{
			models.Version{},
			models.Spec{},
			models.SpecRevisionTag{},
			models.Blob{},
			models.Artifact{},
		} {
			op := tx.Where("project_id = ?", name.ProjectID).
				Where("api_id = ?", name.ApiID).
				Where("version_id = ?", name.VersionID)
			if err := op.Delete(model).Error; err != nil {
				return err
			}

			count += op.RowsAffected
		}

		if count > 1 && !cascade {
			return status.Errorf(codes.FailedPrecondition, "cannot delete child resources in non-cascading mode")
		}

		return nil
	})

	switch status.Code(err) {
	case codes.OK:
		return nil
	case codes.FailedPrecondition:
		return err
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (c *Client) DeleteSpec(ctx context.Context, name names.Spec, cascade bool) error {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, model := range []interface{}{
			models.Spec{},
			models.SpecRevisionTag{},
			models.Blob{},
		} {
			op := tx.Where("project_id = ?", name.ProjectID).
				Where("api_id = ?", name.ApiID).
				Where("version_id = ?", name.VersionID).
				Where("spec_id = ?", name.SpecID)
			if err := op.Delete(model).Error; err != nil {
				return err
			}
		}

		var childCount int64
		for _, model := range []interface{}{
			models.Artifact{},
			models.Blob{},
		} {
			op := tx.Where("project_id = ?", name.ProjectID).
				Where("api_id = ?", name.ApiID).
				Where("version_id = ?", name.VersionID).
				Where("spec_id = ?", name.SpecID)
			if err := op.Delete(model).Error; err != nil {
				return err
			}

			childCount += op.RowsAffected
		}

		if childCount > 0 && !cascade {
			return status.Errorf(codes.FailedPrecondition, "cannot delete child resources in non-cascading mode")
		}

		return nil
	})

	switch status.Code(err) {
	case codes.OK:
		return nil
	case codes.FailedPrecondition:
		return err
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (c *Client) DeleteSpecRevision(ctx context.Context, name names.SpecRevision) error {
	name, err := c.unwrapSpecRevisionTag(ctx, name)
	if err != nil {
		return err
	}

	for _, model := range []interface{}{
		models.Spec{},
		models.SpecRevisionTag{},
	} {
		op := c.db.WithContext(ctx).
			Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID).
			Where("version_id = ?", name.VersionID).
			Where("spec_id = ?", name.SpecID).
			Where("revision_id = ?", name.RevisionID)
		if err := op.Delete(model).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

func (c *Client) DeleteDeployment(ctx context.Context, name names.Deployment, cascade bool) error {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, model := range []interface{}{
			models.Deployment{},
			models.DeploymentRevisionTag{},
		} {
			op := tx.Where("project_id = ?", name.ProjectID).
				Where("api_id = ?", name.ApiID).
				Where("deployment_id = ?", name.DeploymentID)
			if err := op.Delete(model).Error; err != nil {
				return err
			}
		}

		var childCount int64
		for _, model := range []interface{}{
			models.Artifact{},
			models.Blob{},
		} {
			op := tx.Where("project_id = ?", name.ProjectID).
				Where("api_id = ?", name.ApiID).
				Where("deployment_id = ?", name.DeploymentID)
			if err := op.Delete(model).Error; err != nil {
				return err
			}

			childCount += op.RowsAffected
		}

		if childCount > 0 && !cascade {
			return status.Errorf(codes.FailedPrecondition, "cannot delete child resources in non-cascading mode")
		}

		return nil
	})

	switch status.Code(err) {
	case codes.OK:
		return nil
	case codes.FailedPrecondition:
		return err
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (c *Client) DeleteDeploymentRevision(ctx context.Context, name names.DeploymentRevision) error {
	name, err := c.unwrapDeploymentRevisionTag(ctx, name)
	if err != nil {
		return err
	}

	for _, model := range []interface{}{
		models.Deployment{},
		models.DeploymentRevisionTag{},
	} {
		op := c.db.WithContext(ctx).
			Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID).
			Where("deployment_id = ?", name.DeploymentID).
			Where("revision_id = ?", name.RevisionID)
		if err := op.Delete(model).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

func (c *Client) DeleteArtifact(ctx context.Context, name names.Artifact) error {
	for _, model := range []interface{}{
		models.Blob{},
		models.Artifact{},
	} {
		op := c.db.WithContext(ctx).
			Where("project_id = ?", name.ProjectID()).
			Where("api_id = ?", name.ApiID()).
			Where("version_id = ?", name.VersionID()).
			Where("spec_id = ?", name.SpecID()).
			Where("artifact_id = ?", name.ArtifactID())
		if err := op.Delete(model).Error; err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}
