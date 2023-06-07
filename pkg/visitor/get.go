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

package visitor

import (
	"context"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetProject(ctx context.Context,
	client *gapic.AdminClient,
	name names.Project,
	implicitProject *rpc.Project,
	handler ProjectHandler) error {
	project, err := client.GetProject(ctx, &rpc.GetProjectRequest{
		Name: name.String(),
	})
	if err != nil && status.Code(err) == codes.Unimplemented && implicitProject != nil {
		// If the admin service is unavailable, provide a placeholder project.
		// If the project is invalid, downstream actions will fail.
		project = implicitProject
	} else if err != nil {
		return err
	}

	return handler(ctx, project)
}

func GetAPI(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Api,
	handler ApiHandler) error {
	api, err := client.GetApi(ctx, &rpc.GetApiRequest{
		Name: name.String(),
	})
	if err != nil {
		return err
	}

	return handler(ctx, api)
}

func GetDeployment(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Deployment,
	handler DeploymentHandler) error {
	deployment, err := client.GetApiDeployment(ctx, &rpc.GetApiDeploymentRequest{
		Name: name.String(),
	})
	if err != nil {
		return err
	}

	return handler(ctx, deployment)
}

func GetDeploymentRevision(ctx context.Context,
	client *gapic.RegistryClient,
	name names.DeploymentRevision,
	handler DeploymentHandler) error {
	request := &rpc.GetApiDeploymentRequest{
		Name: name.String(),
	}
	deployment, err := client.GetApiDeployment(ctx, request)
	if err != nil {
		return err
	}

	return handler(ctx, deployment)
}

func GetVersion(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Version,
	handler VersionHandler) error {
	version, err := client.GetApiVersion(ctx, &rpc.GetApiVersionRequest{
		Name: name.String(),
	})
	if err != nil {
		return err
	}

	return handler(ctx, version)
}

func GetSpec(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Spec,
	getContents bool,
	handler SpecHandler) error {
	spec, err := client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: name.String(),
	})
	if err != nil {
		return err
	}
	if getContents {
		if err := FetchSpecContents(ctx, client, spec); err != nil {
			return err
		}
	}

	return handler(ctx, spec)
}

func GetSpecRevision(ctx context.Context,
	client *gapic.RegistryClient,
	name names.SpecRevision,
	getContents bool,
	handler SpecHandler) error {
	request := &rpc.GetApiSpecRequest{
		Name: name.String(),
	}
	spec, err := client.GetApiSpec(ctx, request)
	if err != nil {
		return err
	}
	if getContents {
		if err := FetchSpecContents(ctx, client, spec); err != nil {
			return err
		}
	}

	return handler(ctx, spec)
}

func GetArtifact(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Artifact,
	getContents bool,
	handler ArtifactHandler) error {
	artifact, err := client.GetArtifact(ctx, &rpc.GetArtifactRequest{
		Name: name.String(),
	})
	if err != nil {
		return err
	}
	if getContents {
		if err = FetchArtifactContents(ctx, client, artifact); err != nil {
			return err
		}
	}

	return handler(ctx, artifact)
}
