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
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ListProjects(ctx context.Context,
	client *gapic.AdminClient,
	name names.Project,
	implicitProject *rpc.Project,
	pageSize int32,
	filter string,
	handler ProjectHandler) error {
	if id := name.ProjectID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("project_id == '%s'", id)
	}

	it := client.ListProjects(ctx, &rpc.ListProjectsRequest{
		PageSize: pageSize,
		Filter:   filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil && status.Code(err) == codes.Unimplemented && implicitProject != nil {
			// If the admin service is unavailable, provide a placeholder project.
			// If the project is invalid, downstream actions will fail.
			return handler(ctx, implicitProject)
		}

		if err != nil {
			return err
		}

		if err := handler(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func ListAPIs(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Api,
	pageSize int32,
	filter string,
	handler ApiHandler) error {
	if id := name.ApiID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("api_id == '%s'", id)
	}

	it := client.ListApis(ctx, &rpc.ListApisRequest{
		Parent:   name.Parent(),
		PageSize: pageSize,
		Filter:   filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if err := handler(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func ListDeployments(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Deployment,
	pageSize int32,
	filter string,
	handler DeploymentHandler) error {
	if id := name.DeploymentID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("deployment_id == '%s'", id)
	}

	it := client.ListApiDeployments(ctx, &rpc.ListApiDeploymentsRequest{
		Parent:   name.Parent(),
		PageSize: pageSize,
		Filter:   filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if err := handler(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func ListDeploymentRevisions(ctx context.Context,
	client *gapic.RegistryClient,
	name names.DeploymentRevision,
	pageSize int32,
	filter string,
	handler DeploymentHandler) error {
	it := client.ListApiDeploymentRevisions(ctx, &rpc.ListApiDeploymentRevisionsRequest{
		// "@-" indicates a collection of revisions, but we only want to send the resource name to the List RPC.
		Name:     strings.TrimSuffix(name.String(), "@-"),
		PageSize: pageSize,
		Filter:   filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}
		if err := handler(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func ListVersions(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Version,
	pageSize int32,
	filter string,
	handler VersionHandler) error {
	if id := name.VersionID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("version_id == '%s'", id)
	}

	it := client.ListApiVersions(ctx, &rpc.ListApiVersionsRequest{
		Parent:   name.Parent(),
		PageSize: pageSize,
		Filter:   filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if err := handler(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func ListSpecs(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Spec,
	pageSize int32,
	filter string,
	getContents bool,
	handler SpecHandler) error {
	if id := name.SpecID; id != "" && id != "-" {
		if len(filter) > 0 {
			filter += " && "
		}
		filter += fmt.Sprintf("spec_id == '%s'", id)
	}

	it := client.ListApiSpecs(ctx, &rpc.ListApiSpecsRequest{
		Parent:   name.Parent(),
		PageSize: pageSize,
		Filter:   filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if getContents {
			ctx = metadata.AppendToOutgoingContext(ctx, "accept-encoding", "gzip")
			resp, err := client.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
				Name: r.GetName(),
			})
			if err != nil {
				return err
			}
			r.Contents = resp.GetData()
			if mime.IsGZipCompressed(resp.ContentType) {
				r.MimeType = mime.GUnzippedType(r.MimeType)
				r.Contents, err = compress.GUnzippedBytes(r.Contents)
				if err != nil {
					return err
				}
			}
		}

		if err := handler(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func ListSpecRevisions(ctx context.Context,
	client *gapic.RegistryClient,
	name names.SpecRevision,
	pageSize int32,
	filter string,
	getContents bool,
	handler SpecHandler) error {
	it := client.ListApiSpecRevisions(ctx, &rpc.ListApiSpecRevisionsRequest{
		// "@-" indicates a collection of revisions, but we only want to send the resource name to the List RPC.
		Name:     strings.TrimSuffix(name.String(), "@-"),
		PageSize: pageSize,
		Filter:   filter,
	})
	for r, err := it.Next(); err != iterator.Done; r, err = it.Next() {
		if err != nil {
			return err
		}

		if getContents {
			ctx = metadata.AppendToOutgoingContext(ctx, "accept-encoding", "gzip")
			resp, err := client.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
				Name: r.GetName(),
			})
			if err != nil {
				return err
			}
			r.Contents = resp.GetData()
			if mime.IsGZipCompressed(resp.ContentType) {
				r.MimeType = mime.GUnzippedType(r.MimeType)
				r.Contents, err = compress.GUnzippedBytes(r.Contents)
				if err != nil {
					return err
				}
			}
		}

		if err := handler(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func ListArtifacts(ctx context.Context,
	client *gapic.RegistryClient,
	name names.Artifact,
	pageSize int32,
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
		Parent:   name.Parent(),
		PageSize: pageSize,
		Filter:   filter,
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

		if err := handler(ctx, r); err != nil {
			return err
		}
	}
	return nil
}
