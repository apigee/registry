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
	"log"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var indexFilter string

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.PersistentFlags().StringVar(&indexFilter, "filter", "", "filter index arguments")
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Operate on API indexes in the API Registry",
}

func collectInputIndexes(ctx context.Context, client connection.Client, args []string, filter string) ([]string, []*rpc.Index) {
	inputNames := make([]string, 0)
	inputs := make([]*rpc.Index, 0)
	for _, name := range args {
		if m := names.PropertyRegexp().FindStringSubmatch(name); m != nil {
			err := core.ListProperties(ctx, client, m, filter, func(property *rpc.Property) {
				switch v := property.GetValue().(type) {
				case *rpc.Property_MessageValue:
					if v.MessageValue.TypeUrl == "google.cloud.apigee.registry.v1alpha1.Index" {
						vocab := &rpc.Index{}
						err := proto.Unmarshal(v.MessageValue.Value, vocab)
						if err != nil {
							log.Printf("%+v", err)
						} else {
							inputNames = append(inputNames, property.Name)
							inputs = append(inputs, vocab)
						}
					} else {
						log.Printf("skipping, not an index: %s\n", property.Name)
					}
				default:
					log.Printf("skipping, not an index: %s\n", property.Name)
				}
			})
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		}
	}
	return inputNames, inputs
}

func setIndexToProperty(ctx context.Context, client connection.Client, output *rpc.Index, outputPropertyName string) {
	parts := strings.Split(outputPropertyName, "/properties/")
	subject := parts[0]
	relation := parts[1]
	messageData, err := proto.Marshal(output)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	messageData, err = core.GZippedBytes(messageData)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	property := &rpc.Property{
		Subject:  subject,
		Relation: relation,
		Name:     subject + "/properties/" + relation,
		Value: &rpc.Property_MessageValue{
			MessageValue: &any.Any{
				TypeUrl: "google.cloud.apigee.registry.v1alpha1.Index",
				Value:   messageData,
			},
		},
	}
	err = core.SetProperty(ctx, client, property)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}
