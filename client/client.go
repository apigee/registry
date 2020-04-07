package client

import (
	"context"
	"fmt"
	"os"

	"apigov.dev/flame/gapic"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// NewClient creates a new GAPIC client using environment variable settings.
func NewClient() (*gapic.FlameClient, error) {
	var opts []option.ClientOption

	address := os.Getenv("CLI_FLAME_ADDRESS")
	if address != "" {
		opts = append(opts, option.WithEndpoint(address))
	}

	insecure := os.Getenv("CLI_FLAME_INSECURE")
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

	if token := os.Getenv("CLI_FLAME_TOKEN"); token != "" {
		opts = append(opts, option.WithTokenSource(oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: token,
				TokenType:   "Bearer",
			})))
	}
	ctx := context.TODO()
	return gapic.NewFlameClient(ctx, opts...)
}
