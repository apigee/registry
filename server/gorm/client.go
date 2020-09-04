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
	"github.com/apigee/registry/server/storage"
	"github.com/jinzhu/gorm"

	// encapsulate these dependencies
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Client represents a connection to a storage provider.
// In this module, entities are stored using the Cloud Datastore API.
// https://cloud.google.com/datastore/
type Client struct {
	db *gorm.DB
}

// NewClient creates a new database session.
func NewClient(ctx context.Context, gormDB, gormConfig string) (*Client, error) {
	db, err := gorm.Open(gormDB, gormConfig)
	if err != nil {
		panic(err)
	}
	return (&Client{db: db}).ensure(), nil
}

func (c *Client) resetTable(v interface{}) {
	if c.db.HasTable(v) == true {
		c.db.DropTable(v)
	}
	if c.db.HasTable(v) == false {
		c.db.CreateTable(v)
	}
}

func (c *Client) ensureTable(v interface{}) {
	if c.db.HasTable(v) == false {
		c.db.CreateTable(v)
	}
}

func (c *Client) reset() {
	c.resetTable(&models.Project{})
	c.resetTable(&models.Api{})
	c.resetTable(&models.Version{})
	c.resetTable(&models.Spec{})
	c.resetTable(&models.Blob{})
	c.resetTable(&models.Property{})
	c.resetTable(&models.Label{})
	c.resetTable(&models.SpecRevisionTag{})
}

func (c *Client) ensure() *Client {
	c.ensureTable(&models.Project{})
	c.ensureTable(&models.Api{})
	c.ensureTable(&models.Version{})
	c.ensureTable(&models.Spec{})
	c.ensureTable(&models.Blob{})
	c.ensureTable(&models.Property{})
	c.ensureTable(&models.Label{})
	c.ensureTable(&models.SpecRevisionTag{})
	return c
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
func (c *Client) Get(ctx context.Context, k storage.Key, v interface{}) error {
	err := c.db.Where("key = ?", k.(*Key).Name).First(v).Error
	return err
}

// Put puts an entity using the storage client.
func (c *Client) Put(ctx context.Context, k storage.Key, v interface{}) (storage.Key, error) {
	switch r := v.(type) {
	case *models.Project:
		r.Key = k.(*Key).Name
	case *models.Api:
		r.Key = k.(*Key).Name
	case *models.Version:
		r.Key = k.(*Key).Name
	case *models.Spec:
		r.Key = k.(*Key).Name
	case *models.SpecRevisionTag:
		r.Key = k.(*Key).Name
	case *models.Blob:
		r.Key = k.(*Key).Name
	case *models.Property:
		r.Key = k.(*Key).Name
	case *models.Label:
		r.Key = k.(*Key).Name
	}
	rowsAffected := c.db.Model(v).Where("key = ?", k.(*Key).Name).Updates(v).RowsAffected
	if rowsAffected == 0 {
		c.db.Create(v)
	}
	return k, nil
}

// Delete deletes an entity using the storage client.
func (c *Client) Delete(ctx context.Context, k storage.Key) error {
	var err error
	switch k.(*Key).Kind {
	case "Project":
		err = c.db.Delete(&models.Project{}, "key = ?", k.(*Key).Name).Error
	case "Blob":
		err = c.db.Delete(&models.Blob{}, "key = ?", k.(*Key).Name).Error
	case "Spec":
		err = c.db.Delete(&models.Spec{}, "key = ?", k.(*Key).Name).Error
	default:
		return fmt.Errorf("invalid key type (fix in client.go): %s", k.(*Key).Kind)
	}
	if err != nil {
		log.Printf("ignoring error: %+v", err)
	}
	return nil
}

// Run runs a query using the storage client, returning an iterator.
func (c *Client) Run(ctx context.Context, q storage.Query) storage.Iterator {
	limit := q.(*Query).Limit
	cursor := q.(*Query).Cursor

	op := c.db.Limit(limit).Where("key > ?", cursor)
	for _, r := range q.(*Query).Requirements {
		op = op.Where(r.Name+" = ?", r.Value)
	}

	switch q.(*Query).Kind {
	case "Project":
		var v []models.Project
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	case "Api":
		var v []models.Api
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	case "Version":
		var v []models.Version
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	case "Spec":
		var v []models.Spec
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	case "Blob":
		var v []models.Blob
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	case "Property":
		var v []models.Property
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	case "Label":
		var v []models.Label
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	case "SpecRevisionTag":
		var v []models.SpecRevisionTag
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	default:
		log.Printf("Unable to run query for kind %s", q.(*Query).Kind)
		return nil
	}
}
