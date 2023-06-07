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
	"fmt"

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func sum(counts []int64) int64 {
	var s int64
	for _, v := range counts {
		s += v
	}
	return s
}

func reason(counts []int64, tables []interface{}) string {
	r := ""
	for i, v := range counts {
		if v > 0 {
			r += fmt.Sprintf(" %T:%d", tables[i], v)
		}
	}
	return r
}

func (c *Client) DeleteProject(ctx context.Context, name names.Project, cascade bool) error {
	tables := []interface{}{
		models.Project{},
		models.Blob{},
		models.Artifact{},
	}
	counts := make([]int64, len(tables))
	for i, model := range tables {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}
		if _, ok := model.(models.Project); ok && op.RowsAffected == 0 {
			return status.Errorf(codes.NotFound, "%q not found in database", name)
		}
		counts[i] = op.RowsAffected
	}

	if sum(counts) > 1 && !cascade {
		return status.Errorf(codes.FailedPrecondition, "cannot delete child resources of %s in non-cascading mode", name)
	}

	// Tags aren't API resources, so they do not block non-cascading deletes.
	for _, model := range []interface{}{
		models.DeploymentRevisionTag{},
		models.SpecRevisionTag{},
	} {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteApi(ctx context.Context, name names.Api, cascade bool) error {
	tables := []interface{}{
		models.Api{},
		models.Blob{},
		models.Artifact{},
	}

	counts := make([]int64, len(tables))
	for i, model := range tables {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}
		if _, ok := model.(models.Api); ok && op.RowsAffected == 0 {
			return status.Errorf(codes.NotFound, "%q not found in database", name)
		}
		counts[i] = op.RowsAffected
	}

	if sum(counts) > 1 && !cascade {
		return status.Errorf(codes.FailedPrecondition, "cannot delete child resources of %s in non-cascading mode: %s", name, reason(counts, tables))
	}

	for _, model := range []interface{}{
		models.DeploymentRevisionTag{},
		models.SpecRevisionTag{},
	} {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteVersion(ctx context.Context, name names.Version, cascade bool) error {
	tables := []interface{}{
		models.Version{},
		models.Blob{},
		models.Artifact{},
	}
	counts := make([]int64, len(tables))
	for i, model := range tables {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID).
			Where("version_id = ?", name.VersionID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}
		if _, ok := model.(models.Version); ok && op.RowsAffected == 0 {
			return status.Errorf(codes.NotFound, "%q not found in database", name)
		}
		counts[i] = op.RowsAffected
	}

	if sum(counts) > 1 && !cascade {
		return status.Errorf(codes.FailedPrecondition, "cannot delete child resources of %s in non-cascading mode: %s", name, reason(counts, tables))
	}

	for _, model := range []interface{}{
		models.SpecRevisionTag{},
	} {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID).
			Where("version_id = ?", name.VersionID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteSpec(ctx context.Context, name names.Spec, cascade bool) error {
	for _, model := range []interface{}{
		models.Spec{},
		models.SpecRevisionTag{},
		models.Blob{},
	} {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID).
			Where("version_id = ?", name.VersionID).
			Where("spec_id = ?", name.SpecID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}
		if _, ok := model.(models.Spec); ok && op.RowsAffected == 0 {
			return status.Errorf(codes.NotFound, "%q not found in database", name)
		}
	}

	tables := []interface{}{
		models.Artifact{},
		models.Blob{},
	}
	counts := make([]int64, len(tables))
	for i, model := range tables {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID).
			Where("version_id = ?", name.VersionID).
			Where("spec_id = ?", name.SpecID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}
		counts[i] = op.RowsAffected
	}

	if sum(counts) > 0 && !cascade {
		return status.Errorf(codes.FailedPrecondition, "cannot delete child resources of %s in non-cascading mode: %s", name, reason(counts, tables))
	}

	return nil
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
			return grpcErrorForDBError(ctx, errors.Wrapf(err, "delete %s", name))
		}
		if _, ok := model.(models.Spec); ok && op.RowsAffected == 0 {
			return status.Errorf(codes.NotFound, "%q not found in database", name)
		}
	}

	// if we deleted the last revision, return an error to cancel the transaction
	op := c.db.WithContext(ctx).
		Where("project_id = ?", name.ProjectID).
		Where("api_id = ?", name.ApiID).
		Where("version_id = ?", name.VersionID).
		Where("spec_id = ?", name.SpecID).
		Order("revision_create_time desc")
	v := new(models.Spec)
	if err := op.First(v).Error; err == gorm.ErrRecordNotFound {
		return status.Errorf(codes.FailedPrecondition, "cannot delete the only revision: %s", name)
	} else if err != nil {
		return grpcErrorForDBError(ctx, errors.Wrapf(err, "delete %s", name))
	}

	return nil
}

