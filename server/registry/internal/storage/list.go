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
	"github.com/apigee/registry/server/registry/internal/storage/filtering"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var tableFieldsLookup = map[string]map[string]filtering.FieldType{
	"projects":             projectFields,
	"apis":                 apiFields,
	"versions":             versionFields,
	"specs":                specFields,
	"deployments":          deploymentFields,
	"artifacts":            artifactFields,
	"revisioned_artifacts": revisionedArtifactFields,
}

var defaultOrder = map[string]string{
	"projects":             "project_id",
	"apis":                 "project_id, api_id",
	"deployments":          "project_id, api_id, deployment_id, revision_create_time desc",
	"versions":             "project_id, api_id, version_id",
	"specs":                "project_id, api_id, version_id, spec_id, revision_create_time desc",
	"artifacts":            "project_id, api_id, version_id, spec_id, deployment_id, artifact_id, create_time desc",
	"revisioned_artifacts": "project_id, api_id, version_id, spec_id, deployment_id, revision_create_time desc, artifact_id",
}

var projectFields = map[string]filtering.FieldType{
	"name":         filtering.String,
	"project_id":   filtering.String,
	"display_name": filtering.String,
	"description":  filtering.String,
	"create_time":  filtering.Timestamp,
	"update_time":  filtering.Timestamp,
}

var apiFields = map[string]filtering.FieldType{
	"name":                   filtering.String,
	"project_id":             filtering.String,
	"api_id":                 filtering.String,
	"display_name":           filtering.String,
	"description":            filtering.String,
	"create_time":            filtering.Timestamp,
	"update_time":            filtering.Timestamp,
	"availability":           filtering.String,
	"recommended_version":    filtering.String,
	"recommended_deployment": filtering.String,
	"labels":                 filtering.StringMap,
}

var versionFields = map[string]filtering.FieldType{
	"name":         filtering.String,
	"project_id":   filtering.String,
	"api_id":       filtering.String,
	"version_id":   filtering.String,
	"display_name": filtering.String,
	"description":  filtering.String,
	"create_time":  filtering.Timestamp,
	"update_time":  filtering.Timestamp,
	"state":        filtering.String,
	"labels":       filtering.StringMap,
	"primary_spec": filtering.String,
}

var specFields = map[string]filtering.FieldType{
	"name":                 filtering.String,
	"project_id":           filtering.String,
	"api_id":               filtering.String,
	"version_id":           filtering.String,
	"spec_id":              filtering.String,
	"filename":             filtering.String,
	"description":          filtering.String,
	"create_time":          filtering.Timestamp,
	"revision_create_time": filtering.Timestamp,
	"revision_update_time": filtering.Timestamp,
	"mime_type":            filtering.String,
	"size_bytes":           filtering.Int,
	"source_uri":           filtering.String,
	"labels":               filtering.StringMap,
}

var deploymentFields = map[string]filtering.FieldType{
	"name":                 filtering.String,
	"project_id":           filtering.String,
	"api_id":               filtering.String,
	"deployment_id":        filtering.String,
	"display_name":         filtering.String,
	"description":          filtering.String,
	"create_time":          filtering.Timestamp,
	"revision_create_time": filtering.Timestamp,
	"revision_update_time": filtering.Timestamp,
	"api_spec_revision":    filtering.String,
	"endpoint_uri":         filtering.String,
	"external_channel_uri": filtering.String,
	"intended_audience":    filtering.String,
	"access_guidance":      filtering.String,
	"labels":               filtering.StringMap,
}

var artifactFields = map[string]filtering.FieldType{
	"name":          filtering.String,
	"project_id":    filtering.String,
	"api_id":        filtering.String,
	"version_id":    filtering.String,
	"spec_id":       filtering.String,
	"artifact_id":   filtering.String,
	"deployment_id": filtering.String,
	"create_time":   filtering.Timestamp,
	"update_time":   filtering.Timestamp,
	"mime_type":     filtering.String,
	"size_bytes":    filtering.Int,
	"labels":        filtering.StringMap,
}

