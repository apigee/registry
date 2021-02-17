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
	"errors"
	"log"
	"sort"
	"time"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	storage "github.com/apigee/registry/server/storage"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateApiSpec handles the corresponding API request.
func (s *RegistryServer) CreateApiSpec(ctx context.Context, request *rpc.CreateApiSpecRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	spec, err := models.NewSpecFromParentAndSpecID(request.GetParent(), request.GetApiSpecId())
	if err != nil {
		return nil, internalError(err)
	}

	if err := spec.Update(request.GetApiSpec(), nil); err != nil {
		return nil, internalError(err)
	}

	return s.createSpec(ctx, client, spec, request.ApiSpec.GetContents())
}

func (s *RegistryServer) createSpec(ctx context.Context, client storage.Client, spec *models.Spec, contents []byte) (*rpc.ApiSpec, error) {
	q := client.NewQuery(models.SpecEntityName)
	q = q.Require("ProjectID", spec.ProjectID)
	q = q.Require("ApiID", spec.ApiID)
	q = q.Require("VersionID", spec.VersionID)
	q = q.Require("SpecID", spec.SpecID)
	it := client.Run(ctx, q)

	if _, err := it.Next(&models.Spec{}); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "spec %q already exists", spec.ResourceName())
	}

	spec.CreateTime = spec.RevisionUpdateTime
	spec.RevisionCreateTime = spec.RevisionUpdateTime
	spec.Currency = models.IsCurrent
	if err := saveSpec(ctx, client, spec); err != nil {
		return nil, internalError(err)
	}

	if err := saveSpecContents(ctx, client, spec, contents); err != nil {
		return nil, internalError(err)
	}

	response, err := spec.Message(nil, "")
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_CREATED, spec.ResourceNameWithRevision())
	return response, nil
}

func saveSpec(ctx context.Context, client storage.Client, spec *models.Spec) error {
	k := client.NewKey(models.SpecEntityName, spec.ResourceNameWithRevision())
	if _, err := client.Put(ctx, k, spec); err != nil {
		log.Printf("save spec error %+v", err)
		return err
	}

	return nil
}

func saveSpecContents(ctx context.Context, client storage.Client, spec *models.Spec, contents []byte) error {
	blob := models.NewBlobForSpec(spec, contents)
	k := client.NewKey(models.BlobEntityName, spec.ResourceNameWithRevision())
	if _, err := client.Put(ctx, k, blob); err != nil {
		log.Printf("save blob error %+v", err)
		return err
	}

	return nil
}

// DeleteApiSpec handles the corresponding API request.
func (s *RegistryServer) DeleteApiSpec(ctx context.Context, request *rpc.DeleteApiSpecRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	// Validate name and create dummy spec (we just need the ID fields).
	spec, err := models.NewSpecFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if spec.RevisionID != "" {
		return nil, invalidArgumentError(errors.New("specific revisions should be deleted with DeleteSpecRevision"))
	}
	// Delete all revisions of the spec.
	q := client.NewQuery(models.SpecEntityName)
	q = q.Require("ProjectID", spec.ProjectID)
	q = q.Require("ApiID", spec.ApiID)
	q = q.Require("VersionID", spec.VersionID)
	q = q.Require("SpecID", spec.SpecID)
	err = client.DeleteAllMatches(ctx, q)
	// Delete all blobs associated with the spec.
	q = client.NewQuery(models.BlobEntityName)
	q = q.Require("ProjectID", spec.ProjectID)
	q = q.Require("ApiID", spec.ApiID)
	q = q.Require("VersionID", spec.VersionID)
	q = q.Require("SpecID", spec.SpecID)
	err = client.DeleteAllMatches(ctx, q)
	s.notify(rpc.Notification_DELETED, request.GetName())
	return &empty.Empty{}, err
}

// GetApiSpec handles the corresponding API request.
func (s *RegistryServer) GetApiSpec(ctx context.Context, request *rpc.GetApiSpecRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, request.GetName())
	if err != nil {
		if client.IsNotFound(err) {
			return nil, notFoundError(err)
		}
		return nil, internalError(err)
	}
	var blob *models.Blob
	if request.GetView() == rpc.View_FULL {
		blob, _ = fetchBlobForSpec(ctx, client, spec)
	}
	return spec.Message(blob, userSpecifiedRevision)
}

