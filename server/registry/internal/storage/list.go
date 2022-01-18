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

	lock()
	var projects []models.Project
	err = c.db.
		Order("key").
		Offset(token.Offset).
		Limit(100000).
		Find(&projects).Error
	unlock()

	if err != nil {
		return ProjectList{}, status.Error(codes.Internal, err.Error())
	}

	response := ProjectList{
		Projects: make([]models.Project, 0, opts.Size),
	}

	for _, project := range projects {
		match, err := filter.Matches(projectMap(project))
		if err != nil {
			return response, err
		} else if !match {
			token.Offset++
			continue
		}

		if len(response.Projects) < int(opts.Size) {
			response.Projects = append(response.Projects, project)
			token.Offset++
		} else if len(response.Projects) == int(opts.Size) {
			response.Token, err = encodeToken(token)
			if err != nil {
				return response, status.Error(codes.Internal, err.Error())
			}
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

	op := c.db.
		Order("key").
		Offset(token.Offset).
		Limit(100000)

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

	lock()
	var apis []models.Api
	_ = op.Find(&apis).Error
	unlock()

	if err != nil {
		return ApiList{}, status.Error(codes.Internal, err.Error())
	}

	response := ApiList{
		Apis: make([]models.Api, 0, opts.Size),
	}

	for _, api := range apis {
		apiMap, err := apiMap(api)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(apiMap)
		if err != nil {
			return response, err
		} else if !match {
			token.Offset++
			continue
		}

		if len(response.Apis) < int(opts.Size) {
			response.Apis = append(response.Apis, api)
			token.Offset++
		} else if len(response.Apis) == int(opts.Size) {
			response.Token, err = encodeToken(token)
			if err != nil {
				return response, status.Error(codes.Internal, err.Error())
			}
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

	op := c.db.
		Order("key").
		Offset(token.Offset).
		Limit(100000)

	if parent.ProjectID != "-" {
		op = op.Where("project_id = ?", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		op = op.Where("api_id = ?", parent.ApiID)
	}

	lock()
	var versions []models.Version
	_ = op.Find(&versions).Error
	unlock()

	if err != nil {
		return VersionList{}, status.Error(codes.Internal, err.Error())
	}

	response := VersionList{
		Versions: make([]models.Version, 0, opts.Size),
	}

	for _, version := range versions {
		versionMap, err := versionMap(version)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(versionMap)
		if err != nil {
			return response, err
		} else if !match {
			token.Offset++
			continue
		}

		if len(response.Versions) < int(opts.Size) {
			response.Versions = append(response.Versions, version)
			token.Offset++
		} else if len(response.Versions) == int(opts.Size) {
			response.Token, err = encodeToken(token)
			if err != nil {
				return response, status.Error(codes.Internal, err.Error())
			}
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
	op := c.db.Select("specs.*").
		Table("specs").
		// Join missing columns that couldn't be selected in the subquery.
		Joins("JOIN (?) AS grp ON specs.project_id = grp.project_id AND specs.api_id = grp.api_id AND specs.version_id = grp.version_id AND specs.spec_id = grp.spec_id AND specs.revision_create_time = grp.recent_create_time",
			// Select spec names and only their most recent revision_create_time
			// This query cannot select all the columns we want.
			// See: https://stackoverflow.com/questions/7745609/sql-select-only-rows-with-max-value-on-a-column
			c.db.Select("project_id, api_id, version_id, spec_id, MAX(revision_create_time) AS recent_create_time").
				Table("specs").
				Group("project_id, api_id, version_id, spec_id")).
		Order("key").
		Offset(token.Offset).
		Limit(100000)

	if parent.ProjectID != "-" {
		op = op.Where("specs.project_id = ?", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		op = op.Where("specs.api_id = ?", parent.ApiID)
	}
	if parent.VersionID != "-" {
		op = op.Where("specs.version_id = ?", parent.VersionID)
	}

	lock()
	var specs []models.Spec
	_ = op.Scan(&specs).Error
	unlock()

	if err != nil {
		return SpecList{}, status.Error(codes.Internal, err.Error())
	}

	response := SpecList{
		Specs: make([]models.Spec, 0, opts.Size),
	}

	for _, spec := range specs {
		specMap, err := specMap(spec)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(specMap)
		if err != nil {
			return response, err
		} else if !match {
			token.Offset++
			continue
		}

		if len(response.Specs) < int(opts.Size) {
			response.Specs = append(response.Specs, spec)
			token.Offset++
		} else if len(response.Specs) == int(opts.Size) {
			response.Token, err = encodeToken(token)
			if err != nil {
				return response, status.Error(codes.Internal, err.Error())
			}
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

func (c *Client) ListSpecRevisions(ctx context.Context, parent names.Spec, opts PageOptions) (SpecList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return SpecList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	response := SpecList{
		Specs: make([]models.Spec, 0, opts.Size),
	}

	lock()
	err = c.db.
		Where("project_id = ?", parent.ProjectID).
		Where("api_id = ?", parent.ApiID).
		Where("version_id = ?", parent.VersionID).
		Where("spec_id = ?", parent.SpecID).
		Order("revision_create_time desc").
		Offset(token.Offset).
		Limit(int(opts.Size) + 1).
		Find(&response.Specs).Error
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
	op := c.db.Select("deployments.*").
		Table("deployments").
		// Join missing columns that couldn't be selected in the subquery.
		Joins("JOIN (?) AS grp ON deployments.project_id = grp.project_id AND deployments.api_id = grp.api_id AND deployments.deployment_id = grp.deployment_id AND deployments.revision_create_time = grp.recent_create_time",
			// Select deployment names and only their most recent revision_create_time
			// This query cannot select all the columns we want.
			// See: https://stackoverflow.com/questions/7745609/sql-select-only-rows-with-max-value-on-a-column
			c.db.Select("project_id, api_id, deployment_id, MAX(revision_create_time) AS recent_create_time").
				Table("deployments").
				Group("project_id, api_id, deployment_id")).
		Order("key").
		Offset(token.Offset).
		Limit(100000)

	if parent.ProjectID != "-" {
		op = op.Where("deployments.project_id = ?", parent.ProjectID)
	}
	if parent.ApiID != "-" {
		op = op.Where("deployments.api_id = ?", parent.ApiID)
	}

	lock()
	var deployments []models.Deployment
	_ = op.Scan(&deployments).Error
	unlock()

	if err != nil {
		return DeploymentList{}, status.Error(codes.Internal, err.Error())
	}

	response := DeploymentList{
		Deployments: make([]models.Deployment, 0, opts.Size),
	}

	for _, deployment := range deployments {
		deploymentMap, err := deploymentMap(deployment)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(deploymentMap)
		if err != nil {
			return response, err
		} else if !match {
			token.Offset++
			continue
		}

		if len(response.Deployments) < int(opts.Size) {
			response.Deployments = append(response.Deployments, deployment)
			token.Offset++
		} else if len(response.Deployments) == int(opts.Size) {
			response.Token, err = encodeToken(token)
			if err != nil {
				return response, status.Error(codes.Internal, err.Error())
			}
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

func (c *Client) ListDeploymentRevisions(ctx context.Context, parent names.Deployment, opts PageOptions) (DeploymentList, error) {
	token, err := decodeToken(opts.Token)
	if err != nil {
		return DeploymentList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	response := DeploymentList{
		Deployments: make([]models.Deployment, 0, opts.Size),
	}

	lock()
	err = c.db.
		Where("project_id = ?", parent.ProjectID).
		Where("api_id = ?", parent.ApiID).
		Where("deployment_id = ?", parent.DeploymentID).
		Order("revision_create_time desc").
		Offset(token.Offset).
		Limit(int(opts.Size) + 1).
		Find(&response.Deployments).Error
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

	op := c.db.Where(`deployment_id = ''`)
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

	op := c.db.Where(`deployment_id = ''`).
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

	op := c.db.Where(`version_id = ''`).
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

	if parent.ProjectID != "-" && parent.ApiID != "-" {
		if _, err := c.GetApi(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" {
		if _, err := c.GetProject(ctx, parent.Project()); err != nil {
			return ArtifactList{}, err
		}
	}

	op := c.db.Where(`deployment_id = ''`).
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

	op := c.db.Where(`api_id = ''`).
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

	filter, err := filtering.NewFilter(opts.Filter, artifactFields)
	if err != nil {
		return ArtifactList{}, err
	}

	lock()
	var artifacts []models.Artifact
	_ = op.Offset(token.Offset).
		Limit(100000).
		Find(&artifacts).Error
	unlock()

	if err != nil {
		return ArtifactList{}, status.Error(codes.Internal, err.Error())
	}

	response := ArtifactList{
		Artifacts: make([]models.Artifact, 0, opts.Size),
	}

	for _, artifact := range artifacts {
		artifactMap, err := artifactMap(artifact)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(artifactMap)
		if err != nil {
			return response, err
		} else if !match || !include(&artifact) {
			token.Offset++
			continue
		}

		if len(response.Artifacts) < int(opts.Size) {
			response.Artifacts = append(response.Artifacts, artifact)
			token.Offset++
		} else if len(response.Artifacts) == int(opts.Size) {
			response.Token, err = encodeToken(token)
			if err != nil {
				return response, status.Error(codes.Internal, err.Error())
			}
			break
		}
	}

	return response, nil
}

func artifactMap(artifact models.Artifact) (map[string]interface{}, error) {
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
	}, nil
}

func (c *Client) GetSpecTags(ctx context.Context, name names.Spec) ([]models.SpecRevisionTag, error) {
	op := c.db.Where("project_id = ?", name.ProjectID).
		Where("api_id = ?", name.ApiID).
		Where("version_id = ?", name.VersionID)
	if name.SpecID != "-" {
		op = op.Where("spec_id = ?", name.SpecID)
	}

	lock()
	defer unlock()

	tags := make([]models.SpecRevisionTag, 0)
	err := op.Limit(100000).Find(&tags).Error
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return tags, nil
}

func (c *Client) GetDeploymentTags(ctx context.Context, name names.Deployment) ([]models.DeploymentRevisionTag, error) {
	op := c.db.Where("project_id = ?", name.ProjectID).
		Where("api_id = ?", name.ApiID)
	if name.DeploymentID != "-" {
		op = op.Where("deployment_id = ?", name.DeploymentID)
	}

	lock()
	defer unlock()

	tags := make([]models.DeploymentRevisionTag, 0)
	err := op.Limit(100000).Find(&tags).Error
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return tags, nil
}
