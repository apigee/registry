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
	"log"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/api/iterator"
)

func ListProjects(ctx context.Context,
	client *gapic.AdminClient,
	name names.Project,
	filterFlag string,
	handler ProjectHandler) error {
	request := &rpc.ListProjectsRequest{}
	filter := filterFlag
	projectID := name.ProjectID
	if projectID != "" && projectID != "-" {
		filter = "project_id == '" + projectID + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListProjects(ctx, request)
	for {
		project, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(project)
	}
	return nil
}

func ListAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Api,
	filterFlag string,
	handler ApiHandler) error {
	request := &rpc.ListApisRequest{
		Parent: name.Parent(),
	}
	filter := filterFlag
	apiID := name.ApiID
	if apiID != "" && apiID != "-" {
		filter = "api_id == '" + apiID + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListApis(ctx, request)
	for {
		api, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(api)
	}
	return nil
}

func ListDeployments(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Deployment,
	filterFlag string,
	handler DeploymentHandler) error {
	request := &rpc.ListApiDeploymentsRequest{
		Parent: name.Parent(),
	}
	filter := filterFlag
	deploymentID := name.DeploymentID
	if deploymentID != "" && deploymentID != "-" {
		filter = "deployment_id == '" + deploymentID + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListApiDeployments(ctx, request)
	for {
		deployment, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(deployment)
	}
	return nil
}

func ListVersions(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Version,
	filterFlag string,
	handler VersionHandler) error {
	request := &rpc.ListApiVersionsRequest{
		Parent: name.Parent(),
	}
	filter := filterFlag
	versionID := name.VersionID
	if versionID != "" && versionID != "-" {
		filter = "version_id == '" + versionID + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListApiVersions(ctx, request)
	for {
		version, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(version)
	}
	return nil
}

func ListSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Spec,
	filterFlag string,
	handler SpecHandler) error {
	request := &rpc.ListApiSpecsRequest{
		Parent: name.Parent(),
	}
	filter := filterFlag
	specID := name.SpecID
	if specID != "" && specID != "-" {
		filter = "spec_id == '" + specID + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListApiSpecs(ctx, request)
	for {
		spec, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(spec)
	}
	return nil
}

func ListSpecRevisions(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Spec,
	filterFlag string,
	handler SpecHandler) error {
	request := &rpc.ListApiSpecRevisionsRequest{
		Name: name.String(),
	}
	it := client.ListApiSpecRevisions(ctx, request)
	for {
		spec, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(spec)
	}
	return nil
}

func ListArtifacts(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Artifact,
	filterFlag string,
	getContents bool,
	handler ArtifactHandler) error {
	request := &rpc.ListArtifactsRequest{
		Parent: name.Parent(),
	}
	filter := filterFlag
	artifactID := name.ArtifactID()
	log.Printf("ARTIFACT ID %s", artifactID)
	if artifactID != "" && artifactID != "-" {
		if filter != "" {
			filter += " && "
		}
		filter += "artifact_id == '" + artifactID + "'"
	}
	if filter != "" {
		request.Filter = filter
	}
	it := client.ListArtifacts(ctx, request)
	for {
		artifact, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		if getContents {
			req := &rpc.GetArtifactContentsRequest{
				Name: artifact.GetName(),
			}
			resp, err := client.GetArtifactContents(ctx, req)
			if err != nil {
				return err
			}
			artifact.Contents = resp.GetData()
		}
		handler(artifact)
	}
	return nil
}

func ListArtifactsForParent(ctx context.Context,
	client *gapic.RegistryClient,
	parent names.Name,
	handler ArtifactHandler) error {
	request := &rpc.ListArtifactsRequest{
		Parent: parent.String(),
	}
	it := client.ListArtifacts(ctx, request)
	for {
		artifact, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}
		handler(artifact)
	}
	return nil
}
