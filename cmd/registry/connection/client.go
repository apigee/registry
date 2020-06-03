package connection

import (
	"context"
	"fmt"
	"os"

	"apigov.dev/registry/gapic"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// NewClient ...
func NewClient() (*gapic.RegistryClient, error) {
	var opts []option.ClientOption

	address := os.Getenv("CLI_REGISTRY_ADDRESS")
	if address != "" {
		opts = append(opts, option.WithEndpoint(address))
	}

	insecure := os.Getenv("CLI_REGISTRY_INSECURE")
	if insecure != "" {
		if address == "" {
			return nil, fmt.Errorf("Missing address to use with insecure connection")
		}
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		opts = append(opts, option.WithGRPCConn(conn))
	}

	if token := os.Getenv("CLI_REGISTRY_TOKEN"); token != "" {
		opts = append(opts, option.WithTokenSource(oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: token,
				TokenType:   "Bearer",
			})))
	}
	ctx := context.TODO()
	return gapic.NewRegistryClient(ctx, opts...)
}
