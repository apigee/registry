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
	"fmt"

	"github.com/apigee/registry/cmd/registry/conformance"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/conformance"
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

			name := args[0]

			// Ensure that the provided argument is a spec.
			specName, err := names.ParseSpec(name)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatalf("The provided argument %s does not match the regex of a spec", name)
			}

			// List all the styleGuide artifacts in the registry
			artifactName, err := names.ParseArtifact(fmt.Sprintf("projects/%s/locations/%s/artifacts/-", specName.ProjectID, names.Location))
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatalf("Invalid project %q", specName.ProjectID)
			}

			err = core.ListArtifacts(ctx, client, artifactName, styleguideFilter, true, func(artifact *rpc.Artifact) {

				// Unmarshal the contents of the artifact into a style guide
				styleguide := &rpc.StyleGuide{}
				err = proto.Unmarshal(artifact.GetContents(), styleguide)
				if err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Unmarshal() to StyleGuide failed on artifact: %s", artifact.GetName())
					return
				}

				log.Debugf(ctx, "Processing styleguide: %s", styleguide.GetId())

				processStyleGuide(ctx, client, styleguide, specName, filter)
			})

			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list styleguide artifacts")
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	return cmd
}

// processStyleGuide computes and attaches conformance reports as
// artifacts to a spec or a collection of specs.
func processStyleGuide(ctx context.Context,
	client connection.Client,
	styleguide *rpc.StyleGuide,
	spec names.Spec,
	filter string) {

	linterNameToMetadata, err := conformance.GenerateLinterMetadata(styleguide)
	if err != nil {
		log.Errorf(ctx, "Failed generating linter metadata, check styleguide definition, Error: %s", err)
		return
	}

	// Initialize task queue.
	taskQueue, wait := core.WorkerPool(ctx, 16)
	defer wait()

	// Generate tasks.
	err = core.ListSpecs(ctx, client, spec, filter, func(spec *rpc.ApiSpec) {
		// Check if the styleguide definition contains the mime_type of the spec
		for _, supportedType := range styleguide.GetMimeTypes() {
			if supportedType == spec.GetMimeType() {
				// Delegate the task of computing the conformance report for this spec to the worker pool.
				taskQueue <- &conformance.ComputeConformanceTask{
					Client:          client,
					Spec:            spec,
					LintersMetadata: linterNameToMetadata,
					StyleguideId:    styleguide.GetId(),
				}
				break
			}
		}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to list specs")
	}
}
