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
			if len(inputs) != 1 {
				log.Fatalf("%d properties matched. Please specify exactly one for export.", len(inputs))
			}
			vocabulary, err := getVocabulary(inputs[0])
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			err = core.ExportVocabularyToSheet(inputs[0].Name, vocabulary)
		} else if typeURL == "gnostic.metrics.Complexity" {
			err = core.ExportComplexityToSheet("Complexity", inputs)
		} else if typeURL == "google.cloud.apigee.registry.v1alpha1.Corpus" {
			if len(inputs) != 1 {
				log.Fatalf("%d properties matched. Please specify exactly one for export.", len(inputs))
			}
			corpus, err := getCorpus(inputs[0])
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			err = core.ExportCorpusToSheet(inputs[0].Name, corpus)
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
			return vocab, err
		}
	}
	return nil, fmt.Errorf("not a vocabulary: %s", property.Name)
}

func getCorpus(property *rpc.Property) (*rpc.Corpus, error) {
	switch v := property.GetValue().(type) {
	case *rpc.Property_MessageValue:
		if v.MessageValue.TypeUrl == "google.cloud.apigee.registry.v1alpha1.Corpus" {
			corpus := &rpc.Corpus{}
			err := proto.Unmarshal(v.MessageValue.Value, corpus)
			return corpus, err
		}
	}
	return nil, fmt.Errorf("not a corpus: %s", property.Name)
}
