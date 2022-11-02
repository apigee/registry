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
	"strings"

	"github.com/apigee/registry/server/registry/filtering"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// gormOrdering accepts a user-specified order_by string and returns a gorm-compatible equivalent.
// For example, the user-specified string `name,description` returns `key,description`.
// An error is returned if the string is invalid or refers to a field that isn't included in the provided `fields` map.
func gormOrdering(ordering string, fields map[string]filtering.FieldType) (string, error) {
	if ordering == "" {
		return "key", nil
	}

	clauses := make([]string, 0)
	for _, v := range strings.Split(ordering, ",") {
		v = strings.TrimSpace(v)

		// Check if the field is specified in descending order and trim it from the string.
		// After this point only the field name should remain.
		descending := strings.HasSuffix(v, " desc")
		v = strings.TrimSuffix(v, "desc")
		v = strings.TrimSpace(v)

		if strings.Contains(v, " ") {
			return "", status.Errorf(codes.InvalidArgument, "invalid order_by field %q: too many parts", v)
		} else if len(v) == 0 {
			return "", status.Errorf(codes.InvalidArgument, "invalid order_by field %q: missing field name", v)
		}

		// Check if the field is valid for this model type and replace it with the internal name if needed.
		// After this point, the clause should contain an internal field name.
		var clause string
		for field := range fields {
			if field == v && field == "name" {
				clause = "key"
			} else if field == v {
				clause = field
			}
		}
		if clause == "" {
			return ordering, status.Errorf(codes.InvalidArgument, "unknown field name %q", v)
		}

		if descending {
			clause += " desc"
		}

		clauses = append(clauses, clause)
	}

	return strings.Join(clauses, ","), nil
}

// limit returns the database page size to use for a listing request.
func limit(opts PageOptions) int {
	// Without filters, read exactly enough rows to fill the page,
	// plus an extra row to check if another page exists.
	if opts.Filter == "" {
		return int(opts.Size) + 1
	}

	// When filters are present, we should read larger pages.
	return 500
}

// ProjectList contains a page of project resources.
type ProjectList struct {
	Projects []models.Project
	Token    string
}

func (c *Client) ListProjects(ctx context.Context, opts PageOptions) (ProjectList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return ProjectList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ProjectList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return ProjectList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	filter, err := filtering.NewFilter(opts.Filter, filtering.ProjectFields)
	if err != nil {
		return ProjectList{}, err
	}

	order, err := gormOrdering(opts.Order, filtering.ProjectFields)
	if err != nil {
		return ProjectList{}, err
	}

	response := ProjectList{
		Projects: make([]models.Project, 0, opts.Size),
	}

	for {
		var page []models.Project
		op := c.db.WithContext(ctx).Order(order).Limit(limit(opts))
		err := op.Offset(token.Offset).Find(&page).Error

		if err != nil {
			return ProjectList{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "find %#v", token))
		} else if len(page) == 0 {
			break
		}

		for _, v := range page {
			m, err := filtering.ProjectMap(v)
			if err != nil {
				return ProjectList{}, status.Error(codes.Internal, err.Error())
			}
			match, err := filter.Matches(m)
			if err != nil {
				return ProjectList{}, err
			} else if !match {
				token.Offset++
				continue
			}

			if len(response.Projects) == int(opts.Size) {
				response.Token, err = encodeToken(token)
				if err != nil {
					return ProjectList{}, status.Error(codes.Internal, err.Error())
				}
				return response, nil
			}

			token.Offset++
			response.Projects = append(response.Projects, v)
		}
		if op.RowsAffected < int64(opts.Size) {
			break
		}
	}

	return response, nil
}

// ApiList contains a page of api resources.
type ApiList struct {
	Apis  []models.Api
	Token string
}

