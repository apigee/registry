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

package lint

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/apigee/registry/cmd/registry/conformance"
	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint SPEC",
		Short: "Compute lint results for API specs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}
			args[0] = c.FQName(args[0])

			linter, err := cmd.Flags().GetString("linter")
			if err != nil {
				return fmt.Errorf("failed to get linter from flags: %s", err)
			}
			if linter == "" {
				return errors.New("--linter argument cannot be empty")
			}
			if _, err = exec.LookPath(fmt.Sprintf("registry-lint-%s", linter)); err != nil {
				return err
			}

			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				return fmt.Errorf("failed to get filter from flags: %s", err)
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return fmt.Errorf("failed to get dry-run from flags: %s", err)
			}

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			// Initialize task queue.
			jobs, err := cmd.Flags().GetInt("jobs")
			if err != nil {
				return fmt.Errorf("failed to get jobs from flags: %s", err)
			}
			taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, jobs)
			defer wait()

			spec, err := names.ParseSpec(args[0])
			if err != nil {
				return fmt.Errorf("invalid spec pattern %s", args[0])
			}

			// Iterate through a collection of specs and evaluate each.
			return visitor.ListSpecs(ctx, client, spec, filter, false, func(ctx context.Context, spec *rpc.ApiSpec) error {
				taskQueue <- &computeLintTask{
					client: client,
					spec:   spec,
					linter: linter,
					dryRun: dryRun,
				}
				return nil
			})
		},
	}

	cmd.Flags().String("linter", "", "the linter to use")
	_ = cmd.MarkFlagRequired("linter")
	cmd.Flags().String("filter", "", "Filter selected resources")
	cmd.Flags().Bool("dry-run", false, "if set, computation results will only be printed and will not stored in the registry")
	cmd.Flags().Int("jobs", 10, "Number of actions to perform concurrently")
	return cmd
}

type computeLintTask struct {
	client connection.RegistryClient
	spec   *rpc.ApiSpec
	linter string
	dryRun bool
}

func (task *computeLintTask) String() string {
	return fmt.Sprintf("compute %s/lint-%s", task.spec.Name, task.linter)
}

func lintRelation(linter string) string {
	return "lint-" + linter
}

func (task *computeLintTask) Run(ctx context.Context) error {
	root, err := conformance.WriteSpecForLinting(ctx, task.client, task.spec)
	if root != "" {
		defer os.RemoveAll(root)
	}
	if err != nil {
		return err
	}
	linterMetadata := conformance.SimpleLinterMetadata(task.linter)
	response, err := conformance.RunLinter(ctx, root, linterMetadata)
	if err != nil {
		return err
	}
	lint := response.Lint
	if task.dryRun {
		fmt.Println(protojson.Format((lint)))
		return nil
	}
	subject := task.spec.GetName()
	messageData, _ := proto.Marshal(lint)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + lintRelation(task.linter),
		MimeType: mime.MimeTypeForMessageType("google.cloud.apigeeregistry.v1.style.Lint"),
		Contents: messageData,
	}
	return visitor.SetArtifact(ctx, task.client, artifact)
}
