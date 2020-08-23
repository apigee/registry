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

// ErrNotFound is returned when an entity is not found.
func (c *Client) ErrNotFound() error {
	return datastore.ErrNoSuchEntity
}

// Get gets an entity using the storage client.
func (c *Client) Get(ctx context.Context, k *Key, v interface{}) error {
	return c.client.Get(ctx, k, v)
}

// Put puts an entity using the storage client.
func (c *Client) Put(ctx context.Context, k *Key, v interface{}) (*Key, error) {
	return c.client.Put(ctx, k, v)
}

// Delete deletes an entity using the storage client.
func (c *Client) Delete(ctx context.Context, k *Key) error {
	return c.client.Delete(ctx, k)
}

// Run runs a query using the storage client, returning an iterator.
func (c *Client) Run(ctx context.Context, q *Query) *Iterator {
	return c.client.Run(ctx, q)
}
