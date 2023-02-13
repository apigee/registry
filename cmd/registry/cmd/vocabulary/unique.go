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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/google/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func uniqueCommand() *cobra.Command {
	var outputID string
	cmd := &cobra.Command{
		Use:   "unique",
		Short: "Compute the unique subsets of each member of specified vocabularies",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get filter from flags")
			}

			if strings.Contains(outputID, "/") {
				log.Fatal(ctx, "output-id must specify an artifact id (final segment only) and not a full name.")
			}

			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			names, inputs := collectInputVocabularies(ctx, client, args, filter)
			list := vocabulary.FilterCommon(inputs)
			if outputID != "" {
				for i, unique := range list.Vocabularies {
					outputArtifactName := filepath.Dir(names[i]) + "/" + outputID
					setVocabularyToArtifact(ctx, client, unique, outputArtifactName)
				}
			} else {
				fmt.Println(protojson.Format((list)))
			}
		},
	}

	cmd.Flags().StringVar(&outputID, "output-id", "vocabulary-unique", "artifact ID to use when saving each result vocabulary")
	return cmd
}
