// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"
	"log"

	"apigov.dev/registry/models"
	rpc "apigov.dev/registry/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RegistryServer) CreateTag(ctx context.Context, request *rpc.CreateTagRequest) (*rpc.Tag, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	tag, err := models.NewTagFromParentAndTagID(request.GetParent(), request.GetTagId())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	err = tag.Update(request.GetTag())
	k := &datastore.Key{Kind: models.TagEntityName, Name: tag.ResourceName()}
	// fail if tag already exists
	var existingTag models.Tag
	err = client.Get(ctx, k, &existingTag)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, tag.ResourceName()+" already exists")
	}
	tag.CreateTime = tag.UpdateTime
	k, err = client.Put(ctx, k, tag)
	if err != nil {
		return nil, internalError(err)
	}
	return tag.Message()
}

func (s *RegistryServer) DeleteTag(ctx context.Context, request *rpc.DeleteTagRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	// Validate name and create dummy tag (we just need the ID fields).
	_, err = models.NewTagFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete the tag.
	k := &datastore.Key{Kind: models.TagEntityName, Name: request.GetName()}
	err = client.Delete(ctx, k)
	return &empty.Empty{}, internalError(err)
}

func (s *RegistryServer) GetTag(ctx context.Context, request *rpc.GetTagRequest) (*rpc.Tag, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	tag, err := models.NewTagFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	log.Printf("looking for %s", tag.ResourceName())
	k := &datastore.Key{Kind: models.TagEntityName, Name: tag.ResourceName()}
	err = client.Get(ctx, k, tag)
	if err == datastore.ErrNoSuchEntity {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return tag.Message()
}

func (s *RegistryServer) ListTags(ctx context.Context, req *rpc.ListTagsRequest) (*rpc.ListTagsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	q := datastore.NewQuery(models.TagEntityName)
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	p, err := models.NewTagFromParentAndTagID(req.GetParent(), "-")
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if p.ProjectID != "-" && p.ProjectID != "-" {
		q = q.Filter("ProjectID =", p.ProjectID)
	}
	if p.ProductID != "-" && p.ProductID != "-" {
		q = q.Filter("ProductID =", p.ProductID)
	}
	if p.VersionID != "-" && p.VersionID != "-" {
		q = q.Filter("VersionID =", p.VersionID)
	}
	if p.SpecID != "-" && p.SpecID != "-" {
		q = q.Filter("SpecID =", p.SpecID)
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"tag_id", filterArgTypeString},
		})
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	var tagMessages []*rpc.Tag
	var tag models.Tag
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&tag); err == nil; _, err = it.Next(&tag) {
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"tag_id": tag.TagID,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		tagMessage, _ := tag.Message()
		tagMessages = append(tagMessages, tagMessage)
		if len(tagMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListTagsResponse{
		Tags: tagMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(tagMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

func (s *RegistryServer) UpdateTag(ctx context.Context, request *rpc.UpdateTagRequest) (*rpc.Tag, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	tag, err := models.NewTagFromResourceName(request.GetTag().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := &datastore.Key{Kind: models.TagEntityName, Name: request.GetTag().GetName()}
	err = client.Get(ctx, k, tag)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = tag.Update(request.GetTag())
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, tag)
	if err != nil {
		return nil, internalError(err)
	}
	return tag.Message()
}