// ListApiSpecs handles the corresponding API request.
func (s *RegistryServer) ListApiSpecs(ctx context.Context, req *rpc.ListApiSpecsRequest) (*rpc.ListApiSpecsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	q := client.NewQuery(models.SpecEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := names.ParseParentVersion(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if m[1] != "-" {
		q = q.Require("ProjectID", m[1])
	}
	if m[2] != "-" {
		q = q.Require("ApiID", m[2])
	}
	if m[3] != "-" {
		q = q.Require("VersionID", m[3])
	}
	q = q.Require("Currency", models.IsCurrent)
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"project_id", filterArgTypeString},
			{"api_id", filterArgTypeString},
			{"version_id", filterArgTypeString},
			{"spec_id", filterArgTypeString},
			{"filename", filterArgTypeString},
			{"description", filterArgTypeString},
			{"create_time", filterArgTypeTimestamp},
			{"update_time", filterArgTypeTimestamp},
			{"style", filterArgTypeString},
			{"size_bytes", filterArgTypeInt},
			{"source_uri", filterArgTypeString},
		})
	if err != nil {
		return nil, internalError(err)
	}
	var specMessages []*rpc.ApiSpec
	var spec models.Spec
	it := client.Run(ctx, q)
	pageSize := boundPageSize(req.GetPageSize())
	for _, err := it.Next(&spec); err == nil; _, err = it.Next(&spec) {
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id":           spec.ProjectID,
				"api_id":               spec.ApiID,
				"version_id":           spec.VersionID,
				"spec_id":              spec.SpecID,
				"filename":             spec.FileName,
				"description":          spec.Description,
				"create_time":          spec.CreateTime,
				"revision_create_time": spec.RevisionCreateTime,
				"revision_update_time": spec.RevisionUpdateTime,
				"mime_type":            spec.MimeType,
				"size_bytes":           spec.SizeInBytes,
				"source_uri":           spec.SourceURI,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		var blob *models.Blob
		if req.GetView() == rpc.View_FULL {
			blob, _ = fetchBlobForSpec(ctx, client, &spec)
		}
		specMessage, _ := spec.Message(blob, "")
		specMessages = append(specMessages, specMessage)
		if len(specMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListApiSpecsResponse{
		ApiSpecs: specMessages,
	}
	responses.NextPageToken, err = it.GetCursor(len(specMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// UpdateApiSpec handles the corresponding API request.
func (s *RegistryServer) UpdateApiSpec(ctx context.Context, request *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if request.GetApiSpec() == nil {
		return nil, invalidArgumentError(errors.New("spec body is required for updates"))
	}

	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, request.GetApiSpec().GetName())
	if request.GetAllowMissing() && client.IsNotFound(err) {
		spec, err := models.NewSpecFromResourceName(request.ApiSpec.GetName())
		if err != nil {
			return nil, internalError(err)
		}

		if err := spec.Update(request.GetApiSpec(), nil); err != nil {
			return nil, internalError(err)
		}

		return s.createSpec(ctx, client, spec, request.ApiSpec.GetContents())
	} else if err != nil {
		return nil, internalError(err)
	} else if userSpecifiedRevision != "" {
		return nil, invalidArgumentError(errors.New("updates to specific revisions are unsupported"))
	}
	oldRevisionID := spec.RevisionID
	err = spec.Update(request.GetApiSpec(), request.GetUpdateMask())
	if err != nil {
		return nil, internalError(err)
	}
	newRevisionID := spec.RevisionID
	// if the revision changed, get the previously-current revision and mark it as non-current
	if oldRevisionID != newRevisionID {
		k := client.NewKey(models.SpecEntityName, spec.ResourceNameWithSpecifiedRevision(oldRevisionID))
		currentRevision := &models.Spec{}
		client.Get(ctx, k, currentRevision)
		currentRevision.Currency = models.NotCurrent
		_, err = client.Put(ctx, k, currentRevision)
		if err != nil {
			return nil, internalError(err)
		}
		spec.Currency = models.IsCurrent
	}
	k := client.NewKey(models.SpecEntityName, spec.ResourceNameWithRevision())
	spec.Key = spec.ResourceNameWithRevision()
	k, err = client.Put(ctx, k, spec)
	if err != nil {
		return nil, internalError(err)
	}
	// save a blob with the spec contents (but only if the contents were updated)
	if request.GetApiSpec().GetContents() != nil {
		blob := models.NewBlobForSpec(
			spec,
			request.GetApiSpec().GetContents())
		_, err = client.Put(ctx,
			client.NewKey(models.BlobEntityName, spec.ResourceNameWithRevision()),
			blob)
		if err != nil {
			return nil, internalError(err)
		}
	}
	s.notify(rpc.Notification_UPDATED, spec.ResourceNameWithRevision())
	return spec.Message(nil, "")
}

// ListApiSpecRevisions handles the corresponding API request.
func (s *RegistryServer) ListApiSpecRevisions(ctx context.Context, req *rpc.ListApiSpecRevisionsRequest) (*rpc.ListApiSpecRevisionsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	targetSpec, err := models.NewSpecFromResourceName(req.GetName())
	if err != nil {
		return nil, internalError(err)
	}
	q := client.NewQuery(models.SpecEntityName)
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	q = q.Require("ProjectID", targetSpec.ProjectID)
	q = q.Require("ApiID", targetSpec.ApiID)
	q = q.Require("VersionID", targetSpec.VersionID)
	q = q.Require("SpecID", targetSpec.SpecID)
	q = q.Order("-CreateTime")

	var specMessages []*rpc.ApiSpec
	responses := &rpc.ListApiSpecRevisionsResponse{}
	if s.weTrustTheSort {
		var spec models.Spec
		it := client.Run(ctx, q)
		pageSize := boundPageSize(req.GetPageSize())
		for _, err := it.Next(&spec); err == nil; _, err = it.Next(&spec) {
			specMessage, _ := spec.Message(nil, spec.RevisionID)
			specMessages = append(specMessages, specMessage)
			if len(specMessages) == pageSize {
				break
			}
		}
		if err != nil && err != iterator.Done {
			return nil, internalError(err)
		}
		responses.NextPageToken, err = it.GetCursor(len(specMessages))
		if err != nil {
			return nil, internalError(err)
		}
	} else {
		specs := make([]*models.Spec, 0)
		it := client.Run(ctx, q)
		for {
			spec := &models.Spec{}
			_, err := it.Next(spec)
			if err != nil {
				break
			}
			specs = append(specs, spec)
		}
		if err != nil && err != iterator.Done {
			return nil, internalError(err)
		}
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].CreateTime.After(specs[j].CreateTime)
		})
		for _, spec := range specs {
			specMessage, _ := spec.Message(nil, spec.RevisionID)
			specMessages = append(specMessages, specMessage)
		}
		responses.NextPageToken = ""
		err = nil
	}
	responses.Specs = specMessages
	return responses, nil
}

