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
	"context"
	"fmt"
	"io/ioutil"

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

type ArtifactHeader struct {
	Id   string `yaml:"id"`
	Kind string `yaml:"kind"`
}

func readArtifact(ctx context.Context, filename string) (*rpc.Artifact, string, error) {

	yamlBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, "", err
	}

	// get the kind and id from the yaml
	var header ArtifactHeader
	err = yaml.Unmarshal(yamlBytes, &header)
	if err != nil {
		panic(err)
	}

	fmt.Printf("header: %+v\n", header)

	jsonBytes, _ := yaml.YAMLToJSON(yamlBytes)
	artifact := &rpc.Artifact{}
	switch header.Kind {
	case "Manifest", "google.cloud.apigeeregistry.v1.controller.Manifest":
		m := &rpc.Manifest{}
		err = protojson.Unmarshal(jsonBytes, m)
		// validate the manifest
		isValid := true
		for _, resource := range m.GeneratedResources {
			if err := controller.ValidateResourceEntry(resource); err != nil {
				log.FromContext(ctx).WithError(err).Errorf("Invalid manifest entry %v", resource)
				isValid = false
			}
			if err != nil {
				return nil, "", fmt.Errorf("in file %q: %v", filename, err)
			}
		}
		if !isValid {
			log.Fatal(ctx, "Manifest definition contains errors")
		}
		artifact.Contents, _ = proto.Marshal(m)
		artifact.MimeType = core.MimeTypeForMessageType("google.cloud.apigeeregistry.v1.controller.Manifest")
	case "TaxonomyList", "google.cloud.apigeeregistry.v1.apihub.TaxonomyList":

	case "Lifecycle", "google.cloud.apigeeregistry.v1.apihub.Lifecycle":

	default:
	}
	return artifact, header.Id, nil
}

func artifactCommand(ctx context.Context) *cobra.Command {
	var projectID string
	cmd := &cobra.Command{
		Use:   "artifact FILE_PATH --project-id=value",
		Short: "Upload an artifact",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			artifactPath := args[0]
			if artifactPath == "" {
				log.Fatal(ctx, "Please provide a FILE_PATH for an artifact")
			}
			artifact, artifactID, err := readArtifact(ctx, artifactPath)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to read manifest")
			}
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			artifact.Name = "projects/" + projectID + "/locations/global/artifacts/" + artifactID
			log.Debugf(ctx, "Uploading %s", artifact.Name)
			err = core.SetArtifact(ctx, client, artifact)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to save artifact")
			}
		},
	}

	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID to use when saving the artifact")
	_ = cmd.MarkFlagRequired("project-id")
	return cmd
}
