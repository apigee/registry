// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upload

import (
	"fmt"
	"os"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func readStyleGuideProto(filename string) (*rpc.StyleGuide, error) {
	yamlBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return nil, err
	}

	m := &rpc.StyleGuide{}
	err = protojson.Unmarshal(jsonBytes, m)

	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}

	return m, nil
}

func styleGuideCommand() *cobra.Command {
	var projectID string
	cmd := &cobra.Command{
		Use:   "styleguide FILE_PATH --project-id=value",
		Short: "Upload an API style guide",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			styleGuidePath := args[0]
			if styleGuidePath == "" {
				log.Fatal(ctx, "Please provide style guide path")
			}

			styleGuide, err := readStyleGuideProto(styleGuidePath)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to read style guide")
			}
			styleGuideMarshalled, err := proto.Marshal(styleGuide)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to encode style guide")
			}

			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			artifact := &rpc.Artifact{
				Name: "projects/" +
					projectID +
					"/locations/global/artifacts/" +
					styleGuide.GetId(),
				MimeType: patch.MimeTypeForKind("StyleGuide"),
				Contents: styleGuideMarshalled,
			}
			log.Debugf(ctx, "Uploading %s", artifact.Name)
			err = core.SetArtifact(ctx, client, artifact)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to save artifact")
			}
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID to use when storing the styleguide artifact")
	_ = cmd.MarkFlagRequired("project-id")
	return cmd
}
