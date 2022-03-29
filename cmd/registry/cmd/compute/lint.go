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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func lintCommand(ctx context.Context) *cobra.Command {
	var linter string
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Compute lint results for API specs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get filter from flags")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			// Initialize task queue.
			taskQueue, wait := core.WorkerPool(ctx, 16)
			defer wait()

			spec, err := names.ParseSpec(args[0])
			if err != nil {
				return // TODO: Log an error.
			}

			// Iterate through a collection of specs and evaluate each.
			err = core.ListSpecs(ctx, client, spec, filter, func(spec *rpc.ApiSpec) error {
				taskQueue <- &computeLintTask{
					client:   client,
					specName: spec.Name,
					linter:   linter,
				}
				return nil
			})
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list specs")
			}
		},
	}

	cmd.Flags().StringVar(&linter, "linter", "", "The linter to use (aip|spectral|gnostic)")
	return cmd
}

type computeLintTask struct {
	client   connection.Client
	specName string
	linter   string
}

func (task *computeLintTask) String() string {
	return fmt.Sprintf("compute %s/lint-%s", task.specName, task.linter)
}

func lintRelation(linter string) string {
	return "lint-" + linter
}

func (task *computeLintTask) Run(ctx context.Context) error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
	}
	spec, err := task.client.GetApiSpec(ctx, request)
	if err != nil {
		return err
	}
	data, err := core.GetBytesForSpec(ctx, task.client, spec)
	if err != nil {
		return err
	}
	var relation string
	var lint *rpc.Lint
	if core.IsOpenAPIv2(spec.GetMimeType()) || core.IsOpenAPIv3(spec.GetMimeType()) {
		// the default openapi linter is gnostic
		if task.linter == "" {
			task.linter = "gnostic"
		}
		relation = lintRelation(task.linter)
		log.Debugf(ctx, "Computing %s/artifacts/%s", spec.Name, relation)
		lint, err = core.NewLintFromOpenAPI(spec.Name, data, task.linter)
		if err != nil {
			return fmt.Errorf("error processing OpenAPI: %s (%s)", spec.Name, err.Error())
		}
	} else if core.IsDiscovery(spec.GetMimeType()) {
		return fmt.Errorf("unsupported Discovery document: %s", spec.Name)
	} else if core.IsProto(spec.GetMimeType()) && core.IsZipArchive(spec.GetMimeType()) {
		// the default proto linter is the aip linter
		if task.linter == "" {
			task.linter = "aip"
		}
		relation = lintRelation(task.linter)
		log.Debugf(ctx, "Computing %s/artifacts/%s", spec.Name, relation)
		lint, err = core.NewLintFromZippedProtos(spec.Name, data)
		if err != nil {
			return fmt.Errorf("error processing protos: %s (%s)", spec.Name, err.Error())
		}
	} else {
		return fmt.Errorf("we don't know how to lint %s", spec.Name)
	}
	subject := spec.GetName()
	messageData, _ := proto.Marshal(lint)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.Lint"),
		Contents: messageData,
	}
	err = core.SetArtifact(ctx, task.client, artifact)
	if err != nil {
		return err
	}
	return nil
}
