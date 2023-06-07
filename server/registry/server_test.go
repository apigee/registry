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

package registry

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/remote"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	postgresDriver           = "postgres"
	postgresDBConfig         = "host=localhost port=5432 user=registry_tester dbname=registry_test sslmode=disable"
	testRequiresAdminService = "test requires admin service, skipping"
)

var (
	sharedStorage sync.Mutex
	usePostgres   = false
	useRemote     = false
	hostedProject = ""
)

func adminServiceUnavailable() bool {
	return hostedProject != ""
}

func init() {
	flag.BoolVar(&usePostgres, "postgresql", false, "perform server tests using postgresql")
	flag.BoolVar(&useRemote, "remote", false, "perform server tests using a remote server")
	flag.StringVar(&hostedProject, "hosted", "", "perform server tests using a remote server with the specified project and no Admin service")
}

type TestServer interface {
	rpc.AdminServer
	rpc.RegistryServer
}

// defaultTestServer will call server.Close() when test completes
func defaultTestServer(t *testing.T) TestServer {
	t.Helper()
	var err error
	var server *RegistryServer

	if hostedProject != "" {
		p := remote.NewProxyForHostedService(hostedProject)
		if err = p.Open(context.Background()); err != nil {
			t.Fatalf("Setup: failed to connect to remote server: %s", err)
		}
		return p
	}
	if useRemote {
		p := &remote.Proxy{}
		if err = p.Open(context.Background()); err != nil {
			t.Fatalf("Setup: failed to connect to remote server: %s", err)
		}
		return p
	}
	if usePostgres {
		server, err = serverWithPostgres(t)
		if err != nil {
			t.Fatalf("Setup: failed to get server with postgres: %s", err)
		}
	}
	if server == nil {
		if server, err = serverWithSQLite(t); err != nil {
			t.Fatalf("Setup: failed to get server with SQLite: %s", err)
		}
	}

	return server
}

// serverWithSQLite will call server.Close() when test completes
func serverWithSQLite(t *testing.T) (*RegistryServer, error) {
	server, err := New(Config{
		Database: "sqlite3",
		DBConfig: fmt.Sprintf("%s/registry.db", t.TempDir()),
	})
	if server != nil {
		t.Cleanup(server.Close)
	}
	return server, err
}

// serverWithPostgres will call server.Close() when test completes
func serverWithPostgres(t *testing.T) (*RegistryServer, error) {
	sharedStorage.Lock()
	t.Cleanup(sharedStorage.Unlock)

	if err := resetPostgres(); err != nil {
		return nil, fmt.Errorf("failed to reset database: %s", err)
	}

	server, err := New(Config{
		Database: postgresDriver,
		DBConfig: postgresDBConfig,
	})
	if server != nil {
		t.Cleanup(server.Close)
	}
	return server, err
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
