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
	"path/filepath"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/googleapis/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
)

func init() {
	vocabularyCmd.AddCommand(vocabularyVersionsCmd)
	vocabularyVersionsCmd.Flags().String("output_id", "", "id of property to store output.")
}

// vocabularyVersionsCmd represents the vocabulary versions command
var vocabularyVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Compute the differences in API vocabularies associated with successive API versions.",
	Long:  "Compute the differences in API vocabularies associated with successive API versions.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		names, inputs := collectInputVocabularies(ctx, client, args, vocabularyFilter)
		output := vocabulary.Version(inputs, names, "api")
		if true {
			for _, version := range output.Versions {
				log.Printf("%s\n", version.Name)
				newTermsPropertyName := filepath.Dir(version.Name) + "/" + "vocabulary-new"
				setVocabularyToProperty(ctx, client, version.NewTerms, newTermsPropertyName)
				deletedTermsPropertyName := filepath.Dir(version.Name) + "/" + "vocabulary-deleted"
				setVocabularyToProperty(ctx, client, version.DeletedTerms, deletedTermsPropertyName)
			}
		} else {
			core.PrintMessage(output)
		}
	},
}
