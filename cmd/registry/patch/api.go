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

type API struct {
	Header `yaml:",inline"`
	Data   struct {
		DisplayName           string           `yaml:"displayName,omitempty"`
		Description           string           `yaml:"description,omitempty"`
		Availability          string           `yaml:"availability,omitempty"`
		RecommendedVersion    string           `yaml:"recommendedVersion,omitempty"`
		RecommendedDeployment string           `yaml:"recommendedDeployment,omitempty"`
		APIVersions           []*APIVersion    `yaml:"versions,omitempty"`
		APIDeployments        []*APIDeployment `yaml:"deployments,omitempty"`
		Artifacts             []*Artifact      `yaml:"artifacts,omitempty"`
	} `yaml:"data"`
}

type APIVersion struct {
	Header `yaml:",inline"`
	Data   struct {
		DisplayName string      `yaml:"displayName,omitempty"`
		Description string      `yaml:"description,omitempty"`
		APISpecs    []*APISpec  `yaml:"specs,omitempty"`
		Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"data"`
}

type APISpec struct {
	Header `yaml:",inline"`
	Data   struct {
		FileName    string      `yaml:"fileName,omitempty"`
		Description string      `yaml:"description,omitempty"`
		MimeType    string      `yaml:"mimeType,omitempty"`
		SourceURI   string      `yaml:"sourceURI,omitempty"`
		Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"data"`
}

type APIDeployment struct {
	Header `yaml:",inline"`
	Data   struct {
		DisplayName string      `yaml:"displayName,omitempty"`
		Description string      `yaml:"description,omitempty"`
		Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"data"`
}

func newAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) (*API, error) {
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
	api.Data.DisplayName = message.DisplayName
	api.Data.Description = message.Description
	api.Data.Availability = message.Availability
	api.Data.RecommendedVersion = message.RecommendedVersion
	api.Data.RecommendedDeployment = message.RecommendedDeployment
	core.ListVersions(ctx, client, apiName.Version(""), "", func(message *rpc.ApiVersion) {
		version, err2 := newAPIVersion(ctx, client, message)
		// unset these because they can be inferred
		version.APIVersion = ""
		version.Kind = ""
		if err2 == nil {
			api.Data.APIVersions = append(api.Data.APIVersions, version)
		} else {
			err = err2
		}
	})
	core.ListDeployments(ctx, client, apiName.Deployment(""), "", func(message *rpc.ApiDeployment) {
		deployment, err2 := newAPIDeployment(ctx, client, message)
		// unset these because they can be inferred
		deployment.APIVersion = ""
		deployment.Kind = ""
		if err2 == nil {
			api.Data.APIDeployments = append(api.Data.APIDeployments, deployment)
		} else {
			err = err2
		}
	})
	return api, err
}

func newAPIVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiVersion) (*APIVersion, error) {
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
	version.Data.DisplayName = message.DisplayName
	version.Data.Description = message.Description
	core.ListSpecs(ctx, client, versionName.Spec(""), "", func(message *rpc.ApiSpec) {
		spec, err2 := newAPISpec(ctx, client, message)
		// unset these because they can be inferred
		spec.APIVersion = ""
		spec.Kind = ""
		if err2 == nil {
			version.Data.APISpecs = append(version.Data.APISpecs, spec)
		} else {
			err = err2
		}
	})
	return version, err
}

func newAPISpec(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiSpec) (*APISpec, error) {
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
	spec.Data.FileName = message.Filename
	spec.Data.Description = message.Description
	spec.Data.MimeType = message.MimeType
	spec.Data.SourceURI = message.SourceUri
	return spec, nil
}

func newAPIDeployment(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiDeployment) (*APIDeployment, error) {
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
	deployment.Data.DisplayName = message.DisplayName
	deployment.Data.Description = message.Description
	return deployment, nil
}

// ExportAPI writes an API as a YAML file.
func ExportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) {
	bytes, _, err := exportAPI(ctx, client, message)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact")
	} else {
		fmt.Println(string(bytes))
	}
}

func exportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) ([]byte, *Header, error) {
	api, err := newAPI(ctx, client, message)
	if err != nil {
		return nil, nil, err
	}
	b, err := yaml.Marshal(api)
	if err != nil {
		return nil, nil, err
	}
	return b, &api.Header, nil
}
