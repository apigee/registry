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

package apply

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func Command(ctx context.Context) *cobra.Command {
	var fileName string
	var parent string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply patches that add content to the API Registry",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			bytes, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to read file")
			}

			// get the id and kind of artifact from the YAML elements common to all artifacts
			var header patch.Header
			err = yaml.Unmarshal(bytes, &header)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
			}
			log.FromContext(ctx).Infof("%+v", header)
			if header.APIVersion != "registry/v1" {
				log.FromContext(ctx).Fatalf("Unsupported API version: %s", header.APIVersion)
			}
			if header.Kind == "API" {
				var api patch.API
				err = yaml.Unmarshal(bytes, &api)
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to parse YAML")
				}
				log.FromContext(ctx).Infof("%+v", api)

				err = applyApiPatch(ctx, client, &api, parent)
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to apply patch")
				}
			} else {
				log.FromContext(ctx).Fatalf("Unsupported kind: %s", header.Kind)
			}
		},
	}
	cmd.Flags().StringVarP(&fileName, "file", "f", "", "File containing the patch to apply")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent resource for the patch")
	return cmd
}

func applyApiPatch(
	ctx context.Context,
	client connection.Client,
	api *patch.API,
	parent string) error {
	name := fmt.Sprintf("%s/apis/%s", parent, api.Metadata.Name)
	req := &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:        name,
			DisplayName: api.Body.DisplayName,
			Description: api.Body.Description,
		},
		AllowMissing: true,
	}
	_, err := client.UpdateApi(ctx, req)
	if err != nil {
		return err
	}
	for _, versionPatch := range api.Body.APIVersions {
		err := applyApiVersionPatch(ctx, client, versionPatch, name)
		if err != nil {
			return err
		}
	}
	for _, deploymentPatch := range api.Body.APIDeployments {
		err := applyApiDeploymentPatch(ctx, client, deploymentPatch, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyApiVersionPatch(
	ctx context.Context,
	client connection.Client,
	version *patch.APIVersion,
	parent string) error {
	name := fmt.Sprintf("%s/versions/%s", parent, version.Metadata.Name)
	req := &rpc.UpdateApiVersionRequest{
		ApiVersion: &rpc.ApiVersion{
			Name:        name,
			DisplayName: version.Body.DisplayName,
			Description: version.Body.Description,
		},
		AllowMissing: true,
	}
	_, err := client.UpdateApiVersion(ctx, req)
	if err != nil {
		return err
	}
	for _, specPatch := range version.Body.APISpecs {
		err := applyApiSpecPatch(ctx, client, specPatch, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyApiSpecPatch(
	ctx context.Context,
	client connection.Client,
	spec *patch.APISpec,
	parent string) error {
	name := fmt.Sprintf("%s/specs/%s", parent, spec.Metadata.Name)
	req := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:        name,
			Filename:    spec.Body.FileName,
			Description: spec.Body.Description,
		},
		AllowMissing: true,
	}
	_, err := client.UpdateApiSpec(ctx, req)
	return err
}

func applyApiDeploymentPatch(
	ctx context.Context,
	client connection.Client,
	deployment *patch.APIDeployment,
	parent string) error {
	name := fmt.Sprintf("%s/deployments/%s", parent, deployment.Metadata.Name)
	req := &rpc.UpdateApiDeploymentRequest{
		ApiDeployment: &rpc.ApiDeployment{
			Name:        name,
			DisplayName: deployment.Body.DisplayName,
			Description: deployment.Body.Description,
		},
		AllowMissing: true,
	}
	_, err := client.UpdateApiDeployment(ctx, req)
	return err
}
