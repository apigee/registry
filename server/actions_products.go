// Copyright 2020 Google LLC. All Rights Reserved.

package server

import (
	"context"

	"apigov.dev/registry/rpc"
	"apigov.dev/registry/server/models"
	"apigov.dev/registry/server/names"
	"cloud.google.com/go/datastore"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateProduct handles the corresponding API request.
func (s *RegistryServer) CreateProduct(ctx context.Context, request *rpc.CreateProductRequest) (*rpc.Product, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
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
	s.notify(rpc.Notification_CREATED, product.ResourceName())
	return product.Message()
}

// DeleteProduct handles the corresponding API request.
func (s *RegistryServer) DeleteProduct(ctx context.Context, request *rpc.DeleteProductRequest) (*empty.Empty, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
	// Validate name and create dummy product (we just need the ID fields).
	product, err := models.NewProductFromResourceName(request.GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	// Delete children first and then delete the product.
	product.DeleteChildren(ctx, client)
	k := &datastore.Key{Kind: models.ProductEntityName, Name: request.GetName()}
	err = client.Delete(ctx, k)
	s.notify(rpc.Notification_DELETED, request.GetName())
	return &empty.Empty{}, internalError(err)
}

// GetProduct handles the corresponding API request.
func (s *RegistryServer) GetProduct(ctx context.Context, request *rpc.GetProductRequest) (*rpc.Product, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
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

// ListProducts handles the corresponding API request.
func (s *RegistryServer) ListProducts(ctx context.Context, req *rpc.ListProductsRequest) (*rpc.ListProductsResponse, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
	q := datastore.NewQuery(models.ProductEntityName)
	q, err = queryApplyCursor(q, req.GetPageToken())
	if err != nil {
		return nil, internalError(err)
	}
	m, err := names.ParseParentProject(req.GetParent())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	if m[1] != "-" {
		q = q.Filter("ProjectID =", m[1])
	}
	prg, err := createFilterOperator(req.GetFilter(),
		[]filterArg{
			{"product_id", filterArgTypeString},
			{"display_name", filterArgTypeString},
			{"description", filterArgTypeString},
			{"availability", filterArgTypeString},
		})
	if err != nil {
		return nil, internalError(err)
	}
	var productMessages []*rpc.Product
	var product models.Product
	it := client.Run(ctx, q.Distinct())
	pageSize := boundPageSize(req.GetPageSize())
	for _, err = it.Next(&product); err == nil; _, err = it.Next(&product) {
		if prg != nil {
			out, _, err := prg.Eval(map[string]interface{}{
				"product_id":   product.ProductID,
				"display_name": product.DisplayName,
				"description":  product.Description,
				"availability": product.Availability,
			})
			if err != nil {
				return nil, invalidArgumentError(err)
			}
			if !out.Value().(bool) {
				continue
			}
		}
		productMessage, _ := product.Message()
		productMessages = append(productMessages, productMessage)
		if len(productMessages) == pageSize {
			break
		}
	}
	if err != nil && err != iterator.Done {
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

// UpdateProduct handles the corresponding API request.
func (s *RegistryServer) UpdateProduct(ctx context.Context, request *rpc.UpdateProductRequest) (*rpc.Product, error) {
	client, err := s.getDataStoreClient(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	s.releaseDataStoreClient(client)
	product, err := models.NewProductFromResourceName(request.GetProduct().GetName())
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	k := &datastore.Key{Kind: models.ProductEntityName, Name: product.ResourceName()}
	err = client.Get(ctx, k, product)
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
	s.notify(rpc.Notification_UPDATED, product.ResourceName())
	return product.Message()
}