// DeleteApiSpecRevision handles the corresponding API request.
func (s *RegistryServer) DeleteApiSpecRevision(ctx context.Context, request *rpc.DeleteApiSpecRevisionRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	// Delete the spec revision.
	// First, get the revision to delete.
	spec, _, err := fetchSpec(ctx, client, request.GetName())
	if err != nil {
		return nil, internalError(err)
	}
	k := client.NewKey(models.SpecEntityName, spec.ResourceNameWithRevision())
	// If the one we will delete is the current revision, we need to designate a new current revision.
	if spec.Currency == models.IsCurrent {
		// get the most recent non-current revision and make it current
		newKey, newCurrentRevision, err := s.fetchMostRecentNonCurrentRevisionOfSpec(ctx, client, request.GetName())
		if err != nil {
			log.Printf("error %+v", err)
		}
		if err == nil && newCurrentRevision != nil {
			newCurrentRevision.Currency = models.IsCurrent
			client.Put(ctx, newKey, newCurrentRevision)
		}
	}
	err = client.Delete(ctx, k)
	// Delete the blob associated with the spec
	k2 := client.NewKey(models.BlobEntityName, spec.ResourceNameWithRevision())
	err = client.Delete(ctx, k2)
	s.notify(rpc.Notification_DELETED, spec.ResourceNameWithRevision())
	return &empty.Empty{}, err
}

