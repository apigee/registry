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

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Proxy implements a local proxy for a remote Registry server.
// This allows tests written to the generated "RegistryServer" and "AdminServer" interfaces to be run against remote servers.
type Proxy struct {
	adminClient    connection.AdminClient
	registryClient connection.RegistryClient

	rpc.UnimplementedRegistryServer
	rpc.UnimplementedAdminServer
}

func (p *Proxy) Open(ctx context.Context) error {
	var err error
	p.adminClient, err = connection.NewAdminClient(ctx)
	if err != nil {
		return err
	}

	// Delete everything in the remote server to be tested.
	it := p.adminClient.ListProjects(ctx, &rpc.ListProjectsRequest{})
	projectNames := make([]string, 0)
	for p, err := it.Next(); err != iterator.Done; p, err = it.Next() {
		if err != nil {
			return err
		}
		projectNames = append(projectNames, p.Name)
	}
	for _, n := range projectNames {
		err = p.adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{Name: n, Force: true})
		if err != nil {
			return err
		}
	}

	p.registryClient, err = connection.NewClient(ctx)
	return err
}

func (p *Proxy) Close() {
	p.adminClient.Close()
	p.registryClient.Close()
}

// NewProxy creates a proxy with the default configuration for a remote server connection.
func NewProxy() *Proxy {
	return &Proxy{}
}

// Admin

func (p *Proxy) GetStatus(ctx context.Context, req *emptypb.Empty) (*rpc.Status, error) {
	return p.adminClient.GrpcClient().GetStatus(ctx, req)
}

func (p *Proxy) GetStorage(ctx context.Context, req *emptypb.Empty) (*rpc.Storage, error) {
	return p.adminClient.GrpcClient().GetStorage(ctx, req)
}

// Projects

func (p *Proxy) GetProject(ctx context.Context, req *rpc.GetProjectRequest) (*rpc.Project, error) {
	return p.adminClient.GrpcClient().GetProject(ctx, req)
}

func (p *Proxy) ListProjects(ctx context.Context, req *rpc.ListProjectsRequest) (*rpc.ListProjectsResponse, error) {
	return p.adminClient.GrpcClient().ListProjects(ctx, req)
}

func (p *Proxy) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	return p.adminClient.GrpcClient().CreateProject(ctx, req)
}

func (p *Proxy) UpdateProject(ctx context.Context, req *rpc.UpdateProjectRequest) (*rpc.Project, error) {
	return p.adminClient.GrpcClient().UpdateProject(ctx, req)
}

func (p *Proxy) DeleteProject(ctx context.Context, req *rpc.DeleteProjectRequest) (*emptypb.Empty, error) {
	return p.adminClient.GrpcClient().DeleteProject(ctx, req)
}

// Apis

func (p *Proxy) GetApi(ctx context.Context, req *rpc.GetApiRequest) (*rpc.Api, error) {
	return p.registryClient.GrpcClient().GetApi(ctx, req)
}

func (p *Proxy) ListApis(ctx context.Context, req *rpc.ListApisRequest) (*rpc.ListApisResponse, error) {
	return p.registryClient.GrpcClient().ListApis(ctx, req)
}

func (p *Proxy) CreateApi(ctx context.Context, req *rpc.CreateApiRequest) (*rpc.Api, error) {
	return p.registryClient.GrpcClient().CreateApi(ctx, req)
}

func (p *Proxy) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	return p.registryClient.GrpcClient().UpdateApi(ctx, req)
}

func (p *Proxy) DeleteApi(ctx context.Context, req *rpc.DeleteApiRequest) (*emptypb.Empty, error) {
	return p.registryClient.GrpcClient().DeleteApi(ctx, req)
}

// Versions

func (p *Proxy) GetApiVersion(ctx context.Context, req *rpc.GetApiVersionRequest) (*rpc.ApiVersion, error) {
	return p.registryClient.GrpcClient().GetApiVersion(ctx, req)
}

func (p *Proxy) ListApiVersions(ctx context.Context, req *rpc.ListApiVersionsRequest) (*rpc.ListApiVersionsResponse, error) {
	return p.registryClient.GrpcClient().ListApiVersions(ctx, req)
}

func (p *Proxy) CreateApiVersion(ctx context.Context, req *rpc.CreateApiVersionRequest) (*rpc.ApiVersion, error) {
	return p.registryClient.GrpcClient().CreateApiVersion(ctx, req)
}

func (p *Proxy) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	return p.registryClient.GrpcClient().UpdateApiVersion(ctx, req)
}

func (p *Proxy) DeleteApiVersion(ctx context.Context, req *rpc.DeleteApiVersionRequest) (*emptypb.Empty, error) {
	return p.registryClient.GrpcClient().DeleteApiVersion(ctx, req)
}

// Specs

func (p *Proxy) GetApiSpec(ctx context.Context, req *rpc.GetApiSpecRequest) (*rpc.ApiSpec, error) {
	return p.registryClient.GrpcClient().GetApiSpec(ctx, req)
}

func (p *Proxy) ListApiSpecs(ctx context.Context, req *rpc.ListApiSpecsRequest) (*rpc.ListApiSpecsResponse, error) {
	return p.registryClient.GrpcClient().ListApiSpecs(ctx, req)
}

