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

type ApiData struct {
	DisplayName           string           `yaml:"displayName,omitempty"`
	Description           string           `yaml:"description,omitempty"`
	Availability          string           `yaml:"availability,omitempty"`
	RecommendedVersion    string           `yaml:"recommendedVersion,omitempty"`
	RecommendedDeployment string           `yaml:"recommendedDeployment,omitempty"`
	ApiVersions           []*ApiVersion    `yaml:"versions,omitempty"`
	ApiDeployments        []*ApiDeployment `yaml:"deployments,omitempty"`
}

type Api struct {
	Header `yaml:",inline"`
	Data   ApiData `yaml:"data"`
}

func newApi(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) (*Api, error) {
	apiName, err := names.ParseApi(message.Name)
	if err != nil {
		return nil, err
	}
	api := &Api{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "API",
			Metadata: Metadata{
				Name:        apiName.ApiID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: ApiData{
			DisplayName:  message.DisplayName,
			Description:  message.Description,
			Availability: message.Availability,
		},
	}
	api.Data.RecommendedVersion, err = relativeVersionName(apiName, message.RecommendedVersion)
	if err != nil {
		return nil, err
	}
	api.Data.RecommendedDeployment, err = relativeDeploymentName(apiName, message.RecommendedDeployment)
	if err != nil {
		return nil, err
	}
	var innerErr error // TODO: remove when ListVersions accepts a handler that returns errors
	err = core.ListVersions(ctx, client, apiName.Version("-"), "", func(message *rpc.ApiVersion) {
		if innerErr != nil {
			return
		}
		var version *ApiVersion
		version, innerErr = newApiVersion(ctx, client, message)
		if innerErr == nil {
			// unset these because they can be inferred
			version.ApiVersion = ""
			version.Kind = ""
			api.Data.ApiVersions = append(api.Data.ApiVersions, version)
		}
	})
	if innerErr != nil {
		return nil, innerErr
	}
	if err != nil {
		return nil, err
	}
	err = core.ListDeployments(ctx, client, apiName.Deployment("-"), "", func(message *rpc.ApiDeployment) {
		if innerErr != nil {
			return
		}
		var deployment *ApiDeployment
		deployment, innerErr = newApiDeployment(ctx, client, message)
		if innerErr == nil {
			// unset these because they can be inferred
			deployment.ApiVersion = ""
			deployment.Kind = ""
			api.Data.ApiDeployments = append(api.Data.ApiDeployments, deployment)
		}
	})
	if innerErr != nil {
		return nil, innerErr
	}
	return api, err
}

// ExportAPI allows an API to be individually exported as a YAML file.
func ExportAPI(ctx context.Context, client *gapic.RegistryClient, message *rpc.Api) ([]byte, *Header, error) {
	api, err := newApi(ctx, client, message)
	if err != nil {
		return nil, nil, err
	}
	b, err := yaml.Marshal(api)
	if err != nil {
		return nil, nil, err
	}
	return b, &api.Header, nil
}

// TODO: These functions assume that their arguments are valid names and fail the export if they aren't.
// They are used to replace the absolute resource names stored by the API with more reusable relative
// resource names that allow the exported files to be applied to arbitrary projects. This more
// concise form is also more convenient for users to specify. But it requires additional validation.
// Either we 1) only allow relative names in these fields in the YAML representation or
// 2) allow both absolute and relative names in these fields and handle them appropriately.
// We should probably support only one output format; for this, relative names seem preferable.

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

// TODO: The following functions assume that their arguments are truly an ID,
// but there's there's no validation (here or in the caller) to prevent a full
// resource name from being passed in. Users specifying a full resource name will
// end up with a malformed name being stored.

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

func applyApiPatch(ctx context.Context, client connection.Client, bytes []byte, parent string) error {
	var api Api
	err := yaml.Unmarshal(bytes, &api)
	if err != nil {
		return err
	}
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
	for _, versionPatch := range api.Data.ApiVersions {
		err := applyApiVersionPatch(ctx, client, versionPatch, apiName.String())
		if err != nil {
			return err
		}
	}
	for _, deploymentPatch := range api.Data.ApiDeployments {
		err := applyApiDeploymentPatch(ctx, client, deploymentPatch, apiName.String())
		if err != nil {
			return err
		}
	}
	return nil
}
