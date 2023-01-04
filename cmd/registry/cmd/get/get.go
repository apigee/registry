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
			}
			return h.run()
		},
	}

	cmd.Flags().BoolVar(&getContents, "contents", false, "Get resource contents if available")
	cmd.Flags().BoolVar(&getRawContents, "raw", false, "Get raw resource contents if available")
	cmd.Flags().BoolVar(&getPrintedContents, "print", false, "Print resource contents if available")
	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
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
}

func (h *GetHandler) run() error {
	// Define aliases to simplify the subsequent code.
	name := h.name
	ctx := h.ctx
	client := h.client
	adminClient := h.adminClient
	filter := h.filter

	// First try to match collection names.
	if project, err := names.ParseProjectCollection(name); err == nil {
		return core.ListProjects(ctx, adminClient, project, filter, h.printProjectName())
	} else if api, err := names.ParseApiCollection(name); err == nil {
		return core.ListAPIs(ctx, client, api, filter, h.printApiName())
	} else if deployment, err := names.ParseDeploymentCollection(name); err == nil {
		return core.ListDeployments(ctx, client, deployment, filter, h.printApiDeploymentName())
	} else if rev, err := names.ParseDeploymentRevisionCollection(name); err == nil {
		return core.ListDeploymentRevisions(ctx, client, rev, filter, h.printApiDeploymentName())
	} else if version, err := names.ParseVersionCollection(name); err == nil {
		return core.ListVersions(ctx, client, version, filter, h.printApiVersionName())
	} else if spec, err := names.ParseSpecCollection(name); err == nil {
		return core.ListSpecs(ctx, client, spec, filter, h.printApiSpecName())
	} else if rev, err := names.ParseSpecRevisionCollection(name); err == nil {
		return core.ListSpecRevisions(ctx, client, rev, filter, h.printApiSpecName())
	} else if artifact, err := names.ParseArtifactCollection(name); err == nil {
		return core.ListArtifacts(ctx, client, artifact, filter, false, h.printArtifactName())
	}

	// Then try to match resource names containing wildcards, these also are treated as collections.
	if strings.Contains(name, "/-") || strings.Contains(name, "@-") {
		if project, err := names.ParseProject(name); err == nil {
			return core.ListProjects(ctx, adminClient, project, filter, h.printProjectName())
		} else if api, err := names.ParseApi(name); err == nil {
			return core.ListAPIs(ctx, client, api, filter, h.printApiName())
		} else if deployment, err := names.ParseDeployment(name); err == nil {
			return core.ListDeployments(ctx, client, deployment, filter, h.printApiDeploymentName())
		} else if rev, err := names.ParseDeploymentRevision(name); err == nil {
			return core.ListDeploymentRevisions(ctx, client, rev, filter, h.printApiDeploymentName())
		} else if version, err := names.ParseVersion(name); err == nil {
			return core.ListVersions(ctx, client, version, filter, h.printApiVersionName())
		} else if spec, err := names.ParseSpec(name); err == nil {
			return core.ListSpecs(ctx, client, spec, filter, h.printApiSpecName())
		} else if rev, err := names.ParseSpecRevision(name); err == nil {
			return core.ListSpecRevisions(ctx, client, rev, filter, h.printApiSpecName())
		} else if artifact, err := names.ParseArtifact(name); err == nil {
			return core.ListArtifacts(ctx, client, artifact, filter, false, h.printArtifactName())
		}
		return fmt.Errorf("unsupported entity %+v", name)
	}

	// If we get here, name designates an individual resource to be displayed.
	// So if a filter was specified, that's an error.
	if filter != "" {
		return errors.New("--filter must not be specified for a non-collection resource")
	}

	if project, err := names.ParseProject(name); err == nil {
		return core.GetProject(ctx, adminClient, project, h.printProjectDetail())
	} else if api, err := names.ParseApi(name); err == nil {
		return core.GetAPI(ctx, client, api, h.printApiDetail())
	} else if deployment, err := names.ParseDeployment(name); err == nil {
		return core.GetDeployment(ctx, client, deployment, h.printApiDeploymentDetail())
	} else if deployment, err := names.ParseDeploymentRevision(name); err == nil {
		return core.GetDeploymentRevision(ctx, client, deployment, h.printApiDeploymentDetail())
	} else if version, err := names.ParseVersion(name); err == nil {
		return core.GetVersion(ctx, client, version, h.printApiVersionDetail())
	} else if spec, err := names.ParseSpec(name); err == nil {
		// for specs, these options are synonymous
		if h.getPrintedContents || h.getRawContents {
			return core.GetSpec(ctx, client, spec, true, h.writeSpecContents())
		} else {
			return core.GetSpec(ctx, client, spec, false, h.printApiSpecDetail())
		}
	} else if spec, err := names.ParseSpecRevision(name); err == nil {
		// for specs, these options are synonymous
		if h.getPrintedContents || h.getRawContents {
			return core.GetSpecRevision(ctx, client, spec, true, h.writeSpecContents())
		} else {
			return core.GetSpecRevision(ctx, client, spec, false, h.printApiSpecDetail())
		}
	} else if artifact, err := names.ParseArtifact(name); err == nil {
		if h.getPrintedContents {
			return core.GetArtifact(ctx, client, artifact, true, h.printArtifactContents())
		} else if h.getRawContents {
			return core.GetArtifact(ctx, client, artifact, true, h.writeArtifactContents())
		} else {
			return core.GetArtifact(ctx, client, artifact, false, h.printArtifactDetail())
		}
	} else {
		return fmt.Errorf("unsupported entity %+v", name)
	}
}

