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
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"
	"log"
	"strings"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes/any"
	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

		jobQueue := make(chan Runnable, 1024)

		workerCount := 64
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go worker(ctx, jobQueue)
		}

		name := args[0]
		if m := names.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			// iterate through a collection of specs and summarize each
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
			wg.Wait()
		}
	},
}

type summarizeOpenAPIRunnable struct {
	ctx       context.Context
	client    connection.Client
	projectID string
	specName  string
}

func (job *summarizeOpenAPIRunnable) run() error {
	return summarizeSpec(job.ctx, job.client, job.specName, job.projectID)
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}

func resourceNameOfSpec(segments []string) string {
	if len(segments) == 4 {
		return "projects/" + segments[0] +
			"/apis/" + segments[1] +
			"/versions/" + segments[2] +
			"/specs/" + segments[3]
	}
	return ""
}

func getBytesForSpec(spec *rpc.Spec) ([]byte, error) {
	var data []byte
	if strings.Contains(spec.GetStyle(), "+gzip") {
		gr, err := gzip.NewReader(bytes.NewBuffer(spec.GetContents()))
		defer gr.Close()
		data, err = ioutil.ReadAll(gr)
		if err != nil {
			return nil, err
		}
	} else {
		data = spec.GetContents()
	}
	return data, nil
}

func summarizeSpec(ctx context.Context,
	client *gapic.RegistryClient,
	specName string,
	projectID string) error {

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
		data, err := getBytesForSpec(spec)
		if err != nil {
			return nil
		}
		document, err := openapi_v2.ParseDocument(data)
		if err != nil {
			return err
		}
		summary := summarizeOpenAPIv2Document(document)

		property := &rpc.Property{}
		property.Subject = spec.GetName()
		property.Relation = "summary"
		property.Name = property.Subject + "/properties/" + property.Relation
		complexitySummary := &metrics.Complexity{}
		complexitySummary.SchemaCount = summary.SchemaCount
		complexitySummary.SchemaPropertyCount = summary.SchemaPropertyCount
		complexitySummary.PathCount = summary.PathCount
		complexitySummary.GetCount = summary.GetCount
		complexitySummary.PostCount = summary.PostCount
		complexitySummary.PutCount = summary.PutCount
		complexitySummary.DeleteCount = summary.DeleteCount
		messageData, err := proto.Marshal(complexitySummary)
		anyValue := &any.Any{
			TypeUrl: "gnostic.metrics.Complexity",
			Value:   messageData,
		}
		property.Value = &rpc.Property_MessageValue{MessageValue: anyValue}
		err = setProperty(ctx, client, projectID, property)
		if err != nil {
			return err
		}

	}
	if strings.HasPrefix(spec.GetStyle(), "openapi/v3") {
		data, err := getBytesForSpec(spec)
		if err != nil {
			return nil
		}
		document, err := openapi_v3.ParseDocument(data)
		if err != nil {
			return err
		}
		summary := summarizeOpenAPIv3Document(document)

		property := &rpc.Property{}
		property.Subject = spec.GetName()
		property.Relation = "summary"
		property.Name = property.Subject + "/properties/" + property.Relation
		complexitySummary := &metrics.Complexity{}
		complexitySummary.SchemaCount = summary.SchemaCount
		complexitySummary.SchemaPropertyCount = summary.SchemaPropertyCount
		complexitySummary.PathCount = summary.PathCount
		complexitySummary.GetCount = summary.GetCount
		complexitySummary.PostCount = summary.PostCount
		complexitySummary.PutCount = summary.PutCount
		complexitySummary.DeleteCount = summary.DeleteCount
		messageData, err := proto.Marshal(complexitySummary)
		anyValue := &any.Any{
			TypeUrl: "gnostic.metrics.Complexity",
			Value:   messageData,
		}
		property.Value = &rpc.Property_MessageValue{MessageValue: anyValue}
		err = setProperty(ctx, client, projectID, property)
		if err != nil {
			return err
		}

	}
	return nil
}

func setProperty(ctx context.Context, client *gapic.RegistryClient, projectID string, property *rpc.Property) error {
	request := &rpc.CreatePropertyRequest{}
	request.Property = property
	request.PropertyId = property.GetRelation()
	request.Parent = property.GetSubject()
	_, err := client.CreateProperty(ctx, request)
	if err == nil {
		return nil
	}
	code := status.Code(err)
	if code == codes.AlreadyExists {
		request := &rpc.UpdatePropertyRequest{}
		request.Property = property
		_, err := client.UpdateProperty(ctx, request)
		return err
	}
	return err
}