// TagApiSpecRevision handles the corresponding API request.
func (s *RegistryServer) TagApiSpecRevision(ctx context.Context, request *rpc.TagApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, request.GetName())
	if err != nil {
		return nil, internalError(err)
	}
	if userSpecifiedRevision == "" {
		log.Printf("we might not want to support tagging specs with unspecified revisions")
	}
	if request.GetTag() == "" {
		return nil, invalidArgumentError(errors.New("tags cannot be empty"))
	}
	// save the tag
	now := time.Now()
	tag := &models.SpecRevisionTag{
		ProjectID:  spec.ProjectID,
		ApiID:      spec.ApiID,
		VersionID:  spec.VersionID,
		SpecID:     spec.SpecID,
		RevisionID: spec.RevisionID,
		Tag:        request.GetTag(),
		CreateTime: now,
		UpdateTime: now,
	}
	k := client.NewKey(models.SpecRevisionTagEntityName, tag.ResourceNameWithTag())
	k, err = client.Put(ctx, k, tag)
	// send a notification that the tagged spec has been updated
	s.notify(rpc.Notification_UPDATED, spec.ResourceNameWithSpecifiedRevision(request.GetTag()))
	// return the spec using the tag for its name
	return spec.Message(nil, request.GetTag())
}

// RollbackApiSpec handles the corresponding API request.
func (s *RegistryServer) RollbackApiSpec(ctx context.Context, request *rpc.RollbackApiSpecRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	specNameWithRevision := request.GetName() + "@" + request.GetRevisionId()
	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, specNameWithRevision)
	if err != nil {
		// TODO: this should return NotFound if the revision was not found.
		return nil, notFoundError(err)
	}
	if userSpecifiedRevision == "" {
		return nil, invalidArgumentError(errors.New("rollbacks require a specified revision"))
	}
	// The previous current revision needs to be marked non-current.
	oldKey, oldCurrent, err := fetchCurrentRevisionOfSpec(ctx, client, request.GetName())
	if err == nil && oldCurrent != nil {
		oldCurrent.Currency = models.NotCurrent
		_, err = client.Put(ctx, oldKey, oldCurrent)
		if err != nil {
			log.Printf("oops %+v", err)
			return nil, internalError(err)
		}
	}
	// Make the selected revision the current revision by giving it a new RevisionID and saving it
	oldBlobKey := client.NewKey(models.BlobEntityName, spec.ResourceNameWithRevision())
	blob := &models.Blob{}
	err = client.Get(ctx, oldBlobKey, blob)
	if err != nil {
		return nil, internalError(err)
	}
	spec.BumpRevision()
	spec.Currency = models.IsCurrent
	newSpecKey := client.NewKey(models.SpecEntityName, spec.ResourceNameWithRevision())
	_, err = client.Put(ctx, newSpecKey, spec)
	if err != nil {
		return nil, internalError(err)
	}
	// Resave the blob for the current revision with the new RevisionID
	newBlobKey := client.NewKey(models.BlobEntityName, spec.ResourceNameWithRevision())
	blob.RevisionID = spec.RevisionID
	_, err = client.Put(ctx, newBlobKey, blob)
	if err != nil {
		return nil, internalError(err)
	}
	// Send a notification of the new revision.
	s.notify(rpc.Notification_UPDATED, spec.ResourceNameWithRevision())
	return spec.Message(nil, spec.RevisionID)
}

