// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"
	"errors"
	"log"
	"time"

	"apigov.dev/registry/models"
	rpc "apigov.dev/registry/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateSpec handles the corresponding API request.
func (s *RegistryServer) CreateSpec(ctx context.Context, request *rpc.CreateSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, err := models.NewSpecFromParentAndSpecID(request.GetParent(), request.GetSpecId())
	if err != nil {
		return nil, err
	}
	// fail if spec already exists
	q := datastore.NewQuery(models.SpecEntityName)
	q = q.Filter("ProjectID =", spec.ProjectID)
	q = q.Filter("ProductID =", spec.ProductID)
	q = q.Filter("VersionID =", spec.VersionID)
	q = q.Filter("SpecID =", spec.SpecID)
	it := client.Run(ctx, q.Distinct())
	var existingSpec models.Spec
	existingKey, err := it.Next(&existingSpec)
	if existingKey != nil {
		return nil, status.Error(codes.AlreadyExists, spec.ResourceName()+" already exists")
	}
	// save the spec under its full resource@revision name
	err = spec.Update(request.GetSpec())
	spec.CreateTime = spec.UpdateTime
	// the first revision of the spec that we save is also the current one
	spec.IsCurrent = true
	k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceNameWithRevision()}
	k, err = client.Put(ctx, k, spec)
	if err != nil {
		return nil, err
	}
	return spec.Message(rpc.SpecView_BASIC, "")
}

// DeleteSpec handles the corresponding API request.
func (s *RegistryServer) DeleteSpec(ctx context.Context, request *rpc.DeleteSpecRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	// Validate name and create dummy spec (we just need the ID fields).
	spec, err := models.NewSpecFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if spec.RevisionID != "" {
		return nil, invalidArgumentError(errors.New("specific revisions should be deleted with DeleteSpecRevision"))
	}
	// Delete all revisions of the spec.
	q := datastore.NewQuery(models.SpecEntityName)
	q = q.Filter("ProjectID =", spec.ProjectID)
	q = q.Filter("ProductID =", spec.ProductID)
	q = q.Filter("VersionID =", spec.VersionID)
	q = q.Filter("SpecID =", spec.SpecID)
	err = models.DeleteAllMatches(ctx, client, q)
	return &empty.Empty{}, err
}

// GetSpec handles the corresponding API request.
func (s *RegistryServer) GetSpec(ctx context.Context, request *rpc.GetSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, request.GetName())
	if err != nil {
		return nil, err
	}
	return spec.Message(request.GetView(), userSpecifiedRevision)
}

