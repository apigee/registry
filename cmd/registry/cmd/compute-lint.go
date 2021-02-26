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

package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func init() {
	computeCmd.AddCommand(computeLintCmd)
	computeLintCmd.Flags().String("linter", "", "name of linter to use (aip, spectral, gnostic)")
}

var computeLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Compute lint results for API specs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		linter, err := cmd.LocalFlags().GetString("linter")
		if err != nil { // ignore errors
			linter = ""
		}
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		// Initialize task queue.
		taskQueue := make(chan core.Task, 1024)
		workerCount := 16
		for i := 0; i < workerCount; i++ {
			core.WaitGroup().Add(1)
			go core.Worker(ctx, taskQueue)
		}
		// Generate tasks.
		name := args[0]
		if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
			// Iterate through a collection of specs and evaluate each.
			err = core.ListSpecs(ctx, client, m, computeFilter, func(spec *rpc.ApiSpec) {
				taskQueue <- &computeLintTask{
					ctx:      ctx,
					client:   client,
					specName: spec.Name,
					linter:   linter,
				}
			})
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			close(taskQueue)
			core.WaitGroup().Wait()
		}
	},
}

type computeLintTask struct {
	ctx      context.Context
	client   connection.Client
	specName string
	linter   string
}

func (task *computeLintTask) Name() string {
	return fmt.Sprintf("compute %s/lint-%s", task.specName, task.linter)
}

func lintRelation(linter string) string {
	return "lint-" + linter
}

func (task *computeLintTask) Run() error {
	request := &rpc.GetApiSpecRequest{
		Name: task.specName,
		View: rpc.View_FULL,
	}
	spec, err := task.client.GetApiSpec(task.ctx, request)
	if err != nil {
		return err
	}
	var relation string
	var lint *rpc.Lint
	if strings.HasPrefix(spec.GetMimeType(), "openapi") {
		// the default openapi linter is gnostic
		if task.linter == "" {
			task.linter = "gnostic"
		}
		relation = lintRelation(task.linter)
		log.Printf("computing %s/artifacts/%s", spec.Name, relation)
		lint, err = core.NewLintFromOpenAPI(spec.Name, spec.GetContents(), task.linter)
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
		log.Printf("computing %s/artifacts/%s", spec.Name, relation)
		lint, err = core.NewLintFromZippedProtos(spec.Name, spec.GetContents())
		if err != nil {
			return fmt.Errorf("error processing protos: %s (%s)", spec.Name, err.Error())
		}
	} else {
		return fmt.Errorf("we don't know how to lint %s", spec.Name)
	}
	subject := spec.GetName()
	messageData, err := proto.Marshal(lint)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("google.cloud.apigee.registry.applications.v1alpha1.Lint"),
		Contents: messageData,
	}
	err = core.SetArtifact(task.ctx, task.client, artifact)
	if err != nil {
		return err
	}
	return nil
}
