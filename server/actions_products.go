// Copyright 2020 Google Inc. All Rights Reserved.

package server

import (
	"context"
	"time"

	"apigov.dev/flame/models"
	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const productEntityName = "Product"

func (s *server) CreateProduct(ctx context.Context, request *rpc.CreateProductRequest) (*rpc.Product, error) {
	client, err := s.newDataStoreClient(ctx)
	if err != nil {
		return nil, err
	}
	product, err := models.NewProductFromParentAndProductID(request.Parent, request.GetProductId())
	if err != nil {
		return nil, err
	}
	k := &datastore.Key{Kind: productEntityName, Name: product.ResourceName()}
	// fail if product already exists
	var existingProduct models.Product
	err = client.Get(ctx, k, &existingProduct)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, product.ResourceName()+" already exists")
	}
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
	// TODO: delete children
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

	pageSize := req.GetPageSize()
	if pageSize > 1000 {
		pageSize = 1000
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	q := datastore.NewQuery(productEntityName).Limit(int(pageSize))

	cursorStr := req.GetPageToken()
	if cursorStr != "" {
		cursor, err := datastore.DecodeCursor(cursorStr)
		if err != nil {
			return nil, err
		}
		q = q.Start(cursor)
	}

	var products []*models.Product
	var product models.Product
	it := client.Run(ctx, q.Distinct())
	_, err = it.Next(&product)
	for err == nil {
		products = append(products, &product)
		_, err = it.Next(&product)
	}
	if err != iterator.Done {
		return nil, err
	}

	var productMessages []*rpc.Product
	for _, product := range products {
		productMessage, _ := product.Message()
		productMessages = append(productMessages, productMessage)
	}
	responses := &rpc.ListProductsResponse{
		Products: productMessages,
	}

	if len(products) > 0 {
		// Get the cursor for the next page of results.
		nextCursor, err := it.Cursor()
		if err != nil {
			return nil, err
		}
		responses.NextPageToken = nextCursor.String()
	}

	return responses, nil
}

func (s *server) UpdateProduct(ctx context.Context, request *rpc.UpdateProductRequest) (*rpc.Product, error) {
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
