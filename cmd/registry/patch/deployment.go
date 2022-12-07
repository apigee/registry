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
	"bytes"
	"context"
	"strings"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/models"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"gopkg.in/yaml.v3"
)

// ExportAPIDeployment allows an API deployment to be individually exported as a YAML file.
func ExportAPIDeployment(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiDeployment, nested bool) ([]byte, *models.Header, error) {
	api, err := newApiDeployment(ctx, client, message, nested)
	if err != nil {
		return nil, nil, err
	}
	var b bytes.Buffer
	err = yamlEncoder(&b).Encode(api)
	if err != nil {
		return nil, nil, err
	}
	return b.Bytes(), &api.Header, nil
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

func newApiDeployment(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiDeployment, nested bool) (*models.ApiDeployment, error) {
	deploymentName, err := names.ParseDeployment(message.Name)
	if err != nil {
		return nil, err
	}
	revisionName := relativeSpecRevisionName(deploymentName.Api(), message.ApiSpecRevision)
	var artifacts []*models.Artifact
	if nested {
		artifacts, err = collectChildArtifacts(ctx, client, deploymentName.Artifact("-"))
		if err != nil {
			return nil, err
		}
	}
	return &models.ApiDeployment{
		Header: models.Header{
			ApiVersion: RegistryV1,
			Kind:       "Deployment",
			Metadata: models.Metadata{
				Name:        deploymentName.DeploymentID,
				Parent:      names.ExportableName(deploymentName.Parent(), deploymentName.ProjectID),
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: models.ApiDeploymentData{
			DisplayName:        message.DisplayName,
			Description:        message.Description,
			EndpointURI:        message.EndpointUri,
			ExternalChannelURI: message.ExternalChannelUri,
			IntendedAudience:   message.IntendedAudience,
			AccessGuidance:     message.AccessGuidance,
			ApiSpecRevision:    revisionName,
			Artifacts:          artifacts,
		},
	}, nil
}

func applyApiDeploymentPatchBytes(ctx context.Context, client connection.RegistryClient, bytes []byte, parent string) error {
	var deployment models.ApiDeployment
	err := yaml.Unmarshal(bytes, &deployment)
	if err != nil {
		return err
	}
	return applyApiDeploymentPatch(ctx, client, &deployment, parent)
}

func deploymentName(parent, deploymentID string) (names.Deployment, error) {
	api, err := names.ParseApi(parent)
	if err != nil {
		return names.Deployment{}, err
	}
	return api.Deployment(deploymentID), nil
}

func applyApiDeploymentPatch(
	ctx context.Context,
	client connection.RegistryClient,
	deployment *models.ApiDeployment,
	parent string) error {
	name, err := deploymentName(parent, deployment.Metadata.Name)
	if err != nil {
		return err
	}
	req := &rpc.UpdateApiDeploymentRequest{
		ApiDeployment: &rpc.ApiDeployment{
			Name:               name.String(),
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
	req.ApiDeployment.ApiSpecRevision = optionalSpecRevisionName(name, deployment.Data.ApiSpecRevision)
	if err != nil {
		return err
	}
	_, err = client.UpdateApiDeployment(ctx, req)
	if err != nil {
		return err
	}
	for _, artifactPatch := range deployment.Data.Artifacts {
		err = applyArtifactPatch(ctx, client, artifactPatch, name.String())
		if err != nil {
			return err
		}
	}
	return nil
}
