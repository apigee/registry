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
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	metrics "github.com/googleapis/gnostic/metrics"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var sheetPropertyName string

func init() {
	exportCmd.AddCommand(exportSheetCmd)
	exportSheetCmd.PersistentFlags().StringVar(&sheetPropertyName, "as", "", "name of property to hold url of exported sheet")
}

var exportSheetCmd = &cobra.Command{
	Use:   "sheet",
	Short: "Export a specified property to a Google sheet",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var path string
		var err error
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		inputNames, inputs := collectInputProperties(ctx, client, args, exportFilter)
		if len(inputs) == 0 {
			return
		}
		if isInt64Property(inputs[0]) {
			title := "properties/" + inputs[0].GetRelation()
			path, err = core.ExportInt64ToSheet(title, inputs)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			log.Printf("exported int64 %+v to %s", inputs, path)
			saveSheetPath(ctx, client, path, sheetPropertyName)
			return
		}
		typeURL := messageTypeURL(inputs[0])
		if typeURL == "gnostic.metrics.Vocabulary" {
			if len(inputs) != 1 {
				log.Fatalf("%d properties matched. Please specify exactly one for export.", len(inputs))
			}
			vocabulary, err := getVocabulary(inputs[0])
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			path, err = core.ExportVocabularyToSheet(inputs[0].Name, vocabulary)
			log.Printf("exported vocabulary %s to %s", inputs[0].Name, path)
			if sheetPropertyName == "" {
				sheetPropertyName = inputs[0].Name + "-sheet"
			}
			saveSheetPath(ctx, client, path, sheetPropertyName)
		} else if typeURL == "gnostic.metrics.VersionHistory" {
			if len(inputs) != 1 {
				log.Fatalf("please specify exactly one version history to export")
				return
			}
			path, err = core.ExportVersionHistoryToSheet(inputNames[0], inputs[0])
			log.Printf("exported version history %s to %s", inputs[0].Name, path)
			if sheetPropertyName == "" {
				sheetPropertyName = inputs[0].Name + "-sheet"
			}
			saveSheetPath(ctx, client, path, sheetPropertyName)
		} else if typeURL == "gnostic.metrics.Complexity" {
			path, err = core.ExportComplexityToSheet("Complexity", inputs)
			log.Printf("exported complexity %+v to %s", inputs, path)
			saveSheetPath(ctx, client, path, sheetPropertyName)
		} else if typeURL == "google.cloud.apigee.registry.v1alpha1.Index" {
			if len(inputs) != 1 {
				log.Fatalf("%d properties matched. Please specify exactly one for export.", len(inputs))
			}
			index, err := getIndex(inputs[0])
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			path, err = core.ExportIndexToSheet(inputs[0].Name, index)
			log.Printf("exported index %s to %s", inputs[0].Name, path)
			if sheetPropertyName == "" {
				sheetPropertyName = inputs[0].Name + "-sheet"
			}
			saveSheetPath(ctx, client, path, sheetPropertyName)
		} else {
			log.Fatalf("Unknown message type: %s", typeURL)
		}
	},
}

func versionNameOfPropertyName(propertyName string) string {
	n := propertyName
	for i := 0; i < 4; i++ {
		n = filepath.Dir(n)
	}
	return n
}

func collectInputProperties(ctx context.Context, client connection.Client, args []string, filter string) ([]string, []*rpc.Property) {
	inputNames := make([]string, 0)
	inputs := make([]*rpc.Property, 0)
	for _, name := range args {
		if m := names.PropertyRegexp().FindStringSubmatch(name); m != nil {
			err := core.ListProperties(ctx, client, m, filter, true, func(property *rpc.Property) {
				inputNames = append(inputNames, property.Name)
				inputs = append(inputs, property)
			})
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		}
	}
	return inputNames, inputs
}

func isInt64Property(property *rpc.Property) bool {
	switch property.GetValue().(type) {
	case *rpc.Property_Int64Value:
		return true
	default:
		return false
	}
}

func messageTypeURL(property *rpc.Property) string {
	switch v := property.GetValue().(type) {
	case *rpc.Property_MessageValue:
		return v.MessageValue.TypeUrl
	default:
		return ""
	}
}

func getVocabulary(property *rpc.Property) (*metrics.Vocabulary, error) {
	switch v := property.GetValue().(type) {
	case *rpc.Property_MessageValue:
		if v.MessageValue.TypeUrl == "gnostic.metrics.Vocabulary" {
			vocab := &metrics.Vocabulary{}
			err := proto.Unmarshal(v.MessageValue.Value, vocab)
			return vocab, err
		}
	}
	return nil, fmt.Errorf("not a vocabulary: %s", property.Name)
}

func getIndex(property *rpc.Property) (*rpc.Index, error) {
	switch v := property.GetValue().(type) {
	case *rpc.Property_MessageValue:
		if v.MessageValue.TypeUrl == "google.cloud.apigee.registry.v1alpha1.Index" {
			index := &rpc.Index{}
			err := proto.Unmarshal(v.MessageValue.Value, index)
			if err != nil {
				// try unzipping and unmarshaling
				value, err := core.GUnzippedBytes(v.MessageValue.Value)
				if err != nil {
					return nil, err
				}
				err = proto.Unmarshal(value, index)
			}
			return index, err
		}
	}
	return nil, fmt.Errorf("not a index: %s", property.Name)
}

func saveSheetPath(ctx context.Context, client connection.Client, path string, propertyName string) error {
	if path == "" {
		return nil
	}
	parts := strings.Split(propertyName, "/")
	parent := strings.Join(parts[0:len(parts)-2], "/")
	propertyID := parts[len(parts)-1]
	req := &rpc.CreatePropertyRequest{
		Parent:     parent,
		PropertyId: propertyID,
		Property: &rpc.Property{
			Value: &rpc.Property_StringValue{StringValue: path},
		},
	}
	_, err := client.CreateProperty(ctx, req)
	return err
}
