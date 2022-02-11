package patch

import (
	"context"
	"fmt"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

type APIVersion struct {
	Header `yaml:",inline"`
	Data   struct {
		DisplayName string      `yaml:"displayName,omitempty"`
		Description string      `yaml:"description,omitempty"`
		State       string      `yaml:"state,omitempty"`
		APISpecs    []*APISpec  `yaml:"specs,omitempty"`
		Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
	} `yaml:"data"`
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
	version.Data.State = message.State
	err = core.ListSpecs(ctx, client, versionName.Spec(""), "", func(message *rpc.ApiSpec) {
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

func applyApiVersionPatch(
	ctx context.Context,
	client connection.Client,
	version *APIVersion,
	parent string) error {
	name := fmt.Sprintf("%s/versions/%s", parent, version.Metadata.Name)
	req := &rpc.UpdateApiVersionRequest{
		ApiVersion: &rpc.ApiVersion{
			Name:        name,
			DisplayName: version.Data.DisplayName,
			Description: version.Data.Description,
			State:       version.Data.State,
			Labels:      version.Metadata.Labels,
			Annotations: version.Metadata.Annotations,
		},
		AllowMissing: true,
	}
	_, err := client.UpdateApiVersion(ctx, req)
	if err != nil {
		return err
	}
	for _, specPatch := range version.Data.APISpecs {
		err := applyApiSpecPatch(ctx, client, specPatch, name)
		if err != nil {
			return err
		}
	}
	return nil
}
