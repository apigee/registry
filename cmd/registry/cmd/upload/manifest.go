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

	"github.com/apigee/registry/cmd/registry/controller"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func readManifestProto(filename string) (*rpc.Manifest, error) {
	yamlBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	jsonBytes, _ := yaml.YAMLToJSON(yamlBytes)
	m := &rpc.Manifest{}
	err = protojson.Unmarshal(jsonBytes, m)

	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}

	return m, nil
}

func manifestCommand() *cobra.Command {
	var projectID string
	cmd := &cobra.Command{
		Use:   "manifest FILE_PATH --project-id=value",
		Short: "Upload a dependency manifest",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			manifestPath := args[0]
			if manifestPath == "" {
				log.Fatal(ctx, "Please provide manifest-path")
			}

			manifest, err := readManifestProto(manifestPath)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to read manifest")
			}

			// validate the manifest
			errs := controller.ValidateManifest(fmt.Sprintf("projects/%s/locations/global", projectID), manifest)
			if len(errs) > 0 {
				for _, err := range errs {
					log.FromContext(ctx).WithError(err).Errorf("Invalid manifest entry")
				}
				log.Fatal(ctx, "Manifest definition contains errors")
			}

			manifestData, _ := proto.Marshal(manifest)
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}

			artifact := &rpc.Artifact{
				Name:     "projects/" + projectID + "/locations/global/artifacts/" + manifest.GetId(),
				MimeType: core.MimeTypeForMessageType("google.cloud.apigeeregistry.v1.controller.Manifest"),
				Contents: manifestData,
			}
			log.Debugf(ctx, "Uploading %s", artifact.Name)
			err = core.SetArtifact(ctx, client, artifact)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to save artifact")
			}
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID to use when saving the result manifest artifact")
	_ = cmd.MarkFlagRequired("project-id")
	return cmd
}
