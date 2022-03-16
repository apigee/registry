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
	"google.golang.org/api/iterator"
)

func ListProjects(ctx context.Context,
	client *gapic.AdminClient,
	name names.Project,
	filter string,
	handler ProjectHandler) error {
	it := client.ListProjects(ctx, &rpc.ListProjectsRequest{
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if id := name.ProjectID; r.GetName() != name.String() && id != "" && id != "-" {
			continue
		}

		handler(r)
	}
	return nil
}

func ListAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Api,
	filter string,
	handler ApiHandler) error {
	it := client.ListApis(ctx, &rpc.ListApisRequest{
		Parent: name.Parent(),
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if id := name.ApiID; r.GetName() != name.String() && id != "" && id != "-" {
			continue
		}

		handler(r)
	}
	return nil
}

func ListDeployments(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Deployment,
	filter string,
	handler DeploymentHandler) error {
	it := client.ListApiDeployments(ctx, &rpc.ListApiDeploymentsRequest{
		Parent: name.Parent(),
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if id := name.DeploymentID; r.GetName() != name.String() && id != "" && id != "-" {
			continue
		}

		handler(r)
	}
	return nil
}

func ListVersions(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Version,
	filter string,
	handler VersionHandler) error {
	it := client.ListApiVersions(ctx, &rpc.ListApiVersionsRequest{
		Parent: name.Parent(),
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if id := name.VersionID; r.GetName() != name.String() && id != "" && id != "-" {
			continue
		}

		handler(r)
	}
	return nil
}

func ListSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Spec,
	filter string,
	handler SpecHandler) error {
	it := client.ListApiSpecs(ctx, &rpc.ListApiSpecsRequest{
		Parent: name.Parent(),
		Filter: filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if id := name.SpecID; r.GetName() != name.String() && id != "" && id != "-" {
			continue
		}

		handler(r)
	}
	return nil
}

func ListArtifacts(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Artifact,
	filter string,
	getContents bool,
	handler ArtifactHandler) error {
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

		handler(r)
	}
	return nil
}
