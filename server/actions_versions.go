// Copyright 2020 Google Inc. All Rights Reserved.

package server

import (
	"context"
	"time"

	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const versionEntityName = "Version"

func (s *server) CreateVersion(ctx context.Context, request *rpc.CreateVersionRequest) (*rpc.Version, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	version, err := models.NewVersionFromMessage(request.Version)
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: versionEntityName, Name: version.ResourceName()}
	version.CreateTime = time.Now()
	version.UpdateTime = version.CreateTime
	k, err = client.Put(ctx, k, version)
	if err != nil {
		return nil, err
	}
	return version.Message()
}

func (s *server) DeleteVersion(ctx context.Context, request *rpc.DeleteVersionRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	// validate name
	_, err = models.NewVersionFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: versionEntityName, Name: request.GetName()}
	// TODO: delete children
	err = client.Delete(ctx, k)
	return &empty.Empty{}, err
}

func (s *server) GetVersion(ctx context.Context, request *rpc.GetVersionRequest) (*rpc.Version, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	version, err := models.NewVersionFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: versionEntityName, Name: version.ResourceName()}
	err = client.Get(ctx, k, &version)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return version.Message()
}

func (s *server) ListVersions(ctx context.Context, req *rpc.ListVersionsRequest) (*rpc.ListVersionsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	q := datastore.NewQuery(versionEntityName)
	var versions []*models.Version
	_, err = client.GetAll(ctx, q, &versions)
	var versionMessages []*rpc.Version
	for _, version := range versions {
		versionMessage, _ := version.Message()
		versionMessages = append(versionMessages, versionMessage)
	}
	responses := &rpc.ListVersionsResponse{
		Versions: versionMessages,
	}
	return responses, nil
}

func (s *server) UpdateVersion(ctx context.Context, request *rpc.UpdateVersionRequest) (*rpc.Version, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	version, err := models.NewVersionFromResourceName(request.GetVersion().GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: versionEntityName, Name: version.ResourceName()}
	err = client.Get(ctx, k, &version)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = version.Update(request.GetVersion())
	if err != nil {
		return nil, err
	}
	k, err = client.Put(ctx, k, version)
	if err != nil {
		return nil, err
	}
	return version.Message()
}
