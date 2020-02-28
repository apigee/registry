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

const productEntityName = "Product"

func (s *server) CreateProduct(ctx context.Context, request *rpc.CreateProductRequest) (*rpc.Product, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	product, err := models.NewProductFromMessage(request.Product)
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: productEntityName, Name: product.ResourceName()}
	product.CreateTime = time.Now()
	product.UpdateTime = product.CreateTime
	k, err = client.Put(ctx, k, product)
	if err != nil {
		return nil, err
	}
	return product.Message()
}

func (s *server) DeleteProduct(ctx context.Context, request *rpc.DeleteProductRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	// validate name
	_, err = models.NewProductFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: productEntityName, Name: request.GetName()}
	// TODO: delete all the children of this product
	err = client.Delete(ctx, k)
	return &empty.Empty{}, err
}

func (s *server) GetProduct(ctx context.Context, request *rpc.GetProductRequest) (*rpc.Product, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	product, err := models.NewProductFromResourceName(request.GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: productEntityName, Name: product.ResourceName()}
	err = client.Get(ctx, k, &product)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return product.Message()
}

func (s *server) ListProducts(ctx context.Context, req *rpc.ListProductsRequest) (*rpc.ListProductsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	q := datastore.NewQuery(productEntityName)
	var products []*models.Product
	_, err = client.GetAll(ctx, q, &products)
	var productMessages []*rpc.Product
	for _, product := range products {
		productMessage, _ := product.Message()
		productMessages = append(productMessages, productMessage)
	}
	responses := &rpc.ListProductsResponse{
		Products: productMessages,
	}
	return responses, nil
}

func (s *server) UpdateProduct(ctx context.Context, request *rpc.CreateProductRequest) (*rpc.Product, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	product, err := models.NewProductFromResourceName(request.GetProduct().GetName())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: productEntityName, Name: product.ResourceName()}
	err = client.Get(ctx, k, &product)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = product.Update(request.GetProduct())
	if err != nil {
		return nil, err
	}
	k, err = client.Put(ctx, k, product)
	if err != nil {
		return nil, err
	}
	return product.Message()
}
