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

package remote

import (
	"context"
	"errors"
	"log"
	"strings"

	longrunning "cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/wipeout"
	"github.com/apigee/registry/rpc"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrAdminServiceUnavailable = errors.New("admin service is unavailable")

// Proxy implements a local proxy for a remote Registry server.
// This allows tests written to the generated "RegistryServer" and "AdminServer" interfaces to be run against remote servers.
type Proxy struct {
	projectID string

	adminClient    connection.AdminClient
	registryClient connection.RegistryClient

	rpc.UnimplementedRegistryServer
	rpc.UnimplementedAdminServer
}

func (p *Proxy) Open(ctx context.Context) error {
	// If a project ID is set, assume we are connecting to a managed server
	// that provides a single project and does not support the Admin service.
	if p.projectID != "" {
		var err error
		p.registryClient, err = connection.NewRegistryClient(ctx)
		if err != nil {
			return err
		}

		// Delete everything in the remote project to be tested.
		wipeout.Wipeout(ctx, p.registryClient, p.projectID, 1)
		return nil
	}

	// If no project ID is set, assume we are connecting to an unmanaged server
	// with full support for the Admin service.
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
	// Delete after iteration to avoid corrupting the iteration.
	for _, n := range projectNames {
		err = p.adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{Name: n, Force: true})
		if err != nil {
			return err
		}
	}
	p.registryClient, err = connection.NewRegistryClient(ctx)
	return err
}

func (p *Proxy) Close() {
	p.adminClient.Close()
	p.registryClient.Close()
}

// NewProxyForHostedService creates a proxy with the default configuration for a connection
// to a remote server with a single project and no Admin service.
func NewProxyForHostedService(projectID string) *Proxy {
	return &Proxy{projectID: projectID}
}

func (p *Proxy) GetStatus(ctx context.Context, req *emptypb.Empty) (*rpc.Status, error) {
	if p.adminClient == nil {
		return nil, ErrAdminServiceUnavailable
	}
	return p.adminClient.GrpcClient().GetStatus(ctx, req)
}

func (p *Proxy) GetStorage(ctx context.Context, req *emptypb.Empty) (*rpc.Storage, error) {
	if p.adminClient == nil {
		return nil, ErrAdminServiceUnavailable
	}
	return p.adminClient.GrpcClient().GetStorage(ctx, req)
}

func (p *Proxy) MigrateDatabase(ctx context.Context, req *rpc.MigrateDatabaseRequest) (*longrunning.Operation, error) {
	if p.adminClient == nil {
		return nil, ErrAdminServiceUnavailable
	}
	return p.adminClient.GrpcClient().MigrateDatabase(ctx, req)
}

// Projects

func (p *Proxy) GetProject(ctx context.Context, req *rpc.GetProjectRequest) (*rpc.Project, error) {
	if p.adminClient == nil {
		return nil, ErrAdminServiceUnavailable
	}
	return p.adminClient.GrpcClient().GetProject(ctx, req)
}

func (p *Proxy) ListProjects(ctx context.Context, req *rpc.ListProjectsRequest) (*rpc.ListProjectsResponse, error) {
	if p.adminClient == nil {
		return nil, ErrAdminServiceUnavailable
	}
	return p.adminClient.GrpcClient().ListProjects(ctx, req)
}

func (p *Proxy) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	if p.adminClient == nil {
		if req.ProjectId == "my-project" {
			return &rpc.Project{Name: req.ProjectId}, nil
		}
		log.Printf("FIXME CreateProject %+v", req)
		return nil, ErrAdminServiceUnavailable
	}
	return p.adminClient.GrpcClient().CreateProject(ctx, req)
}

func (p *Proxy) UpdateProject(ctx context.Context, req *rpc.UpdateProjectRequest) (*rpc.Project, error) {
	if p.adminClient == nil {
		return nil, ErrAdminServiceUnavailable
	}
	return p.adminClient.GrpcClient().UpdateProject(ctx, req)
}

func (p *Proxy) DeleteProject(ctx context.Context, req *rpc.DeleteProjectRequest) (*emptypb.Empty, error) {
	if p.adminClient == nil {
		return nil, ErrAdminServiceUnavailable
	}
	return p.adminClient.GrpcClient().DeleteProject(ctx, req)
}

// Apis

