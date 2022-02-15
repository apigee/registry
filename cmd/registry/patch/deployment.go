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
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

type APIDeployment struct {
	Header `yaml:",inline"`
	Data   struct {
		DisplayName        string      `yaml:"displayName,omitempty"`
		Description        string      `yaml:"description,omitempty"`
		APISpecRevision    string      `yaml:"apiSpecRevision,omitempty"`
		EndpointURI        string      `yaml:"endpointURI,omitempty"`
		ExternalChannelURI string      `yaml:"externalChannelURI,omitempty"`
		IntendedAudience   string      `yaml:"intendedAudience,omitempty"`
		AccessGuidance     string      `yaml:"accessGuidance,omitempty"`
		Artifacts          []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"data"`
}

// relativeSpecRevisionName returns the versionid+specid if the spec is within the specified API
func relativeSpecRevisionName(apiName names.Api, spec string) (string, error) {
	if strings.HasPrefix(spec, apiName.String()) {
		return strings.TrimPrefix(spec, apiName.String()+"/versions/"), nil
	}
	return spec, nil
}

// optionalSpecRevisionName returns a spec revision name if the subpath is not empty
func optionalSpecRevisionName(deploymentName names.Deployment, subpath string) string {
	if subpath == "" {
		return ""
	}
	return deploymentName.Api().String() + "/versions/" + subpath
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
	deployment.Data.APISpecRevision, err = relativeSpecRevisionName(deploymentName.Api(), message.ApiSpecRevision)
	if err != nil {
		return nil, err
	}
	deployment.Data.EndpointURI = message.EndpointUri
	deployment.Data.ExternalChannelURI = message.ExternalChannelUri
	deployment.Data.IntendedAudience = message.IntendedAudience
	deployment.Data.AccessGuidance = message.AccessGuidance
	return deployment, nil
}

func applyApiDeploymentPatch(
	ctx context.Context,
	client connection.Client,
	deployment *APIDeployment,
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
	req.ApiDeployment.ApiSpecRevision = optionalSpecRevisionName(deploymentName, deployment.Data.APISpecRevision)
	if err != nil {
		return err
	}
	_, err = client.UpdateApiDeployment(ctx, req)
	return err
}
