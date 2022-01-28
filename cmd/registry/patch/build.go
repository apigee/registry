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

package patch

import (
	"context"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/proto"
)

func buildAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) (*API, error) {
	apiName, err := names.ParseApi(message.Name)
	if err != nil {
		return nil, err
	}
	api := &API{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "API",
			Metadata: Metadata{
				Name:        apiName.ApiID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
	}
	api.Body.DisplayName = message.DisplayName
	api.Body.Description = message.Description
	api.Body.Availability = message.Availability
	api.Body.RecommendedVersion = message.RecommendedVersion
	api.Body.RecommendedDeployment = message.RecommendedDeployment
	core.ListVersions(ctx, client, apiName.Version(""), "", func(message *rpc.ApiVersion) {
		version, err2 := buildAPIVersion(ctx, client, message)
		// unset these because they can be inferred
		version.APIVersion = ""
		version.Kind = ""
		if err2 == nil {
			api.Body.APIVersions = append(api.Body.APIVersions, version)
		} else {
			err = err2
		}
	})
	core.ListDeployments(ctx, client, apiName.Deployment(""), "", func(message *rpc.ApiDeployment) {
		deployment, err2 := buildAPIDeployment(ctx, client, message)
		// unset these because they can be inferred
		deployment.APIVersion = ""
		deployment.Kind = ""
		if err2 == nil {
			api.Body.APIDeployments = append(api.Body.APIDeployments, deployment)
		} else {
			err = err2
		}
	})
	return api, err
}

func buildAPIVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiVersion) (*APIVersion, error) {
	versionName, err := names.ParseVersion(message.Name)
	if err != nil {
		return nil, err
	}
	version := &APIVersion{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "APIVersion",
			Metadata: Metadata{
				Name:        versionName.VersionID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
	}
	version.Body.DisplayName = message.DisplayName
	version.Body.Description = message.Description
	core.ListSpecs(ctx, client, versionName.Spec(""), "", func(message *rpc.ApiSpec) {
		spec, err2 := buildAPISpec(ctx, client, message)
		// unset these because they can be inferred
		spec.APIVersion = ""
		spec.Kind = ""
		if err2 == nil {
			version.Body.APISpecs = append(version.Body.APISpecs, spec)
		} else {
			err = err2
		}
	})
	return version, err
}

func buildAPISpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiSpec) (*APISpec, error) {
	specName, err := names.ParseSpec(message.Name)
	if err != nil {
		return nil, err
	}
	spec := &APISpec{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "APISpec",
			Metadata: Metadata{
				Name:        specName.SpecID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
	}
	spec.Body.FileName = message.Filename
	spec.Body.Description = message.Description
	spec.Body.MimeType = message.MimeType
	spec.Body.SourceURI = message.SourceUri
	return spec, nil
}

func buildAPIDeployment(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiDeployment) (*APIDeployment, error) {
	deploymentName, err := names.ParseDeployment(message.Name)
	if err != nil {
		return nil, err
	}
	deployment := &APIDeployment{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "APIDeployment",
			Metadata: Metadata{
				Name:        deploymentName.DeploymentID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
	}
	deployment.Body.DisplayName = message.DisplayName
	deployment.Body.Description = message.Description
	return deployment, nil
}

func buildArtifact(message *rpc.Artifact) (*Artifact, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	artifact := &Artifact{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "Artifact",
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
	}
	return artifact, nil
}

func buildManifest(message *rpc.Artifact) (*Manifest, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	value := &rpc.Manifest{}
	err = proto.Unmarshal(message.Contents, value)
	if err != nil {
		return nil, err
	}
	manifest := &Manifest{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "Manifest",
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
	}

	for _, g := range value.GeneratedResources {
		dependencies := make([]*ManifestDependency, 0)
		for _, d := range g.Dependencies {
			dependencies = append(dependencies,
				&ManifestDependency{
					Pattern: d.Pattern,
					Filter:  d.Filter,
				})
		}
		manifest.Body.GeneratedResources = append(
			manifest.Body.GeneratedResources,
			&ManifestGeneratedResource{
				Pattern:      g.Pattern,
				Filter:       g.Filter,
				Receipt:      g.Receipt,
				Dependencies: dependencies,
				Action:       g.Action,
			})
	}
	return manifest, nil
}
