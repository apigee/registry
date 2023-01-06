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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/models"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var filter string
	var output string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get resources from the API Registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			args[0] = c.FQName(args[0])
			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			adminClient, err := connection.NewAdminClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			h := &getHandler{
				ctx:         ctx,
				client:      client,
				adminClient: adminClient,
				writer:      cmd.OutOrStdout(),
				name:        args[0],
				filter:      filter,
				output:      output,
			}
			err = h.traverse()
			if err != nil {
				return err
			}
			return h.write()
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	cmd.Flags().StringVarP(&output, "output", "o", "names", "Output type (names, yaml, contents)")
	return cmd
}

type getHandler struct {
	ctx         context.Context
	client      connection.RegistryClient
	adminClient connection.AdminClient
	writer      io.Writer
	name        string
	filter      string
	output      string
	results     []interface{} // result values to be returned in a single message
}

func (h *getHandler) traverse() error {
	// Define aliases to simplify the subsequent code.
	name := h.name
	ctx := h.ctx
	client := h.client
	adminClient := h.adminClient
	filter := h.filter

	// First try to match collection names.
	if project, err := names.ParseProjectCollection(name); err == nil {
		return core.ListProjects(ctx, adminClient, project, filter, h.projectHandler())
	} else if api, err := names.ParseApiCollection(name); err == nil {
		return core.ListAPIs(ctx, client, api, filter, h.apiHandler())
	} else if deployment, err := names.ParseDeploymentCollection(name); err == nil {
		return core.ListDeployments(ctx, client, deployment, filter, h.apiDeploymentHandler())
	} else if rev, err := names.ParseDeploymentRevisionCollection(name); err == nil {
		return core.ListDeploymentRevisions(ctx, client, rev, filter, h.apiDeploymentHandler())
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		return core.ListVersions(ctx, client, version, filter, h.apiVersionHandler())
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		return core.ListSpecs(ctx, client, spec, filter, false, h.apiSpecHandler())
	} else if rev, err := names.ParseSpecRevisionCollection(name); err == nil {
		return core.ListSpecRevisions(ctx, client, rev, filter, false, h.apiSpecHandler())
	} else if artifact, err := names.ParseArtifactCollection(name); err == nil {
		return core.ListArtifacts(ctx, client, artifact, filter, false, h.artifactHandler())
	}

	// Then try to match resource names containing wildcards, these also are treated as collections.
	if strings.Contains(name, "/-") || strings.Contains(name, "@-") {
		if project, err := names.ParseProject(name); err == nil {
			return core.ListProjects(ctx, adminClient, project, filter, h.projectHandler())
		} else if api, err := names.ParseApi(name); err == nil {
			return core.ListAPIs(ctx, client, api, filter, h.apiHandler())
		} else if deployment, err := names.ParseDeployment(name); err == nil {
			return core.ListDeployments(ctx, client, deployment, filter, h.apiDeploymentHandler())
		} else if rev, err := names.ParseDeploymentRevision(name); err == nil {
			return core.ListDeploymentRevisions(ctx, client, rev, filter, h.apiDeploymentHandler())
		} else if version, err := names.ParseVersion(name); err == nil {
			return core.ListVersions(ctx, client, version, filter, h.apiVersionHandler())
		} else if spec, err := names.ParseSpec(name); err == nil {
			return core.ListSpecs(ctx, client, spec, filter, false, h.apiSpecHandler())
		} else if rev, err := names.ParseSpecRevision(name); err == nil {
			return core.ListSpecRevisions(ctx, client, rev, filter, false, h.apiSpecHandler())
		} else if artifact, err := names.ParseArtifact(name); err == nil {
			return core.ListArtifacts(ctx, client, artifact, filter, false, h.artifactHandler())
		}
		return fmt.Errorf("unsupported entity %+v", name)
	}

	// If we get here, name designates an individual resource to be displayed.
	// So if a filter was specified, that's an error.
	if filter != "" {
		return errors.New("--filter must not be specified for a non-collection resource")
	}

	if project, err := names.ParseProject(name); err == nil {
		return core.GetProject(ctx, adminClient, project, h.projectHandler())
	} else if api, err := names.ParseApi(name); err == nil {
		return core.GetAPI(ctx, client, api, h.apiHandler())
	} else if deployment, err := names.ParseDeployment(name); err == nil {
		return core.GetDeployment(ctx, client, deployment, h.apiDeploymentHandler())
	} else if deployment, err := names.ParseDeploymentRevision(name); err == nil {
		return core.GetDeploymentRevision(ctx, client, deployment, h.apiDeploymentHandler())
	} else if version, err := names.ParseVersion(name); err == nil {
		return core.GetVersion(ctx, client, version, h.apiVersionHandler())
	} else if spec, err := names.ParseSpec(name); err == nil {
		return core.GetSpec(ctx, client, spec, false, h.apiSpecHandler())
	} else if spec, err := names.ParseSpecRevision(name); err == nil {
		return core.GetSpecRevision(ctx, client, spec, false, h.apiSpecHandler())
	} else if artifact, err := names.ParseArtifact(name); err == nil {
		return core.GetArtifact(ctx, client, artifact, false, h.artifactHandler())
	} else {
		return fmt.Errorf("unsupported entity %+v", name)
	}
}

