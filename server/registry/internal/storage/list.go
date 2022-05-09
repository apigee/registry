// Copyright 2022 Google LLC. All Rights Reservec.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or impliec.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"context"

	"github.com/apigee/registry/server/registry/internal/storage/filtering"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

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

var projectFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "display_name", Type: filtering.String},
	{Name: "description", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "update_time", Type: filtering.Timestamp},
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

	filter, err := filtering.NewFilter(opts.Filter, projectFields)
	if err != nil {
		return ProjectList{}, err
	}

	response := ProjectList{
		Projects: make([]models.Project, 0, opts.Size),
	}

	for {
		lock()
		var page []models.Project
		op := c.db.WithContext(ctx).Order("key").Limit(limit(opts))
		err := op.Offset(token.Offset).Find(&page).Error
		unlock()

		if err != nil {
			return ProjectList{}, status.Error(codes.Internal, err.Error())
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

var apiFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "api_id", Type: filtering.String},
	{Name: "display_name", Type: filtering.String},
	{Name: "description", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "update_time", Type: filtering.Timestamp},
	{Name: "availability", Type: filtering.String},
	{Name: "recommended_version", Type: filtering.String},
	{Name: "recommended_deployment", Type: filtering.String},
	{Name: "labels", Type: filtering.StringMap},
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

	op := c.db.WithContext(ctx).
		Order("key").
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

	response := ApiList{
		Apis: make([]models.Api, 0, opts.Size),
	}

	for {
		lock()
		var page []models.Api
		err := op.Offset(token.Offset).Find(&page).Error
		unlock()

		if err != nil {
			return ApiList{}, status.Error(codes.Internal, err.Error())
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

var versionFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "api_id", Type: filtering.String},
	{Name: "version_id", Type: filtering.String},
	{Name: "display_name", Type: filtering.String},
	{Name: "description", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "update_time", Type: filtering.Timestamp},
	{Name: "state", Type: filtering.String},
	{Name: "labels", Type: filtering.StringMap},
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

	op := c.db.WithContext(ctx).
		Order("key").
		Limit(limit(opts))

	if parent.ProjectID != "-" {
		op = op.Where("project_id = ?", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		op = op.Where("api_id = ?", parent.ApiID)
	}

	response := VersionList{
		Versions: make([]models.Version, 0, opts.Size),
	}

	for {
		lock()
		var page []models.Version
		err := op.Offset(token.Offset).Find(&page).Error
		unlock()

		if err != nil {
			return VersionList{}, status.Error(codes.Internal, err.Error())
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
		"version_id":   version.VersionID,
		"display_name": version.DisplayName,
		"description":  version.Description,
		"create_time":  version.CreateTime,
		"update_time":  version.UpdateTime,
		"state":        version.State,
		"labels":       labels,
	}, nil
}

// SpecList contains a page of spec resources.
type SpecList struct {
	Specs []models.Spec
	Token string
}

var specFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "api_id", Type: filtering.String},
	{Name: "version_id", Type: filtering.String},
	{Name: "spec_id", Type: filtering.String},
	{Name: "filename", Type: filtering.String},
	{Name: "description", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "revision_create_time", Type: filtering.Timestamp},
	{Name: "revision_update_time", Type: filtering.Timestamp},
	{Name: "mime_type", Type: filtering.String},
	{Name: "size_bytes", Type: filtering.Int},
	{Name: "source_uri", Type: filtering.String},
	{Name: "labels", Type: filtering.StringMap},
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
		Order("key").
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

	response := SpecList{
		Specs: make([]models.Spec, 0, opts.Size),
	}

	for {
		lock()
		var page []models.Spec
		err := op.Offset(token.Offset).Find(&page).Error
		unlock()

		if err != nil {
			return SpecList{}, status.Error(codes.Internal, err.Error())
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

	lock()
	err = op.Find(&response.Specs).Error
	unlock()

	if err != nil {
		return SpecList{}, status.Error(codes.Internal, err.Error())
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

var deploymentFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "api_id", Type: filtering.String},
	{Name: "deployment_id", Type: filtering.String},
	{Name: "display_name", Type: filtering.String},
	{Name: "description", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "revision_create_time", Type: filtering.Timestamp},
	{Name: "revision_update_time", Type: filtering.Timestamp},
	{Name: "api_spec_revision", Type: filtering.String},
	{Name: "endpoint_uri", Type: filtering.String},
	{Name: "external_channel_uri", Type: filtering.String},
	{Name: "intended_audience", Type: filtering.String},
	{Name: "access_guidance", Type: filtering.String},
	{Name: "labels", Type: filtering.StringMap},
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
		Order("key").
		Limit(limit(opts))

	if parent.ProjectID != "-" {
		op = op.Where("deployments.project_id = ?", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		op = op.Where("deployments.api_id = ?", parent.ApiID)
	}

	response := DeploymentList{
		Deployments: make([]models.Deployment, 0, opts.Size),
	}

	for {
		lock()
		var page []models.Deployment
		err := op.Offset(token.Offset).Find(&page).Error
		unlock()

		if err != nil {
			return DeploymentList{}, status.Error(codes.Internal, err.Error())
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

	lock()
	err = op.Find(&response.Deployments).Error
	unlock()

	if err != nil {
		return DeploymentList{}, status.Error(codes.Internal, err.Error())
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

var artifactFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "api_id", Type: filtering.String},
	{Name: "version_id", Type: filtering.String},
	{Name: "spec_id", Type: filtering.String},
	{Name: "artifact_id", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "update_time", Type: filtering.Timestamp},
	{Name: "mime_type", Type: filtering.String},
	{Name: "size_bytes", Type: filtering.Int},
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

	return c.listArtifacts(op, opts, func(a *models.Artifact) bool {
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

	return c.listArtifacts(op, opts, func(a *models.Artifact) bool {
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

	return c.listArtifacts(op, opts, func(a *models.Artifact) bool {
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

	return c.listArtifacts(op, opts, func(a *models.Artifact) bool {
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

	return c.listArtifacts(op, opts, func(a *models.Artifact) bool {
		return a.ProjectID != ""
	})
}

func (c *Client) listArtifacts(op *gorm.DB, opts PageOptions, include func(*models.Artifact) bool) (ArtifactList, error) {
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
		lock()
		var page []models.Artifact
		op = op.Order("key").Limit(limit(opts))
		err := op.Offset(token.Offset).Find(&page).Error
		unlock()

		if err != nil {
			return ArtifactList{}, status.Error(codes.Internal, err.Error())
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
	}

	return response, nil
}

func artifactMap(artifact models.Artifact) map[string]interface{} {
	return map[string]interface{}{
		"name":        artifact.Name(),
		"project_id":  artifact.ProjectID,
		"api_id":      artifact.ApiID,
		"version_id":  artifact.VersionID,
		"spec_id":     artifact.SpecID,
		"artifact_id": artifact.ArtifactID,
		"create_time": artifact.CreateTime,
		"update_time": artifact.UpdateTime,
		"mime_type":   artifact.MimeType,
		"size_bytes":  artifact.SizeInBytes,
	}
}
