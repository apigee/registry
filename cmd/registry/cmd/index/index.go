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

package index

import (
	"context"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Operate on API indexes in the API Registry",
	}

	cmd.AddCommand(unionCommand(ctx))

	return cmd
}

func collectInputIndexes(ctx context.Context, client connection.Client, args []string, filter string) ([]string, []*rpc.Index) {
	inputNames := make([]string, 0)
	inputs := make([]*rpc.Index, 0)
	for _, name := range args {
		if artifact, err := names.ParseArtifact(name); err == nil {
			err := core.ListArtifacts(ctx, client, artifact, filter, true, func(artifact *rpc.Artifact) {
				messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
				if err == nil && messageType == "google.cloud.apigeeregistry.applications.v1alpha1.Index" {
					vocab := &rpc.Index{}
					err := proto.Unmarshal(artifact.GetContents(), vocab)
					if err != nil {
						log.FromContext(ctx).WithError(err).Debug("Failed to unmarshal contents")
					} else {
						inputNames = append(inputNames, artifact.Name)
						inputs = append(inputs, vocab)
					}
				} else {
					log.Debugf(ctx, "Skipping, not an index: %s", artifact.Name)
				}
			})
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to list artifacts")
			}
		}
	}
	return inputNames, inputs
}

func setIndexToArtifact(ctx context.Context, client connection.Client, output *rpc.Index, outputArtifactName string) {
	parts := strings.Split(outputArtifactName, "/artifacts/")
	subject := parts[0]
	relation := parts[1]
	messageData, err := proto.Marshal(output)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to marshal output proto")
	}
	messageData, err = core.GZippedBytes(messageData)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to compress artifact contents")
	}
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.Index"),
		Contents: messageData,
	}
	err = core.SetArtifact(ctx, client, artifact)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("Failed to save artifact")
	}
}
