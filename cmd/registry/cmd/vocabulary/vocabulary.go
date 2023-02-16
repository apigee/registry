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

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/types"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	metrics "github.com/google/gnostic/metrics"
)

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

	cmd.PersistentFlags().String("filter", "", "Filter selected resources")
	return cmd
}

func collectInputVocabularies(ctx context.Context, client connection.RegistryClient, args []string, filter string) ([]string, []*metrics.Vocabulary) {
	c, err := connection.ActiveConfig()
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
	}

	inputNames := make([]string, 0)
	inputs := make([]*metrics.Vocabulary, 0)
	for _, name := range args {
		name = c.FQName(name)
		artifact, err := names.ParseArtifact(name)
		if err != nil {
			continue
		}

		err = visitor.ListArtifacts(ctx, client, artifact, filter, true, func(ctx context.Context, artifact *rpc.Artifact) error {
			messageType, err := types.MessageTypeForMimeType(artifact.GetMimeType())
			if err != nil || messageType != "gnostic.metrics.Vocabulary" {
				log.Debugf(ctx, "Skipping, not a vocabulary: %s", artifact.Name)
				return nil
			}

			vocab := &metrics.Vocabulary{}
			if err := proto.Unmarshal(artifact.GetContents(), vocab); err != nil {
				log.FromContext(ctx).WithError(err).Debug("Failed to unmarshal contents")
				return nil
			}

			inputNames = append(inputNames, artifact.Name)
			inputs = append(inputs, vocab)
			return nil
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to list artifacts")
		}
	}

	return inputNames, inputs
}

func setVocabularyToArtifact(ctx context.Context, client connection.RegistryClient, output *metrics.Vocabulary, outputArtifactName string) {
	parts := strings.Split(outputArtifactName, "/artifacts/")
	subject := parts[0]
	relation := parts[1]
	messageData, _ := proto.Marshal(output)
	var err error
	messageData, err = compress.GZippedBytes(messageData)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to compress artifact")
	}
	log.Debugf(ctx, "Saving vocabulary data (%d bytes)", len(messageData))
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: types.MimeTypeForMessageType("gnostic.metrics.Vocabulary+gzip"),
		Contents: messageData,
	}
	err = visitor.SetArtifact(ctx, client, artifact)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to save artifact")
	}
}

func setVersionHistoryToArtifact(ctx context.Context, client connection.RegistryClient, output *metrics.VersionHistory, outputArtifactName string) {
	parts := strings.Split(outputArtifactName, "/artifacts/")
	subject := parts[0]
	relation := parts[1]
	messageData, _ := proto.Marshal(output)
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: types.MimeTypeForMessageType("gnostic.metrics.VersionHistory"),
		Contents: messageData,
	}
	err := visitor.SetArtifact(ctx, client, artifact)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to save artifact")
	}
}
