package server

import (
	"context"

	"apigov.dev/registry/rpc"
	emptypb "github.com/golang/protobuf/ptypes/empty"
)

// GetStatus handles the corresponding API request.
func (s *RegistryServer) GetStatus(ctx context.Context, request *emptypb.Empty) (*rpc.Status, error) {
	status := &rpc.Status{
		Message: "running",
	}
	return status, nil
}
