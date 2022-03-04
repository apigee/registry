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

package compute

import (
	"context"

	"github.com/apigee/registry/cmd/registry/conformance"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

const styleguideFilter = "mime_type.contains('google.cloud.apigeeregistry.applications.v1alpha1.StyleGuide')"

func conformanceCommand(ctx context.Context) *cobra.Command {
	var filter string

	cmd := &cobra.Command{
		Use:   "conformance",
		Short: "Compute lint results for API specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			name, err := names.ParseSpec(args[0])
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Invalid Argument: must specify one or more API specs")
			}

			specs := make([]*rpc.ApiSpec, 0)
			if err := core.ListSpecs(ctx, client, name, filter, func(spec *rpc.ApiSpec) {
				specs = append(specs, spec)
			}); err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list specs")
			}

			guides := make([]*rpc.StyleGuide, 0)
			if err := core.ListArtifacts(ctx, client, name.Project().Artifact("-"), styleguideFilter, true, func(artifact *rpc.Artifact) {
				guide := new(rpc.StyleGuide)
				if err := proto.Unmarshal(artifact.GetContents(), guide); err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Unmarshal() to StyleGuide failed on artifact: %s", artifact.GetName())
					return
				}
				guides = append(guides, guide)
			}); err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list styleguide artifacts")
			}

			for _, guide := range guides {
				log.Debugf(ctx, "Processing styleguide: %s", guide.GetId())
				processStyleGuide(ctx, client, guide, specs)
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	return cmd
}

// processStyleGuide computes and attaches conformance reports as
// artifacts to a spec or a collection of specs.
func processStyleGuide(ctx context.Context, client connection.Client, styleguide *rpc.StyleGuide, specs []*rpc.ApiSpec) {
	linterNameToMetadata, err := conformance.GenerateLinterMetadata(styleguide)
	if err != nil {
		log.Errorf(ctx, "Failed generating linter metadata, check styleguide definition, Error: %s", err)
		return
	}

	taskQueue, wait := core.WorkerPool(ctx, 16)
	defer wait()

	for _, spec := range specs {
		for _, supportedType := range styleguide.GetMimeTypes() {
			if supportedType != spec.GetMimeType() {
				continue // Only compute matching style guides.
			}

			taskQueue <- &conformance.ComputeConformanceTask{
				Client:          client,
				Spec:            spec,
				LintersMetadata: linterNameToMetadata,
				StyleguideId:    styleguide.GetId(),
			}
		}
	}
}
