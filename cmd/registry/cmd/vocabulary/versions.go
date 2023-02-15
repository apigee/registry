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
	"strings"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/google/gnostic/metrics/vocabulary"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

func versionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "Compute the differences in API vocabularies associated with successive API versions",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get filter from flags")
			}
			output, err := cmd.Flags().GetString("output")
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get output from flags")
			}
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			names, inputs := collectInputVocabularies(ctx, client, args, filter)

			parts := strings.Split(names[0], "/")
			parts = parts[0:4]
			parent := strings.Join(parts, "/")

			history := vocabulary.Version(inputs, names, parent)
			if output != "" {
				setVersionHistoryToArtifact(ctx, client, history, output)
			} else {
				fmt.Println(protojson.Format((history)))
			}
		},
	}

	cmd.Flags().String("output", "", "artifact name to use when saving the vocabulary artifact")
	return cmd
}