func (h *getHandler) projectHandler() func(message *rpc.Project) error {
	return func(message *rpc.Project) error {
		switch h.output {
		case "names":
			h.results = append(h.results, message.Name)
			_, err := h.writer.Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			project, err := patch.NewProject(h.ctx, h.client, message)
			if err != nil {
				return err
			}
			h.results = append(h.results, project)
			return nil
		default:
			return newOutputTypeError("projects", h.output)
		}
	}
}

func (h *getHandler) apiHandler() func(message *rpc.Api) error {
	return func(message *rpc.Api) error {
		switch h.output {
		case "names":
			h.results = append(h.results, message.Name)
			_, err := h.writer.Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			api, err := patch.NewApi(h.ctx, h.client, message, false)
			if err != nil {
				return err
			}
			h.results = append(h.results, api)
			return nil
		default:
			return newOutputTypeError("apis", h.output)
		}
	}
}

func (h *getHandler) apiVersionHandler() func(message *rpc.ApiVersion) error {
	return func(message *rpc.ApiVersion) error {
		switch h.output {
		case "names":
			h.results = append(h.results, message.Name)
			_, err := h.writer.Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			version, err := patch.NewApiVersion(h.ctx, h.client, message, false)
			if err != nil {
				return err
			}
			h.results = append(h.results, version)
			return nil
		default:
			return newOutputTypeError("versions", h.output)
		}
	}
}

func (h *getHandler) apiDeploymentHandler() func(message *rpc.ApiDeployment) error {
	return func(message *rpc.ApiDeployment) error {
		switch h.output {
		case "names":
			h.results = append(h.results, message.Name)
			_, err := h.writer.Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			deployment, err := patch.NewApiDeployment(h.ctx, h.client, message, false)
			if err != nil {
				return err
			}
			h.results = append(h.results, deployment)
			return nil
		default:
			return newOutputTypeError("deployments", h.output)
		}
	}
}

func (h *getHandler) apiSpecHandler() func(message *rpc.ApiSpec) error {
	return func(message *rpc.ApiSpec) error {
		switch h.output {
		case "names":
			h.results = append(h.results, message.Name)
			_, err := h.writer.Write([]byte(message.Name + "\n"))
			return err
		case "contents":
			if len(h.results) > 0 {
				return fmt.Errorf("contents can be gotten for at most one spec")
			}
			if err := core.FetchSpecContents(h.ctx, h.client, message); err != nil {
				return err
			}
			contents := message.GetContents()
			if strings.Contains(message.GetMimeType(), "+gzip") {
				contents, _ = core.GUnzippedBytes(contents)
			}
			h.results = append(h.results, contents)
			return nil
		case "yaml":
			spec, err := patch.NewApiSpec(h.ctx, h.client, message, false)
			if err != nil {
				return err
			}
			h.results = append(h.results, spec)
			return nil
		default:
			return newOutputTypeError("specs", h.output)
		}
	}
}

func (h *getHandler) artifactHandler() func(message *rpc.Artifact) error {
	return func(message *rpc.Artifact) error {
		switch h.output {
		case "names":
			h.results = append(h.results, message.Name)
			_, err := h.writer.Write([]byte(message.Name + "\n"))
			return err
		case "contents":
			if len(h.results) > 0 {
				return fmt.Errorf("contents can be gotten for at most one artifact")
			}
			if err := core.FetchArtifactContents(h.ctx, h.client, message); err != nil {
				return err
			}
			h.results = append(h.results, message.GetContents())
			return nil
		case "yaml":
			if err := core.FetchArtifactContents(h.ctx, h.client, message); err != nil {
				return err
			}
			artifact, err := patch.NewArtifact(h.ctx, h.client, message)
			if err != nil {
				return err
			}
			h.results = append(h.results, artifact)
			return nil
		default:
			return newOutputTypeError("artifacts", h.output)
		}
	}
}

func newOutputTypeError(resourceType, outputType string) error {
	return fmt.Errorf("%s do not support the %q output type", resourceType, outputType)
}

func (h *getHandler) write() error {
	if len(h.results) == 0 {
		return fmt.Errorf("no matching results found")
	}
	if h.output == "yaml" {
		var result interface{}
		if len(h.results) == 1 {
			result = h.results[0]
		} else {
			result = &models.List{
				Header: models.Header{ApiVersion: patch.RegistryV1},
				Items:  h.results,
			}
		}
		bytes, err := patch.Encode(result)
		if err != nil {
			return err
		}
		_, err = h.writer.Write(bytes)
		return err
	}
	if h.output == "contents" {
		if len(h.results) == 1 {
			_, err := h.writer.Write(h.results[0].([]byte))
			return err
		}
	}
	return nil
}
