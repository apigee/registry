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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"gopkg.in/yaml.v2"
)

func exportProject(ctx context.Context, client *gapic.RegistryClient, name string) (*Project, error) {
	projectName, err := names.ParseProject(name)
	if err != nil {
		return nil, err
	}
	project := &Project{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "Project",
			Metadata: Metadata{
				Name: projectName.ProjectID,
			},
		},
	}
	core.ListAPIs(ctx, client, projectName.Api(""), "", func(message *rpc.Api) {
		api, err2 := exportAPI(ctx, client, message)
		if err2 == nil {
			project.Spec.APIs = append(project.Spec.APIs, api)
		} else {
			err = err2
		}
	})
	return project, err
}

func exportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) (*API, error) {
	apiName, err := names.ParseApi(message.Name)
	if err != nil {
		return nil, err
	}
	api := &API{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "API",
			Metadata: Metadata{
				Name: apiName.ApiID,
			},
		},
	}
	api.Spec.DisplayName = message.DisplayName
	api.Spec.Description = message.Description
	api.Spec.Availability = message.Availability
	api.Spec.RecommendedVersion = message.RecommendedVersion
	api.Spec.RecommendedDeployment = message.RecommendedDeployment
	core.ListVersions(ctx, client, apiName.Version(""), "", func(message *rpc.ApiVersion) {
		version, err2 := exportAPIVersion(ctx, client, message)
		if err2 == nil {
			api.Spec.APIVersions = append(api.Spec.APIVersions, version)
		} else {
			err = err2
		}
	})
	core.ListDeployments(ctx, client, apiName.Deployment(""), "", func(message *rpc.ApiDeployment) {
		deployment, err2 := exportAPIDeployment(ctx, client, message)
		if err2 == nil {
			api.Spec.APIDeployments = append(api.Spec.APIDeployments, deployment)
		} else {
			err = err2
		}
	})
	return api, err
}

func exportAPIVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiVersion) (*APIVersion, error) {
	versionName, err := names.ParseVersion(message.Name)
	if err != nil {
		return nil, err
	}
	version := &APIVersion{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "APIVersion",
			Metadata: Metadata{
				Name: versionName.VersionID,
			},
		},
	}
	version.Spec.DisplayName = message.DisplayName
	version.Spec.Description = message.Description
	core.ListSpecs(ctx, client, versionName.Spec(""), "", func(message *rpc.ApiSpec) {
		spec, err2 := exportAPISpec(ctx, client, message)
		if err2 != nil {
			version.Spec.APISpecs = append(version.Spec.APISpecs, spec)
		} else {
			err = err2
		}
	})
	return version, err
}

func exportAPISpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiSpec) (*APISpec, error) {
	specName, err := names.ParseSpec(message.Name)
	if err != nil {
		return nil, err
	}
	spec := &APISpec{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "APISpec",
			Metadata: Metadata{
				Name: specName.SpecID,
			},
		},
	}
	spec.Spec.FileName = message.Filename
	spec.Spec.Description = message.Description
	spec.Spec.MimeType = message.MimeType
	spec.Spec.SourceURI = message.SourceUri
	return spec, nil
}

func exportAPIDeployment(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiDeployment) (*APIDeployment, error) {
	deploymentName, err := names.ParseDeployment(message.Name)
	if err != nil {
		return nil, err
	}
	deployment := &APIDeployment{
		Header: Header{
			APIVersion: REGISTRY_V1,
			Kind:       "APIDeployment",
			Metadata: Metadata{
				Name: deploymentName.DeploymentID,
			},
		},
	}
	deployment.Spec.DisplayName = message.DisplayName
	deployment.Spec.Description = message.Description
	return deployment, nil
}

func exportArtifact(message *rpc.Artifact) (*Artifact, error) {
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
	project, err := exportProject(ctx, client, message.Name)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export project")
	}
	b, err := yaml.Marshal(project)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportAPI writes an API as a YAML file.
func ExportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) {
	api, err := exportAPI(ctx, client, message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export api")
	}
	b, err := yaml.Marshal(api)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportAPIVersion writes an API version as a YAML file.
func ExportAPIVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiVersion) {
	version, err := exportAPIVersion(ctx, client, message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export version")
	}
	b, err := yaml.Marshal(version)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportAPISpec writes an API spec as a YAML file.
func ExportAPISpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiSpec) {
	spec, err := exportAPISpec(ctx, client, message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export spec")
	}
	b, err := yaml.Marshal(spec)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportAPIDeployment writes an API deployment as a YAML file.
func ExportAPIDeployment(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiDeployment) {
	deployment, err := exportAPIDeployment(ctx, client, message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export deployment")
	}
	b, err := yaml.Marshal(deployment)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportArtifact writes an artifact as a YAML file.
func ExportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) {
	artifact, err := exportArtifact(message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact")
	}
	b, err := yaml.Marshal(artifact)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}
