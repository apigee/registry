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

package gorm

import (
	"context"
	"errors"

	"github.com/jinzhu/gorm"
)

// Client represents a connection to a storage provider.
// In this module, entities are stored using the Cloud Datastore API.
// https://cloud.google.com/datastore/
type Client struct {
	db *gorm.DB
}

// NewClient creates a new database session.
func NewClient() *Client {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=registry dbname=registry password=iloveapis")
	if err != nil {
		panic(err)
	}
	return &Client{db: db}
}

// Close closes a database session.
func (c *Client) Close() {
	c.db.Close()
}

// ErrNotFound is returned when an entity is not found.
func (c *Client) ErrNotFound() error {
	return errors.New("Not Found")
}

// Get gets an entity using the storage client.
func (c *Client) Get(ctx context.Context, k *Key, v interface{}) error {
	c.db.Where("key = ?", *k).First(v)
	return nil
}
