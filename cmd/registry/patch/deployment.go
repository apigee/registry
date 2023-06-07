// Copyright 2022 Google LLC.
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
	"strings"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/encoding"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"gopkg.in/yaml.v3"
)

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

// resolveSpecRevisionName returns a "full-resolved" spec revision name
// that is relative to the root ("/") and includes a revision ID.
func resolveSpecRevisionName(ctx context.Context, client *gapic.RegistryClient, deploymentName names.Deployment, subpath string) (string, error) {
	if subpath == "" {
		return "", nil
	}
	specName := deploymentName.Api().String() + "/versions/" + subpath
	specRevisionName, err := names.ParseSpecRevision(specName)
	if err == nil && specRevisionName.RevisionID != "" {
		// this already includes a revision, we're good
		return specName, nil
	}
	_, err = names.ParseSpec(specName)
	if err != nil {
		// this isn't a valid spec name, so filter it out
		return "", err
	}
	// Get the latest revision of this spec
	spec, err := client.GetApiSpec(ctx, &rpc.GetApiSpecRequest{Name: specName})
	if err != nil {
		return "", err
	}
	return spec.Name + "@" + spec.RevisionId, nil
}

// NewApiDeployment allows an API deployment to be individually exported as a YAML file.
func NewApiDeployment(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiDeployment, nested bool) (*encoding.ApiDeployment, error) {
	deploymentName, err := names.ParseDeploymentRevision(message.Name)
	if err != nil {
		return nil, err
	}
	revisionName := relativeSpecRevisionName(deploymentName.Api(), message.ApiSpecRevision)
	var artifacts []*encoding.Artifact
	if nested {
		artifacts, err = collectChildArtifacts(ctx, client, deploymentName.Artifact("-"))
		if err != nil {
			return nil, err
		}
	}
	return &encoding.ApiDeployment{
		Header: encoding.Header{
			ApiVersion: encoding.RegistryV1,
			Kind:       "Deployment",
			Metadata: encoding.Metadata{
				Name:        deploymentName.DeploymentID,
				Parent:      names.ExportableName(deploymentName.Parent(), deploymentName.ProjectID),
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: encoding.ApiDeploymentData{
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

func applyApiDeploymentPatchBytes(ctx context.Context, client connection.RegistryClient, bytes []byte, project string, filename string) error {
	var deployment encoding.ApiDeployment
	err := yaml.Unmarshal(bytes, &deployment)
	if err != nil {
		return err
	}
	return applyApiDeploymentPatch(ctx, client, &deployment, project, filename)
}

func deploymentName(parent string, metadata encoding.Metadata) (names.Deployment, error) {
	if metadata.Parent != "" {
		parent = parent + "/" + metadata.Parent
	}
	api, err := names.ParseApi(parent)
	if err != nil {
		return names.Deployment{}, err
	}
	return api.Deployment(metadata.Name), nil
}

func applyApiDeploymentPatch(
	ctx context.Context,
	client connection.RegistryClient,
	deployment *encoding.ApiDeployment,
	parent string,
	filename string) error {
	name, err := deploymentName(parent, deployment.Metadata)
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
	req.ApiDeployment.ApiSpecRevision, err = resolveSpecRevisionName(ctx, client, name, deployment.Data.ApiSpecRevision)
	if err != nil {
		return err
	}
	_, err = client.UpdateApiDeployment(ctx, req)
	if err != nil {
		return fmt.Errorf("UpdateApiDeployment: %s", err)
	}
	for _, artifactPatch := range deployment.Data.Artifacts {
		err = applyArtifactPatch(ctx, client, artifactPatch, name.String(), filename)
		if err != nil {
			return err
		}
	}
	return nil
}
