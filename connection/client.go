// Copyright 2020 Google LLC. All Rights Reserved.
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

package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/apigee/registry/gapic"
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
	} else {
		return nil, fmt.Errorf("rpc error: APG_REGISTRY_ADDRESS must be set")
	}

	insecure := os.Getenv("APG_REGISTRY_INSECURE")
	if insecure != "" {
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
