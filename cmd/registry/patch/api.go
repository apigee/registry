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
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
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

func newAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) (*API, error) {
	apiName, err := names.ParseApi(message.Name)
	if err != nil {
		return nil, err
	}
	api := &API{
		Header: Header{
			APIVersion: RegistryV1,
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
	api.Data.RecommendedVersion, err = relativeVersionName(apiName, message.RecommendedVersion)
	if err != nil {
		return nil, err
	}
	api.Data.RecommendedDeployment, err = relativeDeploymentName(apiName, message.RecommendedDeployment)
	if err != nil {
		return nil, err
	}
	err = core.ListVersions(ctx, client, apiName.Version(""), "", func(message *rpc.ApiVersion) {
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
	if err != nil {
		return nil, err
	}
	err = core.ListDeployments(ctx, client, apiName.Deployment(""), "", func(message *rpc.ApiDeployment) {
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

// ExportAPI allows an API to be individually exported as a YAML file.
func ExportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) ([]byte, *Header, error) {
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

// relativeVersionName returns the version id if the version is within the specified API
func relativeVersionName(apiName names.Api, version string) (string, error) {
	if version == "" {
		return "", nil
	}
	versionName, err := names.ParseVersion(version)
	if err != nil {
		return "", err
	}
	if versionName.Api().String() == apiName.String() {
		return versionName.VersionID, nil
	}
	return version, nil
}

// relativeDeploymentName returns the deployment id if the deployment is within the specified API
func relativeDeploymentName(apiName names.Api, deployment string) (string, error) {
	if deployment == "" {
		return "", nil
	}
	deploymentName, err := names.ParseDeployment(deployment)
	if err != nil {
		return "", err
	}
	if deploymentName.Api().String() == apiName.String() {
		return deploymentName.DeploymentID, nil
	}
	return deployment, nil
}

// optionalVersionName returns a version name if the id is not empty
func optionalVersionName(apiName names.Api, versionID string) string {
	if versionID == "" {
		return ""
	}
	return apiName.Version(versionID).String()
}

// optionalDeploymentName returns a deployment name if the id is not empty
func optionalDeploymentName(apiName names.Api, deploymentID string) string {
	if deploymentID == "" {
		return ""
	}
	return apiName.Deployment(deploymentID).String()
}

func applyApiPatch(
	ctx context.Context,
	client connection.Client,
	api *API,
	parent string) error {
	projectName, err := names.ParseProjectWithLocation(parent)
	if err != nil {
		return err
	}
	apiName := projectName.Api(api.Metadata.Name)
	req := &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:                  apiName.String(),
			DisplayName:           api.Data.DisplayName,
			Description:           api.Data.Description,
			Availability:          api.Data.Availability,
			RecommendedVersion:    optionalVersionName(apiName, api.Data.RecommendedVersion),
			RecommendedDeployment: optionalDeploymentName(apiName, api.Data.RecommendedDeployment),
			Labels:                api.Metadata.Labels,
			Annotations:           api.Metadata.Annotations,
		},
		AllowMissing: true,
	}
	_, err = client.UpdateApi(ctx, req)
	if err != nil {
		return err
	}
	for _, versionPatch := range api.Data.APIVersions {
		err := applyApiVersionPatch(ctx, client, versionPatch, apiName.String())
		if err != nil {
			return err
		}
	}
	for _, deploymentPatch := range api.Data.APIDeployments {
		err := applyApiDeploymentPatch(ctx, client, deploymentPatch, apiName.String())
		if err != nil {
			return err
		}
	}
	return nil
}
