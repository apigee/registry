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

package registry

import (
	"context"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config configures the registry server.
type Config struct {
	Database  string
	DBConfig  string
	LogLevel  string
	LogFormat string
	Notify    bool
	ProjectID string
}

// RegistryServer implements a Registry server.
type RegistryServer struct {
	database      string
	dbConfig      string
	notifyEnabled bool
	projectID     string

	rpc.UnimplementedRegistryServer
	rpc.UnimplementedAdminServer
}

func New(config Config) *RegistryServer {
	s := &RegistryServer{
		database:      config.Database,
		dbConfig:      config.DBConfig,
		notifyEnabled: config.Notify,
		projectID:     config.ProjectID,
	}

	if s.database == "" {
		s.database = "sqlite3"
		s.dbConfig = "/tmp/registry.db"
	}

	return s
}

func (s *RegistryServer) getStorageClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx, s.database, s.dbConfig)
}

func isNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}
