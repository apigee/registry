// Copyright 2020 Google LLC.
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

package storage

import (
	"context"
	"fmt"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/apigee/registry/server/registry/internal/storage/models"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// NewClient creates a new database session using the provided driver and data source name.
// Driver must be one of [ sqlite3, postgres, cloudsqlpostgres ]. DSN format varies per database driver.
//
// PostgreSQL DSN Reference: See "Connection Strings" at https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
// SQLite DSN Reference: See "URI filename examples" at https://www.sqlite.org/c3ref/open.html
func NewClient(ctx context.Context, driver, dsn string) (*Client, error) {
	switch driver {
	case "sqlite3":
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger:      NewGormLogger(ctx),
			PrepareStmt: true,
		})
		if err != nil {
			c := &Client{db: db}
			c.close()
			return nil, grpcErrorForDBError(ctx, err)
		}
		// Sets to 1 to disallow multiple connections on SQLite.
		// Any automatically opened connections, such a connection pool
		// or SetConnMaxLifetime() will not apply post-connect
		// PRAGMA commands (eg. "foreign_keys = ON").
		if err := applyConnectionLimits(db, 1); err != nil {
			c := &Client{db: db}
			c.close()
			return nil, grpcErrorForDBError(ctx, err)
		}
		if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			return nil, grpcErrorForDBError(ctx, err)
		}
		return &Client{db: db}, nil
	case "postgres", "cloudsqlpostgres":
		db, err := gorm.Open(postgres.New(postgres.Config{
			DriverName: driver,
			DSN:        dsn,
		}), &gorm.Config{
			Logger:      NewGormLogger(ctx),
			PrepareStmt: true,
		})
		if err != nil {
			c := &Client{db: db}
			c.close()
			return nil, grpcErrorForDBError(ctx, err)
		}
		if err := applyConnectionLimits(db, 10); err != nil {
			c := &Client{db: db}
			c.close()
			return nil, grpcErrorForDBError(ctx, err)
		}
		return &Client{db: db}, nil
	default:
		return nil, fmt.Errorf("unsupported database %s", driver)
	}
}

// Applies limits to concurrent connections.
func applyConnectionLimits(db *gorm.DB, n int) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(n)
	sqlDB.SetMaxIdleConns(n)
	return nil
}

// Close closes a database session.
func (c *Client) Close() {
	c.close()
}

func (c *Client) close() {
	sqlDB, _ := c.db.DB()
	sqlDB.Close()
}

func (c *Client) ensureTable(ctx context.Context, v interface{}) error {
	if !c.db.Migrator().HasTable(v) {
		if err := c.db.Migrator().CreateTable(v); err != nil {
			return grpcErrorForDBError(ctx, errors.Wrapf(err, "create table %#v", v))
		}
	}
	return nil
}

// EnsureTables ensures that all necessary tables exist in the database.
func (c *Client) EnsureTables(ctx context.Context) error {
	for _, entity := range entities {
		if err := c.ensureTable(ctx, entity); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Migrate(ctx context.Context) error {
	if err := c.db.WithContext(ctx).AutoMigrate(entities...); err != nil {
		return grpcErrorForDBError(ctx, err)
	}

	if err := c.ensureForeignKeys(ctx); err != nil {
		return grpcErrorForDBError(ctx, err)
	}

	if err := c.migrateArtifactsToRevisions(ctx); err != nil {
		return grpcErrorForDBError(ctx, err)
	}

	return nil
}

func (c *Client) ensureForeignKeys(ctx context.Context) (err error) {
	err = c.db.Model(&models.Api{}).
		Where("parent_project_key is null").
		Update("parent_project_key",
			c.db.Model(&models.Project{}).Select("key").
				Where("apis.project_id = projects.project_id")).Error
	if err != nil {
		return err
	}

	err = c.db.Model(&models.Version{}).
		Where("parent_api_key is null").
		Update("parent_api_key",
			c.db.Model(&models.Api{}).Select("key").
				Where("versions.project_id = apis.project_id").
				Where("versions.api_id = apis.api_id")).Error
	if err != nil {
		return err
	}

	err = c.db.Model(&models.Spec{}).
		Where("parent_version_key is null").
		Update("parent_version_key",
			c.db.Model(&models.Version{}).Select("key").
				Where("specs.project_id = versions.project_id").
				Where("specs.api_id = versions.api_id").
				Where("specs.version_id = versions.version_id")).Error
	if err != nil {
		return err
	}

	err = c.db.Model(&models.SpecRevisionTag{}).
		Where("parent_spec_key is null").
		Update("parent_spec_key",
			c.db.Model(&models.Spec{}).Select("key").
				Where("spec_revision_tags.project_id = specs.project_id").
				Where("spec_revision_tags.api_id = specs.api_id").
				Where("spec_revision_tags.version_id = specs.version_id").
				Where("spec_revision_tags.spec_id = specs.spec_id").
				Where("spec_revision_tags.revision_id = specs.revision_id")).Error
	if err != nil {
		return err
	}

	return c.db.Model(&models.Deployment{}).
		Where("parent_api_key is null").
		Update("parent_api_key",
			c.db.Model(&models.Api{}).Select("key").
				Where("deployments.project_id = apis.project_id").
				Where("deployments.api_id = apis.api_id")).Error
}

func (c *Client) migrateArtifactsToRevisions(ctx context.Context) (err error) {
	return c.db.Exec(`
	UPDATE artifacts
	SET revision_id = specs.revision_id
	FROM
		-- ONLY LASTEST SPEC REVISIONS
		(SELECT *
		FROM
			(SELECT *, row_number()
			OVER
			   (partition by project_id,api_id,version_id,spec_id order by revision_create_time desc) as row
			FROM specs) 
		AS rev
		WHERE rev.row = 1) as specs
	WHERE
		artifacts.project_id = specs.project_id 
		AND artifacts.api_id = specs.api_id 
		AND artifacts.version_id = specs.version_id 
		AND artifacts.spec_id = specs.spec_id
		AND artifacts.revision_id is null
	`).Error
}

func (c *Client) DatabaseName(ctx context.Context) string {
	return c.db.WithContext(ctx).Name()
}

func (c *Client) TableNames(ctx context.Context) ([]string, error) {
	var tableNames []string
	switch c.db.WithContext(ctx).Name() {
	case "postgres":
		if err := c.db.WithContext(ctx).Table("information_schema.tables").Where("table_schema = ?", "public").Order("table_name").Pluck("table_name", &tableNames).Error; err != nil {
			return nil, grpcErrorForDBError(ctx, errors.Wrap(err, "tables"))
		}
	case "sqlite":
		if err := c.db.WithContext(ctx).Table("sqlite_schema").Where("type = 'table' AND name NOT LIKE 'sqlite_%'").Order("name").Pluck("name", &tableNames).Error; err != nil {
			return nil, grpcErrorForDBError(ctx, errors.Wrap(err, "tables"))
		}
	default:
		return nil, status.Errorf(codes.Internal, "unsupported database %s", c.db.Name())
	}
	return tableNames, nil
}

func (c *Client) RowCount(ctx context.Context, tableName string) (int64, error) {
	var count int64
	err := c.db.WithContext(ctx).Table(tableName).Count(&count).Error
	return count, grpcErrorForDBError(ctx, errors.Wrapf(err, "count %s", tableName))
}

func (c *Client) Transaction(ctx context.Context, fn func(context.Context, *Client) error) error {
	err := c.db.Transaction(func(tx *gorm.DB) error {
		return fn(ctx, &Client{db: tx})
	})
	return grpcErrorForDBError(ctx, err)
}
