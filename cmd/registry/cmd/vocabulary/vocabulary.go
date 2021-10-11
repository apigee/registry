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
	"strings"

	"github.com/apex/log"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/service/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	metrics "github.com/googleapis/gnostic/metrics"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vocabulary",
		Short: "Operate on API vocabularies in the API Registry",
	}

	cmd.AddCommand(differenceCommand(ctx))
	cmd.AddCommand(intersectionCommand(ctx))
	cmd.AddCommand(unionCommand(ctx))
	cmd.AddCommand(uniqueCommand(ctx))
	cmd.AddCommand(versionsCommand(ctx))

	cmd.PersistentFlags().String("filter", "", "Filter selected resources")
	return cmd
}

func collectInputVocabularies(ctx context.Context, client connection.Client, args []string, filter string) ([]string, []*metrics.Vocabulary) {
	inputNames := make([]string, 0)
	inputs := make([]*metrics.Vocabulary, 0)
	for _, name := range args {
		if m := names.ArtifactRegexp().FindStringSubmatch(name); m != nil {
			err := core.ListArtifacts(ctx, client, m, filter, true, func(artifact *rpc.Artifact) {
				messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
				if err == nil && messageType == "gnostic.metrics.Vocabulary" {
					vocab := &metrics.Vocabulary{}
					err := proto.Unmarshal(artifact.GetContents(), vocab)
					if err != nil {
						log.WithError(err).Debug("Failed to unmarshal contents")
					} else {
						inputNames = append(inputNames, artifact.Name)
						inputs = append(inputs, vocab)
					}
				} else {
					log.Debugf("Skipping, not a vocabulary: %s", artifact.Name)
				}
			})
			if err != nil {
				log.WithError(err).Fatal("Failed to list artifacts")
			}
		}
	}
	return inputNames, inputs
}

func setVocabularyToArtifact(ctx context.Context, client connection.Client, output *metrics.Vocabulary, outputArtifactName string) {
	parts := strings.Split(outputArtifactName, "/artifacts/")
	subject := parts[0]
	relation := parts[1]
	messageData, _ := proto.Marshal(output)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("gnostic.metrics.Vocabulary"),
		Contents: messageData,
	}
	err := core.SetArtifact(ctx, client, artifact)
	if err != nil {
		log.WithError(err).Fatal("Failed to save artifact")
	}
}

func setVersionHistoryToArtifact(ctx context.Context, client connection.Client, output *metrics.VersionHistory, outputArtifactName string) {
	parts := strings.Split(outputArtifactName, "/artifacts/")
	subject := parts[0]
	relation := parts[1]
	messageData, _ := proto.Marshal(output)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("gnostic.metrics.VersionHistory"),
		Contents: messageData,
	}
	err := core.SetArtifact(ctx, client, artifact)
	if err != nil {
		log.WithError(err).Fatal("Failed to save artifact")
	}
}
