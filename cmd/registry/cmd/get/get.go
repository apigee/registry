// Copyright 2020 Google LLC.
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

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/encoding"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Command() *cobra.Command {
	var filter string
	var output string
	var nested bool

	cmd := &cobra.Command{
		Use:   "get PATTERN",
		Short: "Get resources from the API Registry",
		Long: `Get resources from the API Registry.
The PATTERN of an entity must be begin with "project" or "apis" and must match
one of the following qualitied name patterns:

- apis/NAME
- apis/NAME/artifacts/NAME
- apis/NAME/versions/NAME
- apis/NAME/versions/NAME/specs/NAME[@REVISION]
- apis/NAME/versions/NAME/specs/NAME[@REVISION]/artifacts/NAME
- apis/NAME/deployments/NAME[@REVISION]/artifacts/NAME

In each case, a prefix of "project/NAME/locations/global/" is allowed but
not required, as it is assumed from the configuration.

The /NAME is an optional suffix. If not present, all entities of that
type are returned. A dash (-) may be also be used as a wildcard for
the complete NAME - either within a PATTERN or as a suffix.
The @REVISION is optional on Specs or Deployments. If missing, the latest
revision will be returned. A dash may also be used for REVISION
as a wildcard. Partial values with a wildcard are not supported.

The output from this command is always a single type of entity. By default, 
it generates a simple text-based list of fully-qualified names. 
The "--output yaml" or "--nested" parameters will generate a single top-level
YAML element, either the entity itself or an array named "items" that contains
the entities. In addition, if "--nested" is specified, each returned YAML 
element will recursively include all sub-elements within its YAML.

Examples:

Retrieve the names of all apis:

	registry get apis

Retrieve YAML for the latest revision of all specs named "openapi":

	registry get --output yaml apis/-/versions/-/specs/openapi

Retrieve YAML for all deployment revisions of the "bookstore" api:

	registry get --output yaml apis/bookstore/deployments/-@-
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
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

			// default output
			if output == "" {
				if nested {
					output = "yaml"
				} else {
					output = "name"
				}
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
				PageSize:       1000,
				Filter:         filter,
			}); err != nil {
				if status.Code(err) == codes.NotFound {
					fmt.Fprintln(cmd.ErrOrStderr(), "Not Found")
					return nil
				}
				return err
			}
			// Write any accumulated output.
			err = v.write()
			if err != nil && status.Code(err) == codes.NotFound {
				fmt.Fprintln(cmd.ErrOrStderr(), "Not Found")
				return nil
			}
			return err
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output type (name|yaml|contents)")
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
		case "raw":
			v.results = append(v.results, message)
			return nil
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
		case "raw":
			v.results = append(v.results, message)
			return nil
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
		case "raw":
			v.results = append(v.results, message)
			return nil
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
		case "raw":
			v.results = append(v.results, message)
			return nil
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
		case "raw":
			v.results = append(v.results, message)
			return nil
		case "contents":
			if len(v.results) > 0 {
				return fmt.Errorf("contents can be gotten for at most one spec")
			}
			if err := visitor.FetchSpecContents(ctx, v.registryClient, message); err != nil {
				return err
			}
			v.results = append(v.results, message.GetContents())
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
		case "raw":
			v.results = append(v.results, message)
			return nil
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
		return status.Error(codes.NotFound, "no matching results found")
	}
	if v.output == "yaml" {
		var result interface{}
		if len(v.results) == 1 {
			result = v.results[0]
		} else {
			result = &encoding.List{
				Header: encoding.Header{ApiVersion: encoding.RegistryV1},
				Items:  v.results,
			}
		}
		bytes, err := encoding.EncodeYAML(result)
		if err != nil {
			return err
		}
		_, err = v.writer.Write(bytes)
		return err
	}
	if v.output == "raw" {
		if _, err := v.writer.Write([]byte("[")); err != nil {
			return err
		}
		for i, r := range v.results {
			if i > 0 {
				if _, err := v.writer.Write([]byte(",")); err != nil {
					return err
				}
			}
			b, err := protojson.Marshal(r.(proto.Message))
			if err != nil {
				return err
			}
			if _, err := v.writer.Write(b); err != nil {
				return err
			}
		}
		if _, err := v.writer.Write([]byte("]")); err != nil {
			return err
		}
		return nil
	}
	if v.output == "contents" {
		if len(v.results) == 1 {
			_, err := v.writer.Write(v.results[0].([]byte))
			return err
		}
	}
	return nil
}
