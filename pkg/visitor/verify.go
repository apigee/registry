package visitor

import (
	"context"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
)

func VerifyLocation(ctx context.Context, client connection.RegistryClient, location string) error {
	_, err := client.GrpcClient().ListApis(ctx, &rpc.ListApisRequest{
		Parent:   location,
		PageSize: 1,
	})
	return err
}
