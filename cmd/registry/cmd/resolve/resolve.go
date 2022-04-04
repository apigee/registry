// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resolve

import (
	"context"
	"fmt"

	"github.com/apigee/registry/cmd/registry/controller"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func fetchManifest(
	ctx context.Context,
	client connection.Client,
	manifestName string) (*rpc.Manifest, error) {

	manifest := &rpc.Manifest{}
	body, err := client.GetArtifactContents(
		ctx,
		&rpc.GetArtifactContentsRequest{
			Name: manifestName,
		})
	if err != nil {
		return nil, err
	}

	contents := body.GetData()
	err = proto.Unmarshal(contents, manifest)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func Command(ctx context.Context) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "resolve MANIFEST_RESOURCE",
		Short: "resolve the dependencies and update the registry state (experimental)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name, err := names.ParseArtifact(args[0])
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Invalid manifest resource name")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			manifest, err := fetchManifest(ctx, client, name.String())
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to fetch manifest")
			}

			log.Debug(ctx, "Generating the list of actions...")
			actions := controller.ProcessManifest(ctx, client, name.ProjectID(), manifest)

			// The monitoring metrics/dashboards are built on top of the format of the log messages here.
			// Check the metric filters before making any changes to the format.
			// Location: registry/deployments/controller/dashboard/*
			if len(actions) == 0 {
				log.Debug(ctx, "Generated 0 actions. The registry is already in a resolved state.")
				return
			}

			log.Debugf(ctx, "Generated %d actions.", len(actions))

			// If dry_run is set to true, print the generated actions and exit
			if dryRun {
				for _, a := range actions {
					log.Debugf(ctx, "Action: %q", a.Command)
				}
				return
			}

			log.Debug(ctx, "Starting execution...")
			taskQueue, wait := core.WorkerPoolWithWarnings(ctx, 64)
			defer wait()
			// Submit tasks to taskQueue
			for _, a := range actions {
				taskQueue <- &controller.ExecCommandTask{
					Action: a,
					TaskID: fmt.Sprintf("%.8s", uuid.New()),
				}
			}
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "if set, actions will only be printed and not executed")
	return cmd
}
