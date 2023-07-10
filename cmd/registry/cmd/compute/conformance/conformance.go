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

package conformance

import (
	"context"
	"fmt"

	"github.com/apigee/registry/cmd/registry/conformance"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
)

var styleguideFilter = fmt.Sprintf("mime_type.contains('%s')", mime.MimeTypeForKind("StyleGuide"))

func Command() *cobra.Command {
	var filter string
	var jobs int
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "conformance SPEC_REVISION",
		Short: "Compute lint results for API specs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}
			args[0] = c.FQName(args[0])

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}

			name, err := names.ParseSpecRevision(args[0])
			if err != nil {
				return err
			}

			specs := make([]*rpc.ApiSpec, 0)
			if name.RevisionID == "" {
				err = visitor.ListSpecs(ctx, client, name.Spec(), 0, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
					specs = append(specs, spec)
					return nil
				})
			} else {
				err = visitor.ListSpecRevisions(ctx, client, name, 0, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
					specs = append(specs, spec)
					return nil
				})
			}
			if err != nil {
				return err
			}

			guides := make([]*style.StyleGuide, 0)
			if err := visitor.ListArtifacts(ctx, client, name.Project().Artifact("-"), 0, styleguideFilter, true, func(ctx context.Context, artifact *rpc.Artifact) error {
				guide := new(style.StyleGuide)
				if err := patch.UnmarshalContents(artifact.GetContents(), artifact.GetMimeType(), guide); err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Unmarshal() to StyleGuide failed on artifact: %s", artifact.GetName())
					return nil
				}
				guides = append(guides, guide)
				return nil
			}); err != nil {
				return err
			}

			for _, guide := range guides {
				log.Debugf(ctx, "Processing styleguide: %s", guide.GetId())
				processStyleGuide(ctx, client, guide, specs, dryRun, jobs)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "filter selected resources")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if set, computation results will only be printed and will not stored in the registry")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
}

// processStyleGuide computes and attaches conformance reports as
// artifacts to a spec or a collection of specs.
func processStyleGuide(ctx context.Context, client connection.RegistryClient, styleguide *style.StyleGuide, specs []*rpc.ApiSpec, dryRun bool, jobs int) {
	linterNameToMetadata, err := conformance.GenerateLinterMetadata(styleguide)
	if err != nil {
		log.Errorf(ctx, "Failed generating linter metadata, check styleguide definition, Error: %s", err)
		return
	}

	taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
	defer wait()

	for _, spec := range specs {
		for _, supportedType := range styleguide.GetMimeTypes() {
			if supportedType != spec.GetMimeType() {
				continue // Only compute matching style guides.
			}

			// Delegate the task of computing the conformance report for this spec to the worker pool.
			taskQueue <- &conformance.ComputeConformanceTask{
				Client:          client,
				Spec:            spec,
				LintersMetadata: linterNameToMetadata,
				StyleguideId:    styleguide.GetId(),
				DryRun:          dryRun,
			}
		}
	}
}
