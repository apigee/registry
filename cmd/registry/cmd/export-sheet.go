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

package cmd

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	metrics "github.com/googleapis/gnostic/metrics"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var sheetArtifactName string

func init() {
	exportCmd.AddCommand(exportSheetCmd)
	exportSheetCmd.PersistentFlags().StringVar(&sheetArtifactName, "as", "", "name of artifact to hold url of exported sheet")
}

var exportSheetCmd = &cobra.Command{
	Use:   "sheet",
	Short: "Export a specified artifact to a Google sheet",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var path string
		var err error
		ctx := context.TODO()
		client, err := connection.NewClient(ctx)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		inputNames, inputs := collectInputArtifacts(ctx, client, args, exportFilter)
		if len(inputs) == 0 {
			return
		}
		if isInt64Artifact(inputs[0]) {
			title := "artifacts/" + filepath.Base(inputs[0].GetName())
			path, err = core.ExportInt64ToSheet(title, inputs)
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			log.Printf("exported int64 %+v to %s", inputs, path)
			saveSheetPath(ctx, client, path, sheetArtifactName)
			return
		}
		messageType, err := core.MessageTypeForMimeType(inputs[0].GetMimeType())
		if err != nil {
			log.Fatalf("Not a message type: %s", inputs[0].GetMimeType())
		} else if messageType == "gnostic.metrics.Vocabulary" {
			if len(inputs) != 1 {
				log.Fatalf("%d artifacts matched. Please specify exactly one for export.", len(inputs))
			}
			vocabulary, err := getVocabulary(inputs[0])
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			path, err = core.ExportVocabularyToSheet(inputs[0].Name, vocabulary)
			log.Printf("exported vocabulary %s to %s", inputs[0].Name, path)
			if sheetArtifactName == "" {
				sheetArtifactName = inputs[0].Name + "-sheet"
			}
			saveSheetPath(ctx, client, path, sheetArtifactName)
		} else if messageType == "gnostic.metrics.VersionHistory" {
			if len(inputs) != 1 {
				log.Fatalf("please specify exactly one version history to export")
				return
			}
			path, err = core.ExportVersionHistoryToSheet(inputNames[0], inputs[0])
			log.Printf("exported version history %s to %s", inputs[0].Name, path)
			if sheetArtifactName == "" {
				sheetArtifactName = inputs[0].Name + "-sheet"
			}
			saveSheetPath(ctx, client, path, sheetArtifactName)
		} else if messageType == "gnostic.metrics.Complexity" {
			path, err = core.ExportComplexityToSheet("Complexity", inputs)
			log.Printf("exported complexity to %s", path)
			saveSheetPath(ctx, client, path, sheetArtifactName)
		} else if messageType == "google.cloud.apigee.registry.applications.v1alpha1.Index" {
			if len(inputs) != 1 {
				log.Fatalf("%d artifacts matched. Please specify exactly one for export.", len(inputs))
			}
			index, err := getIndex(inputs[0])
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
			path, err = core.ExportIndexToSheet(inputs[0].Name, index)
			log.Printf("exported index %s to %s", inputs[0].Name, path)
			if sheetArtifactName == "" {
				sheetArtifactName = inputs[0].Name + "-sheet"
			}
			saveSheetPath(ctx, client, path, sheetArtifactName)
		} else {
			log.Fatalf("Unknown message type: %s", messageType)
		}
	},
}

func versionNameOfArtifactName(artifactName string) string {
	n := artifactName
	for i := 0; i < 4; i++ {
		n = filepath.Dir(n)
	}
	return n
}

func collectInputArtifacts(ctx context.Context, client connection.Client, args []string, filter string) ([]string, []*rpc.Artifact) {
	inputNames := make([]string, 0)
	inputs := make([]*rpc.Artifact, 0)
	for _, name := range args {
		if m := names.ArtifactRegexp().FindStringSubmatch(name); m != nil {
			err := core.ListArtifacts(ctx, client, m, filter, true, func(artifact *rpc.Artifact) {
				inputNames = append(inputNames, artifact.Name)
				inputs = append(inputs, artifact)
			})
			if err != nil {
				log.Fatalf("%s", err.Error())
			}
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

func messageTypeURL(artifact *rpc.Artifact) string {
	return artifact.GetMimeType()
}

func getVocabulary(artifact *rpc.Artifact) (*metrics.Vocabulary, error) {
	messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
	if err == nil && messageType == "gnostic.metrics.Vocabulary" {
		vocab := &metrics.Vocabulary{}
		err := proto.Unmarshal(artifact.GetContents(), vocab)
		return vocab, err
	}
	return nil, fmt.Errorf("not a vocabulary: %s", artifact.Name)
}

func getIndex(artifact *rpc.Artifact) (*rpc.Index, error) {
	messageType, err := core.MessageTypeForMimeType(artifact.GetMimeType())
	if err == nil && messageType == "google.cloud.apigee.registry.applications.v1alpha1.Index" {
		index := &rpc.Index{}
		err := proto.Unmarshal(artifact.GetContents(), index)
		if err != nil {
			// try unzipping and unmarshaling
			value, err := core.GUnzippedBytes(artifact.GetContents())
			if err != nil {
				return nil, err
			}
			err = proto.Unmarshal(value, index)
		}
		return index, err
	}
	return nil, fmt.Errorf("not a index: %s", artifact.Name)
}

func saveSheetPath(ctx context.Context, client connection.Client, path string, artifactName string) error {
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
