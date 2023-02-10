// Copyright 2020 Google LLC. All Rights Reserved.
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

package get

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/models"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var filter string
	var output string
	var nested bool

	cmd := &cobra.Command{
		Use:   "get PATTERN",
		Short: "Get resources from the API Registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			pattern := c.FQName(args[0])
			registryClient, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			adminClient, err := connection.NewAdminClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			if nested && output != "yaml" {
				return errors.New("--nested is only supported for yaml output")
			}
			// Create the visitor that will perform gets.
			v := &getVisitor{
				registryClient: registryClient,
				adminClient:    adminClient,
				writer:         cmd.OutOrStdout(),
				output:         output,
				nested:         nested,
			}
			// Visit the selected resources.
			if err = visitor.Visit(ctx, v, visitor.VisitorOptions{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
				Pattern:        pattern,
				Filter:         filter,
			}); err != nil {
				return err
			}
			return v.write()
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().StringVarP(&output, "output", "o", "name", "output type (name|yaml|contents)")
	cmd.Flags().BoolVar(&nested, "nested", false, "include nested subresources in YAML output")
	return cmd
}

type getVisitor struct {
	registryClient connection.RegistryClient
	adminClient    connection.AdminClient
	writer         io.Writer
	output         string
	nested         bool
	results        []interface{} // result values to be returned in a single message
}

func (v *getVisitor) ProjectHandler() visitor.ProjectHandler {
	return func(ctx context.Context, message *rpc.Project) error {
		switch v.output {
		case "name":
			v.results = append(v.results, message.Name)
			_, err := v.writer.Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			project, err := patch.NewProject(ctx, v.registryClient, message)
			if err != nil {
				return err
			}
			v.results = append(v.results, project)
			return nil
		default:
			return newOutputTypeError("projects", v.output)
		}
	}
}

func (v *getVisitor) ApiHandler() visitor.ApiHandler {
	return func(ctx context.Context, message *rpc.Api) error {
		switch v.output {
		case "name":
			v.results = append(v.results, message.Name)
			_, err := v.writer.Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			api, err := patch.NewApi(ctx, v.registryClient, message, v.nested)
			if err != nil {
				return err
			}
			v.results = append(v.results, api)
			return nil
		default:
			return newOutputTypeError("apis", v.output)
		}
	}
}

func (v *getVisitor) VersionHandler() visitor.VersionHandler {
	return func(ctx context.Context, message *rpc.ApiVersion) error {
		switch v.output {
		case "name":
			v.results = append(v.results, message.Name)
			_, err := v.writer.Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			version, err := patch.NewApiVersion(ctx, v.registryClient, message, v.nested)
			if err != nil {
				return err
			}
			v.results = append(v.results, version)
			return nil
		default:
			return newOutputTypeError("versions", v.output)
		}
	}
}

func (v *getVisitor) DeploymentHandler() visitor.DeploymentHandler {
	return func(ctx context.Context, message *rpc.ApiDeployment) error {
		switch v.output {
		case "name":
			v.results = append(v.results, message.Name)
			_, err := v.writer.Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			deployment, err := patch.NewApiDeployment(ctx, v.registryClient, message, v.nested)
			if err != nil {
				return err
			}
			v.results = append(v.results, deployment)
			return nil
		default:
			return newOutputTypeError("deployments", v.output)
		}
	}
}

func (v *getVisitor) DeploymentRevisionHandler() visitor.DeploymentHandler {
	return v.DeploymentHandler()
}

func (v *getVisitor) SpecHandler() visitor.SpecHandler {
	return func(ctx context.Context, message *rpc.ApiSpec) error {
		switch v.output {
		case "name":
			v.results = append(v.results, message.Name)
			_, err := v.writer.Write([]byte(message.Name + "\n"))
			return err
		case "contents":
			if len(v.results) > 0 {
				return fmt.Errorf("contents can be gotten for at most one spec")
			}
			if err := visitor.FetchSpecContents(ctx, v.registryClient, message); err != nil {
				return err
			}
			contents := message.GetContents()
			if strings.Contains(message.GetMimeType(), "+gzip") {
				contents, _ = compress.GUnzippedBytes(contents)
			}
			v.results = append(v.results, contents)
			return nil
		case "yaml":
			spec, err := patch.NewApiSpec(ctx, v.registryClient, message, v.nested)
			if err != nil {
				return err
			}
			v.results = append(v.results, spec)
			return nil
		default:
			return newOutputTypeError("specs", v.output)
		}
	}
}

func (v *getVisitor) SpecRevisionHandler() visitor.SpecHandler {
	return v.SpecHandler()
}

func (v *getVisitor) ArtifactHandler() visitor.ArtifactHandler {
	return func(ctx context.Context, message *rpc.Artifact) error {
		switch v.output {
		case "name":
			v.results = append(v.results, message.Name)
			_, err := v.writer.Write([]byte(message.Name + "\n"))
			return err
		case "contents":
			if len(v.results) > 0 {
				return fmt.Errorf("contents can be gotten for at most one artifact")
			}
			if err := visitor.FetchArtifactContents(ctx, v.registryClient, message); err != nil {
				return err
			}
			v.results = append(v.results, message.GetContents())
			return nil
		case "yaml":
			if err := visitor.FetchArtifactContents(ctx, v.registryClient, message); err != nil {
				return err
			}
			artifact, err := patch.NewArtifact(ctx, v.registryClient, message)
			if err != nil {
				return err
			}
			v.results = append(v.results, artifact)
			return nil
		default:
			return newOutputTypeError("artifacts", v.output)
		}
	}
}

func newOutputTypeError(resourceType, outputType string) error {
	return fmt.Errorf("%s do not support the %q output type", resourceType, outputType)
}

func (v *getVisitor) write() error {
	if len(v.results) == 0 {
		return fmt.Errorf("no matching results found")
	}
	if v.output == "yaml" {
		var result interface{}
		if len(v.results) == 1 {
			result = v.results[0]
		} else {
			result = &models.List{
				Header: models.Header{ApiVersion: patch.RegistryV1},
				Items:  v.results,
			}
		}
		bytes, err := patch.Encode(result)
		if err != nil {
			return err
		}
		_, err = v.writer.Write(bytes)
		return err
	}
	if v.output == "contents" {
		if len(v.results) == 1 {
			_, err := v.writer.Write(v.results[0].([]byte))
			return err
		}
	}
	return nil
}
