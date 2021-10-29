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

	"github.com/apigee/registry/server/registry/internal/storage/filtering"
	"github.com/apigee/registry/server/registry/internal/storage/gorm"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

func (d *Client) ListDeployments(ctx context.Context, parent names.Api, opts PageOptions) (DeploymentList, error) {
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
		if _, err := d.GetApi(ctx, parent); err != nil {
			return DeploymentList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" {
		if _, err := d.GetProject(ctx, parent.Project()); err != nil {
			return DeploymentList{}, err
		}
	}

	filter, err := filtering.NewFilter(opts.Filter, deploymentFields)
	if err != nil {
		return DeploymentList{}, err
	}

	it := d.GetRecentDeploymentRevisions(ctx, token.Offset, parent.ProjectID, parent.ApiID)
	response := DeploymentList{
		Deployments: make([]models.Deployment, 0, opts.Size),
	}

	deployment := new(models.Deployment)
	for _, err = it.Next(deployment); err == nil; _, err = it.Next(deployment) {
		deploymentMap, err := deploymentMap(*deployment)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(deploymentMap)
		if err != nil {
			return response, err
		} else if !match {
			token.Offset++
			continue
		} else if len(response.Deployments) == int(opts.Size) {
			break
		}

		response.Deployments = append(response.Deployments, *deployment)
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

func (d *Client) GetDeployment(ctx context.Context, name names.Deployment) (*models.Deployment, error) {
	normal := name.Normal()
	q := d.NewQuery(gorm.DeploymentEntityName)
	q = q.Require("ProjectID", normal.ProjectID)
	q = q.Require("ApiID", normal.ApiID)
	q = q.Require("DeploymentID", normal.DeploymentID)
	q = q.Descending("RevisionCreateTime")

	it := d.Run(ctx, q)
	deployment := &models.Deployment{}
	if _, err := it.Next(deployment); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return deployment, nil
}

func (d *Client) DeleteDeployment(ctx context.Context, name names.Deployment) error {
	for _, entityName := range []string{
		gorm.DeploymentEntityName,
		gorm.DeploymentRevisionTagEntityName,
		gorm.ArtifactEntityName,
	} {
		q := d.NewQuery(entityName)
		q = q.Require("ProjectID", name.ProjectID)
		q = q.Require("ApiID", name.ApiID)
		q = q.Require("DeploymentID", name.DeploymentID)
		if err := d.Delete(ctx, q); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

func (d *Client) GetDeploymentTags(ctx context.Context, name names.Deployment) ([]*models.DeploymentRevisionTag, error) {
	q := d.NewQuery(gorm.DeploymentRevisionTagEntityName)
	q = q.Require("ProjectID", name.ProjectID)
	q = q.Require("ApiID", name.ApiID)
	if name.DeploymentID != "-" {
		q = q.Require("DeploymentID", name.DeploymentID)
	}

	var (
		tags = make([]*models.DeploymentRevisionTag, 0)
		tag  = new(models.DeploymentRevisionTag)
		it   = d.Run(ctx, q)
		err  error
	)

	for _, err = it.Next(tag); err == nil; _, err = it.Next(tag) {
		tags = append(tags, tag)
	}
	if err != nil && err != iterator.Done {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return tags, nil
}
