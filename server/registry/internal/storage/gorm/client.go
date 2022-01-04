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
	"sort"
	"sync"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// A complete list of entities used by the storage system
// and managed using gorm features.
var entities = []interface{}{
	&models.Project{},
	&models.Api{},
	&models.Version{},
	&models.Spec{},
	&models.SpecRevisionTag{},
	&models.Deployment{},
	&models.DeploymentRevisionTag{},
	&models.Artifact{},
	&models.Blob{},
}

// Client represents a connection to a storage provider.
type Client struct {
	db *gorm.DB
}

var mutex sync.Mutex
var disableMutex bool

func lock() {
	if !disableMutex {
		mutex.Lock()
	}
}

func unlock() {
	if !disableMutex {
		mutex.Unlock()
	}
}

// NewClient creates a new database session using the provided driver and data source name.
// Driver must be one of [ sqlite3, postgres, cloudsqlpostgres ]. DSN format varies per database driver.
//
// PostgreSQL DSN Reference: See "Connection Strings" at https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
// SQLite DSN Reference: See "URI filename examples" at https://www.sqlite.org/c3ref/open.html
func NewClient(ctx context.Context, driver, dsn string) (*Client, error) {
	lock()
	switch driver {
	case "sqlite3":
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: NewGormLogger(ctx),
		})
		if err != nil {
			c := &Client{db: db}
			c.close()
			unlock()
			return nil, err
		}
		unlock()
		// empirically, it does not seem safe to disable the mutex for sqlite3,
		// which might make sense since sqlite database access is in-process.
		disableMutex = false
		fmt.Printf("DATABASE %+v\n", db.Name())
		return &Client{db: db}, nil
	case "postgres", "cloudsqlpostgres":
		db, err := gorm.Open(postgres.New(postgres.Config{
			DriverName: driver,
			DSN:        dsn,
		}), &gorm.Config{
			Logger: NewGormLogger(ctx),
		})
		if err != nil {
			c := &Client{db: db}
			c.close()
			unlock()
			return nil, err
		}
		unlock()
		// postgres runs in a separate process and seems to have no problems
		// with concurrent access and modifications.
		disableMutex = true
		fmt.Printf("DATABASE %+v\n", db.Name())
		return &Client{db: db}, nil
	default:
		unlock()
		return nil, fmt.Errorf("unsupported database %s", driver)
	}
}

// Close closes a database session.
func (c *Client) Close() {
	lock()
	defer unlock()
	c.close()
}

func (c *Client) close() {
	sqlDB, _ := c.db.DB()
	sqlDB.Close()
}

func (c *Client) ensureTable(v interface{}) error {
	lock()
	defer unlock()
	if !c.db.Migrator().HasTable(v) {
		if err := c.db.Migrator().CreateTable(v); err != nil {
			return err
		}
	}
	return nil
}

// EnsureTables ensures that all necessary tables exist in the database.
func (c *Client) EnsureTables() error {
	for _, entity := range entities {
		if err := c.ensureTable(entity); err != nil {
			return err
		}
	}
	return nil
}

// IsNotFound returns true if an error is due to an entity not being found.
func (c *Client) IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

// Get gets an entity using the storage client.
func (c *Client) Get(ctx context.Context, k *Key, v interface{}) error {
	lock()
	defer unlock()
	return c.db.Where("key = ?", k.Name).First(v).Error
}

// Put puts an entity using the storage client.
func (c *Client) Put(ctx context.Context, k *Key, v interface{}) (*Key, error) {
	lock()
	defer unlock()
	switch r := v.(type) {
	case *models.Project:
		r.Key = k.Name
	case *models.Api:
		r.Key = k.Name
	case *models.Version:
		r.Key = k.Name
	case *models.Spec:
		r.Key = k.Name
	case *models.SpecRevisionTag:
		r.Key = k.Name
	case *models.Deployment:
		r.Key = k.Name
	case *models.DeploymentRevisionTag:
		r.Key = k.Name
	case *models.Artifact:
		r.Key = k.Name
	case *models.Blob:
		r.Key = k.Name
	}
	_ = c.db.Transaction(func(tx *gorm.DB) error {
		// Update all fields from model: https://gorm.io/docs/update.html#Update-Selected-Fields
		rowsAffected := tx.Model(v).Select("*").Where("key = ?", k.Name).Updates(v).RowsAffected
		if rowsAffected == 0 {
			tx.Create(v)
		}
		return nil
	})
	return k, nil
}

// Delete deletes all entities matching a query.
func (c *Client) Delete(ctx context.Context, q *Query) error {
	op := c.db
	for _, r := range q.Requirements {
		op = op.Where(r.Name+" = ?", r.Value)
	}
	switch q.Kind {
	case "Project":
		return op.Delete(models.Project{}).Error
	case "Api":
		return op.Delete(models.Api{}).Error
	case "Version":
		return op.Delete(models.Version{}).Error
	case "Spec":
		return op.Delete(models.Spec{}).Error
	case "SpecRevisionTag":
		return op.Delete(models.SpecRevisionTag{}).Error
	case "Deployment":
		return op.Delete(models.Deployment{}).Error
	case "DeploymentRevisionTag":
		return op.Delete(models.DeploymentRevisionTag{}).Error
	case "Artifact":
		return op.Delete(models.Artifact{}).Error
	case "Blob":
		return op.Delete(models.Blob{}).Error
	}
	return nil
}

