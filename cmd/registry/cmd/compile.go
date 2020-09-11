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

	"github.com/apigee/registry/cmd/registry/tools"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func init() {
	rootCmd.AddCommand(compileCmd)
}

// compileCmd represents the compile command
var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile API specs.",
	Long:  `Compile API specs.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		log.Printf("compile called %+v", args)
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		// Initialize job queue.
		jobQueue := make(chan tools.Runnable, 1024)
		workerCount := 64
		for i := 0; i < workerCount; i++ {
			tools.WaitGroup().Add(1)
			go tools.Worker(ctx, jobQueue)
		}
		// Generate jobs.
		name := args[0]
		if m := names.SpecRegexp().FindStringSubmatch(name); m != nil {
			err = listSpecs(ctx, client, m, func(spec *rpc.Spec) {
				jobQueue <- &compileSpecRunnable{
					ctx:      ctx,
					client:   client,
					specName: spec.Name,
				}
			})
			close(jobQueue)
			tools.WaitGroup().Wait()
		}
	},
}

type compileSpecRunnable struct {
	ctx      context.Context
	client   connection.Client
	specName string
}

func (job *compileSpecRunnable) Run() error {
	request := &rpc.GetSpecRequest{
		Name: job.specName,
		View: rpc.SpecView_FULL,
	}
	spec, err := job.client.GetSpec(job.ctx, request)
	if err != nil {
		return err
	}
	name := spec.GetName()
	log.Printf("compiling %s", name)
	parent, specID := tools.ParentAndIdOfResourceNamed(name)
	var newSpecID string
	if strings.HasSuffix(specID, ".pb") {
		return nil // spec is already compiled
	} else if strings.HasSuffix(specID, ".yaml") {
		newSpecID = strings.Replace(specID, ".yaml", ".pb", 1)
	} else if strings.HasSuffix(specID, ".json") {
		newSpecID = strings.Replace(specID, ".json", ".pb", 1)
	} else {
		newSpecID = specID + ".pb"
	}
	data, err := tools.GetBytesForSpec(spec)
	if err != nil {
		return nil
	}
	var document proto.Message
	if strings.HasPrefix(spec.GetStyle(), "openapi/v2") {
		document, err = openapi_v2.ParseDocument(data)
	} else if strings.HasPrefix(spec.GetStyle(), "openapi/v3") {
		document, err = openapi_v3.ParseDocument(data)
	} else {
		return fmt.Errorf("we don't know how to compile %s", spec.Name)
	}
	if err != nil {
		return fmt.Errorf("invalid OpenAPI: %s", spec.Name)
	}
	return tools.UploadBytesForSpec(
		job.ctx,
		job.client,
		parent,
		newSpecID,
		spec.GetStyle(),
		document)
}
