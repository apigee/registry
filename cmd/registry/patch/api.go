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

func applyApiPatch(
	ctx context.Context,
	client connection.Client,
	api *API,
	parent string) error {
	name := fmt.Sprintf("%s/apis/%s", parent, api.Metadata.Name)
	req := &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:                  name,
			DisplayName:           api.Data.DisplayName,
			Description:           api.Data.Description,
			Availability:          api.Data.Availability,
			RecommendedVersion:    api.Data.RecommendedVersion,
			RecommendedDeployment: api.Data.RecommendedDeployment,
			Labels:                api.Metadata.Labels,
			Annotations:           api.Metadata.Annotations,
		},
		AllowMissing: true,
	}
	_, err := client.UpdateApi(ctx, req)
	if err != nil {
		return err
	}
	for _, versionPatch := range api.Data.APIVersions {
		err := applyApiVersionPatch(ctx, client, versionPatch, name)
		if err != nil {
			return err
		}
	}
	for _, deploymentPatch := range api.Data.APIDeployments {
		err := applyApiDeploymentPatch(ctx, client, deploymentPatch, name)
		if err != nil {
			return err
		}
	}
	return nil
}
