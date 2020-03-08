// Copyright 2020 Google Inc. All Rights Reserved.

package server

import (
	"context"

	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const fileEntityName = "File"

func (s *server) CreateFile(ctx context.Context, request *rpc.CreateFileRequest) (*rpc.File, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	file, err := models.NewFileFromParentAndFileID(request.GetParent(), request.GetFileId())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: fileEntityName, Name: file.ResourceName()}
	// fail if file already exists
	var existingFile models.File
	err = client.Get(ctx, k, &existingFile)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, file.ResourceName()+" already exists")
	}
	file.CreateTime = file.UpdateTime
	err = file.Update(request.GetFile())
	k, err = client.Put(ctx, k, file)
	if err != nil {
		return nil, err
	}
	return file.Message(rpc.FileView_FULL)
}

func (s *server) DeleteFile(ctx context.Context, request *rpc.DeleteFileRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
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
	defer client.Close()
	file, err := models.NewFileFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: fileEntityName, Name: file.ResourceName()}
	err = client.Get(ctx, k, file)
	if err == datastore.ErrNoSuchEntity {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return file.Message(request.GetView())
}

func (s *server) ListFiles(ctx context.Context, req *rpc.ListFilesRequest) (*rpc.ListFilesResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	q := datastore.NewQuery(fileEntityName)
	q = queryApplyPageSize(q, req.GetPageSize())
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := models.ParseParentSpec(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	q = q.Filter("ProjectID =", m[1])
	q = q.Filter("ProductID =", m[2])
	q = q.Filter("VersionID =", m[3])
	q = q.Filter("SpecID =", m[4])
	var fileMessages []*rpc.File
	var file models.File
	it := client.Run(ctx, q.Distinct())
	_, err = it.Next(&file)
	for err == nil {
		fileMessage, _ := file.Message(req.GetView())
		fileMessages = append(fileMessages, fileMessage)
		_, err = it.Next(&file)
	}
	if err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListFilesResponse{
		Files: fileMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(fileMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

func (s *server) UpdateFile(ctx context.Context, request *rpc.UpdateFileRequest) (*rpc.File, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
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
	return file.Message(rpc.FileView_FULL)
}
