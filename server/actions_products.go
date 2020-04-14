// Copyright 2020 Google LLC. All Rights Reserved.

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

func (s *FlameServer) CreateProduct(ctx context.Context, request *rpc.CreateProductRequest) (*rpc.Product, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	product, err := models.NewProductFromParentAndProductID(request.GetParent(), request.GetProductId())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := &datastore.Key{Kind: models.ProductEntityName, Name: product.ResourceName()}
	// fail if product already exists
	var existingProduct models.Product
	err = client.Get(ctx, k, &existingProduct)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, product.ResourceName()+" already exists")
	}
	err = product.Update(request.GetProduct())
	product.CreateTime = product.UpdateTime
	k, err = client.Put(ctx, k, product)
	if err != nil {
		return nil, internalError(err)
	}
	return product.Message()
}

func (s *FlameServer) DeleteProduct(ctx context.Context, request *rpc.DeleteProductRequest) (*empty.Empty, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	// Validate name and create dummy product (we just need the ID fields).
	product, err := models.NewProductFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete children first and then delete the product.
	product.DeleteChildren(ctx, client)
	k := &datastore.Key{Kind: models.ProductEntityName, Name: request.GetName()}
	err = client.Delete(ctx, k)
	return &empty.Empty{}, internalError(err)
}

func (s *FlameServer) GetProduct(ctx context.Context, request *rpc.GetProductRequest) (*rpc.Product, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	product, err := models.NewProductFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := &datastore.Key{Kind: models.ProductEntityName, Name: product.ResourceName()}
	err = client.Get(ctx, k, product)
	if err == datastore.ErrNoSuchEntity {
		return nil, status.Error(codes.NotFound, "not found")
	} else if err != nil {
		return nil, internalError(err)
	}
	return product.Message()
}

func (s *FlameServer) ListProducts(ctx context.Context, req *rpc.ListProductsRequest) (*rpc.ListProductsResponse, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	q := datastore.NewQuery(models.ProductEntityName)
	q = queryApplyPageSize(q, req.GetPageSize())
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := models.ParseParentProject(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	q = q.Filter("ProjectID =", m[1])
	var productMessages []*rpc.Product
	var product models.Product
	it := client.Run(ctx, q.Distinct())
	_, err = it.Next(&product)
	for err == nil {
		productMessage, _ := product.Message()
		productMessages = append(productMessages, productMessage)
		_, err = it.Next(&product)
	}
	if err != iterator.Done {
		return nil, internalError(err)
	}
	responses := &rpc.ListProductsResponse{
		Products: productMessages,
	}
	responses.NextPageToken, err = iteratorGetCursor(it, len(productMessages))
	if err != nil {
		return nil, internalError(err)
	}
	return responses, nil
}

func (s *FlameServer) UpdateProduct(ctx context.Context, request *rpc.UpdateProductRequest) (*rpc.Product, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	defer client.Close()
	product, err := models.NewProductFromResourceName(request.GetProduct().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := &datastore.Key{Kind: models.ProductEntityName, Name: product.ResourceName()}
	err = client.Get(ctx, k, &product)
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	err = product.Update(request.GetProduct())
	if err != nil {
		return nil, internalError(err)
	}
	k, err = client.Put(ctx, k, product)
	if err != nil {
		return nil, internalError(err)
	}
	return product.Message()
}
