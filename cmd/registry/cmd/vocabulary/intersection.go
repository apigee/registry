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

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/googleapis/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
)

func intersectionCommand(ctx context.Context) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "intersection",
		Short: "Compute the intersection of specified API vocabularies",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.Fatalf("Failed to get filter from flags: %s", err)
			}

			ctx := context.Background()
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			_, inputs := collectInputVocabularies(ctx, client, args, filter)
			vocab := vocabulary.Intersection(inputs)
			if output != "" {
				setVocabularyToArtifact(ctx, client, vocab, output)
			} else {
				core.PrintMessage(vocab)
			}
		},
	}

	cmd.Flags().String("output", "", "Artifact name to use when saving the vocabulary artifact")
	cmd.MarkFlagRequired("output")
	return cmd
}
