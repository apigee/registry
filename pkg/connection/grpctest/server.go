// Copyright 2022 Google LLC. All Rights Reserved.
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
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"

	"github.com/apigee/registry/pkg/config"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/server/registry"
	"google.golang.org/grpc"
)

// NewIfNoAddress will create a RegistryServer served by a
// basic grpc.Server if env var APG_REGISTRY_ADDRESS is
// not set, see NewServer() for details. If APG_REGISTRY_ADDRESS
// is set, returns nil as client will connect to that remote address.
// Example:
// func TestXXX(t *testing.T) {
// 	 l, err := grpctest.NewIfNoAddress(registry.Config{})
// 	 if err != nil {
// 		t.Fatal(err)
// 	 }
// 	 defer l.Close()
//   ... run test here ...
// }
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
// func TestMain(m *testing.M) {
// 	grpctest.TestMain(m, registry.Config{})
// }
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

// NewServer creates a RegistryServer served by a basic grpc.Server.
// If rc.Database and rc.DBConfig are blank, a RegistryServer using
// sqlite3 on a tmpDir is automatically created.
// APG_REGISTRY_ADDRESS and APG_REGISTRY_INSECURE
// env vars are set for the client to connect to the created grpc service.
// Call Close() when done to close server and clean up tmpDir as needed.
// Example:
// func TestXXX(t *testing.T) {
// 	 l, err := grpctest.NewServer(registry.Config{})
// 	 if err != nil {
// 		t.Fatal(err)
// 	 }
// 	 defer l.Close()
//   ... run test here ...
// }
func NewServer(rc registry.Config) (*Server, error) {
	s := &Server{}
	var err error
	if rc.Database == "" {
		rc.Database = "sqlite3"
	}
	if rc.Database == "sqlite3" && rc.DBConfig == "" {
		f, err := ioutil.TempFile("", "registry.db.*")
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
	config, err := connection.ActiveConfig()
	if err != nil {
		return nil, err
	}
	config.Insecure = true
	config.Address = addr
	connection.SetConfig(config)

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
