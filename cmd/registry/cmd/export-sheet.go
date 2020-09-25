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

func init() {
	exportCmd.AddCommand(exportSheetCmd)
}

// exportSheetCmd represents the export sheet command
var exportSheetCmd = &cobra.Command{
	Use:   "sheet",
	Short: "Export a specified property to a Google sheet.",
	Long:  "Export a specified property to a Google sheet.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		_, inputs := collectInputProperties(ctx, client, args, exportFilter)
		if len(inputs) == 0 {
			return
		}
		typeURL := messageTypeURL(inputs[0])
		if typeURL == "gnostic.metrics.Vocabulary" {
			if len(inputs) == 1 {
				vocabulary, err := getVocabulary(inputs[0])
				if err != nil {
					log.Fatalf("%s", err.Error())
				}
				err = core.ExportVocabularyToSheet(inputs[0].Name, vocabulary)
			} else {
				log.Fatalf("%d properties matched. Please specify exactly one for export.", len(inputs))
			}
		} else if typeURL == "gnostic.metrics.Complexity" {
			name := "Complexity"
			sheetsClient, err := core.NewSheetsClient("")
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			sheet, err := sheetsClient.CreateSheet(name, []string{"Complexity"})
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			_, err = sheetsClient.FormatHeaderRow(sheet.Sheets[0].Properties.SheetId)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			rows := make([][]interface{}, 0)
			rows = append(rows, rowForLabeledComplexity("", "", nil))
			for _, input := range inputs {
				complexity, err := getComplexity(input)
				if err != nil {
					log.Fatalf("%s", err.Error())
				}
				parts := strings.Split(input.Name, "/") // use to get api_id [3] and version_id [5]
				rows = append(rows, rowForLabeledComplexity(parts[3], parts[5], complexity))
			}
			_, err = sheetsClient.Update(fmt.Sprintf("Complexity"), rows)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}

			log.Printf("exported to %+v\n", sheet.SpreadsheetUrl)
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
			err := core.ListProperties(ctx, client, m, filter, func(property *rpc.Property) {
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
			if err != nil {
				return nil, err
			} else {
				return vocab, nil
			}
		} else {
			return nil, fmt.Errorf("not a vocabulary: %s", property.Name)
		}
	default:
		return nil, fmt.Errorf("not a vocabulary: %s", property.Name)
	}
}

func getComplexity(property *rpc.Property) (*metrics.Complexity, error) {
	switch v := property.GetValue().(type) {
	case *rpc.Property_MessageValue:
		if v.MessageValue.TypeUrl == "gnostic.metrics.Complexity" {
			value := &metrics.Complexity{}
			err := proto.Unmarshal(v.MessageValue.Value, value)
			if err != nil {
				return nil, err
			} else {
				return value, nil
			}
		} else {
			return nil, fmt.Errorf("not a complexity: %s", property.Name)
		}
	default:
		return nil, fmt.Errorf("not a complexity: %s", property.Name)
	}
}

func rowForLabeledComplexity(api, version string, c *metrics.Complexity) []interface{} {
	if c == nil {
		return []interface{}{
			"api",
			"version",
			"schemas",
			"schema properties",
			"paths",
			"gets",
			"posts",
			"puts",
			"deletes",
		}
	}
	return []interface{}{api, version, c.SchemaCount, c.SchemaPropertyCount, c.PathCount, c.GetCount, c.PostCount, c.PutCount, c.DeleteCount}
}
