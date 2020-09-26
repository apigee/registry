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
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/datastore"
	"github.com/apigee/registry/server/gorm"
	"github.com/apigee/registry/server/storage"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/soheilhy/cmux"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Config configures the registry server.
type Config struct {
	Database string `yaml:"database"`
	DBConfig string `yaml:"dbconfig"`
	Notify   bool   `yaml:"notify"`
	Log      string `yaml:"log"`
}

const (
	loggingFatal = iota
	loggingError = iota
	loggingWarn  = iota
	loggingInfo  = iota
	loggingDebug = iota
)

// RegistryServer implements a Registry server.
type RegistryServer struct {
	// Uncomment the following line when adding new methods.
	// rpc.UnimplementedRegistryServer

	database            string // configured
	dbConfig            string // configured
	enableNotifications bool   // configured
	logging             string // configured

	projectID      string // computed
	weTrustTheSort bool   // computed
	loggingLevel   int    // computed
}

func newRegistryServer(config *Config) *RegistryServer {
	s := &RegistryServer{}
	if config != nil {
		s.database = config.Database
		s.dbConfig = config.DBConfig
		s.enableNotifications = config.Notify
		switch strings.ToUpper(config.Log) {
		case "FATAL":
			s.loggingLevel = loggingFatal
		case "ERRORS":
			s.loggingLevel = loggingError
		case "WARNINGS":
			s.loggingLevel = loggingWarn
		case "INFO":
			s.loggingLevel = loggingInfo
		case "DEBUG":
			s.loggingLevel = loggingDebug
		default:
			s.loggingLevel = loggingFatal
		}
	}
	return s
}

func (s *RegistryServer) getStorageClient(ctx context.Context) (storage.Client, error) {
	// Cloud Datastore is the default.
	if s.database == "" || s.database == "datastore" {
		s.weTrustTheSort = true
		projectID, err := s.getProjectID()
		if err != nil {
			return nil, err
		}
		return datastore.NewClient(ctx, projectID)
	}
	// If we're not using Cloud Datastore, attempt to connect to a database using gorm.
	s.weTrustTheSort = false
	return gorm.NewClient(ctx, s.database, s.dbConfig)
}

// if we had one client per handler, this would close the client.
func (s *RegistryServer) releaseStorageClient(client storage.Client) {
	client.Close()
}

func (s *RegistryServer) getProjectID() (string, error) {
	if s.projectID != "" {
		return s.projectID, nil
	}
	ctx := context.TODO()
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}
	s.projectID = credentials.ProjectID
	if s.projectID == "" {
		s.projectID = os.Getenv("REGISTRY_PROJECT_IDENTIFIER")
	}
	if s.projectID == "" {
		return "", fmt.Errorf("unable to determine project ID")
	}
	return s.projectID, nil
}

var serverMutex sync.Mutex
var serverSerialization bool

// RunServer runs the Registry server on a specified port
func RunServer(port string, config *Config) error {
	// Construct Registry API server (request handler).
	r := newRegistryServer(config)
	// Construct gRPC server.
	loggingHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if serverSerialization {
			serverMutex.Lock()
			defer serverMutex.Unlock()
		}
		method := filepath.Base(info.FullMethod)
		if r.loggingLevel >= loggingInfo {
			log.Printf(">> %s", method)
		}
		resp, err := handler(ctx, req)
		if err != nil {
			if isNotFound(err) { // only log "not found" at DEBUG and higher
				if r.loggingLevel >= loggingDebug {
					log.Printf("[%s] %s", method, err.Error())
				}
			} else { // log all other problems at ERROR and higher
				if r.loggingLevel >= loggingError {
					log.Printf("[%s] %s", method, err.Error())
				}
			}
		}
		return resp, err
	}
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(loggingHandler))
	reflection.Register(grpcServer)
	rpc.RegisterRegistryServer(grpcServer, r)
	// Create a listener and use it to run the server.
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	log.Printf("registry-server listening on %s", port)
	// Use a cmux to route incoming requests by protocol.
	m := cmux.New(listener)
	// Match gRPC requests and serve them in a goroutine.
	grpcListener := m.Match(cmux.HTTP2())
	go grpcServer.Serve(grpcListener)
	// Match HTTP1 requests (including gRPC-Web) and serve them in a goroutine.
	httpListener := m.Match(cmux.HTTP1Fast())
	httpServer := &http.Server{
		Handler: &httpHandler{grpcWebServer: grpcweb.WrapServer(grpcServer)},
	}
	go httpServer.Serve(httpListener)
	// Run the mux server.
	return m.Serve()
}

type httpHandler struct {
	grpcWebServer *grpcweb.WrappedGrpcServer
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// handle gRPC Web requests
	if h.grpcWebServer.IsGrpcWebRequest(r) {
		h.grpcWebServer.ServeHTTP(w, r)
		return
	}
	// handle any other requests
	log.Printf("%+v", r)
	http.NotFound(w, r)
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.NotFound
}