func (p *Proxy) CreateApiSpec(ctx context.Context, req *rpc.CreateApiSpecRequest) (*rpc.ApiSpec, error) {
	return p.registryClient.GrpcClient().CreateApiSpec(ctx, req)
}

func (p *Proxy) UpdateApiSpec(ctx context.Context, req *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	return p.registryClient.GrpcClient().UpdateApiSpec(ctx, req)
}

func (p *Proxy) DeleteApiSpec(ctx context.Context, req *rpc.DeleteApiSpecRequest) (*emptypb.Empty, error) {
	return p.registryClient.GrpcClient().DeleteApiSpec(ctx, req)
}

func (p *Proxy) GetApiSpecContents(ctx context.Context, req *rpc.GetApiSpecContentsRequest) (*httpbody.HttpBody, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	for k, v := range md {
		ctx = metadata.AppendToOutgoingContext(ctx, k, v[0])
	}
	return p.registryClient.GrpcClient().GetApiSpecContents(ctx, req)
}

func (p *Proxy) ListApiSpecRevisions(ctx context.Context, req *rpc.ListApiSpecRevisionsRequest) (*rpc.ListApiSpecRevisionsResponse, error) {
	return p.registryClient.GrpcClient().ListApiSpecRevisions(ctx, req)
}

func (p *Proxy) TagApiSpecRevision(ctx context.Context, req *rpc.TagApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	return p.registryClient.GrpcClient().TagApiSpecRevision(ctx, req)
}

func (p *Proxy) DeleteApiSpecRevision(ctx context.Context, req *rpc.DeleteApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	return p.registryClient.GrpcClient().DeleteApiSpecRevision(ctx, req)
}

func (p *Proxy) RollbackApiSpec(ctx context.Context, req *rpc.RollbackApiSpecRequest) (*rpc.ApiSpec, error) {
	return p.registryClient.GrpcClient().RollbackApiSpec(ctx, req)
}

// Deployments

func (p *Proxy) GetApiDeployment(ctx context.Context, req *rpc.GetApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	return p.registryClient.GrpcClient().GetApiDeployment(ctx, req)
}

func (p *Proxy) ListApiDeployments(ctx context.Context, req *rpc.ListApiDeploymentsRequest) (*rpc.ListApiDeploymentsResponse, error) {
	return p.registryClient.GrpcClient().ListApiDeployments(ctx, req)
}

func (p *Proxy) CreateApiDeployment(ctx context.Context, req *rpc.CreateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	return p.registryClient.GrpcClient().CreateApiDeployment(ctx, req)
}

func (p *Proxy) UpdateApiDeployment(ctx context.Context, req *rpc.UpdateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	return p.registryClient.GrpcClient().UpdateApiDeployment(ctx, req)
}

func (p *Proxy) DeleteApiDeployment(ctx context.Context, req *rpc.DeleteApiDeploymentRequest) (*emptypb.Empty, error) {
	return p.registryClient.GrpcClient().DeleteApiDeployment(ctx, req)
}

func (p *Proxy) ListApiDeploymentRevisions(ctx context.Context, req *rpc.ListApiDeploymentRevisionsRequest) (*rpc.ListApiDeploymentRevisionsResponse, error) {
	return p.registryClient.GrpcClient().ListApiDeploymentRevisions(ctx, req)
}

func (p *Proxy) TagApiDeploymentRevision(ctx context.Context, req *rpc.TagApiDeploymentRevisionRequest) (*rpc.ApiDeployment, error) {
	return p.registryClient.GrpcClient().TagApiDeploymentRevision(ctx, req)
}

func (p *Proxy) DeleteApiDeploymentRevision(ctx context.Context, req *rpc.DeleteApiDeploymentRevisionRequest) (*rpc.ApiDeployment, error) {
	return p.registryClient.GrpcClient().DeleteApiDeploymentRevision(ctx, req)
}

func (p *Proxy) RollbackApiDeployment(ctx context.Context, req *rpc.RollbackApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	return p.registryClient.GrpcClient().RollbackApiDeployment(ctx, req)
}

// Artifacts

func (p *Proxy) GetArtifact(ctx context.Context, req *rpc.GetArtifactRequest) (*rpc.Artifact, error) {
	return p.registryClient.GrpcClient().GetArtifact(ctx, req)
}

func (p *Proxy) ListArtifacts(ctx context.Context, req *rpc.ListArtifactsRequest) (*rpc.ListArtifactsResponse, error) {
	return p.registryClient.GrpcClient().ListArtifacts(ctx, req)
}

func (p *Proxy) CreateArtifact(ctx context.Context, req *rpc.CreateArtifactRequest) (*rpc.Artifact, error) {
	return p.registryClient.GrpcClient().CreateArtifact(ctx, req)
}

func (p *Proxy) ReplaceArtifact(ctx context.Context, req *rpc.ReplaceArtifactRequest) (*rpc.Artifact, error) {
	return p.registryClient.GrpcClient().ReplaceArtifact(ctx, req)
}

func (p *Proxy) DeleteArtifact(ctx context.Context, req *rpc.DeleteArtifactRequest) (*emptypb.Empty, error) {
	return p.registryClient.GrpcClient().DeleteArtifact(ctx, req)
}

func (p *Proxy) GetArtifactContents(ctx context.Context, req *rpc.GetArtifactContentsRequest) (*httpbody.HttpBody, error) {
	return p.registryClient.GrpcClient().GetArtifactContents(ctx, req)
}
