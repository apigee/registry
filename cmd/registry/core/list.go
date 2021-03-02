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
	"google.golang.org/api/iterator"
)

func ListProjects(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	handler ProjectHandler) error {
	request := &rpc.ListProjectsRequest{}
	filter := filterFlag
	if len(segments) == 2 && segments[1] != "-" {
		filter = "project_id == '" + segments[1] + "'"
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
	segments []string,
	filterFlag string,
	handler ApiHandler) error {
	request := &rpc.ListApisRequest{
		Parent: "projects/" + segments[1],
	}
	filter := filterFlag
	if len(segments) == 3 && segments[2] != "-" {
		filter = "api_id == '" + segments[2] + "'"
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

func ListVersions(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	filterFlag string,
	handler VersionHandler) error {
	request := &rpc.ListApiVersionsRequest{
		Parent: "projects/" + segments[1] + "/apis/" + segments[2],
	}
	filter := filterFlag
	if len(segments) == 4 && segments[3] != "-" {
		filter = "version_id == '" + segments[3] + "'"
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
	segments []string,
	filterFlag string,
	handler SpecHandler) error {
	request := &rpc.ListApiSpecsRequest{
		Parent: "projects/" + segments[1] + "/apis/" + segments[2] + "/versions/" + segments[3],
	}
	filter := filterFlag
	if len(segments) > 4 && segments[4] != "-" {
		filter = "spec_id == '" + segments[4] + "'"
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
	segments []string,
	filterFlag string,
	handler SpecHandler) error {
	request := &rpc.ListApiSpecRevisionsRequest{
		Name: "projects/" + segments[1] +
			"/apis/" + segments[2] +
			"/versions/" + segments[3] +
			"/specs/" + segments[4],
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
	segments []string,
	filterFlag string,
	getContents bool,
	handler ArtifactHandler) error {
	parent := "projects/" + segments[1]
	if segments[3] != "" {
		parent += "/apis/" + segments[3]
		if segments[5] != "" {
			parent += "/versions/" + segments[5]
			if segments[7] != "" {
				parent += "/specs/" + segments[7]
			}
		}
	}
	request := &rpc.ListArtifactsRequest{
		Parent: parent,
		View:   rpc.View_BASIC,
	}
	if getContents {
		request.View = rpc.View_FULL
	}
	filter := filterFlag
	if len(segments) == 9 && segments[8] != "-" {
		if filter != "" {
			filter += " && "
		}
		filter += "artifact_id == '" + segments[8] + "'"
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
		handler(artifact)
	}
	return nil
}

func ListArtifactsForParent(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler ArtifactHandler) error {
	parent := "projects/" + segments[1]
	if len(segments) > 2 {
		parent += "/apis/" + segments[2]
		if len(segments) > 3 {
			parent += "/versions/" + segments[3]
			if len(segments) > 4 {
				parent += "/specs/" + segments[4]
			}
		}
	}
	request := &rpc.ListArtifactsRequest{
		Parent: parent,
		View:   rpc.View_BASIC,
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