// Run runs a query using the storage client, returning an iterator.
func (c *Client) Run(ctx context.Context, q *Query) *Iterator {
	lock()
	defer unlock()

	// Filtering is currently implemented by skipping iterator elements that
	// don't match the filter criteria, and expects to only reach the end of
	// the iterator if there are no more resources to consider. Previously,
	// the entire table would be read into memory. This limit should maintain
	// that behavior until we improve our iterator implementation.
	op := c.db.Offset(q.Offset).Limit(100000)
	for _, r := range q.Requirements {
		op = op.Where(r.Name+" = ?", r.Value)
	}

	if order := q.Order; order != "" {
		op = op.Order(order)
	} else {
		op = op.Order("key")
	}

	switch q.Kind {
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
	case "SpecRevisionTag":
		var v []models.SpecRevisionTag
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	case "Deployment":
		var v []models.Deployment
		_ = op.Find(&v).Error
		return &Iterator{Client: c, Values: v, Index: 0}
	case "DeploymentRevisionTag":
		var v []models.DeploymentRevisionTag
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
	default:
		return nil
	}
}

func (c *Client) GetRecentSpecRevisions(ctx context.Context, offset int32, projectID, apiID, versionID string) *Iterator {
	lock()
	defer unlock()

	// Select all columns from `specs` table specifically.
	// We do not want to select duplicates from the joined subquery result.
	op := c.db.Select("specs.*").
		Table("specs").
		// Join missing columns that couldn't be selected in the subquery.
		Joins("JOIN (?) AS grp ON specs.project_id = grp.project_id AND specs.api_id = grp.api_id AND specs.version_id = grp.version_id AND specs.spec_id = grp.spec_id AND specs.revision_create_time = grp.recent_create_time",
			// Select spec names and only their most recent revision_create_time
			// This query cannot select all the columns we want.
			// See: https://stackoverflow.com/questions/7745609/sql-select-only-rows-with-max-value-on-a-column
			c.db.Select("project_id, api_id, version_id, spec_id, MAX(revision_create_time) AS recent_create_time").
				Table("specs").
				Group("project_id, api_id, version_id, spec_id")).
		Order("key").
		Offset(int(offset)).
		Limit(100000)

	if projectID != "-" {
		op = op.Where("specs.project_id = ?", projectID)
	}
	if apiID != "-" {
		op = op.Where("specs.api_id = ?", apiID)
	}
	if versionID != "-" {
		op = op.Where("specs.version_id = ?", versionID)
	}

	var v []models.Spec
	_ = op.Scan(&v).Error
	return &Iterator{Client: c, Values: v, Index: 0}
}

func (c *Client) GetRecentDeploymentRevisions(ctx context.Context, offset int32, projectID, apiID string) *Iterator {
	lock()
	defer unlock()

	// Select all columns from `deployments` table specifically.
	// We do not want to select duplicates from the joined subquery result.
	op := c.db.Select("deployments.*").
		Table("deployments").
		// Join missing columns that couldn't be selected in the subquery.
		Joins("JOIN (?) AS grp ON deployments.project_id = grp.project_id AND deployments.api_id = grp.api_id AND deployments.deployment_id = grp.deployment_id AND deployments.revision_create_time = grp.recent_create_time",
			// Select deployment names and only their most recent revision_create_time
			// This query cannot select all the columns we want.
			// See: https://stackoverflow.com/questions/7745609/sql-select-only-rows-with-max-value-on-a-column
			c.db.Select("project_id, api_id, deployment_id, MAX(revision_create_time) AS recent_create_time").
				Table("deployments").
				Group("project_id, api_id, deployment_id")).
		Order("key").
		Offset(int(offset)).
		Limit(100000)

	if projectID != "-" {
		op = op.Where("deployments.project_id = ?", projectID)
	}
	if apiID != "-" {
		op = op.Where("deployments.api_id = ?", apiID)
	}

	var v []models.Deployment
	_ = op.Scan(&v).Error
	return &Iterator{Client: c, Values: v, Index: 0}
}

func (c *Client) Migrate(kind string) error {
	return c.db.AutoMigrate(entities...)
}

func (c *Client) GetStatus() (*rpc.Status, error) {
	var dbSize int64
	if c.db.Name() == "postgres" {
		c.db.Raw("SELECT pg_database_size(current_database())").Scan(&dbSize)
	} else if c.db.Name() == "sqlite" {
		c.db.Raw("SELECT SUM(pgsize) FROM dbstat").Scan(&dbSize)
	}
	var tableNames []string
	if c.db.Name() == "postgres" {
		if err := c.db.Table("information_schema.tables").Where("table_schema = ?", "public").Pluck("table_name", &tableNames).Error; err != nil {
			panic(err)
		}
	} else if c.db.Name() == "sqlite" {
		if err := c.db.Table("sqlite_schema").Where("type = 'table' AND name NOT LIKE 'sqlite_%'").Pluck("name", &tableNames).Error; err != nil {
			panic(err)
		}
	}
	sort.Strings(tableNames)
	tables := make([]*rpc.Status_Table, 0)
	for _, tableName := range tableNames {
		var entityCount int64
		err := c.db.Table(tableName).Count(&entityCount).Error
		if err != nil {
			return nil, err
		}
		var tableSize int64
		if c.db.Name() == "postgres" {
			c.db.Raw("SELECT pg_total_relation_size(?)", tableName).Scan(&tableSize)
		} else if c.db.Name() == "sqlite" {
			c.db.Raw("SELECT SUM(pgsize) FROM dbstat WHERE name = ?", tableName).Scan(&tableSize)
		}
		tables = append(tables,
			&rpc.Status_Table{
				Name:        tableName,
				EntityCount: entityCount,
				SizeBytes:   tableSize,
			},
		)
	}
	return &rpc.Status{
		Message:   "running",
		SizeBytes: dbSize,
		Tables:    tables,
	}, nil
}