// Summary ...
type Summary struct {
	Title               string
	Description         string
	Version             string
	SchemaCount         int32
	SchemaPropertyCount int32
	PathCount           int32
	GetCount            int32
	PostCount           int32
	PutCount            int32
	DeleteCount         int32
	TagCount            int32
	Extensions          []string
}

// NewSummary ...
func NewSummary() *Summary {
	s := &Summary{}
	s.Extensions = make([]string, 0)
	return s
}

func (s *Summary) addExtension(name string) {
	for _, n := range s.Extensions {
		if n == name {
			return
		}
	}
	s.Extensions = append(s.Extensions, name)
}

func truncateString(str string, num int) string {
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		return str[0:num] + "..."
	}
	return str
}

func summarizeOpenAPIv2Document(document *openapi_v2.Document) *Summary {
	summary := NewSummary()

	if document.Info != nil {
		summary.Title = document.Info.Title
		summary.Description = truncateString(document.Info.Description, 240)
		summary.Version = document.Info.Version
	}

	if document.Definitions != nil && document.Definitions.AdditionalProperties != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			summarizeSchema(summary, pair.Value)
		}
	}

	for _, pair := range document.Paths.Path {
		summary.PathCount++
		v := pair.Value
		if v.Get != nil {
			summary.GetCount++
			summarizeOperation(summary, v.Get)
		}
		if v.Post != nil {
			summary.PostCount++
			summarizeOperation(summary, v.Post)
		}
		if v.Put != nil {
			summary.PutCount++
			summarizeOperation(summary, v.Put)
		}
		if v.Delete != nil {
			summary.DeleteCount++
			summarizeOperation(summary, v.Delete)
		}
	}
	for _, tag := range document.Tags {
		summary.TagCount++
		summarizeVendorExtension(summary, tag.VendorExtension)
	}
	return summary
}

func summarizeOperation(summary *Summary, operation *openapi_v2.Operation) {
	summarizeVendorExtension(summary, operation.Responses.VendorExtension)
	summarizeVendorExtension(summary, operation.VendorExtension)
}

func summarizeSchema(summary *Summary, schema *openapi_v2.Schema) {
	summary.SchemaCount++
	if schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			summary.SchemaPropertyCount++
			summarizeSchema(summary, pair.Value)
		}
	}
	summarizeVendorExtension(summary, schema.VendorExtension)
}

func summarizeVendorExtension(summary *Summary, vendorExtension []*openapi_v2.NamedAny) {
	if len(vendorExtension) > 0 {
		for _, v := range vendorExtension {
			summary.addExtension(v.Name)
		}
	}
}

func summarizeOpenAPIv3Document(document *openapi_v3.Document) *Summary {
	summary := NewSummary()

	if document.Info != nil {
		summary.Title = document.Info.Title
		summary.Description = truncateString(document.Info.Description, 240)
		summary.Version = document.Info.Version
	}

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
			summarizeOpenAPIv3Operation(summary, v.Get)
		}
		if v.Post != nil {
			summary.PostCount++
			summarizeOpenAPIv3Operation(summary, v.Post)
		}
		if v.Put != nil {
			summary.PutCount++
			summarizeOpenAPIv3Operation(summary, v.Put)
		}
		if v.Delete != nil {
			summary.DeleteCount++
			summarizeOpenAPIv3Operation(summary, v.Delete)
		}
	}
	for _, tag := range document.Tags {
		summary.TagCount++
		summarizeOpenAPIv3VendorExtension(summary, tag.SpecificationExtension)
	}
	return summary
}

func summarizeOpenAPIv3Schema(summary *Summary, schemaOrReference *openapi_v3.SchemaOrReference) {
	summary.SchemaCount++
	schema := schemaOrReference.GetSchema()
	if schema != nil && schema.Properties != nil {
		for _, pair := range schema.Properties.AdditionalProperties {
			summary.SchemaPropertyCount++
			summarizeOpenAPIv3Schema(summary, pair.Value)
		}
	}
	if schema != nil {
		summarizeOpenAPIv3VendorExtension(summary, schema.SpecificationExtension)
	}
}

func summarizeOpenAPIv3VendorExtension(summary *Summary, vendorExtension []*openapi_v3.NamedAny) {
	if len(vendorExtension) > 0 {
		for _, v := range vendorExtension {
			summary.addExtension(v.Name)
		}
	}
}

func summarizeOpenAPIv3Operation(summary *Summary, operation *openapi_v3.Operation) {
	summarizeOpenAPIv3VendorExtension(summary, operation.Responses.SpecificationExtension)
	summarizeOpenAPIv3VendorExtension(summary, operation.SpecificationExtension)
}