func (c *Client) ListApis(ctx context.Context, parent names.Project, opts PageOptions) (ApiList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return ApiList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ApiList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return ApiList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	op := c.db.WithContext(ctx).
		Limit(limit(opts))

	if parent.ProjectID != "-" {
		op = op.Where("project_id = ?", parent.ProjectID)
		if _, err := c.GetProject(ctx, parent); err != nil {
			return ApiList{}, err
		}
	}

	filter, err := filtering.NewFilter(opts.Filter, filtering.ApiFields)
	if err != nil {
		return ApiList{}, err
	}

	if order, err := gormOrdering(opts.Order, filtering.ApiFields); err != nil {
		return ApiList{}, err
	} else {
		op = op.Order(order)
	}

	response := ApiList{
		Apis: make([]models.Api, 0, opts.Size),
	}

	for {
		var page []models.Api
		err := op.Offset(token.Offset).Find(&page).Error

		if err != nil {
			return ApiList{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "find %#v", token))
		} else if len(page) == 0 {
			break
		}

		for _, v := range page {
			m, err := filtering.ApiMap(v)
			if err != nil {
				return ApiList{}, status.Error(codes.Internal, err.Error())
			}

			match, err := filter.Matches(m)
			if err != nil {
				return ApiList{}, err
			} else if !match {
				token.Offset++
				continue
			}

			if len(response.Apis) == int(opts.Size) {
				response.Token, err = encodeToken(token)
				if err != nil {
					return ApiList{}, status.Error(codes.Internal, err.Error())
				}
				return response, nil
			}

			token.Offset++
			response.Apis = append(response.Apis, v)
		}
		if op.RowsAffected < int64(opts.Size) {
			break
		}
	}

	return response, nil
}

// VersionList contains a page of version resources.
type VersionList struct {
	Versions []models.Version
	Token    string
}

func (c *Client) ListVersions(ctx context.Context, parent names.Api, opts PageOptions) (VersionList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return VersionList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return VersionList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return VersionList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" {
		if _, err := c.GetApi(ctx, parent); err != nil {
			return VersionList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return VersionList{}, err
		}
	}

	filter, err := filtering.NewFilter(opts.Filter, filtering.VersionFields)
	if err != nil {
		return VersionList{}, err
	}

	op := c.db.WithContext(ctx).Limit(limit(opts))
	if parent.ProjectID != "-" {
		op = op.Where("project_id = ?", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		op = op.Where("api_id = ?", parent.ApiID)
	}

	if order, err := gormOrdering(opts.Order, filtering.VersionFields); err != nil {
		return VersionList{}, err
	} else {
		op = op.Order(order)
	}

	response := VersionList{
		Versions: make([]models.Version, 0, opts.Size),
	}

	for {
		var page []models.Version
		err := op.Offset(token.Offset).Find(&page).Error

		if err != nil {
			return VersionList{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "find %#v", token))
		} else if len(page) == 0 {
			break
		}

		for _, v := range page {
			m, err := filtering.VersionMap(v)
			if err != nil {
				return VersionList{}, status.Error(codes.Internal, err.Error())
			}

			match, err := filter.Matches(m)
			if err != nil {
				return VersionList{}, err
			} else if !match {
				token.Offset++
				continue
			}

			if len(response.Versions) == int(opts.Size) {
				response.Token, err = encodeToken(token)
				if err != nil {
					return VersionList{}, status.Error(codes.Internal, err.Error())
				}
				return response, nil
			}

			token.Offset++
			response.Versions = append(response.Versions, v)
		}
		if op.RowsAffected < int64(opts.Size) {
			break
		}
	}

	return response, nil
}

// SpecList contains a page of spec resources.
type SpecList struct {
	Specs []models.Spec
	Token string
}