// fetchSpec gets the stored model of a Spec.
func fetchSpec(
	ctx context.Context,
	client storage.Client,
	name string,
) (*models.Spec, string, error) {
	spec, err := models.NewSpecFromResourceName(name)
	if err != nil {
		return nil, "", err
	}
	// if there's no revision, get the current revision
	if spec.RevisionID == "" {
		_, spec, err := fetchCurrentRevisionOfSpec(ctx, client, name)
		if err != nil {
			return nil, "", err
		}
		return spec, "", nil
	}
	// since a revision was specified, get the spec by revision
	// if the revision reference is a tag, resolve the tag
	var resourceName string
	var revisionName string
	specRevisionTag := &models.SpecRevisionTag{}
	k0 := client.NewKey(models.SpecRevisionTagEntityName, spec.ResourceNameWithRevision())
	err = client.Get(ctx, k0, specRevisionTag)
	if client.IsNotFound(err) {
		// if there is no tag, just use the provided revision
		resourceName = spec.ResourceNameWithRevision()
		revisionName = spec.RevisionID
	} else if err != nil {
		return nil, "", err
	} else {
		// if there is a tag, use the revision that the tag references
		resourceName = specRevisionTag.ResourceNameWithRevision()
		revisionName = specRevisionTag.Tag
	}
	// now that we know the revision, use it get the spec
	k := client.NewKey(models.SpecEntityName, resourceName)
	err = client.Get(ctx, k, spec)
	if client.IsNotFound(err) {
		return nil, revisionName, err
	} else if err != nil {
		return nil, revisionName, err
	}
	return spec, revisionName, nil
}

// fetchMostRecentNonCurrentRevisionOfSpec gets the most recent revision that's not current.
func (s *RegistryServer) fetchMostRecentNonCurrentRevisionOfSpec(
	ctx context.Context,
	client storage.Client,
	name string,
) (storage.Key, *models.Spec, error) {
	pattern, err := models.NewSpecFromResourceName(name)
	if err != nil {
		return nil, nil, err
	}
	// note that we ignore any specified RevisionID
	q := client.NewQuery(models.SpecEntityName)
	q = q.Require("ProjectID", pattern.ProjectID)
	q = q.Require("ApiID", pattern.ApiID)
	q = q.Require("VersionID", pattern.VersionID)
	q = q.Require("SpecID", pattern.SpecID)
	q = q.Require("Currency", models.NotCurrent)
	q = q.Order("-CreateTime")
	it := client.Run(ctx, q)

	if s.weTrustTheSort {
		spec := &models.Spec{}
		k, err := it.Next(spec)
		if err != nil {
			return nil, nil, client.NotFoundError()
		}
		return k, spec, nil
	} else {
		specs := make([]*models.Spec, 0)
		for {
			spec := &models.Spec{}
			_, err := it.Next(spec)
			if err != nil {
				break
			}
			specs = append(specs, spec)
		}
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].CreateTime.After(specs[j].CreateTime)
		})
		k := client.NewKey("Spec", specs[0].Key)
		return k, specs[0], nil
	}
}

// fetchCurrentRevisionOfSpec gets the current revision.
func fetchCurrentRevisionOfSpec(
	ctx context.Context,
	client storage.Client,
	name string,
) (storage.Key, *models.Spec, error) {
	pattern, err := models.NewSpecFromResourceName(name)
	if err != nil {
		return nil, nil, err
	}
	// note that we ignore any specified RevisionID
	q := client.NewQuery(models.SpecEntityName)
	q = q.Require("ProjectID", pattern.ProjectID)
	q = q.Require("ApiID", pattern.ApiID)
	q = q.Require("VersionID", pattern.VersionID)
	q = q.Require("SpecID", pattern.SpecID)
	q = q.Require("Currency", models.IsCurrent)
	it := client.Run(ctx, q)
	spec := &models.Spec{}
	k, err := it.Next(spec)
	if err != nil {
		return nil, nil, client.NotFoundError()
	}
	return k, spec, nil
}

// fetchBlobForSpec gets the blob containing the spec contents.
func fetchBlobForSpec(
	ctx context.Context,
	client storage.Client,
	spec *models.Spec) (*models.Blob, error) {
	blob := &models.Blob{}
	k := client.NewKey(models.BlobEntityName, spec.ResourceNameWithRevision())
	err := client.Get(ctx, k, blob)
	return blob, err
}
