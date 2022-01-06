// Copyright 2021 Google LLC. All Rights Reserved.
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
// See the License for the revisionific language governing permissions and
// limitations under the License.

package storage

import (
	"context"

	"github.com/apigee/registry/server/registry/internal/storage/gorm"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *Client) ListDeploymentRevisions(ctx context.Context, parent names.Deployment, opts PageOptions) (DeploymentList, error) {
	q := d.NewQuery(gorm.DeploymentEntityName)
	q = q.Require("ProjectID", parent.ProjectID)
	q = q.Require("ApiID", parent.ApiID)
	q = q.Require("DeploymentID", parent.DeploymentID)
	q = q.Descending("RevisionCreateTime")

	token, err := decodeToken(opts.Token)
	if err != nil {
		return DeploymentList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return DeploymentList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	}

	q = q.ApplyOffset(token.Offset)

	it := d.Run(ctx, q)
	response := DeploymentList{
		Deployments: make([]models.Deployment, 0, opts.Size),
	}

	revision := new(models.Deployment)
	for _, err = it.Next(revision); err == nil; _, err = it.Next(revision) {
		token.Offset++

		response.Deployments = append(response.Deployments, *revision)
		if len(response.Deployments) == int(opts.Size) {
			break
		}
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

func (d *Client) SaveDeploymentRevision(ctx context.Context, revision *models.Deployment) error {
	return d.Client.SaveDeploymentRevision(ctx, revision)
}

func (d *Client) GetDeploymentRevision(ctx context.Context, name names.DeploymentRevision) (*models.Deployment, error) {
	return d.Client.GetDeploymentRevision(ctx, name)
}

func (d *Client) DeleteDeploymentRevision(ctx context.Context, name names.DeploymentRevision) error {
	return d.Client.DeleteDeploymentRevision(ctx, name)
}

func (d *Client) SaveDeploymentRevisionTag(ctx context.Context, tag *models.DeploymentRevisionTag) error {
	return d.Client.SaveDeploymentRevisionTag(ctx, tag)
}
