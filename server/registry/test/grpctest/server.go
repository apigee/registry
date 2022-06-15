package grpctest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewIfNoAddress will create a RegistryServer served by a
// basic grpc.Server if env var APG_REGISTRY_ADDRESS is
// not set, otherwise will simply return a dummy Server
// assuming the tests will connect to a the remote address.
func NewIfNoAddress(rc registry.Config) (*Server, error) {
	addr := os.Getenv("APG_REGISTRY_ADDRESS")
	if addr != "" {
		return &Server{}, nil
	}
	return NewServer(registry.Config{})
}

// TestMain can delegate here in packages that wish to
// use TestMain for tests using a RegistryServer. This
// uses NewIfNoAddress() semantics.
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

	rs, err := registry.New(rc)
	if err != nil {
		return nil, err
	}

	gl.Listener, err = net.Listen("tcp", "localhost:0") // random port
	if err != nil {
		gl.Close()
		return nil, err
	}

	gl.Server = grpc.NewServer()
	reflection.Register(gl.Server)
	rpc.RegisterRegistryServer(gl.Server, rs)
	rpc.RegisterAdminServer(gl.Server, rs)

	go func() {
		if err := gl.Server.Serve(gl.Listener); err != nil {
			log.Fatal(err)
		}
	}()

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
	TmpDir   string
}

// Close will gracefully stop the server and remove tmpDir
func (g *Server) Close() {
	if g.Server != nil {
		g.Server.GracefulStop()
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
