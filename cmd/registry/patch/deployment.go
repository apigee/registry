package patch

import (
	"context"
	"fmt"

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
	deployment.Data.APISpecRevision = message.ApiSpecRevision
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
	name := fmt.Sprintf("%s/deployments/%s", parent, deployment.Metadata.Name)
	req := &rpc.UpdateApiDeploymentRequest{
		ApiDeployment: &rpc.ApiDeployment{
			Name:               name,
			DisplayName:        deployment.Data.DisplayName,
			Description:        deployment.Data.Description,
			ApiSpecRevision:    deployment.Data.APISpecRevision,
			EndpointUri:        deployment.Data.EndpointURI,
			ExternalChannelUri: deployment.Data.ExternalChannelURI,
			IntendedAudience:   deployment.Data.IntendedAudience,
			AccessGuidance:     deployment.Data.AccessGuidance,
			Labels:             deployment.Metadata.Labels,
			Annotations:        deployment.Metadata.Annotations,
		},
		AllowMissing: true,
	}
	_, err := client.UpdateApiDeployment(ctx, req)
	return err
}
