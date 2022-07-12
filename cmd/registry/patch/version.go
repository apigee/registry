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
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/yaml"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

func newApiVersion(ctx context.Context, client *gapic.RegistryClient, message *rpc.ApiVersion) (*yaml.ApiVersion, error) {
	versionName, err := names.ParseVersion(message.Name)
	if err != nil {
		return nil, err
	}

	specs := make([]*yaml.ApiSpec, 0)
	if err = core.ListSpecs(ctx, client, versionName.Spec("-"), "", func(message *rpc.ApiSpec) error {
		spec, err := newApiSpec(message)
		if err != nil {
			return err
		}
		// unset these because they can be inferred
		spec.ApiVersion = ""
		spec.Kind = ""
		specs = append(specs, spec)
		return nil
	}); err != nil {
		return nil, err
	}

	return &yaml.ApiVersion{
		Header: yaml.Header{
			ApiVersion: RegistryV1,
			Kind:       "ApiVersion",
			Metadata: yaml.Metadata{
				Name:        versionName.VersionID,
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: yaml.ApiVersionData{
			DisplayName: message.DisplayName,
			Description: message.Description,
			State:       message.State,
			ApiSpecs:    specs,
		},
	}, nil
}

func applyApiVersionPatch(
	ctx context.Context,
	client connection.RegistryClient,
	version *yaml.ApiVersion,
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
