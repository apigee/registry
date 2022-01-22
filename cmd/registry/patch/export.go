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

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"gopkg.in/yaml.v2"
)

func exportProject(name string) *Project {
	project := &Project{}
	project.Kind = "Project"
	project.Metadata.Name = name
	return project
}

func exportAPI(message *rpc.Api) *API {
	api := &API{}
	api.Kind = "API"
	api.Metadata.Name = message.Name
	return api
}

func exportAPIVersion(message *rpc.ApiVersion) *APIVersion {
	version := &APIVersion{}
	version.Kind = "APIVersion"
	version.Metadata.Name = message.Name
	return version
}

func exportAPISpec(message *rpc.ApiSpec) *APISpec {
	spec := &APISpec{}
	spec.Kind = "APISpec"
	spec.Spec.FileName = message.Filename
	return spec
}

func exportAPIDeployment(message *rpc.ApiDeployment) *APIDeployment {
	deployment := &APIDeployment{}
	deployment.Kind = "APIDeployment"
	deployment.Metadata.Name = message.Name
	return deployment
}

func exportArtifact(message *rpc.Artifact) *Artifact {
	artifact := &Artifact{}
	artifact.Kind = "Artifact"
	artifact.Metadata.Name = message.Name
	return artifact
}

// ExportProject writes a project as a YAML file.
func ExportProject(ctx context.Context, client *gapic.RegistryClient, adminClient *gapic.AdminClient, message *rpc.Project) {
	project := exportProject(message.Name)
	b, err := yaml.Marshal(project)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportAPI writes an API as a YAML file.
func ExportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) {
	api := exportAPI(message)
	b, err := yaml.Marshal(api)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportAPIVersion writes an API version as a YAML file.
func ExportAPIVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiVersion) {
	api := exportAPIVersion(message)
	b, err := yaml.Marshal(api)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportAPISpec writes an API spec as a YAML file.
func ExportAPISpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiSpec) {
	api := exportAPISpec(message)
	b, err := yaml.Marshal(api)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportAPIDeployment writes an API deployment as a YAML file.
func ExportAPIDeployment(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiDeployment) {
	api := exportAPIDeployment(message)
	b, err := yaml.Marshal(api)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}

// ExportArtifact writes an artifact as a YAML file.
func ExportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) {
	api := exportArtifact(message)
	b, err := yaml.Marshal(api)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal doc as YAML")
	}
	fmt.Println(string(b))
}