var revisionedArtifactFields = map[string]filtering.FieldType{
	"name":                 filtering.String,
	"project_id":           filtering.String,
	"api_id":               filtering.String,
	"version_id":           filtering.String,
	"spec_id":              filtering.String,
	"artifact_id":          filtering.String,
	"deployment_id":        filtering.String,
	"create_time":          filtering.Timestamp,
	"update_time":          filtering.Timestamp,
	"mime_type":            filtering.String,
	"size_bytes":           filtering.Int,
	"revision_create_time": filtering.Timestamp,
	"revision_update_time": filtering.Timestamp,
}

// gormOrdering accepts a user-specified order_by string and returns a gorm-compatible equivalent.
// For example, the user-specified string `name,description` returns `key,description`.
// An error is returned if the string is invalid or refers to a field that isn't included in the `fields` map.
func gormOrdering(ordering, table string) (string, error) {
	fields, ok := tableFieldsLookup[table]
	if !ok {
		return "", status.Errorf(codes.Internal, "unknown order table: %q", table)
	}
	if ordering == "" {
		ordering = defaultOrder[table]
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
			return ordering, status.Errorf(codes.InvalidArgument, "unknown field name %q in %q", v, table)
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

	// When filters are present, read max page size
	return 1000
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

	filter, err := filtering.NewFilter(opts.Filter, projectFields)
	if err != nil {
		return ProjectList{}, err
	}

	order, err := gormOrdering(opts.Order, "projects")
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
			match, err := filter.Matches(projectMap(v))
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

	filter, err := filtering.NewFilter(opts.Filter, apiFields)
	if err != nil {
		return ApiList{}, err
	}

	if order, err := gormOrdering(opts.Order, "apis"); err != nil {
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
			m, err := apiMap(v)
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

func apiMap(api models.Api) (map[string]interface{}, error) {
	labels, err := api.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":                api.Name(),
		"project_id":          api.ProjectID,
		"api_id":              api.ApiID,
		"display_name":        api.DisplayName,
		"description":         api.Description,
		"create_time":         api.CreateTime,
		"update_time":         api.UpdateTime,
		"availability":        api.Availability,
		"recommended_version": api.RecommendedVersion,
		"labels":              labels,
	}, nil
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

	filter, err := filtering.NewFilter(opts.Filter, versionFields)
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

	if order, err := gormOrdering(opts.Order, "versions"); err != nil {
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
			m, err := versionMap(v)
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

func versionMap(version models.Version) (map[string]interface{}, error) {
	labels, err := version.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":         version.Name(),
		"project_id":   version.ProjectID,
		"api_id":       version.ApiID,
		"version_id":   version.VersionID,
		"display_name": version.DisplayName,
		"description":  version.Description,
		"create_time":  version.CreateTime,
		"update_time":  version.UpdateTime,
		"state":        version.State,
		"labels":       labels,
		"primary_spec": version.PrimarySpec,
	}, nil
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

	filter, err := filtering.NewFilter(opts.Filter, specFields)
	if err != nil {
		return SpecList{}, err
	}

	op := c.db.WithContext(ctx).Select("specs.*").Table("specs").
		// select latest spec revision
		Joins(`join (?) latest
		ON specs.project_id = latest.project_id
		AND specs.api_id = latest.api_id
		AND specs.version_id = latest.version_id
		AND specs.spec_id = latest.spec_id
		AND specs.revision_id = latest.revision_id`, c.latestSpecRevisionsQuery(ctx)).
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

	if order, err := gormOrdering(opts.Order, "specs"); err != nil {
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
			m, err := specMap(v)
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

func specMap(spec models.Spec) (map[string]interface{}, error) {
	labels, err := spec.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":                 spec.Name(),
		"project_id":           spec.ProjectID,
		"api_id":               spec.ApiID,
		"version_id":           spec.VersionID,
		"spec_id":              spec.SpecID,
		"filename":             spec.FileName,
		"description":          spec.Description,
		"revision_id":          spec.RevisionID,
		"create_time":          spec.CreateTime,
		"revision_create_time": spec.RevisionCreateTime,
		"revision_update_time": spec.RevisionUpdateTime,
		"mime_type":            spec.MimeType,
		"size_bytes":           spec.SizeInBytes,
		"hash":                 spec.Hash,
		"source_uri":           spec.SourceURI,
		"labels":               labels,
	}, nil
}

func (c *Client) ListSpecRevisions(ctx context.Context, parent names.SpecRevision, opts PageOptions) (SpecList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	// Check existence of the deepest fully specified resource in the parent name.
	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" && parent.SpecID != "-" && parent.RevisionID != "-" {
		if _, err := c.GetSpecRevision(ctx, parent); err != nil {
			return SpecList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" && parent.SpecID != "-" {
		if _, err := c.GetSpec(ctx, parent.Spec()); err != nil {
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

	filter, err := filtering.NewFilter(opts.Filter, specFields)
	if err != nil {
		return SpecList{}, err
	}

	op := c.db.WithContext(ctx).
		Offset(token.Offset).
		Limit(int(opts.Size) + 1)

	if id := parent.ProjectID; id != "-" {
		op = op.Where("specs.project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("specs.api_id = ?", id)
	}
	if id := parent.VersionID; id != "-" {
		op = op.Where("specs.version_id = ?", id)
	}
	if id := parent.SpecID; id != "-" {
		op = op.Where("specs.spec_id = ?", id)
	}
	if id := parent.RevisionID; id != "-" && id != "" { // select specific spec revision
		op = op.Where("specs.revision_id = ?", id)
	}

	if order, err := gormOrdering(opts.Order, "specs"); err != nil {
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
			m, err := specMap(v)
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

	filter, err := filtering.NewFilter(opts.Filter, deploymentFields)
	if err != nil {
		return DeploymentList{}, err
	}

	op := c.db.WithContext(ctx).Select("deployments.*").Table("deployments").
		// select latest deployment revision
		Joins(`join (?) latest
		ON deployments.project_id = latest.project_id
		AND deployments.api_id = latest.api_id
		AND deployments.deployment_id = latest.deployment_id
		AND deployments.revision_id = latest.revision_id`, c.latestDeploymentRevisionsQuery(ctx)).
		Limit(limit(opts))

	if parent.ProjectID != "-" {
		op = op.Where("deployments.project_id = ?", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		op = op.Where("deployments.api_id = ?", parent.ApiID)
	}

	if order, err := gormOrdering(opts.Order, "deployments"); err != nil {
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
			m, err := deploymentMap(v)
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

func deploymentMap(deployment models.Deployment) (map[string]interface{}, error) {
	labels, err := deployment.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":                 deployment.Name(),
		"project_id":           deployment.ProjectID,
		"api_id":               deployment.ApiID,
		"deployment_id":        deployment.DeploymentID,
		"revision_id":          deployment.RevisionID,
		"display_name":         deployment.DisplayName,
		"description":          deployment.Description,
		"create_time":          deployment.CreateTime,
		"revision_create_time": deployment.RevisionCreateTime,
		"revision_update_time": deployment.RevisionUpdateTime,
		"api_spec_revision":    deployment.ApiSpecRevision,
		"endpoint_uri":         deployment.EndpointURI,
		"external_channel_uri": deployment.ExternalChannelURI,
		"intended_audience":    deployment.IntendedAudience,
		"access_guidance":      deployment.AccessGuidance,
		"labels":               labels,
	}, nil
}

func (c *Client) ListDeploymentRevisions(ctx context.Context, parent names.DeploymentRevision, opts PageOptions) (DeploymentList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return DeploymentList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return DeploymentList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	filter, err := filtering.NewFilter(opts.Filter, specFields)
	if err != nil {
		return DeploymentList{}, err
	}

	// Check existence of the deepest fully specified resource in the parent name.
	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.DeploymentID != "-" && parent.RevisionID != "-" {
		if _, err := c.GetDeploymentRevision(ctx, parent); err != nil {
			return DeploymentList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.DeploymentID != "-" {
		if _, err := c.GetDeployment(ctx, parent.Deployment()); err != nil {
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
		Offset(token.Offset).
		Limit(int(opts.Size) + 1)

	if id := parent.ProjectID; id != "-" {
		op = op.Where("deployments.project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("deployments.api_id = ?", id)
	}
	if id := parent.DeploymentID; id != "-" {
		op = op.Where("deployments.deployment_id = ?", id)
	}
	if id := parent.RevisionID; id != "-" && id != "" { // select specific spec revision
		op = op.Where("deployments.revision_id = ?", id)
	}

	if order, err := gormOrdering(opts.Order, "deployments"); err != nil {
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
			m, err := deploymentMap(v)
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

// ArtifactList contains a page of artifact resources.
type ArtifactList struct {
	Artifacts []models.Artifact
	Token     string
}

func (c *Client) ListSpecArtifacts(ctx context.Context, parent names.Spec, opts PageOptions) (ArtifactList, error) {
	return c.ListSpecRevisionArtifacts(ctx, parent.Revision(""), opts)
}

func (c *Client) ListSpecRevisionArtifacts(ctx context.Context, parent names.SpecRevision, opts PageOptions) (ArtifactList, error) {
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

	// if any explicit parent specified, ensure it exists
	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" && parent.SpecID != "-" {
		specAndRev := strings.Split(parent.SpecID, "@")
		if specAndRev[0] != "-" {
			if _, err := c.GetSpec(ctx, parent.Spec()); err != nil {
				return ArtifactList{}, err
			}
		}
		if len(specAndRev) > 1 && specAndRev[1] != "-" {
			if _, err := c.GetSpecRevision(ctx, parent); err != nil {
				return ArtifactList{}, err
			}
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

	op := c.db.WithContext(ctx)
	if id := parent.ProjectID; id != "-" {
		op = op.Where("artifacts.project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("artifacts.api_id = ?", id)
	}
	if id := parent.VersionID; id != "-" {
		op = op.Where("artifacts.version_id = ?", id)
	}
	if id := parent.SpecID; id != "-" {
		op = op.Where("artifacts.spec_id = ?", id)
	}
	if id := parent.RevisionID; id != "-" && id != "" { // select specific spec revision
		op = op.Where("artifacts.revision_id = ?", id)
	}
	orderTable := "artifacts"
	if id := parent.RevisionID; id == "" { // select latest spec revision
		op = op.Select("artifacts.*,revision_create_time").Table("artifacts").
			Where(`artifacts.deployment_id = ''`).
			Joins(`join (?) latest
			ON artifacts.project_id = latest.project_id
			AND artifacts.api_id = latest.api_id
			AND artifacts.version_id = latest.version_id
			AND artifacts.spec_id = latest.spec_id
			AND artifacts.revision_id = latest.revision_id`, c.latestSpecRevisionsQuery(ctx))
		orderTable = "revisioned_artifacts"
	}

	if order, err := gormOrdering(opts.Order, orderTable); err != nil {
		return ArtifactList{}, err
	} else {
		op = op.Order(order)
	}

	return c.listArtifacts(ctx, op, opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != "" && a.VersionID != "" && a.SpecID != "" && a.RevisionID != ""
	})
}

// This query only returns the most recent revision rows for each unique spec from the
// specs table. Additional criteria may be added to restrict this query further and it
// may be joined with dependant tables (eg. artifacts, blobs) to ensure that only the
// rows in those tables that are associated with the more recent spec revision are matched.
func (c *Client) latestSpecRevisionsQuery(ctx context.Context) *gorm.DB {
	inner := c.db.Table("specs").
		Select("*, row_number() over(partition by project_id,api_id,version_id,spec_id order by revision_create_time desc) as rn")
	return c.db.WithContext(ctx).
		Table("(?) as t", inner).
		Where("t.rn = 1")
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

	if order, err := gormOrdering(opts.Order, "artifacts"); err != nil {
		return ArtifactList{}, err
	} else {
		op = op.Order(order)
	}

	return c.listArtifacts(ctx, op, opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != "" && a.VersionID != ""
	})
}

func (c *Client) ListDeploymentArtifacts(ctx context.Context, parent names.Deployment, opts PageOptions) (ArtifactList, error) {
	return c.ListDeploymentRevisionArtifacts(ctx, parent.Revision(""), opts)
}

func (c *Client) ListDeploymentRevisionArtifacts(ctx context.Context, parent names.DeploymentRevision, opts PageOptions) (ArtifactList, error) {
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

	// if any explicit parent specified, ensure it exists
	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.DeploymentID != "-" {
		deploymentAndRev := strings.Split(parent.DeploymentID, "@")
		if deploymentAndRev[0] != "-" {
			if _, err := c.GetDeployment(ctx, parent.Deployment()); err != nil {
				return ArtifactList{}, err
			}
		}
		if len(deploymentAndRev) > 1 && deploymentAndRev[1] != "-" {
			if _, err := c.GetDeploymentRevision(ctx, parent); err != nil {
				return ArtifactList{}, err
			}
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

	op := c.db.WithContext(ctx)
	if id := parent.ProjectID; id != "-" {
		op = op.Where("artifacts.project_id = ?", id)
	}
	if id := parent.ApiID; id != "-" {
		op = op.Where("artifacts.api_id = ?", id)
	}
	if id := parent.DeploymentID; id != "-" {
		op = op.Where("artifacts.deployment_id = ?", id)
	}
	if id := parent.RevisionID; id != "-" && id != "" { // select specific deployment revision
		op = op.Where("artifacts.revision_id = ?", id)
	}
	orderTable := "artifacts"
	if id := parent.RevisionID; id == "" { // select latest deployment revision
		op = op.Select("artifacts.*,revision_create_time").Table("artifacts").
			Where(`artifacts.spec_id = ''`).
			Joins(`join (?) latest
			ON artifacts.project_id = latest.project_id
			AND artifacts.api_id = latest.api_id
			AND artifacts.deployment_id = latest.deployment_id
			AND artifacts.revision_id = latest.revision_id`, c.latestDeploymentRevisionsQuery(ctx))
		orderTable = "revisioned_artifacts"
	}

	if order, err := gormOrdering(opts.Order, orderTable); err != nil {
		return ArtifactList{}, err
	} else {
		op = op.Order(order)
	}

	return c.listArtifacts(ctx, op, opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != "" && a.DeploymentID != "" && a.RevisionID != ""
	})
}

// This query only returns the most recent revision rows for each unique deployment from the
// deployment table. Additional criteria may be added to restrict this query further and it
// may be joined with dependant tables (eg. artifacts, blobs) to ensure that only the
// rows in those tables that are associated with the more recent deployment revision are matched.
func (c *Client) latestDeploymentRevisionsQuery(ctx context.Context) *gorm.DB {
	inner := c.db.Table("deployments").
		Select("*, row_number() over(partition by project_id,api_id,deployment_id order by revision_create_time desc) as rn")
	return c.db.WithContext(ctx).
		Table("(?) as t", inner).
		Where("t.rn = 1")
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

	if order, err := gormOrdering(opts.Order, "artifacts"); err != nil {
		return ArtifactList{}, err
	} else {
		op = op.Order(order)
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

	if order, err := gormOrdering(opts.Order, "artifacts"); err != nil {
		return ArtifactList{}, err
	} else {
		op = op.Order(order)
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

	filter, err := filtering.NewFilter(opts.Filter, artifactFields)
	if err != nil {
		return ArtifactList{}, err
	}

	response := ArtifactList{
		Artifacts: make([]models.Artifact, 0, opts.Size),
	}

	for {
		var page []models.Artifact
		op.Limit(limit(opts))
		err := op.Offset(token.Offset).Find(&page).Error

		if err != nil {
			return ArtifactList{}, grpcErrorForDBError(ctx, errors.Wrapf(err, "find %#v", token))
		} else if len(page) == 0 {
			break
		}

		for _, v := range page {
			m := artifactMap(v)
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

func artifactMap(artifact models.Artifact) map[string]interface{} {
	return map[string]interface{}{
		"name":          artifact.Name(),
		"project_id":    artifact.ProjectID,
		"api_id":        artifact.ApiID,
		"version_id":    artifact.VersionID,
		"spec_id":       artifact.SpecID,
		"deployment_id": artifact.DeploymentID,
		"artifact_id":   artifact.ArtifactID,
		"create_time":   artifact.CreateTime,
		"update_time":   artifact.UpdateTime,
		"mime_type":     artifact.MimeType,
		"size_bytes":    artifact.SizeInBytes,
	}
}
