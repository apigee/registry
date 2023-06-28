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
	"errors"
	"log"
	"net"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
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
	NoMigrate bool
}

// RegistryServer implements a Registry server.
type RegistryServer struct {
	database      string
	dbConfig      string
	notifyEnabled bool
	projectID     string
	storageClient *storage.Client
	pubSubClient  *pubsub.Client

	rpc.UnimplementedRegistryServer
	rpc.UnimplementedAdminServer
}

func New(config Config) (*RegistryServer, error) {
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

	var err error
	ctx := context.Background()
	s.storageClient, err = storage.NewClient(ctx, s.database, s.dbConfig)
	if err != nil {
		return nil, err
	}
	if err := s.storageClient.EnsureTables(ctx); err != nil {
		return nil, err
	}
	if !config.NoMigrate {
		if err := s.storageClient.Migrate(ctx); err != nil {
			return nil, err
		}
	}

	if s.notifyEnabled {
		s.pubSubClient, err = pubsub.NewClient(ctx, s.projectID)
		if err != nil {
			return nil, err
		}
		if _, err := s.pubSubClient.CreateTopic(ctx, TopicName); err != nil && status.Code(err) != codes.AlreadyExists {
			return nil, err
		}
	}

	return s, nil
}

func (s *RegistryServer) getStorageClient(ctx context.Context) (*storage.Client, error) {
	if s.storageClient == nil {
		return nil, errors.New("no storageClient")
	}
	return s.storageClient, nil
}

var mutex sync.Mutex

func (s *RegistryServer) runInTransaction(ctx context.Context, fn func(ctx context.Context, db *storage.Client) error) error {
	s.begin()
	defer s.end()
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return status.Error(codes.Unavailable, err.Error())
	}
	return db.Transaction(ctx, fn)
}

func (s *RegistryServer) begin() {
	if s.database == "sqlite3" {
		mutex.Lock()
	}
}

func (s *RegistryServer) end() {
	if s.database == "sqlite3" {
		mutex.Unlock()
	}
}

func (s *RegistryServer) getPubSubClient(ctx context.Context) (*pubsub.Client, error) {
	if s.pubSubClient == nil {
		return nil, errors.New("no pubSubClient")
	}
	return s.pubSubClient, nil
}

func (s *RegistryServer) Close() {
	s.storageClient.Close()
	if s.pubSubClient != nil {
		s.pubSubClient.Topic(TopicName).Flush()
	}
}

func isNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}

// GRPCListen starts a net.Listener and grpc.Server for this RegistryServer.
// Caller is responsible for stopping server.
func (rs *RegistryServer) ServeGRPC(addr *net.TCPAddr, opt ...grpc.ServerOption) (net.Listener, *grpc.Server, error) {
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	s := grpc.NewServer(opt...)
	reflection.Register(s)
	rpc.RegisterRegistryServer(s, rs)
	rpc.RegisterAdminServer(s, rs)

	go func() {
		if err := s.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()

	return l, s, err
}
