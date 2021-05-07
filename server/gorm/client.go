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
	"bytes"
	"context"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/storage"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// Client represents a connection to a storage provider.
type Client struct {
	db *gorm.DB
}

var mutex sync.Mutex
var disableMutex bool

func mylock() {
	if !disableMutex {
		mutex.Lock()
	}
}

func myunlock() {
	if !disableMutex {
		mutex.Unlock()
	}
}

func config() *gorm.Config {
	return &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // https://gorm.io/docs/logger.html
	}
}

var clientCount int
var clientTotal int

var openErrorCount int

// NewClient creates a new database session.
func NewClient(ctx context.Context, gormDBName, gormConfig string) (*Client, error) {
	mylock()
	clientCount++
	clientTotal++
	switch gormDBName {
	case "sqlite3":
		if (!cgoEnabled) {
			myunlock()
			return nil, fmt.Errorf("%s is unavailable. Please recompile with CGO_ENABLED=1 or use a different database.", gormDBName)
		}
		db, err := gorm.Open(sqlite.Open(gormConfig), config())
		if err != nil {
			openErrorCount++
			log.Printf("OPEN ERROR %d", openErrorCount)
			(&Client{db: db}).close()
			myunlock()
			return nil, err
		}
		myunlock()
		// empirically, it does not seem safe to disable the mutex for sqlite3,
		// which might make sense since sqlite database access is in-process.
		//disableMutex = true
		c := (&Client{db: db}).ensure()
		return c, nil
	case "postgres", "cloudsqlpostgres":
		db, err := gorm.Open(postgres.New(postgres.Config{
			DriverName: gormDBName,
			DSN:        gormConfig,
		}), config())
		if err != nil {
			openErrorCount++
			log.Printf("OPEN ERROR %d", openErrorCount)
			(&Client{db: db}).close()
			myunlock()
			return nil, err
		}
		myunlock()
		// postgres runs in a separate process and seems to have no problems
		// with concurrent access and modifications.
		disableMutex = true
		c := (&Client{db: db}).ensure()
		return c, nil
	default:
		myunlock()
		return nil, fmt.Errorf("Unsupported database %s", gormDBName)
	}
}

// Close closes a database session.
func (c *Client) Close() {
	mylock()
	defer myunlock()
	c.close()
}

func (c *Client) close() {
	clientCount--
	sqlDB, _ := c.db.DB()
	sqlDB.Close()
}

func (c *Client) resetTable(v interface{}) {
	mylock()
	defer myunlock()
	if c.db.Migrator().HasTable(v) == true {
		c.db.Migrator().DropTable(v)
	}
	if c.db.Migrator().HasTable(v) == false {
		c.db.Migrator().CreateTable(v)
	}
}

func (c *Client) ensureTable(v interface{}) {
	mylock()
	defer myunlock()
	if c.db.Migrator().HasTable(v) == false {
		c.db.Migrator().CreateTable(v)
	}
}

func (c *Client) reset() {
	c.resetTable(&models.Project{})
	c.resetTable(&models.Api{})
	c.resetTable(&models.Version{})
	c.resetTable(&models.Spec{})
	c.resetTable(&models.Blob{})
	c.resetTable(&models.Artifact{})
	c.resetTable(&models.SpecRevisionTag{})
}

func (c *Client) ensure() *Client {
	c.ensureTable(&models.Project{})
	c.ensureTable(&models.Api{})
	c.ensureTable(&models.Version{})
	c.ensureTable(&models.Spec{})
	c.ensureTable(&models.Blob{})
	c.ensureTable(&models.Artifact{})
	c.ensureTable(&models.SpecRevisionTag{})
	return c
}

// IsNotFound returns true if an error is due to an entity not being found.
func (c *Client) IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

// NotFoundError is the error returned when an entity is not found.
func (c *Client) NotFoundError() error {
	return gorm.ErrRecordNotFound
}

// Get gets an entity using the storage client.
func (c *Client) Get(ctx context.Context, k storage.Key, v interface{}) error {
	mylock()
	defer myunlock()
	return c.db.Where("key = ?", k.(*Key).Name).First(v).Error
}

// Put puts an entity using the storage client.
func (c *Client) Put(ctx context.Context, k storage.Key, v interface{}) (storage.Key, error) {
	mylock()
	defer myunlock()
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
	case *models.Artifact:
		r.Key = k.(*Key).Name
	}
	c.db.Transaction(
		func(tx *gorm.DB) error {
			// Update all fields from model: https://gorm.io/docs/update.html#Update-Selected-Fields
			rowsAffected := tx.Model(v).Select("*").Where("key = ?", k.(*Key).Name).Updates(v).RowsAffected
			if rowsAffected == 0 {
				err := tx.Create(v).Error
				if err != nil {
					log.Printf("CREATE ERROR %s", err.Error())
				}
			}
			return nil
		})
	return k, nil
}

// Delete deletes an entity using the storage client.
func (c *Client) Delete(ctx context.Context, k storage.Key) error {
	mylock()
	defer myunlock()
	var err error
	switch k.(*Key).Kind {
	case "Project":
		err = c.db.Delete(&models.Project{}, "key = ?", k.(*Key).Name).Error
	case "Api":
		err = c.db.Delete(&models.Api{}, "key = ?", k.(*Key).Name).Error
	case "Version":
		err = c.db.Delete(&models.Version{}, "key = ?", k.(*Key).Name).Error
	case "Spec":
		err = c.db.Delete(&models.Spec{}, "key = ?", k.(*Key).Name).Error
	case "SpecRevisionTag":
		err = c.db.Delete(&models.SpecRevisionTag{}, "key = ?", k.(*Key).Name).Error
	case "Blob":
		err = c.db.Delete(&models.Blob{}, "key = ?", k.(*Key).Name).Error
	case "Artifact":
		err = c.db.Delete(&models.Artifact{}, "key = ?", k.(*Key).Name).Error
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
	mylock()
	defer myunlock()
	cursor := q.(*Query).Cursor

	op := c.db.Where("key > ?", cursor)
	for _, r := range q.(*Query).Requirements {
		op = op.Where(r.Name+" = ?", r.Value)
	}
	op = op.Order("key")

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
	case "Artifact":
		var v []models.Artifact
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
