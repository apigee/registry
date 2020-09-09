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
		if m := names.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			// Iterate through a collection of specs and summarize each.
			err = listSpecs(ctx, client, segments, func(spec *rpc.Spec) {
				m := names.SpecRegexp().FindAllStringSubmatch(spec.Name, -1)
				if m != nil {
					jobQueue <- &summarizeOpenAPIRunnable{
						ctx:       ctx,
						client:    client,
						specName:  spec.Name,
						projectID: segments[1],
					}
				}
			})
			close(jobQueue)
			tools.WaitGroup().Wait()
		}
	},
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}

type summarizeOpenAPIRunnable struct {
	ctx       context.Context
	client    connection.Client
	projectID string
	specName  string
}

func (job *summarizeOpenAPIRunnable) Run() error {
	ctx := job.ctx
	client := job.client
	specName := job.specName
	projectID := job.projectID
	name := specName

	request := &rpc.GetSpecRequest{
		Name: name,
		View: rpc.SpecView_FULL,
	}
	spec, err := client.GetSpec(ctx, request)
	if err != nil {
		return err
	}
	log.Printf("summarizing %s", spec.Name)
	if strings.HasPrefix(spec.GetStyle(), "openapi/v2") {
		data, err := tools.GetBytesForSpec(spec)
		if err != nil {
			return nil
		}
		document, err := openapi_v2.ParseDocument(data)
		if err != nil {
			return fmt.Errorf("invalid OpenAPI: %s", spec.Name)
		}
		summary := summarizeOpenAPIv2Document(document)
		property := &rpc.Property{}
		property.Subject = spec.GetName()
		property.Relation = "summary"
		property.Name = property.Subject + "/properties/" + property.Relation
		messageData, err := proto.Marshal(summary)
		anyValue := &any.Any{
			TypeUrl: "gnostic.metrics.Complexity",
			Value:   messageData,
		}
		property.Value = &rpc.Property_MessageValue{MessageValue: anyValue}
		err = tools.SetProperty(ctx, client, projectID, property)
		if err != nil {
			return err
		}
	}
	if strings.HasPrefix(spec.GetStyle(), "openapi/v3") {
		data, err := tools.GetBytesForSpec(spec)
		if err != nil {
			return nil
		}
		document, err := openapi_v3.ParseDocument(data)
		if err != nil {
			return fmt.Errorf("invalid OpenAPI: %s", spec.Name)
		}
		summary := summarizeOpenAPIv3Document(document)
		property := &rpc.Property{}
		property.Subject = spec.GetName()
		property.Relation = "summary"
		property.Name = property.Subject + "/properties/" + property.Relation
		messageData, err := proto.Marshal(summary)
		anyValue := &any.Any{
			TypeUrl: "gnostic.metrics.Complexity",
			Value:   messageData,
		}
		property.Value = &rpc.Property_MessageValue{MessageValue: anyValue}
		err = tools.SetProperty(ctx, client, projectID, property)
		if err != nil {
			return err
		}
	}
	return nil
}

func summarizeOpenAPIv2Document(document *openapi_v2.Document) *metrics.Complexity {
	summary := &metrics.Complexity{}
	if document.Definitions != nil && document.Definitions.AdditionalProperties != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			summarizeOpenAPIv2Schema(summary, pair.Value)
		}
	}
	for _, pair := range document.Paths.Path {
		summary.PathCount++
		v := pair.Value
		if v.Get != nil {
			summary.GetCount++
		}
		if v.Post != nil {
			summary.PostCount++
		}
		if v.Put != nil {
			summary.PutCount++
		}
		if v.Delete != nil {
			summary.DeleteCount++
		}
	}
	return summary
}

func summarizeOpenAPIv2Schema(summary *metrics.Complexity, schema *openapi_v2.Schema) {
	summary.SchemaCount++
	if schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			summary.SchemaPropertyCount++
			summarizeOpenAPIv2Schema(summary, pair.Value)
		}
	}
}

func summarizeOpenAPIv3Document(document *openapi_v3.Document) *metrics.Complexity {
	summary := &metrics.Complexity{}
	if document.Components != nil && document.Components.Schemas != nil {
		for _, pair := range document.Components.Schemas.AdditionalProperties {
			summarizeOpenAPIv3Schema(summary, pair.Value)
		}
	}
	for _, pair := range document.Paths.Path {
		summary.PathCount++
		v := pair.Value
		if v.Get != nil {
			summary.GetCount++
		}
		if v.Post != nil {
			summary.PostCount++
		}
		if v.Put != nil {
			summary.PutCount++
		}
		if v.Delete != nil {
			summary.DeleteCount++
		}
	}
	return summary
}

func summarizeOpenAPIv3Schema(summary *metrics.Complexity, schemaOrReference *openapi_v3.SchemaOrReference) {
	summary.SchemaCount++
	schema := schemaOrReference.GetSchema()
	if schema != nil && schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			summary.SchemaPropertyCount++
			summarizeOpenAPIv3Schema(summary, pair.Value)
		}
	}
}
