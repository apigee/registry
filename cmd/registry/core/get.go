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

package core

import (
	"context"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

func GetProject(ctx context.Context,
	client *gapic.AdminClient,
	name names.Project,
	handler ProjectHandler) (*rpc.Project, error) {
	request := &rpc.GetProjectRequest{
		Name: name.String(),
	}
	project, err := client.GetProject(ctx, request)
	if err != nil {
		return nil, err
	}

	return project, handler(project)
}

func GetAPI(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Api,
	handler ApiHandler) (*rpc.Api, error) {
	request := &rpc.GetApiRequest{
		Name: name.String(),
	}
	api, err := client.GetApi(ctx, request)
	if err != nil {
		return nil, err
	}

	return api, handler(api)
}

func GetDeployment(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Deployment,
	handler DeploymentHandler) (*rpc.ApiDeployment, error) {
	request := &rpc.GetApiDeploymentRequest{
		Name: name.String(),
	}
	deployment, err := client.GetApiDeployment(ctx, request)
	if err != nil {
		return nil, err
	}

	return deployment, handler(deployment)
}

func GetVersion(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Version,
	handler VersionHandler) (*rpc.ApiVersion, error) {
	request := &rpc.GetApiVersionRequest{
		Name: name.String(),
	}
	version, err := client.GetApiVersion(ctx, request)
	if err != nil {
		return nil, err
	}

	return version, handler(version)
}

func GetSpec(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Spec,
	getContents bool,
	handler SpecHandler) (*rpc.ApiSpec, error) {
	request := &rpc.GetApiSpecRequest{
		Name: name.String(),
	}
	spec, err := client.GetApiSpec(ctx, request)
	if err != nil {
		return nil, err
	}
	if getContents {
		request := &rpc.GetApiSpecContentsRequest{
			Name: spec.GetName(),
		}
		contents, err := client.GetApiSpecContents(ctx, request)
		if err != nil {
			return nil, err
		}
		spec.Contents = contents.GetData()
		spec.MimeType = contents.GetContentType()
	}

	return spec, handler(spec)
}

func GetArtifact(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Artifact,
	getContents bool,
	handler ArtifactHandler) (*rpc.Artifact, error) {
	request := &rpc.GetArtifactRequest{
		Name: name.String(),
	}
	artifact, err := client.GetArtifact(ctx, request)
	if err != nil {
		return nil, err
	}
	if getContents {
		request := &rpc.GetArtifactContentsRequest{
			Name: artifact.GetName(),
		}
		contents, err := client.GetArtifactContents(ctx, request)
		if err != nil {
			return nil, err
		}
		artifact.Contents = contents.GetData()
		artifact.MimeType = contents.GetContentType()
	}

	return artifact, handler(artifact)
}