func (c *Client) ListSpecs(ctx context.Context, parent names.Version, opts PageOptions) (SpecList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" {
		if _, err := c.GetVersion(ctx, parent); err != nil {
			return SpecList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID == "-" {
		if _, err := c.GetApi(ctx, parent.Api()); err != nil {
			return SpecList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.VersionID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return SpecList{}, err
		}
	}

	filter, err := filtering.NewFilter(opts.Filter, filtering.SpecFields)
	if err != nil {
		return SpecList{}, err
	}

	// Select all columns from `specs` table specifically.
	// We do not want to select duplicates from the joined subquery result.
	op := c.db.WithContext(ctx).Select("specs.*").
		Table("specs").
		// Join missing columns that couldn't be selected in the subquery.
		Joins("JOIN (?) AS grp ON specs.project_id = grp.project_id AND specs.api_id = grp.api_id AND specs.version_id = grp.version_id AND specs.spec_id = grp.spec_id AND specs.revision_create_time = grp.recent_create_time",
			// Select spec names and only their most recent revision_create_time
			// This query cannot select all the columns we want.
			// See: https://stackoverflow.com/questions/7745609/sql-select-only-rows-with-max-value-on-a-column
			c.db.WithContext(ctx).Select("project_id, api_id, version_id, spec_id, MAX(revision_create_time) AS recent_create_time").
				Table("specs").
				Group("project_id, api_id, version_id, spec_id")).
		Limit(limit(opts))

	if parent.ProjectID != "-" {
		op = op.Where("specs.project_id = ?", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		op = op.Where("specs.api_id = ?", parent.ApiID)
	}
	if parent.VersionID != "-" {
		op = op.Where("specs.version_id = ?", parent.VersionID)
	}

	if order, err := gormOrdering(opts.Order, filtering.SpecFields); err != nil {
		return SpecList{}, err
	} else {
		op = op.Order(order)
	}

	response := SpecList{
		Specs: make([]models.Spec, 0, opts.Size),
	}

	for {
		var page []models.Spec
		err := op.Offset(token.Offset).Find(&page).Error

		if err != nil {
			return SpecList{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "find %#v", token))
		} else if len(page) == 0 {
			break
		}

		for _, v := range page {
			m, err := filtering.SpecMap(v)
			if err != nil {
				return SpecList{}, status.Error(codes.Internal, err.Error())
			}

			match, err := filter.Matches(m)
			if err != nil {
				return SpecList{}, err
			} else if !match {
				token.Offset++
				continue
			}

			if len(response.Specs) == int(opts.Size) {
				response.Token, err = encodeToken(token)
				if err != nil {
					return SpecList{}, status.Error(codes.Internal, err.Error())
				}
				return response, nil
			}

			token.Offset++
			response.Specs = append(response.Specs, v)
		}
		if op.RowsAffected < int64(opts.Size) {
			break
		}
	}

	return response, nil
}

func (c *Client) ListSpecRevisions(ctx context.Context, parent names.Spec, opts PageOptions) (SpecList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	// Check existence of the deepest fully specified resource in the parent name.
	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" && parent.SpecID != "-" {
		if _, err := c.GetSpec(ctx, parent); err != nil {
			return SpecList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" && parent.SpecID == "-" {
		if _, err := c.GetVersion(ctx, parent.Version()); err != nil {
			return SpecList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID == "-" && parent.SpecID == "-" {
		if _, err := c.GetApi(ctx, parent.Api()); err != nil {
			return SpecList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.VersionID == "-" && parent.SpecID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return SpecList{}, err
		}
	}

	op := c.db.WithContext(ctx).
		Order("revision_create_time desc").
		Offset(token.Offset).
		Limit(int(opts.Size) + 1)

	if id := parent.ProjectID; id != "-" {
		op = op.Where("project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("api_id = ?", id)
	}
	if id := parent.VersionID; id != "-" {
		op = op.Where("version_id = ?", id)
	}
	if id := parent.SpecID; id != "-" {
		op = op.Where("spec_id = ?", id)
	}

	response := SpecList{
		Specs: make([]models.Spec, 0, opts.Size),
	}

	err = op.Find(&response.Specs).Error

	if err != nil {
		return SpecList{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "find %#v", token))
	}

	// Trim the response and return a page token if too many resources were found.
	if len(response.Specs) > int(opts.Size) {
		token.Offset += int(opts.Size)
		response.Specs = response.Specs[:opts.Size]
		response.Token, err = encodeToken(token)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

// DeploymentList contains a page of deployment resources.
type DeploymentList struct {
	Deployments []models.Deployment
	Token       string
}

func (c *Client) ListDeployments(ctx context.Context, parent names.Api, opts PageOptions) (DeploymentList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return DeploymentList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return DeploymentList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return DeploymentList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" {
		if _, err := c.GetApi(ctx, parent); err != nil {
			return DeploymentList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return DeploymentList{}, err
		}
	}

	filter, err := filtering.NewFilter(opts.Filter, filtering.DeploymentFields)
	if err != nil {
		return DeploymentList{}, err
	}

	// Select all columns from `deployments` table specifically.
	// We do not want to select duplicates from the joined subquery result.
	op := c.db.WithContext(ctx).Select("deployments.*").
		Table("deployments").
		// Join missing columns that couldn't be selected in the subquery.
		Joins("JOIN (?) AS grp ON deployments.project_id = grp.project_id AND deployments.api_id = grp.api_id AND deployments.deployment_id = grp.deployment_id AND deployments.revision_create_time = grp.recent_create_time",
			// Select deployment names and only their most recent revision_create_time
			// This query cannot select all the columns we want.
			// See: https://stackoverflow.com/questions/7745609/sql-select-only-rows-with-max-value-on-a-column
			c.db.WithContext(ctx).Select("project_id, api_id, deployment_id, MAX(revision_create_time) AS recent_create_time").
				Table("deployments").
				Group("project_id, api_id, deployment_id")).
		Limit(limit(opts))

	if parent.ProjectID != "-" {
		op = op.Where("deployments.project_id = ?", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		op = op.Where("deployments.api_id = ?", parent.ApiID)
	}

	if order, err := gormOrdering(opts.Order, filtering.DeploymentFields); err != nil {
		return DeploymentList{}, err
	} else {
		op = op.Order(order)
	}

	response := DeploymentList{
		Deployments: make([]models.Deployment, 0, opts.Size),
	}

	for {
		var page []models.Deployment
		err := op.Offset(token.Offset).Find(&page).Error

		if err != nil {
			return DeploymentList{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "find %#v", token))
		} else if len(page) == 0 {
			break
		}

		for _, v := range page {
			m, err := filtering.DeploymentMap(v)
			if err != nil {
				return DeploymentList{}, status.Error(codes.Internal, err.Error())
			}

			match, err := filter.Matches(m)
			if err != nil {
				return DeploymentList{}, err
			} else if !match {
				token.Offset++
				continue
			}

			if len(response.Deployments) == int(opts.Size) {
				response.Token, err = encodeToken(token)
				if err != nil {
					return DeploymentList{}, status.Error(codes.Internal, err.Error())
				}
				return response, nil
			}

			token.Offset++
			response.Deployments = append(response.Deployments, v)
		}
		if op.RowsAffected < int64(opts.Size) {
			break
		}
	}

	return response, nil
}

func (c *Client) ListDeploymentRevisions(ctx context.Context, parent names.Deployment, opts PageOptions) (DeploymentList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return DeploymentList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.DeploymentID != "-" {
		if _, err := c.GetDeployment(ctx, parent); err != nil {
			return DeploymentList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.DeploymentID == "-" {
		if _, err := c.GetApi(ctx, parent.Api()); err != nil {
			return DeploymentList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.DeploymentID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return DeploymentList{}, err
		}
	}

	op := c.db.WithContext(ctx).
		Order("revision_create_time desc").
		Offset(token.Offset).
		Limit(int(opts.Size) + 1)

	if id := parent.ProjectID; id != "-" {
		op = op.Where("project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("api_id = ?", id)
	}
	if id := parent.DeploymentID; id != "-" {
		op = op.Where("deployment_id = ?", id)
	}

	response := DeploymentList{
		Deployments: make([]models.Deployment, 0, opts.Size),
	}

	err = op.Find(&response.Deployments).Error

	if err != nil {
		return DeploymentList{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "find %#v", token))
	}

	// Trim the response and return a page token if too many resources were found.
	if len(response.Deployments) > int(opts.Size) {
		token.Offset += int(opts.Size)
		response.Deployments = response.Deployments[:opts.Size]
		response.Token, err = encodeToken(token)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

// ArtifactList contains a page of artifact resources.
type ArtifactList struct {
	Artifacts []models.Artifact
	Token     string
}

func (c *Client) ListSpecArtifacts(ctx context.Context, parent names.Spec, opts PageOptions) (ArtifactList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" && parent.SpecID != "-" {
		if _, err := c.GetSpec(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" && parent.SpecID == "-" {
		if _, err := c.GetVersion(ctx, parent.Version()); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID == "-" && parent.SpecID == "-" {
		if _, err := c.GetApi(ctx, parent.Api()); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.VersionID == "-" && parent.SpecID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return ArtifactList{}, err
		}
	}

	op := c.db.WithContext(ctx).Where(`deployment_id = ''`)
	if id := parent.ProjectID; id != "-" {
		op = op.Where("project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("api_id = ?", id)
	}
	if id := parent.VersionID; id != "-" {
		op = op.Where("version_id = ?", id)
	}
	if id := parent.SpecID; id != "-" {
		op = op.Where("spec_id = ?", id)
	}

	return c.listArtifacts(ctx, op, opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != "" && a.VersionID != "" && a.SpecID != ""
	})
}

func (c *Client) ListVersionArtifacts(ctx context.Context, parent names.Version, opts PageOptions) (ArtifactList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" {
		if _, err := c.GetVersion(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID == "-" {
		if _, err := c.GetApi(ctx, parent.Api()); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.VersionID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return ArtifactList{}, err
		}
	}

	op := c.db.WithContext(ctx).
		Where(`deployment_id = ''`).
		Where(`spec_id = ''`)
	if id := parent.ProjectID; id != "-" {
		op = op.Where("project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("api_id = ?", id)
	}
	if id := parent.VersionID; id != "-" {
		op = op.Where("version_id = ?", id)
	}

	return c.listArtifacts(ctx, op, opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != "" && a.VersionID != ""
	})
}

func (c *Client) ListDeploymentArtifacts(ctx context.Context, parent names.Deployment, opts PageOptions) (ArtifactList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.DeploymentID != "-" {
		if _, err := c.GetDeployment(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.DeploymentID == "-" {
		if _, err := c.GetApi(ctx, parent.Api()); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.DeploymentID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return ArtifactList{}, err
		}
	}

	op := c.db.WithContext(ctx).
		Where(`version_id = ''`).
		Where(`spec_id = ''`)
	if id := parent.ProjectID; id != "-" {
		op = op.Where("project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("api_id = ?", id)
	}
	if id := parent.DeploymentID; id != "-" {
		op = op.Where("deployment_id = ?", id)
	}

	return c.listArtifacts(ctx, op, opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != "" && a.DeploymentID != ""
	})
}

func (c *Client) ListApiArtifacts(ctx context.Context, parent names.Api, opts PageOptions) (ArtifactList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" {
		if _, err := c.GetApi(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return ArtifactList{}, err
		}
	}

	op := c.db.WithContext(ctx).
		Where(`deployment_id = ''`).
		Where(`version_id = ''`).
		Where(`spec_id = ''`)
	if id := parent.ProjectID; id != "-" {
		op = op.Where("project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("api_id = ?", id)
	}

	return c.listArtifacts(ctx, op, opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != ""
	})
}

func (c *Client) ListProjectArtifacts(ctx context.Context, parent names.Project, opts PageOptions) (ArtifactList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	if err := token.ValidateOrder(opts.Order); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid order_by %q: %s", opts.Order, err)
	} else {
		token.Order = opts.Order
	}

	op := c.db.WithContext(ctx).
		Where(`api_id = ''`).
		Where(`deployment_id = ''`).
		Where(`version_id = ''`).
		Where(`spec_id = ''`)
	if id := parent.ProjectID; id != "-" {
		op = op.Where("project_id = ?", id)
		if _, err := c.GetProject(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	}

	return c.listArtifacts(ctx, op, opts, func(a *models.Artifact) bool {
		return a.ProjectID != ""
	})
}

func (c *Client) listArtifacts(ctx context.Context, op *gorm.DB, opts PageOptions, include func(*models.Artifact) bool) (ArtifactList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	} else {
		token.Filter = opts.Filter
	}

	filter, err := filtering.NewFilter(opts.Filter, filtering.ArtifactFields)
	if err != nil {
		return ArtifactList{}, err
	}

	if order, err := gormOrdering(opts.Order, filtering.ArtifactFields); err != nil {
		return ArtifactList{}, err
	} else {
		op = op.Order(order)
	}

	response := ArtifactList{
		Artifacts: make([]models.Artifact, 0, opts.Size),
	}

	for {
		var page []models.Artifact
		op = op.Order("key").Limit(limit(opts))
		err := op.Offset(token.Offset).Find(&page).Error

		if err != nil {
			return ArtifactList{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "find %#v", token))
		} else if len(page) == 0 {
			break
		}

		for _, v := range page {
			m, err := filtering.ArtifactMap(v)
			if err != nil {
				return ArtifactList{}, status.Error(codes.Internal, err.Error())
			}
			match, err := filter.Matches(m)
			if err != nil {
				return ArtifactList{}, err
			} else if !match || !include(&v) {
				token.Offset++
				continue
			}

			if len(response.Artifacts) == int(opts.Size) {
				response.Token, err = encodeToken(token)
				if err != nil {
					return ArtifactList{}, status.Error(codes.Internal, err.Error())
				}
				return response, nil
			}

			token.Offset++
			response.Artifacts = append(response.Artifacts, v)
		}
		if op.RowsAffected < int64(opts.Size) {
			break
		}
	}

	return response, nil
}
