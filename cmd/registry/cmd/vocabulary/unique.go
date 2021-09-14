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
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/googleapis/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
)

func uniqueCommand(ctx context.Context) *cobra.Command {
	var outputID string
	cmd := &cobra.Command{
		Use:   "unique",
		Short: "Compute the unique subsets of each member of specified vocabularies",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.WithError(err).Fatal("Failed to get filter from flags")
			}

			if strings.Contains(outputID, "/") {
				log.Fatal("output_id must specify an artifact id (final segment only) and not a full name.")
			}

			ctx := context.Background()
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.WithError(err).Fatal("Failed to get client")
			}
			names, inputs := collectInputVocabularies(ctx, client, args, filter)
			list := vocabulary.FilterCommon(inputs)
			if outputID != "" {
				for i, unique := range list.Vocabularies {
					outputArtifactName := filepath.Dir(names[i]) + "/" + outputID
					setVocabularyToArtifact(ctx, client, unique, outputArtifactName)
				}
			} else {
				core.PrintMessage(list)
			}
		},
	}

	cmd.Flags().StringVar(&outputID, "output_id", "vocabulary-unique", "Artifact ID to use when saving each result vocabulary")
	return cmd
}
