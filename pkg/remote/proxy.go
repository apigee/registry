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

package remote

import (
	"context"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Proxy implements a local proxy for a remote Registry server.
// This allows tests written to the generated "RegistryServer" and "AdminServer" interfaces to be run against remote servers.
type Proxy struct {
	adminClient    connection.AdminClient
	registryClient connection.Client

	rpc.UnimplementedRegistryServer
	rpc.UnimplementedAdminServer
}

func (p *Proxy) Open(ctx context.Context) error {
	var err error
	p.adminClient, err = connection.NewAdminClient(ctx)
	if err != nil {
		return err
	}

	p.registryClient, err = connection.NewClient(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (p *Proxy) Close() {
	p.adminClient.Close()
	p.registryClient.Close()
}

// NewProxy creates a proxy with the default configuration for a remote server connection.
func NewProxy() *Proxy {
	return &Proxy{}
}

func (p *Proxy) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	return p.adminClient.CreateProject(ctx, req)
}

func (p *Proxy) DeleteProject(ctx context.Context, req *rpc.DeleteProjectRequest) (*emptypb.Empty, error) {
	return nil, p.adminClient.DeleteProject(ctx, req)
}

func (p *Proxy) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	return p.registryClient.UpdateApi(ctx, req)
}

func (p *Proxy) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	return p.registryClient.UpdateApiVersion(ctx, req)
}

func (p *Proxy) UpdateApiSpec(ctx context.Context, req *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	return p.registryClient.UpdateApiSpec(ctx, req)
}

func (p *Proxy) UpdateApiDeployment(ctx context.Context, req *rpc.UpdateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	return p.registryClient.UpdateApiDeployment(ctx, req)
}

func (p *Proxy) CreateArtifact(ctx context.Context, req *rpc.CreateArtifactRequest) (*rpc.Artifact, error) {
	return p.registryClient.CreateArtifact(ctx, req)
}
