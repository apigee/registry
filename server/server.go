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
	"context"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/internal/storage"
	"github.com/google/uuid"

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
}

// RegistryServer implements a Registry server.
type RegistryServer struct {
	database      string
	dbConfig      string
	notifyEnabled bool
	loggingLevel  string
	loggingFormat string
	projectID     string

	rpc.UnimplementedRegistryServer
}

func New(config Config) *RegistryServer {
	s := &RegistryServer{
		database:      config.Database,
		dbConfig:      config.DBConfig,
		loggingLevel:  config.LogLevel,
		loggingFormat: config.LogFormat,
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

// Start runs the Registry server using the provided listener.
// It blocks until the context is cancelled.
func (s *RegistryServer) Start(ctx context.Context, listener net.Listener) {
	var (
		logInterceptor = grpc.UnaryInterceptor(s.loggingInterceptor)
		grpcServer     = grpc.NewServer(logInterceptor)
	)

	reflection.Register(grpcServer)
	rpc.RegisterRegistryServer(grpcServer, s)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	// Block until the context is cancelled.
	<-ctx.Done()
}

func (s *RegistryServer) logger(ctx context.Context) log.Interface {
	return log.FromContext(ctx)
}

func (s *RegistryServer) loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger := new(log.Logger)
	switch s.loggingLevel {
	case "debug":
		logger.Level = log.DebugLevel
	case "info":
		logger.Level = log.InfoLevel
	case "warn":
		logger.Level = log.WarnLevel
	case "error":
		logger.Level = log.ErrorLevel
	case "fatal":
		logger.Level = log.FatalLevel
	default:
		logger.Level = log.InfoLevel
	}

	switch s.loggingFormat {
	case "json":
		logger.Handler = json.Default
	case "text":
		logger.Handler = text.Default
	default:
		logger.Handler = text.Default
	}

	reqInfo := log.Fields{
		"request_id": fmt.Sprintf("%.8s", uuid.New()),
		"method":     filepath.Base(info.FullMethod),
	}

	type (
		resourceOp   interface{ GetName() string }
		collectionOp interface{ GetParent() string }
	)

	switch r := req.(type) {
	case resourceOp:
		reqInfo["resource"] = r.GetName()
	case collectionOp:
		reqInfo["collection"] = r.GetParent()
	}

	// Bind request-scoped attributes to the context logger before handling the request.
	reqLogger := logger.WithFields(reqInfo)
	ctx = log.NewContext(ctx, reqLogger)

	reqLogger.Info("Handling request.")
	start := time.Now()
	resp, err := handler(ctx, req)

	respInfo := log.Fields{
		"duration":    time.Since(start),
		"status_code": status.Code(err),
	}

	if r, ok := resp.(resourceOp); err == nil && ok {
		respInfo["resource"] = r.GetName()
	}

	// Bind response details before logging a response.
	respLogger := reqLogger.WithFields(respInfo)

	// Error messages may include a status code, but we want to log messages and codes separately.
	if err != nil {
		st, _ := status.FromError(err)
		unwrapped := errors.New(st.Message())
		respLogger = respLogger.WithError(unwrapped)
	}

	switch status.Code(err) {
	case codes.OK:
		respLogger.Info("Success.")
	case codes.Internal:
		respLogger.Error("Internal error.")
	case codes.Unknown:
		respLogger.Error("Unknown error.")
	default:
		respLogger.Info("User error.")
	}

	return resp, err
}

func isNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}
