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
	"fmt"
	"log"

	"github.com/apigee/registry/server/models"
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

func (c *Client) reset() {
	project := &models.Project{}
	c.db.DropTable(&project)
	if c.db.HasTable(&project) == false {
		c.db.CreateTable(&project)
	}
}

// Close closes a database session.
func (c *Client) Close() {
	c.db.Close()
}

// IsNotFound returns true if an error is due to an entity not being found.
func (c *Client) IsNotFound(err error) bool {
	return gorm.IsRecordNotFoundError(err)
}

// Get gets an entity using the storage client.
func (c *Client) Get(ctx context.Context, k *Key, v interface{}) error {
	err := c.db.Where("key = ?", k.Name).First(v).Error
	return err
}

// Put puts an entity using the storage client.
func (c *Client) Put(ctx context.Context, k *Key, v interface{}) (*Key, error) {
	switch r := v.(type) {
	case *models.Project:
		r.Key = k.Name
	}
	log.Printf("calling CREATE")
	if c.db.Model(v).Where("key = ?", k.Name).Updates(v).RowsAffected == 0 {
		c.db.Create(v)
	}
	return k, nil
}

// Delete deletes an entity using the storage client.
func (c *Client) Delete(ctx context.Context, k *Key) error {
	var v interface{}
	switch k.Kind {
	case "project":
		v = &models.Project{Key: k.Name}
	default:
		return fmt.Errorf("invalid key type: %s", k.Kind)
	}
	c.db.Delete(v)
	return nil
}

// Run runs a query using the storage client, returning an iterator.
func (c *Client) Run(ctx context.Context, q *Query) *Iterator {
	return nil
}