func (p *Proxy) GetApi(ctx context.Context, req *rpc.GetApiRequest) (*rpc.Api, error) {
	req, _ = proto.Clone(req).(*rpc.GetApiRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().GetApi(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) ListApis(ctx context.Context, req *rpc.ListApisRequest) (*rpc.ListApisResponse, error) {
	req, _ = proto.Clone(req).(*rpc.ListApisRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	req.Filter = p.hostedFilter(req.Filter)
	response, err := p.registryClient.GrpcClient().ListApis(ctx, req)
	if err != nil {
		return nil, err
	}
	for i := range response.Apis {
		response.Apis[i].Name = p.testResourceName(response.Apis[i].Name)
	}
	return response, err
}

func (p *Proxy) CreateApi(ctx context.Context, req *rpc.CreateApiRequest) (*rpc.Api, error) {
	req, _ = proto.Clone(req).(*rpc.CreateApiRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	response, err := p.registryClient.GrpcClient().CreateApi(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	req, _ = proto.Clone(req).(*rpc.UpdateApiRequest)
	if req.Api != nil {
		req.Api.Name = p.hostedResourceName(req.Api.Name)
	}
	response, err := p.registryClient.GrpcClient().UpdateApi(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) DeleteApi(ctx context.Context, req *rpc.DeleteApiRequest) (*emptypb.Empty, error) {
	req, _ = proto.Clone(req).(*rpc.DeleteApiRequest)
	req.Name = p.hostedResourceName(req.Name)
	return p.registryClient.GrpcClient().DeleteApi(ctx, req)
}

// Versions

func (p *Proxy) GetApiVersion(ctx context.Context, req *rpc.GetApiVersionRequest) (*rpc.ApiVersion, error) {
	req, _ = proto.Clone(req).(*rpc.GetApiVersionRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().GetApiVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) ListApiVersions(ctx context.Context, req *rpc.ListApiVersionsRequest) (*rpc.ListApiVersionsResponse, error) {
	req, _ = proto.Clone(req).(*rpc.ListApiVersionsRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	req.Filter = p.hostedFilter(req.Filter)
	response, err := p.registryClient.GrpcClient().ListApiVersions(ctx, req)
	if err != nil {
		return nil, err
	}
	for i := range response.ApiVersions {
		response.ApiVersions[i].Name = p.testResourceName(response.ApiVersions[i].Name)
	}
	return response, err
}

func (p *Proxy) CreateApiVersion(ctx context.Context, req *rpc.CreateApiVersionRequest) (*rpc.ApiVersion, error) {
	req, _ = proto.Clone(req).(*rpc.CreateApiVersionRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	response, err := p.registryClient.GrpcClient().CreateApiVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	req, _ = proto.Clone(req).(*rpc.UpdateApiVersionRequest)
	if req.ApiVersion != nil {
		req.ApiVersion.Name = p.hostedResourceName(req.ApiVersion.Name)
	}
	response, err := p.registryClient.GrpcClient().UpdateApiVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) DeleteApiVersion(ctx context.Context, req *rpc.DeleteApiVersionRequest) (*emptypb.Empty, error) {
	req, _ = proto.Clone(req).(*rpc.DeleteApiVersionRequest)
	req.Name = p.hostedResourceName(req.Name)
	return p.registryClient.GrpcClient().DeleteApiVersion(ctx, req)
}

// Specs

func (p *Proxy) GetApiSpec(ctx context.Context, req *rpc.GetApiSpecRequest) (*rpc.ApiSpec, error) {
	req, _ = proto.Clone(req).(*rpc.GetApiSpecRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().GetApiSpec(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) ListApiSpecs(ctx context.Context, req *rpc.ListApiSpecsRequest) (*rpc.ListApiSpecsResponse, error) {
	req, _ = proto.Clone(req).(*rpc.ListApiSpecsRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	req.Filter = p.hostedFilter(req.Filter)
	response, err := p.registryClient.GrpcClient().ListApiSpecs(ctx, req)
	if err != nil {
		return nil, err
	}
	for i := range response.ApiSpecs {
		response.ApiSpecs[i].Name = p.testResourceName(response.ApiSpecs[i].Name)
	}
	return response, err
}

func (p *Proxy) CreateApiSpec(ctx context.Context, req *rpc.CreateApiSpecRequest) (*rpc.ApiSpec, error) {
	req, _ = proto.Clone(req).(*rpc.CreateApiSpecRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	response, err := p.registryClient.GrpcClient().CreateApiSpec(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) UpdateApiSpec(ctx context.Context, req *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	req, _ = proto.Clone(req).(*rpc.UpdateApiSpecRequest)
	if req.ApiSpec != nil {
		req.ApiSpec.Name = p.hostedResourceName(req.ApiSpec.Name)
	}
	response, err := p.registryClient.GrpcClient().UpdateApiSpec(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) DeleteApiSpec(ctx context.Context, req *rpc.DeleteApiSpecRequest) (*emptypb.Empty, error) {
	req, _ = proto.Clone(req).(*rpc.DeleteApiSpecRequest)
	req.Name = p.hostedResourceName(req.Name)
	return p.registryClient.GrpcClient().DeleteApiSpec(ctx, req)
}

func (p *Proxy) GetApiSpecContents(ctx context.Context, req *rpc.GetApiSpecContentsRequest) (*httpbody.HttpBody, error) {
	req, _ = proto.Clone(req).(*rpc.GetApiSpecContentsRequest)
	req.Name = p.hostedResourceName(req.Name)
	md, _ := metadata.FromIncomingContext(ctx)
	for k, v := range md {
		ctx = metadata.AppendToOutgoingContext(ctx, k, v[0])
	}
	return p.registryClient.GrpcClient().GetApiSpecContents(ctx, req)
}

func (p *Proxy) ListApiSpecRevisions(ctx context.Context, req *rpc.ListApiSpecRevisionsRequest) (*rpc.ListApiSpecRevisionsResponse, error) {
	req, _ = proto.Clone(req).(*rpc.ListApiSpecRevisionsRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().ListApiSpecRevisions(ctx, req)
	if err != nil {
		return nil, err
	}
	for i := range response.ApiSpecs {
		response.ApiSpecs[i].Name = p.testResourceName(response.ApiSpecs[i].Name)
	}
	return response, err
}

func (p *Proxy) TagApiSpecRevision(ctx context.Context, req *rpc.TagApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	req, _ = proto.Clone(req).(*rpc.TagApiSpecRevisionRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().TagApiSpecRevision(ctx, req)
	if err == nil {
		response.Name = p.testResourceName(response.Name)
	}
	return response, err
}

func (p *Proxy) DeleteApiSpecRevision(ctx context.Context, req *rpc.DeleteApiSpecRevisionRequest) (*rpc.ApiSpec, error) {
	req, _ = proto.Clone(req).(*rpc.DeleteApiSpecRevisionRequest)
	req.Name = p.hostedResourceName(req.Name)
	return p.registryClient.GrpcClient().DeleteApiSpecRevision(ctx, req)
}

func (p *Proxy) RollbackApiSpec(ctx context.Context, req *rpc.RollbackApiSpecRequest) (*rpc.ApiSpec, error) {
	req, _ = proto.Clone(req).(*rpc.RollbackApiSpecRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().RollbackApiSpec(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

// Deployments

func (p *Proxy) GetApiDeployment(ctx context.Context, req *rpc.GetApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	req, _ = proto.Clone(req).(*rpc.GetApiDeploymentRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().GetApiDeployment(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) ListApiDeployments(ctx context.Context, req *rpc.ListApiDeploymentsRequest) (*rpc.ListApiDeploymentsResponse, error) {
	req, _ = proto.Clone(req).(*rpc.ListApiDeploymentsRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	req.Filter = p.hostedFilter(req.Filter)
	response, err := p.registryClient.GrpcClient().ListApiDeployments(ctx, req)
	if err != nil {
		return nil, err
	}
	for i := range response.ApiDeployments {
		response.ApiDeployments[i].Name = p.testResourceName(response.ApiDeployments[i].Name)
	}
	return response, err
}

func (p *Proxy) CreateApiDeployment(ctx context.Context, req *rpc.CreateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	req, _ = proto.Clone(req).(*rpc.CreateApiDeploymentRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	response, err := p.registryClient.GrpcClient().CreateApiDeployment(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) UpdateApiDeployment(ctx context.Context, req *rpc.UpdateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	req, _ = proto.Clone(req).(*rpc.UpdateApiDeploymentRequest)
	if req.ApiDeployment != nil {
		req.ApiDeployment.Name = p.hostedResourceName(req.ApiDeployment.Name)
	}
	response, err := p.registryClient.GrpcClient().UpdateApiDeployment(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) DeleteApiDeployment(ctx context.Context, req *rpc.DeleteApiDeploymentRequest) (*emptypb.Empty, error) {
	req, _ = proto.Clone(req).(*rpc.DeleteApiDeploymentRequest)
	req.Name = p.hostedResourceName(req.Name)
	return p.registryClient.GrpcClient().DeleteApiDeployment(ctx, req)
}

func (p *Proxy) ListApiDeploymentRevisions(ctx context.Context, req *rpc.ListApiDeploymentRevisionsRequest) (*rpc.ListApiDeploymentRevisionsResponse, error) {
	req, _ = proto.Clone(req).(*rpc.ListApiDeploymentRevisionsRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().ListApiDeploymentRevisions(ctx, req)
	if err != nil {
		return nil, err
	}
	for i := range response.ApiDeployments {
		response.ApiDeployments[i].Name = p.testResourceName(response.ApiDeployments[i].Name)
	}
	return response, err
}

func (p *Proxy) TagApiDeploymentRevision(ctx context.Context, req *rpc.TagApiDeploymentRevisionRequest) (*rpc.ApiDeployment, error) {
	req, _ = proto.Clone(req).(*rpc.TagApiDeploymentRevisionRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().TagApiDeploymentRevision(ctx, req)
	if err == nil {
		response.Name = p.testResourceName(response.Name)
	}
	return response, err
}

func (p *Proxy) DeleteApiDeploymentRevision(ctx context.Context, req *rpc.DeleteApiDeploymentRevisionRequest) (*rpc.ApiDeployment, error) {
	req, _ = proto.Clone(req).(*rpc.DeleteApiDeploymentRevisionRequest)
	req.Name = p.hostedResourceName(req.Name)
	return p.registryClient.GrpcClient().DeleteApiDeploymentRevision(ctx, req)
}

func (p *Proxy) RollbackApiDeployment(ctx context.Context, req *rpc.RollbackApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	req, _ = proto.Clone(req).(*rpc.RollbackApiDeploymentRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().RollbackApiDeployment(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

// Artifacts

func (p *Proxy) GetArtifact(ctx context.Context, req *rpc.GetArtifactRequest) (*rpc.Artifact, error) {
	req, _ = proto.Clone(req).(*rpc.GetArtifactRequest)
	req.Name = p.hostedResourceName(req.Name)
	response, err := p.registryClient.GrpcClient().GetArtifact(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) ListArtifacts(ctx context.Context, req *rpc.ListArtifactsRequest) (*rpc.ListArtifactsResponse, error) {
	req, _ = proto.Clone(req).(*rpc.ListArtifactsRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	req.Filter = p.hostedFilter(req.Filter)
	response, err := p.registryClient.GrpcClient().ListArtifacts(ctx, req)
	if err != nil {
		return nil, err
	}
	for i := range response.Artifacts {
		response.Artifacts[i].Name = p.testResourceName(response.Artifacts[i].Name)
	}
	return response, err
}

func (p *Proxy) CreateArtifact(ctx context.Context, req *rpc.CreateArtifactRequest) (*rpc.Artifact, error) {
	req, _ = proto.Clone(req).(*rpc.CreateArtifactRequest)
	req.Parent = p.hostedResourceName(req.Parent)
	response, err := p.registryClient.GrpcClient().CreateArtifact(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) ReplaceArtifact(ctx context.Context, req *rpc.ReplaceArtifactRequest) (*rpc.Artifact, error) {
	req, _ = proto.Clone(req).(*rpc.ReplaceArtifactRequest)
	if req.Artifact != nil {
		req.Artifact.Name = p.hostedResourceName(req.Artifact.Name)
	}
	response, err := p.registryClient.GrpcClient().ReplaceArtifact(ctx, req)
	if err != nil {
		return nil, err
	}
	response.Name = p.testResourceName(response.Name)
	return response, err
}

func (p *Proxy) DeleteArtifact(ctx context.Context, req *rpc.DeleteArtifactRequest) (*emptypb.Empty, error) {
	req, _ = proto.Clone(req).(*rpc.DeleteArtifactRequest)
	req.Name = p.hostedResourceName(req.Name)
	return p.registryClient.GrpcClient().DeleteArtifact(ctx, req)
}

func (p *Proxy) GetArtifactContents(ctx context.Context, req *rpc.GetArtifactContentsRequest) (*httpbody.HttpBody, error) {
	req, _ = proto.Clone(req).(*rpc.GetArtifactContentsRequest)
	req.Name = p.hostedResourceName(req.Name)
	return p.registryClient.GrpcClient().GetArtifactContents(ctx, req)
}

// helpers

func (p *Proxy) hostedResourceName(name string) string {
	if p.projectID != "" {
		return strings.Replace(name, "projects/my-project", "projects/"+p.projectID, 1)
	}
	return name
}

func (p *Proxy) hostedFilter(filter string) string {
	if p.projectID != "" {
		return strings.ReplaceAll(filter, "projects/my-project", "projects/"+p.projectID)
	}
	return filter
}

func (p *Proxy) testResourceName(name string) string {
	if p.projectID != "" {
		return strings.Replace(name, "projects/"+p.projectID, "projects/my-project", 1)
	}
	return name
}
