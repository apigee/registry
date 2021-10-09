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

	"github.com/apigee/registry/servers/registry/internal/storage/filtering"
	"github.com/apigee/registry/servers/registry/internal/storage/gorm"
	"github.com/apigee/registry/servers/registry/internal/storage/models"
	"github.com/apigee/registry/servers/registry/names"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProjectList contains a page of project resources.
type ProjectList struct {
	Projects []models.Project
	Token    string
}

var projectFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "display_name", Type: filtering.String},
	{Name: "description", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "update_time", Type: filtering.Timestamp},
}

func (d *Client) ListProjects(ctx context.Context, opts PageOptions) (ProjectList, error) {
	q := d.NewQuery(gorm.ProjectEntityName)

	token, err := decodeToken(opts.Token)
	if err != nil {
		return ProjectList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ProjectList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	q = q.ApplyOffset(token.Offset)

	filter, err := filtering.NewFilter(opts.Filter, projectFields)
	if err != nil {
		return ProjectList{}, err
	}

	it := d.Run(ctx, q)
	response := ProjectList{
		Projects: make([]models.Project, 0, opts.Size),
	}

	project := new(models.Project)
	for _, err = it.Next(project); err == nil; _, err = it.Next(project) {
		match, err := filter.Matches(projectMap(*project))
		if err != nil {
			return response, err
		} else if !match {
			token.Offset++
			continue
		} else if len(response.Projects) == int(opts.Size) {
			break
		}

		response.Projects = append(response.Projects, *project)
		token.Offset++
	}
	if err != nil && err != iterator.Done {
		return response, status.Error(codes.Internal, err.Error())
	}

	if err == nil {
		response.Token, err = encodeToken(token)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

func projectMap(p models.Project) map[string]interface{} {
	return map[string]interface{}{
		"name":         p.Name(),
		"project_id":   p.ProjectID,
		"display_name": p.DisplayName,
		"description":  p.Description,
		"create_time":  p.CreateTime,
		"update_time":  p.UpdateTime,
	}
}

func (d *Client) GetProject(ctx context.Context, name names.Project) (*models.Project, error) {
	project := new(models.Project)
	k := d.NewKey(gorm.ProjectEntityName, name.String())
	if err := d.Get(ctx, k, project); d.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "project %q not found in database", name)
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return project, nil
}

func (d *Client) SaveProject(ctx context.Context, project *models.Project) error {
	k := d.NewKey(gorm.ProjectEntityName, project.Name())
	if _, err := d.Put(ctx, k, project); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *Client) DeleteProject(ctx context.Context, name names.Project) error {
	for _, entityName := range []string{
		gorm.ProjectEntityName,
		gorm.ApiEntityName,
		gorm.VersionEntityName,
		gorm.SpecEntityName,
		gorm.SpecRevisionTagEntityName,
		gorm.ArtifactEntityName,
		gorm.BlobEntityName,
	} {
		q := d.NewQuery(entityName)
		q = q.Require("ProjectID", name.ProjectID)
		if err := d.Delete(ctx, q); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}
