// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"context"

	"cloud.google.com/go/datastore"
	"github.com/apigee/registry/server/storage"
)

// Client represents a connection to a storage provider.
// In this module, entities are stored using the Cloud Datastore API.
// https://cloud.google.com/datastore/
type Client struct {
	client *datastore.Client
}

var globalClient *Client

// NewClient creates a new storage client.
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	if globalClient != nil {
		return globalClient, nil
	}
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	globalClient = &Client{client: client}
	return globalClient, nil
}

// Close closes the database connection.
func (c *Client) Close() {
	// Do nothing because we maintain a single global connection.
}

// IsNotFound returns true if an error is due to an entity not being found.
func (c *Client) IsNotFound(err error) bool {
	return err == datastore.ErrNoSuchEntity
}

// NotFoundError is the error returned when an entity is not found.
func (c *Client) NotFoundError() error {
	return datastore.ErrNoSuchEntity
}

// Get gets an entity using the storage client.
func (c *Client) Get(ctx context.Context, k storage.Key, v interface{}) error {
	return c.client.Get(ctx, k.(*Key).key, v)
}

// Put puts an entity using the storage client.
func (c *Client) Put(ctx context.Context, k storage.Key, v interface{}) (storage.Key, error) {
	key, err := c.client.Put(ctx, k.(*Key).key, v)
	return &Key{key: key}, err
}

// Delete deletes an entity using the storage client.
func (c *Client) Delete(ctx context.Context, k storage.Key) error {
	return c.client.Delete(ctx, k.(*Key).key)
}

// Run runs a query using the storage client, returning an iterator.
func (c *Client) Run(ctx context.Context, q storage.Query) storage.Iterator {
	return &Iterator{iterator: c.client.Run(ctx, q.(*Query).query.Distinct())}
}
