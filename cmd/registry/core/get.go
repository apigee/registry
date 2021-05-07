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
)

func GetProject(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler ProjectHandler) (*rpc.Project, error) {
	request := &rpc.GetProjectRequest{
		Name: "projects/" + segments[1],
	}
	project, err := client.GetProject(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(project)
	}
	return project, nil
}

func GetAPI(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler ApiHandler) (*rpc.Api, error) {
	request := &rpc.GetApiRequest{
		Name: "projects/" + segments[1] + "/apis/" + segments[2],
	}
	api, err := client.GetApi(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(api)
	}
	return api, nil
}

func GetVersion(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	handler VersionHandler) (*rpc.ApiVersion, error) {
	request := &rpc.GetApiVersionRequest{
		Name: "projects/" + segments[1] + "/apis/" + segments[2] + "/versions/" + segments[3],
	}
	version, err := client.GetApiVersion(ctx, request)
	if err != nil {
		return nil, err
	}
	if handler != nil {
		handler(version)
	}
	return version, nil
}

func GetSpec(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	getContents bool,
	handler SpecHandler) (*rpc.ApiSpec, error) {
	request := &rpc.GetApiSpecRequest{
		Name: "projects/" + segments[1] + "/apis/" + segments[2] + "/versions/" + segments[3] + "/specs/" + segments[4],
	}
	spec, err := client.GetApiSpec(ctx, request)
	if err != nil {
		return nil, err
	}
	if getContents {
		request := &rpc.GetApiSpecContentsRequest{
			Name: fmt.Sprintf("%s/contents", spec.GetName()),
		}
		contents, err := client.GetApiSpecContents(ctx, request)
		if err != nil {
			return nil, err
		}
		spec.Contents = contents.GetData()
		spec.MimeType = contents.GetContentType()
	}
	if handler != nil {
		handler(spec)
	}
	return spec, nil
}

func GetArtifact(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string,
	getContents bool,
	handler ArtifactHandler) (*rpc.Artifact, error) {
	request := &rpc.GetArtifactRequest{}
	if segments[3] == "" {
		request.Name = "projects/" + segments[1]
	} else if segments[5] == "" {
		request.Name = "projects/" + segments[1] + "/apis/" + segments[3]
	} else if segments[7] == "" {
		request.Name = "projects/" + segments[1] + "/apis/" + segments[3] + "/versions/" + segments[5]
	} else {
		request.Name = "projects/" + segments[1] + "/apis/" + segments[3] + "/versions/" + segments[5] + "/specs/" + segments[7]
	}
	request.Name += "/artifacts/" + segments[8]

	artifact, err := client.GetArtifact(ctx, request)
	if err != nil {
		return nil, err
	}
	if getContents {
		request := &rpc.GetArtifactContentsRequest{
			Name: fmt.Sprintf("%s/contents", artifact.GetName()),
		}
		contents, err := client.GetArtifactContents(ctx, request)
		if err != nil {
			return nil, err
		}
		artifact.Contents = contents.GetData()
		artifact.MimeType = contents.GetContentType()
	}
	if handler != nil {
		handler(artifact)
	}
	return artifact, nil
}
