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
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/googleapis/gnostic/compiler"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	openapi_v3 "github.com/googleapis/gnostic/openapiv3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

// vocabularyCmd represents the vocabulary command
var vocabularyCmd = &cobra.Command{
	Use:   "vocabulary",
	Short: "Compute the vocabulary of an API spec.",
	Long:  `Compute the vocabulary of an API spec.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		log.Printf("vocabulary called %+v", args)
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		name := args[0]
		if m := names.SpecRegexp().FindAllStringSubmatch(name, -1); m != nil {
			segments := m[0]
			if sliceContainsString(segments, "-") {
				// iterate through a collection of specs and vocabulary each
				completions := make(chan int)
				processes := 0
				err = listSpecs(ctx, client, segments, func(spec *rpc.Spec) {
					fmt.Println(spec.Name)
					m := names.SpecRegexp().FindAllStringSubmatch(spec.Name, -1)
					if m != nil {
						processes++
						go func() {
							vocabularySpec(ctx, client, m[0])
							completions <- 1
						}()
					}
				})
				for i := 0; i < processes; i++ {
					<-completions
				}

			} else {
				err := vocabularySpec(ctx, client, segments)
				if err != nil {
					log.Printf("%s", err.Error())
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(vocabularyCmd)
}

func vocabularySpec(ctx context.Context,
	client *gapic.RegistryClient,
	segments []string) error {

	name := tools.ResourceNameOfSpec(segments[1:])
	request := &rpc.GetSpecRequest{
		Name: name,
		View: rpc.SpecView_FULL,
	}
	spec, err := client.GetSpec(ctx, request)
	if err != nil {
		return err
	}

	log.Printf("computing vocabulary of %s", spec.Name)
	if strings.HasPrefix(spec.GetStyle(), "openapi/v2") {
		data, err := tools.GetBytesForSpec(spec)
		if err != nil {
			return nil
		}
		info, err := compiler.ReadInfoFromBytes(spec.GetName(), data)
		if err != nil {
			return err
		}
		document, err := openapi_v2.NewDocument(info, compiler.NewContextWithExtensions("$root", nil, nil))
		if err != nil {
			return err
		}
		log.Printf("%+v", document)

		vocabulary := tools.ProcessDocumentV2(document)
		log.Printf("%+v", vocabulary)

		projectID := segments[1]
		property := &rpc.Property{}
		property.Subject = spec.GetName()
		property.Relation = "vocabulary"
		messageData, err := proto.Marshal(vocabulary)
		anyValue := &any.Any{
			TypeUrl: "Vocabulary",
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
		info, err := compiler.ReadInfoFromBytes(spec.GetName(), data)
		if err != nil {
			return err
		}
		document, err := openapi_v3.NewDocument(info, compiler.NewContextWithExtensions("$root", nil, nil))
		if err != nil {
			return err
		}
		vocabulary := tools.ProcessDocumentV3(document)

		projectID := segments[1]

		log.Printf("%s", projectID)
		property := &rpc.Property{}
		property.Subject = spec.GetName()
		property.Relation = "vocabulary"
		messageData, err := proto.Marshal(vocabulary)
		anyValue := &any.Any{
			TypeUrl: "Vocabulary",
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
