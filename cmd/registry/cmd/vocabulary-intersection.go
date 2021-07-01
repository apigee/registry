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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/googleapis/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
)

func init() {
	vocabularyCmd.AddCommand(vocabularyIntersectionCmd)
	vocabularyIntersectionCmd.Flags().String("output", "", "name of artifact where output should be stored")
}

var vocabularyIntersectionCmd = &cobra.Command{
	Use:   "intersection",
	Short: "Compute the intersection of specified API vocabularies",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flagset := cmd.LocalFlags()
		outputArtifactName, err := flagset.GetString("output")
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		ctx := context.Background()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		_, inputs := collectInputVocabularies(ctx, client, args, vocabularyFilter)
		output := vocabulary.Intersection(inputs)
		if outputArtifactName != "" {
			setVocabularyToArtifact(ctx, client, output, outputArtifactName)
		} else {
			core.PrintMessage(output)
		}
	},
}
