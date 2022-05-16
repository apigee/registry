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
	"strings"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

type ApiDeploymentData struct {
	DisplayName        string `yaml:"displayName,omitempty"`
	Description        string `yaml:"description,omitempty"`
	ApiSpecRevision    string `yaml:"apiSpecRevision,omitempty"`
	EndpointURI        string `yaml:"endpointURI,omitempty"`
	ExternalChannelURI string `yaml:"externalChannelURI,omitempty"`
	IntendedAudience   string `yaml:"intendedAudience,omitempty"`
	AccessGuidance     string `yaml:"accessGuidance,omitempty"`
}

type ApiDeployment struct {
	Header `yaml:",inline"`
	Data   ApiDeploymentData `yaml:"data"`
}

// relativeSpecRevisionName returns the versionid+specid if the spec is within the specified API
func relativeSpecRevisionName(apiName names.Api, spec string) string {
	if spec == "" {
		return ""
	}
	if strings.HasPrefix(spec, apiName.String()) {
		return strings.TrimPrefix(spec, apiName.String()+"/versions/")
	}
	return spec
}

// optionalSpecRevisionName returns a spec revision name if the subpath is not empty
func optionalSpecRevisionName(deploymentName names.Deployment, subpath string) string {
	if subpath == "" {
		return ""
	}
	return deploymentName.Api().String() + "/versions/" + subpath
}

func newApiDeployment(message *rpc.ApiDeployment) (*ApiDeployment, error) {
	deploymentName, err := names.ParseDeployment(message.Name)
	if err != nil {
		return nil, err
	}
	revisionName := relativeSpecRevisionName(deploymentName.Api(), message.ApiSpecRevision)
	return &ApiDeployment{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "ApiDeployment",
			Metadata: Metadata{
				Name:        deploymentName.DeploymentID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: ApiDeploymentData{
			DisplayName:        message.DisplayName,
			Description:        message.Description,
			EndpointURI:        message.EndpointUri,
			ExternalChannelURI: message.ExternalChannelUri,
			IntendedAudience:   message.IntendedAudience,
			AccessGuidance:     message.AccessGuidance,
			ApiSpecRevision:    revisionName,
		},
	}, nil
}

func applyApiDeploymentPatch(
	ctx context.Context,
	client connection.Client,
	deployment *ApiDeployment,
	parent string) error {
	apiName, err := names.ParseApi(parent)
	if err != nil {
		return err
	}
	deploymentName := apiName.Deployment(deployment.Metadata.Name)
	req := &rpc.UpdateApiDeploymentRequest{
		ApiDeployment: &rpc.ApiDeployment{
			Name:               deploymentName.String(),
			DisplayName:        deployment.Data.DisplayName,
			Description:        deployment.Data.Description,
			EndpointUri:        deployment.Data.EndpointURI,
			ExternalChannelUri: deployment.Data.ExternalChannelURI,
			IntendedAudience:   deployment.Data.IntendedAudience,
			AccessGuidance:     deployment.Data.AccessGuidance,
			Labels:             deployment.Metadata.Labels,
			Annotations:        deployment.Metadata.Annotations,
		},
		AllowMissing: true,
	}
	req.ApiDeployment.ApiSpecRevision = optionalSpecRevisionName(deploymentName, deployment.Data.ApiSpecRevision)
	if err != nil {
		return err
	}
	_, err = client.UpdateApiDeployment(ctx, req)
	return err
}
