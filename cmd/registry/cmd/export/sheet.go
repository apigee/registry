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

package export

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	metrics "github.com/google/gnostic/metrics"
)

func sheetCommand() *cobra.Command {
	var (
		filter   string
		artifact string
	)

	cmd := &cobra.Command{
		Use:   "sheet",
		Short: "Export a specified artifact to a Google sheet",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get config")
			}
			for i := range args {
				args[i] = c.FQName(args[i])
			}

			var path string
			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			inputNames, inputs := collectInputArtifacts(ctx, client, args, filter)
			if len(inputs) == 0 {
				return
			}
			if isInt64Artifact(inputs[0]) {
				title := "artifacts/" + filepath.Base(inputs[0].GetName())
				path, err = core.ExportInt64ToSheet(ctx, title, inputs)
				if err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Failed to export int64 %+v", inputs)
					return
				}
				log.Debugf(ctx, "Exported int64 %+v to %s", inputs, path)
				_ = saveSheetPath(ctx, client, path, artifact)
				return
			}
			messageType, err := core.MessageTypeForMimeType(inputs[0].GetMimeType())
			if err != nil {
				log.Fatalf(ctx, "Not a message type: %s", inputs[0].GetMimeType())
			} else if messageType == "gnostic.metrics.Vocabulary" {
				if len(inputs) != 1 {
					log.Fatalf(ctx, "%d artifacts matched. Please specify exactly one for export.", len(inputs))
				}
				vocabulary, err := getVocabulary(inputs[0])
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to get vocabulary")
				}
				path, err = core.ExportVocabularyToSheet(ctx, inputs[0].Name, vocabulary)
				if err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Failed to export vocabulary %s", inputs[0].Name)
					return
				}
				log.Debugf(ctx, "Exported vocabulary %s to %s", inputs[0].Name, path)
				if artifact == "" {
					artifact = inputs[0].Name + "-sheet"
				}
				_ = saveSheetPath(ctx, client, path, artifact)
			} else if messageType == "gnostic.metrics.VersionHistory" {
				if len(inputs) != 1 {
					log.Fatalf(ctx, "Please specify exactly one version history to export")
					return
				}
				path, err = core.ExportVersionHistoryToSheet(ctx, inputNames[0], inputs[0])
				if err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Failed to export version history %s", inputs[0].Name)
					return
				}
				log.Debugf(ctx, "Exported version history %s to %s", inputs[0].Name, path)
				if artifact == "" {
					artifact = inputs[0].Name + "-sheet"
				}
				_ = saveSheetPath(ctx, client, path, artifact)
			} else if messageType == "gnostic.metrics.Complexity" {
				path, err = core.ExportComplexityToSheet(ctx, "Complexity", inputs)
				if err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Failed to export complexity")
					return
				}
				log.Debugf(ctx, "Exported complexity to %s", path)
				_ = saveSheetPath(ctx, client, path, artifact)
			} else if messageType == "google.cloud.apigeeregistry.applications.v1alpha1.Index" {
				if len(inputs) != 1 {
					log.Fatalf(ctx, "%d artifacts matched. Please specify exactly one for export.", len(inputs))
				}
				index, err := getIndex(inputs[0])
				if err != nil {
					log.FromContext(ctx).WithError(err).Fatal("Failed to get index")
				}
				path, err = core.ExportIndexToSheet(ctx, inputs[0].Name, index)
				if err != nil {
					log.FromContext(ctx).WithError(err).Debugf("Failed to export index %+v", inputs[0].Name)
					return
				}
				log.Debugf(ctx, "Exported index %s to %s", inputs[0].Name, path)
				if artifact == "" {
					artifact = inputs[0].Name + "-sheet"
				}
				_ = saveSheetPath(ctx, client, path, artifact)
			} else {
				log.Fatalf(ctx, "Unknown message type: %s", messageType)
			}
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Filter selected resources")
	cmd.Flags().StringVar(&artifact, "as", "", "Artifact ID to use when saving the result sheet URL")
	return cmd
}

func collectInputArtifacts(ctx context.Context, client connection.RegistryClient, args []string, filter string) ([]string, []*rpc.Artifact) {
	inputNames := make([]string, 0)
	inputs := make([]*rpc.Artifact, 0)
	for _, name := range args {
		artifact, err := names.ParseArtifact(name)
		if err != nil {
			continue
		}

		err = core.ListArtifacts(ctx, client, artifact, filter, true, func(artifact *rpc.Artifact) error {
			inputNames = append(inputNames, artifact.Name)
			inputs = append(inputs, artifact)
			return nil
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("Failed to list artifacts")
		}
	}

	return inputNames, inputs
}

func isInt64Artifact(artifact *rpc.Artifact) bool {
	if artifact.GetMimeType() != "text/plain" {
		return false
	}
	_, err := strconv.ParseInt(string(artifact.GetContents()), 10, 64)
	return err == nil
}

func getVocabulary(artifact *rpc.Artifact) (*metrics.Vocabulary, error) {
	messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
	if err == nil && strings.HasPrefix(messageType, "gnostic.metrics.Vocabulary") {
		vocab := &metrics.Vocabulary{}
		contents := artifact.GetContents()
		if core.IsGZipCompressed(artifact.GetMimeType()) {
			contents, err = core.GUnzippedBytes(contents)
			if err != nil {
				return nil, err
			}
		}
		err := proto.Unmarshal(contents, vocab)
		return vocab, err
	}
	return nil, fmt.Errorf("not a vocabulary: %s", artifact.Name)
}

func getIndex(artifact *rpc.Artifact) (*rpc.Index, error) {
	messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
	if err == nil && messageType == "google.cloud.apigeeregistry.applications.v1alpha1.Index" {
		index := &rpc.Index{}
		err := proto.Unmarshal(artifact.GetContents(), index)
		if err != nil {
			// try unzipping and unmarshaling
			value, err := core.GUnzippedBytes(artifact.GetContents())
			if err != nil {
				return nil, err
			}
			_ = proto.Unmarshal(value, index)
		}
		return index, err
	}
	return nil, fmt.Errorf("not a index: %s", artifact.Name)
}

func saveSheetPath(ctx context.Context, client connection.RegistryClient, path string, artifactName string) error {
	if path == "" {
		return nil
	}
	parts := strings.Split(artifactName, "/")
	parent := strings.Join(parts[0:len(parts)-2], "/")
	artifactID := parts[len(parts)-1]
	req := &rpc.CreateArtifactRequest{
		Parent:     parent,
		ArtifactId: artifactID,
		Artifact: &rpc.Artifact{MimeType: "text/plain",
			Contents: []byte(path),
		},
	}
	_, err := client.CreateArtifact(ctx, req)
	return err
}