func (c *Client) DeleteDeployment(ctx context.Context, name names.Deployment, cascade bool) error {
	for _, model := range []interface{}{
		models.Deployment{},
		models.DeploymentRevisionTag{},
	} {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID).
			Where("deployment_id = ?", name.DeploymentID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}
		if _, ok := model.(models.Deployment); ok && op.RowsAffected == 0 {
			return status.Errorf(codes.NotFound, "%q not found in database", name)
		}
	}

	tables := []interface{}{
		models.Artifact{},
		models.Blob{},
	}
	counts := make([]int64, len(tables))
	for i, model := range tables {
		op := c.db.WithContext(ctx).Where("project_id = ?", name.ProjectID).
			Where("api_id = ?", name.ApiID).
			Where("deployment_id = ?", name.DeploymentID)
		if err := op.Delete(model).Error; err != nil {
			return err
		}

		counts[i] += op.RowsAffected
	}

	if sum(counts) > 0 && !cascade {
		return status.Errorf(codes.FailedPrecondition, "cannot delete child resources of %s in non-cascading mode: %s", name, reason(counts, tables))
	}

	return nil
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
			return grpcErrorForDBError(ctx, errors.Wrapf(err, "delete %s", name))
		}
		if _, ok := model.(models.Deployment); ok && op.RowsAffected == 0 {
			return status.Errorf(codes.NotFound, "%q not found in database", name)
		}
	}

	// if we deleted the last revision, return an error to cancel the transaction
	op := c.db.WithContext(ctx).
		Where("project_id = ?", name.ProjectID).
		Where("api_id = ?", name.ApiID).
		Where("deployment_id = ?", name.DeploymentID).
		Order("revision_create_time desc")
	v := new(models.Deployment)
	if err := op.First(v).Error; err == gorm.ErrRecordNotFound {
		return status.Errorf(codes.FailedPrecondition, "cannot delete the only revision: %s", name)
	} else if err != nil {
		return grpcErrorForDBError(ctx, errors.Wrapf(err, "delete %s", name))
	}

	return nil
}

func (c *Client) DeleteArtifact(ctx context.Context, name names.Artifact) error {
	for _, model := range []interface{}{
		models.Artifact{},
		models.Blob{},
	} {
		op := c.db.WithContext(ctx).
			Where("project_id = ?", name.ProjectID()).
			Where("api_id = ?", name.ApiID()).
			Where("version_id = ?", name.VersionID()).
			Where("spec_id = ?", name.SpecID()).
			Where("deployment_id = ?", name.DeploymentID()).
			Where("artifact_id = ?", name.ArtifactID())
		if err := op.Delete(model).Error; err != nil {
			return grpcErrorForDBError(ctx, errors.Wrapf(err, "delete %s", name))
		}
		if _, ok := model.(models.Artifact); ok && op.RowsAffected == 0 {
			return status.Errorf(codes.NotFound, "%q not found in database", name)
		}
	}

	return nil
}
