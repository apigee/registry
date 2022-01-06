// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Artifact 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the artifactific language governing permissions and
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

func (d *Client) ListSpecArtifacts(ctx context.Context, parent names.Spec, opts PageOptions) (ArtifactList, error) {
	q := d.NewQuery(gorm.ArtifactEntityName)
	q = q.Require("DeploymentID", "")

	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	q = q.ApplyOffset(token.Offset)

	if id := parent.ProjectID; id != "-" {
		q = q.Require("ProjectID", id)
	}
	if id := parent.ApiID; id != "-" {
		q = q.Require("ApiID", id)
	}
	if id := parent.VersionID; id != "-" {
		q = q.Require("VersionID", id)
	}
	if id := parent.SpecID; id != "-" {
		q = q.Require("SpecID", id)
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" && parent.SpecID != "-" {
		if _, err := d.GetSpec(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" && parent.SpecID == "-" {
		if _, err := d.GetVersion(ctx, parent.Version()); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID == "-" && parent.SpecID == "-" {
		if _, err := d.GetApi(ctx, parent.Api()); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.VersionID == "-" && parent.SpecID == "-" {
		if _, err := d.GetProject(ctx, parent.Project()); err != nil {
			return ArtifactList{}, err
		}
	}

	return d.listArtifacts(ctx, d.Run(ctx, q), opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != "" && a.VersionID != "" && a.SpecID != ""
	})
}

func (d *Client) ListVersionArtifacts(ctx context.Context, parent names.Version, opts PageOptions) (ArtifactList, error) {
	q := d.NewQuery(gorm.ArtifactEntityName)
	q = q.Require("SpecID", "")
	q = q.Require("DeploymentID", "")

	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	q = q.ApplyOffset(token.Offset)

	if id := parent.ProjectID; id != "-" {
		q = q.Require("ProjectID", id)
	}
	if id := parent.ApiID; id != "-" {
		q = q.Require("ApiID", id)
	}
	if id := parent.VersionID; id != "-" {
		q = q.Require("VersionID", id)
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID != "-" {
		if _, err := d.GetVersion(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.VersionID == "-" {
		if _, err := d.GetApi(ctx, parent.Api()); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.VersionID == "-" {
		if _, err := d.GetProject(ctx, parent.Project()); err != nil {
			return ArtifactList{}, err
		}
	}

	return d.listArtifacts(ctx, d.Run(ctx, q), opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != "" && a.VersionID != ""
	})
}

func (d *Client) ListDeploymentArtifacts(ctx context.Context, parent names.Deployment, opts PageOptions) (ArtifactList, error) {
	q := d.NewQuery(gorm.ArtifactEntityName)
	q = q.Require("VersionID", "")

	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	q = q.ApplyOffset(token.Offset)

	if id := parent.ProjectID; id != "-" {
		q = q.Require("ProjectID", id)
	}
	if id := parent.ApiID; id != "-" {
		q = q.Require("ApiID", id)
	}
	if id := parent.DeploymentID; id != "-" {
		q = q.Require("DeploymentID", id)
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" && parent.DeploymentID != "-" {
		if _, err := d.GetDeployment(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID != "-" && parent.DeploymentID == "-" {
		if _, err := d.GetApi(ctx, parent.Api()); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" && parent.DeploymentID == "-" {
		if _, err := d.GetProject(ctx, parent.Project()); err != nil {
			return ArtifactList{}, err
		}
	}

	return d.listArtifacts(ctx, d.Run(ctx, q), opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != "" && a.DeploymentID != ""
	})
}

func (d *Client) ListApiArtifacts(ctx context.Context, parent names.Api, opts PageOptions) (ArtifactList, error) {
	q := d.NewQuery(gorm.ArtifactEntityName)
	q = q.Require("DeploymentID", "")
	q = q.Require("VersionID", "")

	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	q = q.ApplyOffset(token.Offset)

	if id := parent.ProjectID; id != "-" {
		q = q.Require("ProjectID", id)
	}
	if id := parent.ApiID; id != "-" {
		q = q.Require("ApiID", id)
	}

	if parent.ProjectID != "-" && parent.ApiID != "-" {
		if _, err := d.GetApi(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	} else if parent.ProjectID != "-" && parent.ApiID == "-" {
		if _, err := d.GetProject(ctx, parent.Project()); err != nil {
			return ArtifactList{}, err
		}
	}

	return d.listArtifacts(ctx, d.Run(ctx, q), opts, func(a *models.Artifact) bool {
		return a.ProjectID != "" && a.ApiID != ""
	})
}

func (d *Client) ListProjectArtifacts(ctx context.Context, parent names.Project, opts PageOptions) (ArtifactList, error) {
	q := d.NewQuery(gorm.ArtifactEntityName)
	q = q.Require("ApiID", "")

	token, err := decodeToken(opts.Token)
	if err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ArtifactList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	q = q.ApplyOffset(token.Offset)

	if id := parent.ProjectID; id != "-" {
		q = q.Require("ProjectID", id)
		if _, err := d.GetProject(ctx, parent); err != nil {
			return ArtifactList{}, err
		}
	}

	return d.listArtifacts(ctx, d.Run(ctx, q), opts, func(a *models.Artifact) bool {
		return a.ProjectID != ""
	})
}

func (d *Client) listArtifacts(ctx context.Context, it *gorm.Iterator, opts PageOptions, include func(*models.Artifact) bool) (ArtifactList, error) {
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

	artifact := new(models.Artifact)
	for _, err = it.Next(artifact); err == nil; _, err = it.Next(artifact) {
		artifactMap, err := artifactMap(*artifact)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(artifactMap)
		if err != nil {
			return response, err
		} else if !match || !include(artifact) {
			token.Offset++
			continue
		} else if len(response.Artifacts) == int(opts.Size) {
			break
		}

		response.Artifacts = append(response.Artifacts, *artifact)
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

func (d *Client) SaveArtifact(ctx context.Context, artifact *models.Artifact) error {
	k := d.NewKey(gorm.ArtifactEntityName, artifact.Name())
	if _, err := d.Put(ctx, k, artifact); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *Client) SaveArtifactContents(ctx context.Context, artifact *models.Artifact, contents []byte) error {
	blob := models.NewBlobForArtifact(artifact, contents)
	k := d.NewKey(gorm.BlobEntityName, artifact.Name())
	if _, err := d.Put(ctx, k, blob); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *Client) GetArtifact(ctx context.Context, name names.Artifact) (*models.Artifact, error) {
	artifact := new(models.Artifact)
	k := d.NewKey(gorm.ArtifactEntityName, name.String())
	if err := d.Get(ctx, k, artifact); d.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "artifact %q not found in database", name)
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return artifact, nil
}

func (d *Client) GetArtifactContents(ctx context.Context, name names.Artifact) (*models.Blob, error) {
	blob := new(models.Blob)
	k := d.NewKey(gorm.BlobEntityName, name.String())
	if err := d.Get(ctx, k, blob); d.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "artifact contents %q not found", name)
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return blob, nil
}

func (d *Client) DeleteArtifact(ctx context.Context, name names.Artifact) error {
	for _, entityName := range []string{
		gorm.ArtifactEntityName,
		gorm.BlobEntityName,
	} {
		q := d.NewQuery(entityName)
		q = q.Require("ProjectID", name.ProjectID())
		q = q.Require("ApiID", name.ApiID())
		q = q.Require("VersionID", name.VersionID())
		q = q.Require("SpecID", name.SpecID())
		q = q.Require("ArtifactID", name.ArtifactID())
		if err := d.Delete(ctx, q); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}
