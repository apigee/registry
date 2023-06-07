// Copyright 2022 Google LLC.
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

package seeder

import (
	"context"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
)

type Client struct {
	*gapic.RegistryClient
	*gapic.AdminClient
}

func (c Client) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	return c.AdminClient.CreateProject(ctx, req)
}

func (c Client) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	return c.RegistryClient.UpdateApi(ctx, req)
}

func (c Client) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	return c.RegistryClient.UpdateApiVersion(ctx, req)
}

func (c Client) UpdateApiSpec(ctx context.Context, req *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	return c.RegistryClient.UpdateApiSpec(ctx, req)
}

func (c Client) UpdateApiDeployment(ctx context.Context, req *rpc.UpdateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	return c.RegistryClient.UpdateApiDeployment(ctx, req)
}

func (c Client) CreateArtifact(ctx context.Context, req *rpc.CreateArtifactRequest) (*rpc.Artifact, error) {
	return c.RegistryClient.CreateArtifact(ctx, req)
}

func (c Client) ReplaceArtifact(ctx context.Context, req *rpc.ReplaceArtifactRequest) (*rpc.Artifact, error) {
	return c.RegistryClient.ReplaceArtifact(ctx, req)
}
