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

package server

import (
	"flag"
	"fmt"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	postgresDriver   = "postgres"
	postgresDBConfig = "host=localhost port=5432 user=registry_tester dbname=registry_test sslmode=disable"
)

var (
	sharedStorage sync.Mutex
	usePostgres   = false
)

func init() {
	flag.BoolVar(&usePostgres, "postgresql", false, "perform server tests using postgresql")
}

func defaultTestServer(t *testing.T) *RegistryServer {
	t.Helper()

	if !usePostgres {
		return serverWithSQLite(t)
	}

	server, err := serverWithPostgres(t)
	if err != nil {
		t.Errorf("Setup: failed to get server with postgres: %s", err)
		t.Log("Falling back to server with SQLite storage")
		return serverWithSQLite(t)
	}

	return server
}

func serverWithSQLite(t *testing.T) *RegistryServer {
	return New(Config{
		Database: "sqlite3",
		DBConfig: fmt.Sprintf("%s/registry.db", t.TempDir()),
	})
}

func serverWithPostgres(t *testing.T) (*RegistryServer, error) {
	sharedStorage.Lock()
	t.Cleanup(sharedStorage.Unlock)

	if err := resetPostgres(); err != nil {
		return nil, fmt.Errorf("failed to reset database: %s", err)
	}

	return New(Config{
		Database: postgresDriver,
		DBConfig: postgresDBConfig,
	}), nil
}

func resetPostgres() error {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DriverName: postgresDriver,
		DSN:        postgresDBConfig,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %s", err)
	}

	if err := db.Exec("DROP owned BY registry_tester").Error; err != nil {
		return fmt.Errorf("failed to drop test user's contents: %s", err)
	}

	if sqlDB, err := db.DB(); err != nil {
		return fmt.Errorf("failed to get database for closing: %s", err)
	} else if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close test database: %s", err)
	}

	return nil
}
