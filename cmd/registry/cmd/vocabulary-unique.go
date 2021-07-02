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
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/googleapis/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
)

func init() {
	vocabularyCmd.AddCommand(vocabularyUniqueCmd)
	vocabularyUniqueCmd.Flags().String("output_id", "vocabulary-unique", "id of artifact to store output.")
}

var vocabularyUniqueCmd = &cobra.Command{
	Use:   "unique",
	Short: "Compute the unique subsets of each member of specified vocabularies",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flagset := cmd.LocalFlags()
		outputArtifactID, err := flagset.GetString("output_id")
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		if strings.Contains(outputArtifactID, "/") {
			log.Fatal("output_id must specify an artifact id (final segment only) and not a full name.")
		}
		ctx := context.Background()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		names, inputs := collectInputVocabularies(ctx, client, args, vocabularyFilter)
		output := vocabulary.FilterCommon(inputs)
		if outputArtifactID != "" {
			for i, unique := range output.Vocabularies {
				outputArtifactName := filepath.Dir(names[i]) + "/" + outputArtifactID
				setVocabularyToArtifact(ctx, client, unique, outputArtifactName)
			}
		} else {
			core.PrintMessage(output)
		}
	},
}
