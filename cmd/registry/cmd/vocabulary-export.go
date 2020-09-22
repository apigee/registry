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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	metrics "github.com/googleapis/gnostic/metrics"
	"github.com/spf13/cobra"
)

func init() {
	vocabularyCmd.AddCommand(vocabularyExportCmd)
}

// vocabularyExportCmd represents the vocabulary export command
var vocabularyExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a specified API vocabulary.",
	Long:  "Export a specified API vocabulary.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		names, inputs := collectInputs(ctx, client, args, vocabularyFilter)
		if len(names) == 1 {
			err = exportVocabularyToSheet(names[0], inputs[0])
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
		} else {
			log.Fatalf("%d vocabularies matched. Please specify exactly one for export.", len(names))
		}
	},
}

func exportVocabularyToSheet(name string, vocabulary *metrics.Vocabulary) error {
	sheetsClient, err := core.NewSheetsClient("")
	if err != nil {
		return err
	}
	sheet, err := sheetsClient.CreateSheet(name, []string{"Everything", "Schemas", "Properties", "Operations", "Parameters"})
	if err != nil {
		return err
	}
	{
		rows := make([][]interface{}, 0)
		for _, wc := range vocabulary.Schemas {
			rows = append(rows, rowForLabeledWordCount("schema", wc))
		}
		for _, wc := range vocabulary.Properties {
			rows = append(rows, rowForLabeledWordCount("property", wc))
		}
		for _, wc := range vocabulary.Operations {
			rows = append(rows, rowForLabeledWordCount("operation", wc))
		}
		for _, wc := range vocabulary.Parameters {
			rows = append(rows, rowForLabeledWordCount("parameter", wc))
		}
		_, err = sheetsClient.Update(fmt.Sprintf("Everything!A1:C%d", len(rows)), rows)
		if err != nil {
			return err
		}
	}
	// create update function as a closure to simplify multiple uses (below)
	updateSheet := func(title string, pairs []*metrics.WordCount) error {
		rows := make([][]interface{}, 0)
		for _, wc := range pairs {
			rows = append(rows, rowForWordCount(wc))
		}
		_, err = sheetsClient.Update(fmt.Sprintf("%s!A1:C%d", title, len(rows)), rows)
		return err
	}
	updateSheet("Schemas", vocabulary.Schemas)
	updateSheet("Properties", vocabulary.Properties)
	updateSheet("Operations", vocabulary.Operations)
	updateSheet("Parameters", vocabulary.Parameters)
	log.Printf("exported to %+v\n", sheet.SpreadsheetUrl)
	return nil
}

func rowForLabeledWordCount(kind string, wc *metrics.WordCount) []interface{} {
	return []interface{}{kind, wc.Word, wc.Count}
}

func rowForWordCount(wc *metrics.WordCount) []interface{} {
	return []interface{}{wc.Word, wc.Count}
}
