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

package vocabulary

import (
	"context"
	"log"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/googleapis/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
)

func init() {
	vocabularyVersionsCmd.Flags().String("output", "", "name of artifact where output should be stored")
}

var vocabularyVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Compute the differences in API vocabularies associated with successive API versions",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		flagset := cmd.LocalFlags()
		outputArtifactName, err := flagset.GetString("output")
		ctx := context.Background()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		names, inputs := collectInputVocabularies(ctx, client, args, vocabularyFilter)

		parts := strings.Split(names[0], "/")
		parts = parts[0:4]
		parent := strings.Join(parts, "/")

		output := vocabulary.Version(inputs, names, parent)
		if outputArtifactName != "" {
			setVersionHistoryToArtifact(ctx, client, output, outputArtifactName)
		} else {
			core.PrintMessage(output)
		}
	},
}
