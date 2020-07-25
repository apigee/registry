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

// Client is a client of the Registry API
type Client = *gapic.RegistryClient

// NewClient creates a new GAPIC client using environment variable settings.
func NewClient(ctx context.Context) (Client, error) {
	var opts []option.ClientOption

	address := os.Getenv("APG_REGISTRY_ADDRESS")
	if address != "" {
		opts = append(opts, option.WithEndpoint(address))
	}

	insecure := os.Getenv("APG_REGISTRY_INSECURE")
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

	if token := os.Getenv("APG_REGISTRY_TOKEN"); token != "" {
		opts = append(opts, option.WithTokenSource(oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: token,
				TokenType:   "Bearer",
			})))
	}
	return gapic.NewRegistryClient(ctx, opts...)
}
