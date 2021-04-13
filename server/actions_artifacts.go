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

package server

import (
	"context"
	"fmt"
	"log"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/models"
	storage "github.com/apigee/registry/server/storage"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateArtifact handles the corresponding API request.
func (s *RegistryServer) CreateArtifact(ctx context.Context, req *rpc.CreateArtifactRequest) (*rpc.Artifact, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	artifact, err := models.NewArtifactFromParentAndArtifactID(req.GetParent(), req.GetArtifactId())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// create a blob for the artifact contents
	blob := models.NewBlobForArtifact(artifact, nil)
	err = artifact.Update(req.GetArtifact(), blob)
	k := client.NewKey(models.ArtifactEntityName, artifact.ResourceName())
	// fail if artifact already exists
	existingArtifact := &models.Artifact{}
	err = client.Get(ctx, k, existingArtifact)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, artifact.ResourceName()+" already exists")
	}
	artifact.CreateTime = artifact.UpdateTime
	k, err = client.Put(ctx, k, artifact)
	if err != nil {
		return nil, internalError(err)
	}
	if blob != nil {
		k2 := client.NewKey(models.BlobEntityName, artifact.ResourceName())
		_, err = client.Put(ctx,
			k2,
			blob)
	}
	if err != nil {
		log.Printf("save blob error %+v", err)
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_CREATED, artifact.ResourceName())
	return artifact.Message(blob)
}

// DeleteArtifact handles the corresponding API request.
func (s *RegistryServer) DeleteArtifact(ctx context.Context, req *rpc.DeleteArtifactRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	// Validate name and create dummy artifact (we just need the ID fields).
	_, err = models.NewArtifactFromResourceName(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete the artifact.
	k := client.NewKey(models.ArtifactEntityName, req.GetName())
	err = client.Delete(ctx, k)
	// Delete any blob associated with the artifact.
	k2 := client.NewKey(models.BlobEntityName, req.GetName())
	err = client.Delete(ctx, k2)
	s.notify(rpc.Notification_DELETED, req.GetName())
	return &empty.Empty{}, internalError(err)
}

// GetArtifact handles the corresponding API request.
func (s *RegistryServer) GetArtifact(ctx context.Context, req *rpc.GetArtifactRequest) (*rpc.Artifact, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	artifact, err := models.NewArtifactFromResourceName(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ArtifactEntityName, artifact.ResourceName())
	err = client.Get(ctx, k, artifact)
	var blob *models.Blob
	if req.GetView() == rpc.View_FULL {
		blob, _ = fetchBlobForArtifact(ctx, client, artifact)
	}
	if client.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return artifact.Message(blob)
}

// ListArtifacts handles the corresponding API request.
func (s *RegistryServer) ListArtifacts(ctx context.Context, req *rpc.ListArtifactsRequest) (*rpc.ListArtifactsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	q := client.NewQuery(models.ArtifactEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	p, err := models.NewArtifactFromParentAndArtifactID(req.GetParent(), "-")
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if p.ProjectID != "-" {
		q = q.Require("ProjectID", p.ProjectID)
	}
	if p.ApiID != "-" {
		q = q.Require("ApiID", p.ApiID)
	}
	if p.VersionID != "-" {
		q = q.Require("VersionID", p.VersionID)
	}
	if p.SpecID != "-" {
		q = q.Require("SpecID", p.SpecID)
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"project_id", filterArgTypeString},
			{"api_id", filterArgTypeString},
			{"version_id", filterArgTypeString},
			{"spec_id", filterArgTypeString},
			{"artifact_id", filterArgTypeString},
			{"create_time", filterArgTypeTimestamp},
			{"update_time", filterArgTypeTimestamp},
		})
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	var artifactMessages []*rpc.Artifact
	var artifact models.Artifact
	it := client.Run(ctx, q)
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&artifact); err == nil; _, err = it.Next(&artifact) {
		// don't allow wildcarded-names to be empty
		if p.SpecID == "-" && artifact.SpecID == "" {
			continue
		}
		if p.VersionID == "-" && artifact.VersionID == "" {
			continue
		}
		if p.ApiID == "-" && artifact.ApiID == "" {
			continue
		}
		if p.ProjectID == "-" && artifact.ProjectID == "" {
			continue
		}
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id":  artifact.ProjectID,
				"api_id":      artifact.ApiID,
				"version_id":  artifact.VersionID,
				"spec_id":     artifact.SpecID,
				"artifact_id": artifact.ArtifactID,
				"create_time": artifact.CreateTime,
				"update_time": artifact.UpdateTime,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if v, ok := out.Value().(bool); !ok {
				return nil, invalidArgumentError(fmt.Errorf("expression does not evaluate to a boolean (instead yielding %T)", out.Value()))
			} else if !v {
				continue
			}
		}
		var blob *models.Blob
		if req.GetView() == rpc.View_FULL {
			blob, _ = fetchBlobForArtifact(ctx, client, &artifact)
		}
		artifactMessage, _ := artifact.Message(blob)
		artifactMessages = append(artifactMessages, artifactMessage)
		if len(artifactMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListArtifactsResponse{
		Artifacts: artifactMessages,
	}
	responses.NextPageToken, err = it.GetCursor()
	if responses.NextPageToken == req.PageToken {
		responses.NextPageToken = ""
	}
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// ReplaceArtifact handles the corresponding API request.
func (s *RegistryServer) ReplaceArtifact(ctx context.Context, req *rpc.ReplaceArtifactRequest) (*rpc.Artifact, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	artifact, err := models.NewArtifactFromResourceName(req.GetArtifact().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := client.NewKey(models.ArtifactEntityName, req.GetArtifact().GetName())
	err = client.Get(ctx, k, artifact)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	blob := models.NewBlobForArtifact(artifact, nil)
	err = artifact.Update(req.GetArtifact(), blob)
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, artifact)
	if err != nil {
		return nil, internalError(err)
	}
	if blob != nil {
		k2 := client.NewKey(models.BlobEntityName, artifact.ResourceName())
		_, err = client.Put(ctx,
			k2,
			blob)
	}
	s.notify(rpc.Notification_UPDATED, artifact.ResourceName())
	return artifact.Message(nil)
}

// TODO: remove this function
func contentsForArtifact(artifact *rpc.Artifact) []byte {
	return artifact.Contents
}

// fetchBlobForArtifact gets the blob containing the artifact contents.
func fetchBlobForArtifact(
	ctx context.Context,
	client storage.Client,
	artifact *models.Artifact) (*models.Blob, error) {
	blob := &models.Blob{}
	k := client.NewKey(models.BlobEntityName, artifact.ResourceName())
	err := client.Get(ctx, k, blob)
	return blob, err
}
