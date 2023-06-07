// Copyright 2020 Google LLC.
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

	"github.com/apigee/registry/gapic"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func clientOptions(config Config) ([]option.ClientOption, error) {
	var opts []option.ClientOption
	if config.Address == "" {
		return nil, fmt.Errorf("rpc error: address must be set")
	}
	opts = append(opts, option.WithEndpoint(config.Address))
	if config.Insecure {
		conn, err := grpc.Dial(config.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		opts = append(opts, option.WithGRPCConn(conn))
	}
	if config.Token != "" {
		opts = append(opts, option.WithTokenSource(oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: config.Token,
				TokenType:   "Bearer",
			})))
	}
	return opts, nil
}

// RegistryClient is a client of the Registry API
type RegistryClient = *gapic.RegistryClient

// NewRegistryClient creates a new client using the active Config.
func NewRegistryClient(ctx context.Context) (RegistryClient, error) {
	c, err := ActiveConfig()
	if err != nil {
		return nil, err
	}
	return NewRegistryClientWithSettings(ctx, c)
}

// NewRegistryClientWithSettings creates a client with specified Config.
func NewRegistryClientWithSettings(ctx context.Context, config Config) (RegistryClient, error) {
	opts, err := clientOptions(config)
	if err != nil {
		return nil, err
	}
	client, err := gapic.NewRegistryClient(ctx, opts...)
	if err != nil && !config.Insecure && config.Token == "" {
		err = fmt.Errorf("registry token missing, attempted gcloud credentials: %w", err)
	}
	return client, err
}

type AdminClient = *gapic.AdminClient

// NewAdminClient creates a new client using the active Config.
func NewAdminClient(ctx context.Context) (AdminClient, error) {
	c, err := ActiveConfig()
	if err != nil {
		return nil, err
	}
	return NewAdminClientWithSettings(ctx, c)
}

// NewAdminClientWithSettings creates a client with specified Config.
func NewAdminClientWithSettings(ctx context.Context, config Config) (AdminClient, error) {
	opts, err := clientOptions(config)
	if err != nil {
		return nil, err
	}
	return gapic.NewAdminClient(ctx, opts...)
}

type ProvisioningClient = *gapic.ProvisioningClient

// NewAdminClient creates a new client using the active Config.
func NewProvisioningClient(ctx context.Context) (ProvisioningClient, error) {
	c, err := ActiveConfig()
	if err != nil {
		return nil, err
	}
	return NewProvisioningClientWithSettings(ctx, c)
}

// NewAdminClientWithSettings creates a GAPIC client with specified Config.
func NewProvisioningClientWithSettings(ctx context.Context, config Config) (ProvisioningClient, error) {
	opts, err := clientOptions(config)
	if err != nil {
		return nil, err
	}
	return gapic.NewProvisioningClient(ctx, opts...)
}
