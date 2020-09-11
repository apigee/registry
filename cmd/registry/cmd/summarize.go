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
	"github.com/golang/protobuf/ptypes/any"
	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func init() {
	rootCmd.AddCommand(summarizeCmd)
}

// summarizeCmd represents the summarize command
var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Summarize API specs.",
	Long:  `Summarize API specs.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		log.Printf("summarize called %+v", args)
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
			// Iterate through a collection of specs and summarize each.
			err = listSpecs(ctx, client, m, func(spec *rpc.Spec) {
				jobQueue <- &summarizeOpenAPIRunnable{
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

type summarizeOpenAPIRunnable struct {
	ctx      context.Context
	client   connection.Client
	specName string
}

func (job *summarizeOpenAPIRunnable) Run() error {
	request := &rpc.GetSpecRequest{
		Name: job.specName,
		View: rpc.SpecView_FULL,
	}
	spec, err := job.client.GetSpec(job.ctx, request)
	if err != nil {
		return err
	}
	log.Printf("summarizing %s", spec.Name)
	data, err := tools.GetBytesForSpec(spec)
	if err != nil {
		return nil
	}
	var summary *metrics.Complexity
	if strings.HasPrefix(spec.GetStyle(), "openapi/v2") {
		document, err := openapi_v2.ParseDocument(data)
		if err != nil {
			return fmt.Errorf("invalid OpenAPI: %s", spec.Name)
		}
		summary = tools.SummarizeOpenAPIv2Document(document)
	} else if strings.HasPrefix(spec.GetStyle(), "openapi/v3") {
		document, err := openapi_v3.ParseDocument(data)
		if err != nil {
			return fmt.Errorf("invalid OpenAPI: %s", spec.Name)
		}
		summary = tools.SummarizeOpenAPIv3Document(document)
	} else {
		return fmt.Errorf("we don't know how to summarize %s", spec.Name)
	}
	subject := spec.GetName()
	relation := "summary"
	messageData, err := proto.Marshal(summary)
	property := &rpc.Property{
		Subject:  subject,
		Relation: relation,
		Name:     subject + "/properties/" + relation,
		Value: &rpc.Property_MessageValue{
			MessageValue: &any.Any{
				TypeUrl: "gnostic.metrics.Complexity",
				Value:   messageData,
			},
		},
	}
	err = tools.SetProperty(job.ctx, job.client, property)
	if err != nil {
		return err
	}
	return nil
}
