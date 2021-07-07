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
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	metrics "github.com/googleapis/gnostic/metrics"
)

var vocabularyFilter string

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vocabulary",
		Short: "Operate on API vocabularies in the API Registry",
	}

	cmd.AddCommand(differenceCommand())
	cmd.AddCommand(intersectionCommand())
	cmd.AddCommand(unionCommand())
	cmd.AddCommand(uniqueCommand())
	cmd.AddCommand(versionsCommand())

	// TODO: Remove the global state.
	cmd.PersistentFlags().StringVar(&vocabularyFilter, "filter", "", "filter vocabulary arguments")
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
						log.Printf("%+v", err)
					} else {
						inputNames = append(inputNames, artifact.Name)
						inputs = append(inputs, vocab)
					}
				} else {
					log.Printf("skipping, not a vocabulary: %s\n", artifact.Name)
				}
			})
			if err != nil {
				log.Fatalf("%s", err.Error())
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
		log.Fatalf("%s", err.Error())
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
		log.Fatalf("%s", err.Error())
	}
}
