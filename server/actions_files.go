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

const fileEntityName = "File"

func (s *server) CreateFile(ctx context.Context, request *rpc.CreateFileRequest) (*rpc.File, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	file, err := models.NewFileFromMessage(request.File)
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: fileEntityName, Name: file.ResourceName()}
	file.CreateTime = time.Now()
	file.UpdateTime = file.CreateTime
	k, err = client.Put(ctx, k, file)
	if err != nil {
		return nil, err
	}
	return file.Message()
}

func (s *server) DeleteFile(ctx context.Context, request *rpc.DeleteFileRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	// validate name
	_, err = models.NewFileFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: fileEntityName, Name: request.GetName()}
	// TODO: delete children
	err = client.Delete(ctx, k)
	return &empty.Empty{}, err
}

func (s *server) GetFile(ctx context.Context, request *rpc.GetFileRequest) (*rpc.File, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	file, err := models.NewFileFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: fileEntityName, Name: file.ResourceName()}
	err = client.Get(ctx, k, &file)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return file.Message()
}

func (s *server) ListFiles(ctx context.Context, req *rpc.ListFilesRequest) (*rpc.ListFilesResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	q := datastore.NewQuery(fileEntityName)
	var files []*models.File
	_, err = client.GetAll(ctx, q, &files)
	var fileMessages []*rpc.File
	for _, file := range files {
		fileMessage, _ := file.Message()
		fileMessages = append(fileMessages, fileMessage)
	}
	responses := &rpc.ListFilesResponse{
		Files: fileMessages,
	}
	return responses, nil
}

func (s *server) UpdateFile(ctx context.Context, request *rpc.UpdateFileRequest) (*rpc.File, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	file, err := models.NewFileFromResourceName(request.GetFile().GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: fileEntityName, Name: file.ResourceName()}
	err = client.Get(ctx, k, &file)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = file.Update(request.GetFile())
	if err != nil {
		return nil, err
	}
	k, err = client.Put(ctx, k, file)
	if err != nil {
		return nil, err
	}
	return file.Message()
}
