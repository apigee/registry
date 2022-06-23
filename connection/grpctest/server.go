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
	addr := os.Getenv("APG_REGISTRY_ADDRESS")
	if addr != "" {
		log.Printf("Client will use remote registry at: %s", addr)
		return nil, nil
	}
	log.Println("Client will use an embedded registry with a SQLite3 database")
	return NewServer(registry.Config{})
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
	os.Exit(m.Run())
}

// NewServer creates a RegistryServer served by a basic grpc.Server.
// If rc.Database and rc.DBConfig are blank, a RegistryServer using
// sqlite3 on a tmpDir is automatically created.
// APG_REGISTRY_ADDRESS, APG_REGISTRY_AUDIENCES, and APG_REGISTRY_INSECURE
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
	gl := &Server{}
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
		gl.TmpDir = f.Name()
	}

	gl.Registry, err = registry.New(rc)
	if err != nil {
		return nil, err
	}

	gl.Listener, gl.Server, err = gl.Registry.ServeGRPC(&net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 0}) // random port
	if err != nil {
		return nil, err
	}

	// set for internal client
	addr := fmt.Sprintf("localhost:%d", gl.Port())
	os.Setenv("APG_REGISTRY_ADDRESS", addr)
	os.Setenv("APG_REGISTRY_AUDIENCES", fmt.Sprintf("http://%s", addr))
	os.Setenv("APG_REGISTRY_INSECURE", "1")

	return gl, nil
}

// Server wraps a gRPC Server for a RegistryServer
type Server struct {
	Server   *grpc.Server
	Listener net.Listener
	Registry *registry.RegistryServer
	TmpDir   string
}

// Close will gracefully stop the server and remove tmpDir
func (g *Server) Close() {
	if g.Server != nil {
		g.Server.GracefulStop()
	}
	if g.Registry != nil {
		g.Registry.Close()
	}
	if g.TmpDir != "" {
		os.RemoveAll(g.TmpDir)
		g.TmpDir = ""
	}
}

// Address returns the address of the listener
func (g *Server) Address() string {
	if g.Port() == 0 {
		return ""
	}
	return fmt.Sprintf("localhost:%d", g.Port())
}

// Port returns the port of the listener
func (g *Server) Port() int {
	if g.Listener == nil {
		return 0
	}
	return g.Listener.Addr().(*net.TCPAddr).Port
}
