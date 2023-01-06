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
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/cmd/registry/types"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/models"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Command() *cobra.Command {
	var getContents bool
	var getRawContents bool
	var getPrintedContents bool
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
			if getContents && getRawContents || getContents && getPrintedContents || getRawContents && getPrintedContents {
				log.FromContext(ctx).Fatal("Please use at most one of --print, --raw, and --contents.")
			}
			if getContents {
				getPrintedContents = true
			}
			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			adminClient, err := connection.NewAdminClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			h := &GetHandler{
				cmd:                cmd,
				ctx:                ctx,
				client:             client,
				adminClient:        adminClient,
				name:               args[0],
				filter:             filter,
				getPrintedContents: getPrintedContents,
				getRawContents:     getRawContents,
				output:             output,
			}
			err = h.run()
			if err != nil {
				return err
			}
			if len(h.results) == 1 {
				bytes, err := patch.Encode(h.results[0])
				if err != nil {
					return err
				}
				_, err = h.cmd.OutOrStdout().Write(bytes)
				return err
			} else if len(h.results) > 1 {
				list := &models.List{
					Header: models.Header{ApiVersion: patch.RegistryV1},
					Items:  h.results,
				}
				bytes, err := patch.Encode(list)
				if err != nil {
					return err
				}
				_, err = h.cmd.OutOrStdout().Write(bytes)
				return err
			} else {
				return nil
			}
		},
	}

	cmd.Flags().BoolVar(&getContents, "contents", false, "Get resource contents if available")
	cmd.Flags().BoolVar(&getRawContents, "raw", false, "Get raw resource contents if available")
	cmd.Flags().BoolVar(&getPrintedContents, "print", false, "Print resource contents if available")
	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	cmd.Flags().StringVarP(&output, "output", "o", "names", "Output type (names, yaml)")
	return cmd
}

type GetHandler struct {
	cmd                *cobra.Command
	ctx                context.Context
	client             connection.RegistryClient
	adminClient        connection.AdminClient
	name               string
	filter             string
	getPrintedContents bool
	getRawContents     bool
	output             string
	results            []interface{}
}

func (h *GetHandler) run() error {
	// Define aliases to simplify the subsequent code.
	name := h.name
	ctx := h.ctx
	client := h.client
	adminClient := h.adminClient
	filter := h.filter

	// Initialize a slice of results.
	h.results = make([]interface{}, 0)

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

func (h *GetHandler) projectHandler() func(message *rpc.Project) error {
	return func(message *rpc.Project) error {
		switch h.output {
		case "names":
			_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			project, err := patch.NewProject(h.ctx, h.client, message)
			if err != nil {
				return err
			}
			h.results = append(h.results, project)
			return nil
		default:
			return fmt.Errorf("invalid output type %s", h.output)
		}
	}
}

func (h *GetHandler) apiHandler() func(message *rpc.Api) error {
	return func(message *rpc.Api) error {
		switch h.output {
		case "names":
			_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			api, err := patch.NewApi(h.ctx, h.client, message, false)
			if err != nil {
				return err
			}
			h.results = append(h.results, api)
			return nil
		default:
			return fmt.Errorf("invalid output type %s", h.output)
		}
	}
}

func (h *GetHandler) apiVersionHandler() func(message *rpc.ApiVersion) error {
	return func(message *rpc.ApiVersion) error {
		switch h.output {
		case "names":
			_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			version, err := patch.NewApiVersion(h.ctx, h.client, message, false)
			if err != nil {
				return err
			}
			h.results = append(h.results, version)
			return nil
		default:
			return fmt.Errorf("invalid output type %s", h.output)
		}
	}
}

func (h *GetHandler) apiDeploymentHandler() func(message *rpc.ApiDeployment) error {
	return func(message *rpc.ApiDeployment) error {
		switch h.output {
		case "names":
			_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			deployment, err := patch.NewApiDeployment(h.ctx, h.client, message, false)
			if err != nil {
				return err
			}
			h.results = append(h.results, deployment)
			return nil
		default:
			return fmt.Errorf("invalid output type %s", h.output)
		}
	}
}

func (h *GetHandler) apiSpecHandler() func(message *rpc.ApiSpec) error {
	return func(message *rpc.ApiSpec) error {
		// for specs, these options are synonymous
		if h.getPrintedContents || h.getRawContents {
			name, err := names.ParseSpecRevision(message.Name)
			if err != nil {
				return err
			}
			if err = core.GetSpecRevision(h.ctx, h.client, name, true, func(spec *rpc.ApiSpec) error {
				message = spec
				return nil
			}); err != nil {
				return err
			}
			contents := message.GetContents()
			if strings.Contains(message.GetMimeType(), "+gzip") {
				contents, _ = core.GUnzippedBytes(contents)
			}
			_, err = h.cmd.OutOrStdout().Write(contents)
			return err
		}
		switch h.output {
		case "names":
			_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			spec, err := patch.NewApiSpec(h.ctx, h.client, message, false)
			if err != nil {
				return err
			}
			h.results = append(h.results, spec)
			return nil
		default:
			return fmt.Errorf("invalid output type %s", h.output)
		}
	}
}

func (h *GetHandler) artifactHandler() func(message *rpc.Artifact) error {
	return func(message *rpc.Artifact) error {
		if h.getPrintedContents || h.getRawContents || h.output == "yaml" {
			name, err := names.ParseArtifact(message.Name)
			if err != nil {
				return err
			}
			if err = core.GetArtifact(h.ctx, h.client, name, true, func(artifact *rpc.Artifact) error {
				message = artifact
				return nil
			}); err != nil {
				return err
			}
		}
		if h.getPrintedContents {
			if types.IsPrintableType(message.GetMimeType()) {
				fmt.Fprintf(h.cmd.OutOrStdout(), "%s\n", string(message.GetContents()))
				return nil
			}
			message, err := getArtifactMessageContents(message)
			if err != nil {
				return err
			}
			fmt.Fprintf(h.cmd.OutOrStdout(), "%s\n", protojson.Format(message))
			return nil
		} else if h.getRawContents {
			_, err := h.cmd.OutOrStdout().Write(message.GetContents())
			return err
		}
		switch h.output {
		case "names":
			_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
			return err
		case "yaml":
			artifact, err := patch.NewArtifact(h.ctx, h.client, message)
			if err != nil {
				return err
			}
			h.results = append(h.results, artifact)
			return nil
		default:
			return fmt.Errorf("invalid output type %s", h.output)
		}
	}
}

func getArtifactMessageContents(artifact *rpc.Artifact) (proto.Message, error) {
	message, err := types.MessageForMimeType(artifact.GetMimeType())
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(artifact.GetContents(), message); err != nil {
		return nil, err
	}
	return message, nil
}