// ListSpecs handles the corresponding API request.
func (s *RegistryServer) ListSpecs(ctx context.Context, req *rpc.ListSpecsRequest) (*rpc.ListSpecsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	q := datastore.NewQuery(models.SpecEntityName)
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := models.ParseParentVersion(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if m[1] != "-" {
		q = q.Filter("ProjectID =", m[1])
	}
	if m[2] != "-" {
		q = q.Filter("ProductID =", m[2])
	}
	if m[3] != "-" {
		q = q.Filter("VersionID =", m[3])
	}
	q = q.Filter("IsCurrent =", true)
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"project_id", filterArgTypeString},
			{"product_id", filterArgTypeString},
			{"version_id", filterArgTypeString},
			{"spec_id", filterArgTypeString},
			{"filename", filterArgTypeString},
			{"description", filterArgTypeString},
			{"style", filterArgTypeString},
		})
	if err != nil {
		return nil, internalError(err)
	}
	var specMessages []*rpc.Spec
	var spec models.Spec
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for _, err := it.Next(&spec); err == nil; _, err = it.Next(&spec) {
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"project_id":  spec.ProjectID,
				"product_id":  spec.ProductID,
				"version_id":  spec.VersionID,
				"spec_id":     spec.SpecID,
				"filename":    spec.FileName,
				"description": spec.Description,
				"style":       spec.Style,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		specMessage, _ := spec.Message(req.GetView(), "")
		specMessages = append(specMessages, specMessage)
		if len(specMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListSpecsResponse{
		Specs: specMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(specMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// UpdateSpec handles the corresponding API request.
func (s *RegistryServer) UpdateSpec(ctx context.Context, request *rpc.UpdateSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, request.GetSpec().GetName())
	if err != nil {
		return nil, err
	}
	if userSpecifiedRevision != "" {
		return nil, invalidArgumentError(errors.New("updates to specific revisions are unsupported"))
	}
	oldRevisionID := spec.RevisionID
	err = spec.Update(request.GetSpec())
	if err != nil {
		return nil, err
	}
	newRevisionID := spec.RevisionID
	// if the revision changed, get the previously-current revision and mark it as non-current
	if oldRevisionID != newRevisionID {
		k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceNameWithSpecifiedRevision(oldRevisionID)}
		var currentRevision models.Spec
		client.Get(ctx, k, &currentRevision)
		currentRevision.IsCurrent = false
		_, err = client.Put(ctx, k, &currentRevision)
		if err != nil {
			return nil, err
		}
		spec.IsCurrent = true
	}
	k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceNameWithRevision()}
	k, err = client.Put(ctx, k, spec)
	if err != nil {
		return nil, err
	}
	return spec.Message(rpc.SpecView_BASIC, "")
}

// ListSpecRevisions handles the corresponding API request.
func (s *RegistryServer) ListSpecRevisions(ctx context.Context, req *rpc.ListSpecRevisionsRequest) (*rpc.ListSpecRevisionsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	targetSpec, err := models.NewSpecFromResourceName(req.GetName())
	if err != nil {
		return nil, err
	}
	q := datastore.NewQuery(models.SpecEntityName)
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	q = q.Filter("ProjectID =", targetSpec.ProjectID)
	q = q.Filter("ProductID =", targetSpec.ProductID)
	q = q.Filter("VersionID =", targetSpec.VersionID)
	q = q.Filter("SpecID =", targetSpec.SpecID)
	q = q.Order("-CreateTime")
	var specMessages []*rpc.Spec
	var spec models.Spec
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for _, err := it.Next(&spec); err == nil; _, err = it.Next(&spec) {
		specMessage, _ := spec.Message(rpc.SpecView_BASIC, spec.RevisionID)
		specMessages = append(specMessages, specMessage)
		if len(specMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListSpecRevisionsResponse{
		Specs: specMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(specMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// DeleteSpecRevision handles the corresponding API request.
func (s *RegistryServer) DeleteSpecRevision(ctx context.Context, request *rpc.DeleteSpecRevisionRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	// Validate name and create dummy spec (we just need the ID fields).
	_, err = models.NewSpecFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete the spec revision.
	// First, get the revision to delete.
	k := &datastore.Key{Kind: models.SpecEntityName, Name: request.GetName()}
	var revisionToDelete models.Spec
	client.Get(ctx, k, &revisionToDelete)
	// If the one we will delete is the current revision, we need to designate a new current revision.
	if revisionToDelete.IsCurrent {
		// get the most recent non-current revision and make it current
		newKey, newCurrentRevision, err := fetchMostRecentNonCurrentRevisionOfSpec(ctx, client, request.GetName())
		if err != nil {
			log.Printf("error %+v", err)
		}
		if err == nil && newCurrentRevision != nil {
			newCurrentRevision.IsCurrent = true
			client.Put(ctx, newKey, newCurrentRevision)
		}
	}
	err = client.Delete(ctx, k)
	return &empty.Empty{}, err
}

// TagSpecRevision handles the corresponding API request.
func (s *RegistryServer) TagSpecRevision(ctx context.Context, request *rpc.TagSpecRevisionRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, request.GetName())
	if err != nil {
		return nil, err
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
		ProductID:  spec.ProductID,
		VersionID:  spec.VersionID,
		SpecID:     spec.SpecID,
		RevisionID: spec.RevisionID,
		Tag:        request.GetTag(),
		CreateTime: now,
		UpdateTime: now,
	}
	k := &datastore.Key{Kind: models.SpecRevisionTagEntityName, Name: tag.ResourceNameWithTag()}
	k, err = client.Put(ctx, k, tag)
	// return the spec using the tag for its name
	return spec.Message(rpc.SpecView_BASIC, request.GetTag())
}

// ListSpecRevisionTags handles the corresponding API request.
func (s *RegistryServer) ListSpecRevisionTags(ctx context.Context, req *rpc.ListSpecRevisionTagsRequest) (*rpc.ListSpecRevisionTagsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	targetSpec, err := models.NewSpecFromResourceName(req.GetName())
	if err != nil {
		return nil, err
	}
	q := datastore.NewQuery(models.SpecRevisionTagEntityName)
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	q = q.Filter("ProjectID =", targetSpec.ProjectID)
	q = q.Filter("ProductID =", targetSpec.ProductID)
	q = q.Filter("VersionID =", targetSpec.VersionID)
	q = q.Filter("SpecID =", targetSpec.SpecID)
	var tagMessages []*rpc.SpecRevisionTag
	tag := models.SpecRevisionTag{}
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for _, err := it.Next(&tag); err == nil; _, err = it.Next(&tag) {
		tagMessage, _ := tag.Message()
		tagMessages = append(tagMessages, tagMessage)
		if len(tagMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListSpecRevisionTagsResponse{
		Tags: tagMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(tagMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

// RollbackSpec handles the corresponding API request.
func (s *RegistryServer) RollbackSpec(ctx context.Context, request *rpc.RollbackSpecRequest) (*rpc.Spec, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	spec, userSpecifiedRevision, err := fetchSpec(ctx, client, request.GetName())
	if err != nil {
		return nil, err
	}
	if userSpecifiedRevision == "" {
		return nil, invalidArgumentError(errors.New("rollbacks require a specified revision"))
	}
	// The previous current revision needs to be marked non-current.
	oldKey, oldCurrent, err := fetchCurrentRevisionOfSpec(ctx, client, request.GetName())
	if err == nil && oldCurrent != nil {
		oldCurrent.IsCurrent = false
		_, err = client.Put(ctx, oldKey, oldCurrent)
		if err != nil {
			return nil, err
		}
	}
	// Make the selected revision the current revision by giving it a new RevisionID and saving it
	spec.BumpRevision()
	spec.IsCurrent = true
	k := &datastore.Key{Kind: models.SpecEntityName, Name: spec.ResourceNameWithRevision()}
	k, err = client.Put(ctx, k, spec)
	return spec.Message(rpc.SpecView_BASIC, spec.RevisionID)
}

// fetchSpec gets the stored model of a Spec.
func fetchSpec(
	ctx context.Context,
	client *datastore.Client,
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
	var specRevisionTag models.SpecRevisionTag
	k0 := &datastore.Key{Kind: models.SpecRevisionTagEntityName, Name: spec.ResourceNameWithRevision()}
	err = client.Get(ctx, k0, &specRevisionTag)
	if err == datastore.ErrNoSuchEntity {
		// if there is no tag, just use the provided revision
		resourceName = spec.ResourceNameWithRevision()
		revisionName = spec.RevisionID
	} else if err != nil {
		return nil, "", internalError(err)
	} else {
		// if there is a tag, use the revision that the tag references
		resourceName = specRevisionTag.ResourceNameWithRevision()
		revisionName = specRevisionTag.Tag
	}
	// now that we know the revision, use it get the spec
	k := &datastore.Key{Kind: models.SpecEntityName, Name: resourceName}
	err = client.Get(ctx, k, spec)
	if err == datastore.ErrNoSuchEntity {
		return nil, revisionName, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, revisionName, internalError(err)
	}
	return spec, revisionName, nil
}

// fetchMostRecentNonCurrentRevisionOfSpec gets the most recent revision that's not current.
func fetchMostRecentNonCurrentRevisionOfSpec(
	ctx context.Context,
	client *datastore.Client,
	name string,
) (*datastore.Key, *models.Spec, error) {
	pattern, err := models.NewSpecFromResourceName(name)
	if err != nil {
		return nil, nil, err
	}
	// note that we ignore any specified RevisionID
	q := datastore.NewQuery(models.SpecEntityName)
	q = q.Filter("ProjectID =", pattern.ProjectID)
	q = q.Filter("ProductID =", pattern.ProductID)
	q = q.Filter("VersionID =", pattern.VersionID)
	q = q.Filter("SpecID =", pattern.SpecID)
	q = q.Filter("IsCurrent =", false)
	q = q.Order("-CreateTime")
	it := client.Run(ctx, q.Distinct())
	spec := &models.Spec{}
	k, err := it.Next(spec)
	if err != nil {
		return nil, nil, err
	}
	return k, spec, nil
}

// fetchCurrentRevisionOfSpec gets the current revision.
func fetchCurrentRevisionOfSpec(
	ctx context.Context,
	client *datastore.Client,
	name string,
) (*datastore.Key, *models.Spec, error) {
	pattern, err := models.NewSpecFromResourceName(name)
	if err != nil {
		return nil, nil, err
	}
	// note that we ignore any specified RevisionID
	q := datastore.NewQuery(models.SpecEntityName)
	q = q.Filter("ProjectID =", pattern.ProjectID)
	q = q.Filter("ProductID =", pattern.ProductID)
	q = q.Filter("VersionID =", pattern.VersionID)
	q = q.Filter("SpecID =", pattern.SpecID)
	q = q.Filter("IsCurrent =", true)
	it := client.Run(ctx, q.Distinct())
	spec := &models.Spec{}
	k, err := it.Next(spec)
	if err != nil {
		return nil, nil, err
	}
	return k, spec, nil
}
