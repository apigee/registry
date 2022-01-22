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
	"fmt"
	"os"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"gopkg.in/yaml.v2"
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
	api.Spec.DisplayName = message.DisplayName
	api.Spec.Description = message.Description
	api.Spec.Availability = message.Availability
	api.Spec.RecommendedVersion = message.RecommendedVersion
	api.Spec.RecommendedDeployment = message.RecommendedDeployment
	core.ListVersions(ctx, client, apiName.Version(""), "", func(message *rpc.ApiVersion) {
		version, err2 := buildAPIVersion(ctx, client, message)
		// unset these because they can be inferred
		version.APIVersion = ""
		version.Kind = ""
		if err2 == nil {
			api.Spec.APIVersions = append(api.Spec.APIVersions, version)
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
			api.Spec.APIDeployments = append(api.Spec.APIDeployments, deployment)
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
	version.Spec.DisplayName = message.DisplayName
	version.Spec.Description = message.Description
	core.ListSpecs(ctx, client, versionName.Spec(""), "", func(message *rpc.ApiSpec) {
		spec, err2 := buildAPISpec(ctx, client, message)
		// unset these because they can be inferred
		spec.APIVersion = ""
		spec.Kind = ""
		if err2 == nil {
			version.Spec.APISpecs = append(version.Spec.APISpecs, spec)
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
	spec.Spec.FileName = message.Filename
	spec.Spec.Description = message.Description
	spec.Spec.MimeType = message.MimeType
	spec.Spec.SourceURI = message.SourceUri
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
	deployment.Spec.DisplayName = message.DisplayName
	deployment.Spec.Description = message.Description
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

// ExportProject writes a project as a YAML file.
func ExportProject(ctx context.Context, client *gapic.RegistryClient, adminClient *gapic.AdminClient, message *rpc.Project) {
	projectName, err := names.ParseProject(message.Name)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to parse project name")
	}
	apisdir := "apis"
	err = os.MkdirAll(apisdir, 0777)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to create output directory")
	}
	core.ListAPIs(ctx, client, projectName.Api(""), "", func(message *rpc.Api) {
		api, err := buildAPI(ctx, client, message)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to export api as YAML")
			return
		}
		b, err := yaml.Marshal(api)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to marshal api as YAML")
			return
		}
		filename := fmt.Sprintf("%s/%s.yaml", apisdir, api.Header.Metadata.Name)
		err = os.WriteFile(filename, b, 0644)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to write output YAML")
		}
	})
}

// ExportAPI writes an API as a YAML file.
func ExportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) {
	api, err := buildAPI(ctx, client, message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export api")
	}
	b, err := yaml.Marshal(api)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportArtifact writes an artifact as a YAML file.
func ExportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) {
	artifact, err := buildArtifact(message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact")
	}
	b, err := yaml.Marshal(artifact)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}
