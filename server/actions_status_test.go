package server

import (
	"context"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetStatus(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)

	req := &empty.Empty{}
	want := &rpc.Status{
		Message: "running",
	}

	got, err := server.GetStatus(ctx, req)
	if err != nil {
		t.Fatalf("GetStatus(%+v) returned error: %s", req, err)
	}

	if !cmp.Equal(want, got, protocmp.Transform()) {
		t.Errorf("GetStatus(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(want, got, protocmp.Transform()))
	}
}
