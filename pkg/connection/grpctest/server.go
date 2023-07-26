// Copyright 2022 Google LLC.
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

package grpctest

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/config"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewIfNoAddress will create a RegistryServer served by a
// basic grpc.Server if env var REGISTRY_ADDRESS is
// not set, see NewServer() for details. If REGISTRY_ADDRESS
// is set, returns nil as client will connect to that remote address.
// Example:
//
//	func TestXXX(t *testing.T) {
//		 l, err := grpctest.NewIfNoAddress(registry.Config{})
//		 if err != nil {
//			t.Fatal(err)
//		 }
//		 defer l.Close()
//	  ... run test here ...
//	}
func NewIfNoAddress(rc registry.Config) (*Server, error) {
	// error doesn't matter here as the point is to allow fallback
	c, _ := connection.ReadConfig("")
	if c.Address != "" {
		log.Printf("Client will use remote registry at: %s", c.Address)
		return nil, nil
	}
	log.Println("Client will use an embedded registry with a SQLite3 database")
	return NewServer(rc)
}

// TestMain can delegate here in packages that wish to
// use TestMain() for tests relying on a RegistryServer.
// Note: As TestMain() is run once per TestXXX() function,
// this will create a new database for the function,
// not per t.Run() within it. See NewIfNoAddress() for details.
// Example:
//
//	func TestMain(m *testing.M) {
//		grpctest.TestMain(m, registry.Config{})
//	}
func TestMain(m *testing.M, rc registry.Config) {
	l, err := NewIfNoAddress(rc)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// ensure tests don't use local config files
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	origConfigPath := config.Directory
	config.Directory = tmpDir
	defer func() { config.Directory = origConfigPath }()
	os.Exit(m.Run())
}

// SetupRegistry connects to a registry instance using the default config,
// deletes the project if it exists, seeds the registry for testing, and
// returns the connections used. It also registers cleanup handlers on the
// test to close the connections and delete the project.
func SetupRegistry(ctx context.Context, t *testing.T, projectID string, seeds []seeder.RegistryResource) (*gapic.RegistryClient, *gapic.AdminClient) {
	t.Helper()
	projectID = strings.TrimPrefix(projectID, "projects/")

	registryClient, err := connection.NewRegistryClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	t.Cleanup(func() { registryClient.Close() })

	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	t.Cleanup(func() { adminClient.Close() })

	delProjReq := &rpc.DeleteProjectRequest{
		Name:  "projects/" + projectID,
		Force: true,
	}
	err = adminClient.DeleteProject(ctx, delProjReq)
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Failed DeleteProject(%v): %s", delProjReq, err.Error())
	}
	t.Cleanup(func() { _ = adminClient.DeleteProject(ctx, delProjReq) })

	if len(seeds) == 0 {
		_, err = adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
			ProjectId: projectID,
			Project: &rpc.Project{
				DisplayName: "Test",
			},
		})
		if err != nil {
			t.Fatalf("Failed to create project: %+v", err)
		}
	}

	seedClient := seeder.Client{
		RegistryClient: registryClient,
		AdminClient:    adminClient,
	}
	if err := seeder.SeedRegistry(ctx, seedClient, seeds...); err != nil {
		t.Fatalf("Setup: failed to seed registry: %s", err)
	}

	return registryClient, adminClient
}

// NewServer creates a RegistryServer served by a basic grpc.Server.
// If rc.Database and rc.DBConfig are blank, a RegistryServer using
// sqlite3 on a tmpDir is automatically created.
// registry.address and registry.insecure configuration values are set
// to cause the in-process client to connect to the created grpc service.
// Call Close() when done to close server and clean up tmpDir as needed.
// Example:
//
//	func TestXXX(t *testing.T) {
//		 l, err := grpctest.NewServer(registry.Config{})
//		 if err != nil {
//			t.Fatal(err)
//		 }
//		 defer l.Close()
//	  ... run test here ...
//	}
func NewServer(rc registry.Config) (*Server, error) {
	s := &Server{}
	var err error
	if rc.Database == "" {
		rc.Database = "sqlite3"
	}
	if rc.Database == "sqlite3" && rc.DBConfig == "" {
		f, err := os.CreateTemp("", "registry.db.*")
		if err != nil {
			return nil, err
		}
		rc.DBConfig = f.Name()
		s.TmpDir = f.Name()
	}

	s.Registry, err = registry.New(rc)
	if err != nil {
		return nil, err
	}

	s.Listener, s.Server, err = s.Registry.ServeGRPC(&net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 0}) // random port
	if err != nil {
		return nil, err
	}

	// set registry configuration to use test server
	addr := fmt.Sprintf("localhost:%d", s.Port())
	conf, err := connection.ActiveConfig()
	if err != nil {
		if ve, ok := err.(config.ValidationError); ok {
			if ve.Field != "registry.address" {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	conf.Insecure = true
	conf.Address = addr
	connection.SetConfig(conf)

	return s, nil
}

// Server wraps a gRPC Server for a RegistryServer
type Server struct {
	Server   *grpc.Server
	Listener net.Listener
	Registry *registry.RegistryServer
	TmpDir   string
}

// Close will gracefully stop the server and remove tmpDir
func (s *Server) Close() {
	if s.Server != nil {
		s.Server.GracefulStop()
	}
	if s.Registry != nil {
		s.Registry.Close()
	}
	if s.TmpDir != "" {
		os.RemoveAll(s.TmpDir)
		s.TmpDir = ""
	}
}

// Address returns the address of the listener
func (s *Server) Address() string {
	if s.Port() == 0 {
		return ""
	}
	return fmt.Sprintf("localhost:%d", s.Port())
}

// Port returns the port of the listener
func (s *Server) Port() int {
	if s.Listener == nil {
		return 0
	}
	return s.Listener.Addr().(*net.TCPAddr).Port
}
