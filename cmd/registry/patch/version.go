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
)

type ApiVersionData struct {
	DisplayName string      `yaml:"displayName,omitempty"`
	Description string      `yaml:"description,omitempty"`
	State       string      `yaml:"state,omitempty"`
	ApiSpecs    []*ApiSpec  `yaml:"specs,omitempty"`
	Artifacts   []*Artifact `yaml:"artifacts,omitempty"`
}

type ApiVersion struct {
	Header `yaml:",inline"`
	Data   ApiVersionData `yaml:"data"`
}

func newApiVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiVersion) (*ApiVersion, error) {
	versionName, err := names.ParseVersion(message.Name)
	if err != nil {
		return nil, err
	}
	version := &ApiVersion{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "ApiVersion",
			Metadata: Metadata{
				Name:        versionName.VersionID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: ApiVersionData{
			DisplayName: message.DisplayName,
			Description: message.Description,
			State:       message.State,
		},
	}

	var innerErr error // TODO: remove when ListSpecs accepts a handler that returns errors
	err = core.ListSpecs(ctx, client, versionName.Spec("-"), "", func(message *rpc.ApiSpec) {
		var spec *ApiSpec
		spec, innerErr = newApiSpec(ctx, client, message)
		if innerErr == nil {
			// unset these because they can be inferred
			spec.ApiVersion = ""
			spec.Kind = ""
			version.Data.ApiSpecs = append(version.Data.ApiSpecs, spec)
		}
	})
	if innerErr != nil {
		return nil, innerErr
	}
	return version, err
}

func applyApiVersionPatch(
	ctx context.Context,
	client connection.Client,
	version *ApiVersion,
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
	for _, specPatch := range version.Data.ApiSpecs {
		err := applyApiSpecPatch(ctx, client, specPatch, name)
		if err != nil {
			return err
		}
	}
	return nil
}
