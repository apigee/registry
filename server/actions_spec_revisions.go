package server

import (
	"context"
	"fmt"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/apigee/registry/server/storage"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
)

// ListApiSpecRevisions handles the corresponding API request.
func (s *RegistryServer) ListApiSpecRevisions(ctx context.Context, req *rpc.ListApiSpecRevisionsRequest) (*rpc.ListApiSpecRevisionsResponse, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetPageSize() < 0 {
		return nil, invalidArgumentError(fmt.Errorf("invalid page_size %d: must not be negative", req.GetPageSize()))
	}

	name, err := names.ParseSpec(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	q := client.NewQuery(storage.SpecEntityName)
	q = q.Require("ProjectID", name.ProjectID)
	q = q.Require("ApiID", name.ApiID)
	q = q.Require("VersionID", name.VersionID)
	q = q.Require("SpecID", name.SpecID)
	q = q.Order("-RevisionCreateTime")
	q, err = q.ApplyCursor(req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}

	pageSize := boundPageSize(req.GetPageSize())
	response := &rpc.ListApiSpecRevisionsResponse{
		Specs: make([]*rpc.ApiSpec, 0, pageSize),
	}

	var spec models.Spec
	it := client.Run(ctx, q)
	for _, err := it.Next(&spec); err == nil; _, err = it.Next(&spec) {
		specMessage, err := spec.BasicMessage(spec.RevisionName())
		if err != nil {
			continue
		}
		response.Specs = append(response.GetSpecs(), specMessage)

		// Exit when page is full and before the iterator moves forward again.
		// This ensures the cursor is returned in the correct position.
		if len(response.GetSpecs()) >= pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}

	response.NextPageToken, err = it.GetCursor(len(response.GetSpecs()))
	if err != nil {
		return nil, internalError(err)
	}

	return response, nil
}

// DeleteApiSpecRevision handles the corresponding API request.
func (s *RegistryServer) DeleteApiSpecRevision(ctx context.Context, req *rpc.DeleteApiSpecRevisionRequest) (*empty.Empty, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	name, err := names.ParseSpecRevision(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	revision, err := getSpecRevision(ctx, client, name)
	if err != nil {
		return nil, err
	}

	// Parse the retrieved spec revision name, which has a non-tag revision ID.
	// This is necessary to ensure the actual revision is deleted.
	name, err = names.ParseSpecRevision(revision.RevisionName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	// If the one we will delete is the current revision, we need to designate a new current revision.
	if revision.Currency == models.IsCurrent {
		// get the most recent non-current revision and make it current
		newKey, newCurrentRevision, err := s.fetchMostRecentNonCurrentRevisionOfSpec(ctx, client, name.Spec())
		if err == nil && newCurrentRevision != nil {
			newCurrentRevision.Currency = models.IsCurrent
			_, _ = client.Put(ctx, newKey, newCurrentRevision)
		}
	}

	if err := client.Delete(ctx, client.NewKey(models.BlobEntityName, name.String())); err != nil {
		return nil, internalError(err)
	}

	if err := client.Delete(ctx, client.NewKey(storage.SpecEntityName, name.String())); err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_DELETED, name.String())
	return &empty.Empty{}, nil
}

// TagApiSpecRevision handles the corresponding API request.
func (s *RegistryServer) TagApiSpecRevision(ctx context.Context, req *rpc.TagApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetTag() == "" {
		return nil, invalidArgumentError(fmt.Errorf("invalid tag %q, must not be empty", req.GetTag()))
	}

	// Parse the requested spec revision name, which may include a tag name.
	name, err := names.ParseSpecRevision(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	revision, err := getSpecRevision(ctx, client, name)
	if err != nil {
		return nil, err
	}

	// Parse the retrieved spec revision name, which has a non-tag revision ID.
	// This is necessary to ensure the new tag is associated with a revision ID, not another tag.
	name, err = names.ParseSpecRevision(revision.RevisionName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	tag := models.NewSpecRevisionTag(name, req.GetTag())
	if err := saveSpecRevisionTag(ctx, client, tag); err != nil {
		return nil, err
	}

	message, err := revision.BasicMessage(tag.String())
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_UPDATED, tag.String())
	return message, nil
}

// RollbackApiSpec handles the corresponding API request.
func (s *RegistryServer) RollbackApiSpec(ctx context.Context, req *rpc.RollbackApiSpecRequest) (*rpc.ApiSpec, error) {
	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)

	if req.GetRevisionId() == "" {
		return nil, invalidArgumentError(fmt.Errorf("invalid revision ID %q, must not be empty", req.GetRevisionId()))
	}

	parent, err := names.ParseSpec(req.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}

	current, err := getSpec(ctx, client, parent)
	if err != nil {
		return nil, err
	}

	// Mark the current revision as non-current.
	current.Currency = models.NotCurrent
	if err := saveSpecRevision(ctx, client, current); err != nil {
		return nil, err
	}

	// Get the target spec revision to use as a base for the new rollback revision.
	name := parent.Revision(req.GetRevisionId())
	target, err := getSpecRevision(ctx, client, name)
	if err != nil {
		return nil, err
	}

	// Save a new rollback revision based on the target revision.
	rollback := target.NewRevision()
	if err := saveSpecRevision(ctx, client, rollback); err != nil {
		return nil, err
	}

	blob, err := getSpecRevisionContents(ctx, client, name)
	if err != nil {
		return nil, err
	}

	// Save a new copy of the target revision blob for the rollback revision.
	blob.RevisionID = name.RevisionID
	if err := saveSpecRevisionContents(ctx, client, rollback, blob.Contents); err != nil {
		return nil, err
	}

	message, err := rollback.BasicMessage(rollback.RevisionName())
	if err != nil {
		return nil, internalError(err)
	}

	s.notify(rpc.Notification_UPDATED, rollback.RevisionName())
	return message, nil
}

func saveSpecRevision(ctx context.Context, client storage.Client, spec *models.Spec) error {
	k := client.NewKey(storage.SpecEntityName, spec.RevisionName())
	if _, err := client.Put(ctx, k, spec); err != nil {
		return internalError(err)
	}

	return nil
}

func getSpecRevision(ctx context.Context, client storage.Client, name names.SpecRevision) (*models.Spec, error) {
	spec := new(models.Spec)

	// If the provided revision ID is a known tag, use the associated revision for lookup.
	tag := new(models.SpecRevisionTag)
	if err := client.Get(ctx, client.NewKey(storage.SpecRevisionTagEntityName, name.String()), tag); err == nil {
		name.RevisionID = tag.RevisionID
	} else if !client.IsNotFound(err) {
		return nil, internalError(err)
	}

	k := client.NewKey(storage.SpecEntityName, name.String())
	if err := client.Get(ctx, k, spec); client.IsNotFound(err) {
		return nil, notFoundError(fmt.Errorf("spec revision %q not found", name))
	} else if err != nil {
		return nil, internalError(err)
	}

	return spec, nil
}

func saveSpecRevisionContents(ctx context.Context, client storage.Client, spec *models.Spec, contents []byte) error {
	blob := models.NewBlobForSpec(spec, contents)
	k := client.NewKey(models.BlobEntityName, spec.RevisionName())
	if _, err := client.Put(ctx, k, blob); err != nil {
		return internalError(err)
	}

	return nil
}

func getSpecRevisionContents(ctx context.Context, client storage.Client, name names.SpecRevision) (*models.Blob, error) {
	// If the provided revision ID is a known tag, use the associated revision for lookup.
	tag := new(models.SpecRevisionTag)
	if err := client.Get(ctx, client.NewKey(storage.SpecRevisionTagEntityName, name.String()), tag); err == nil {
		name.RevisionID = tag.RevisionID
	} else if !client.IsNotFound(err) {
		return nil, internalError(err)
	}

	blob := new(models.Blob)
	k := client.NewKey(models.BlobEntityName, name.String())
	if err := client.Get(ctx, k, blob); client.IsNotFound(err) {
		return nil, notFoundError(fmt.Errorf("spec revision contents %q not found", name.String()))
	} else if err != nil {
		return nil, internalError(err)
	}

	return blob, nil
}

func saveSpecRevisionTag(ctx context.Context, client storage.Client, tag *models.SpecRevisionTag) error {
	k := client.NewKey(storage.SpecRevisionTagEntityName, tag.String())
	if _, err := client.Put(ctx, k, tag); err != nil {
		return internalError(err)
	}

	return nil
}