func (h *GetHandler) printProjectName() func(message *rpc.Project) error {
	return func(message *rpc.Project) error {
		_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
		return err
	}
}

func (h *GetHandler) printApiName() func(message *rpc.Api) error {
	return func(message *rpc.Api) error {
		_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
		return err
	}
}

func (h *GetHandler) printApiVersionName() func(message *rpc.ApiVersion) error {
	return func(message *rpc.ApiVersion) error {
		_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
		return err
	}
}

func (h *GetHandler) printApiDeploymentName() func(message *rpc.ApiDeployment) error {
	return func(message *rpc.ApiDeployment) error {
		_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
		return err
	}
}

func (h *GetHandler) printApiSpecName() func(message *rpc.ApiSpec) error {
	return func(message *rpc.ApiSpec) error {
		_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
		return err
	}
}

func (h *GetHandler) printArtifactName() func(message *rpc.Artifact) error {
	return func(message *rpc.Artifact) error {
		_, err := h.cmd.OutOrStdout().Write([]byte(message.Name + "\n"))
		return err
	}
}

func (h *GetHandler) printProjectDetail() func(message *rpc.Project) error {
	return func(message *rpc.Project) error {
		bytes, _, err := patch.PatchForProject(h.ctx, h.client, message)
		if err != nil {
			return err
		}
		_, err = h.cmd.OutOrStdout().Write(bytes)
		return err
	}
}

func (h *GetHandler) printApiDetail() func(message *rpc.Api) error {
	return func(message *rpc.Api) error {
		bytes, _, err := patch.PatchForApi(h.ctx, h.client, message, false)
		if err != nil {
			return err
		}
		_, err = h.cmd.OutOrStdout().Write(bytes)
		return err
	}
}

func (h *GetHandler) printApiVersionDetail() func(message *rpc.ApiVersion) error {
	return func(message *rpc.ApiVersion) error {
		bytes, _, err := patch.PatchForApiVersion(h.ctx, h.client, message, false)
		if err != nil {
			return err
		}
		_, err = h.cmd.OutOrStdout().Write(bytes)
		return err
	}
}

func (h *GetHandler) printApiDeploymentDetail() func(message *rpc.ApiDeployment) error {
	return func(message *rpc.ApiDeployment) error {
		bytes, _, err := patch.PatchForApiDeployment(h.ctx, h.client, message, false)
		if err != nil {
			return err
		}
		_, err = h.cmd.OutOrStdout().Write(bytes)
		return err
	}
}

func (h *GetHandler) printApiSpecDetail() func(message *rpc.ApiSpec) error {
	return func(message *rpc.ApiSpec) error {
		bytes, _, err := patch.PatchForApiSpec(h.ctx, h.client, message, false)
		if err != nil {
			return err
		}
		_, err = h.cmd.OutOrStdout().Write(bytes)
		return err
	}
}

func (h *GetHandler) printArtifactDetail() func(message *rpc.Artifact) error {
	return func(message *rpc.Artifact) error {
		bytes, _, err := patch.PatchForArtifact(h.ctx, h.client, message)
		if err != nil {
			return err
		}
		_, err = h.cmd.OutOrStdout().Write(bytes)
		return err
	}
}

func (h *GetHandler) writeSpecContents() func(message *rpc.ApiSpec) error {
	return func(message *rpc.ApiSpec) error {
		contents := message.GetContents()
		if strings.Contains(message.GetMimeType(), "+gzip") {
			contents, _ = core.GUnzippedBytes(contents)
		}
		_, err := h.cmd.OutOrStdout().Write(contents)
		return err
	}
}

func (h *GetHandler) printArtifactContents() func(artifact *rpc.Artifact) error {
	return func(artifact *rpc.Artifact) error {
		if types.IsPrintableType(artifact.GetMimeType()) {
			fmt.Fprintf(h.cmd.OutOrStdout(), "%s\n", string(artifact.GetContents()))
			return nil
		}
		message, err := getArtifactMessageContents(artifact)
		if err != nil {
			return err
		}
		fmt.Fprintf(h.cmd.OutOrStdout(), "%s\n", protojson.Format(message))
		return nil
	}
}

func (h *GetHandler) writeArtifactContents() func(artifact *rpc.Artifact) error {
	return func(artifact *rpc.Artifact) error {
		_, err := h.cmd.OutOrStdout().Write(artifact.GetContents())
		return err
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
