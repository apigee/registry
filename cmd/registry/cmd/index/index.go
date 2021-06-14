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
	"log"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var indexFilter string

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Operate on API indexes in the API Registry",
	}

	cmd.AddCommand(indexUnionCmd)

	// TODO: Remove the global state.
	cmd.PersistentFlags().StringVar(&indexFilter, "filter", "", "filter index arguments")
	return cmd
}

func collectInputIndexes(ctx context.Context, client connection.Client, args []string, filter string) ([]string, []*rpc.Index) {
	inputNames := make([]string, 0)
	inputs := make([]*rpc.Index, 0)
	for _, name := range args {
		if m := names.ArtifactRegexp().FindStringSubmatch(name); m != nil {
			err := core.ListArtifacts(ctx, client, m, filter, true, func(artifact *rpc.Artifact) {
				messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
				if err == nil && messageType == "google.cloud.apigee.registry.applications.v1alpha1.Index" {
					vocab := &rpc.Index{}
					err := proto.Unmarshal(artifact.GetContents(), vocab)
					if err != nil {
						log.Printf("%+v", err)
					} else {
						inputNames = append(inputNames, artifact.Name)
						inputs = append(inputs, vocab)
					}
				} else {
					log.Printf("skipping, not an index: %s\n", artifact.Name)
				}
			})
			if err != nil {
				log.Fatalf("%s", err.Error())
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
		log.Fatalf("%s", err.Error())
	}
	messageData, err = core.GZippedBytes(messageData)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	artifact := &rpc.Artifact{
		Name:     subject + "/artifacts/" + relation,
		MimeType: core.MimeTypeForMessageType("google.cloud.apigee.registry.applications.v1alpha1.Index"),
		Contents: messageData,
	}
	err = core.SetArtifact(ctx, client, artifact)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}
