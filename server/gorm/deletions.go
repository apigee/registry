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

package gorm

import (
	"context"

	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/apigee/registry/server/storage"
)

// DeleteAllMatches deletes all entities matching a query.
func (c *Client) DeleteAllMatches(ctx context.Context, q storage.Query) error {
	op := c.db
	for _, r := range q.(*Query).Requirements {
		op = op.Where(r.Name+" = ?", r.Value)
	}
	switch q.(*Query).Kind {
	case "Project":
		return op.Delete(models.Project{}).Error
	case "Api":
		return op.Delete(models.Api{}).Error
	case "Version":
		return op.Delete(models.Version{}).Error
	case "Spec":
		return op.Delete(models.Spec{}).Error
	case "Blob":
		return op.Delete(models.Blob{}).Error
	case "Artifact":
		return op.Delete(models.Artifact{}).Error
	case "SpecRevisionTag":
		return op.Delete(models.SpecRevisionTag{}).Error
	}
	return nil
}

// DeleteChildrenOfProject deletes all the children of a project.
func (c *Client) DeleteChildrenOfProject(ctx context.Context, project names.Project) error {
	entityNames := []string{
		models.ArtifactEntityName,
		models.BlobEntityName,
		models.SpecEntityName,
		models.SpecRevisionTagEntityName,
		storage.VersionEntityName,
		storage.ApiEntityName,
	}
	for _, entityName := range entityNames {
		q := c.NewQuery(entityName)
		q = q.Require("ProjectID", project.ProjectID)
		err := c.DeleteAllMatches(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteChildrenOfApi deletes all the children of a api.
func (c *Client) DeleteChildrenOfApi(ctx context.Context, api names.Api) error {
	for _, entityName := range []string{
		models.BlobEntityName,
		models.SpecEntityName,
		storage.VersionEntityName,
	} {
		q := c.NewQuery(entityName)
		q = q.Require("ProjectID", api.ProjectID)
		q = q.Require("ApiID", api.ApiID)
		err := c.DeleteAllMatches(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteChildrenOfVersion deletes all the children of a version.
func (c *Client) DeleteChildrenOfVersion(ctx context.Context, version names.Version) error {
	for _, entityName := range []string{
		models.BlobEntityName,
		models.SpecEntityName,
	} {
		q := c.NewQuery(entityName)
		q = q.Require("ProjectID", version.ProjectID)
		q = q.Require("ApiID", version.ApiID)
		q = q.Require("VersionID", version.VersionID)
		err := c.DeleteAllMatches(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteChildrenOfSpec deletes all the children of a spec.
func (c *Client) DeleteChildrenOfSpec(ctx context.Context, spec *models.Spec) error {
	q := c.NewQuery(models.BlobEntityName)
	q = q.Require("ProjectID", spec.ProjectID)
	q = q.Require("ApiID", spec.ApiID)
	q = q.Require("VersionID", spec.VersionID)
	q = q.Require("SpecID", spec.SpecID)
	return c.DeleteAllMatches(ctx, q)
}
