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
	"fmt"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/api/iterator"
)

func ListProjects(ctx context.Context,
	client *gapic.AdminClient,
	name names.Project,
	filter string,
	handler ProjectHandler) error {
	if id := name.ProjectID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("project_id == '%s'", id)
	}

	it := client.ListProjects(ctx, &rpc.ListProjectsRequest{
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if err := handler(r); err != nil {
			return err
		}
	}
	return nil
}

func ListAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Api,
	filter string,
	handler ApiHandler) error {
	if id := name.ApiID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("api_id == '%s'", id)
	}

	it := client.ListApis(ctx, &rpc.ListApisRequest{
		Parent: name.Parent(),
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if err := handler(r); err != nil {
			return err
		}
	}
	return nil
}

func ListDeployments(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Deployment,
	filter string,
	handler DeploymentHandler) error {
	if id := name.DeploymentID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("deployment_id == '%s'", id)
	}

	it := client.ListApiDeployments(ctx, &rpc.ListApiDeploymentsRequest{
		Parent: name.Parent(),
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if err := handler(r); err != nil {
			return err
		}
	}
	return nil
}

func ListDeploymentRevisions(ctx context.Context,
	client *gapic.RegistryClient,
	name names.DeploymentRevision,
	filter string,
	handler DeploymentHandler) error {
	it := client.ListApiDeploymentRevisions(ctx, &rpc.ListApiDeploymentRevisionsRequest{
		Name: name.String(),
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}
		if err := handler(r); err != nil {
			return err
		}
	}
	return nil
}

func ListVersions(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Version,
	filter string,
	handler VersionHandler) error {
	if id := name.VersionID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("version_id == '%s'", id)
	}

	it := client.ListApiVersions(ctx, &rpc.ListApiVersionsRequest{
		Parent: name.Parent(),
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if err := handler(r); err != nil {
			return err
		}
	}
	return nil
}

func ListSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Spec,
	filter string,
	handler SpecHandler) error {
	if id := name.SpecID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("spec_id == '%s'", id)
	}

	it := client.ListApiSpecs(ctx, &rpc.ListApiSpecsRequest{
		Parent: name.Parent(),
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if err := handler(r); err != nil {
			return err
		}
	}
	return nil
}

func ListSpecRevisions(ctx context.Context,
	client *gapic.RegistryClient,
	name names.SpecRevision,
	filter string,
	handler SpecHandler) error {
	it := client.ListApiSpecRevisions(ctx, &rpc.ListApiSpecRevisionsRequest{
		Name: name.String(),
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if err := handler(r); err != nil {
			return err
		}
	}
	return nil
}

func ListArtifacts(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Artifact,
	filter string,
	getContents bool,
	handler ArtifactHandler) error {
	if id := name.ArtifactID(); id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("artifact_id == '%s'", id)
	}

	it := client.ListArtifacts(ctx, &rpc.ListArtifactsRequest{
		Parent: name.Parent(),
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if getContents {
			resp, err := client.GetArtifactContents(ctx, &rpc.GetArtifactContentsRequest{
				Name: r.GetName(),
			})
			if err != nil {
				return err
			}
			r.Contents = resp.GetData()
		}

		if err := handler(r); err != nil {
			return err
		}
	}
	return nil
}
