// Copyright 2022 Google LLC
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
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/cmd/registry/scoring"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func artifactCommand() *cobra.Command {
	var parent string
	cmd := &cobra.Command{
		Use:   "artifact FILE_PATH --parent=value",
		Short: "Upload an artifact",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			artifactFilePath := args[0]
			if artifactFilePath == "" {
				log.Fatal(ctx, "Please provide a FILE_PATH for an artifact")
			}
			artifact, err := buildArtifact(ctx, parent, artifactFilePath)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to read artifact")
			}
			client, err := connection.NewClient(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to get client")
			}
			log.Debugf(ctx, "Uploading %s", artifact.Name)
			if err = core.SetArtifact(ctx, client, artifact); err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to save artifact")
			}
		},
	}
	cmd.Flags().StringVar(&parent, "parent", "", "Parent resource for the artifact")
	_ = cmd.MarkFlagRequired("parent")
	return cmd
}

func buildArtifact(ctx context.Context, parent string, filename string) (*rpc.Artifact, error) {
	yamlBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// get the id and kind of artifact from the YAML elements common to all artifacts
	type ArtifactHeader struct {
		Id   string `yaml:"id"`
		Kind string `yaml:"kind"`
	}
	var header ArtifactHeader
	if err = yaml.Unmarshal(yamlBytes, &header); err != nil {
		return nil, err
	}

	// read the specified kind of artifact
	jsonBytes, _ := yaml.YAMLToJSON(yamlBytes) // to use protojson.Unmarshal()
	var artifact *rpc.Artifact
	switch header.Kind {
	case "DisplaySettings", patch.DisplaySettingsMimeType:
		artifact, err = buildDisplaySettingsArtifact(ctx, jsonBytes)
	case "Lifecycle", patch.LifecycleMimeType:
		artifact, err = buildLifecycleArtifact(ctx, jsonBytes)
	case "Manifest", patch.ManifestMimeType:
		artifact, err = buildManifestArtifact(ctx, parent, jsonBytes)
	case "ReferenceList", patch.ReferenceListMimeType:
		artifact, err = buildReferenceListArtifact(ctx, jsonBytes)
	case "Score", patch.ScoreMimeType:
		artifact, err = buildScoreArtifact(ctx, jsonBytes)
	case "ScoreCard", patch.ScoreCardMimeType:
		artifact, err = buildScoreCardArtifact(ctx, jsonBytes)
	case "ScoreCardDefinition", patch.ScoreCardDefinitionMimeType:
		artifact, err = buildScoreCardDefinitionArtifact(ctx, parent, jsonBytes)
	case "ScoreDefinition", patch.ScoreDefinitionMimeType:
		artifact, err = buildScoreDefinitionArtifact(ctx, parent, jsonBytes)
	case "TaxonomyList", patch.TaxonomyListMimeType:
		artifact, err = buildTaxonomyListArtifact(ctx, jsonBytes)
	default:
		err = fmt.Errorf("unsupported artifact type %s", header.Kind)
	}
	if err != nil {
		return nil, err
	}

	// set the artifact name before returning
	artifact.Name = fmt.Sprintf("%s/artifacts/%s", parent, header.Id)
	return artifact, nil
}

func buildDisplaySettingsArtifact(ctx context.Context, jsonBytes []byte) (*rpc.Artifact, error) {
	m := &rpc.DisplaySettings{}
	if err := protojson.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	artifactBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &rpc.Artifact{
		Contents: artifactBytes,
		MimeType: patch.DisplaySettingsMimeType,
	}, nil
}

func buildLifecycleArtifact(ctx context.Context, jsonBytes []byte) (*rpc.Artifact, error) {
	m := &rpc.Lifecycle{}
	if err := protojson.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	artifactBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &rpc.Artifact{
		Contents: artifactBytes,
		MimeType: patch.LifecycleMimeType,
	}, nil
}

func buildManifestArtifact(ctx context.Context, parent string, jsonBytes []byte) (*rpc.Artifact, error) {
	m := &rpc.Manifest{}
	if err := protojson.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	errs := controller.ValidateManifest(ctx, parent, m)
	if count := len(errs); count > 0 {
		for _, err := range errs {
			log.FromContext(ctx).WithError(err).Error("Manifest error")
		}
		return nil, fmt.Errorf("manifest definition contains %d error(s): see logs for details", count)
	}
	artifactBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &rpc.Artifact{
		Contents: artifactBytes,
		MimeType: patch.ManifestMimeType,
	}, nil
}

func buildReferenceListArtifact(ctx context.Context, jsonBytes []byte) (*rpc.Artifact, error) {
	m := &rpc.ReferenceList{}
	if err := protojson.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	artifactBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &rpc.Artifact{
		Contents: artifactBytes,
		MimeType: patch.ReferenceListMimeType,
	}, nil
}

func buildTaxonomyListArtifact(ctx context.Context, jsonBytes []byte) (*rpc.Artifact, error) {
	m := &rpc.TaxonomyList{}
	if err := protojson.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	artifactBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &rpc.Artifact{
		Contents: artifactBytes,
		MimeType: patch.TaxonomyListMimeType,
	}, nil
}

func buildScoreDefinitionArtifact(ctx context.Context, parent string, jsonBytes []byte) (*rpc.Artifact, error) {
	m := &rpc.ScoreDefinition{}
	if err := protojson.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	errs := scoring.ValidateScoreDefinition(ctx, parent, m)
	if count := len(errs); count > 0 {
		for _, err := range errs {
			log.FromContext(ctx).WithError(err).Error("ScoreDefinition error")
		}
		return nil, fmt.Errorf("ScoreDefinition contains %d error(s): see logs for details", count)
	}

	artifactBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &rpc.Artifact{
		Contents: artifactBytes,
		MimeType: patch.ScoreDefinitionMimeType,
	}, nil
}

func buildScoreCardDefinitionArtifact(ctx context.Context, parent string, jsonBytes []byte) (*rpc.Artifact, error) {
	m := &rpc.ScoreCardDefinition{}
	if err := protojson.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	errs := scoring.ValidateScoreCardDefinition(ctx, parent, m)
	if count := len(errs); count > 0 {
		for _, err := range errs {
			log.FromContext(ctx).WithError(err).Error("ScoreCardDefinition error")
		}
		return nil, fmt.Errorf("ScoreCardDefinition contains %d error(s): see logs for details", count)
	}

	artifactBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &rpc.Artifact{
		Contents: artifactBytes,
		MimeType: patch.ScoreCardDefinitionMimeType,
	}, nil
}

func buildScoreArtifact(ctx context.Context, jsonBytes []byte) (*rpc.Artifact, error) {
	m := &rpc.Score{}
	if err := protojson.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	artifactBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &rpc.Artifact{
		Contents: artifactBytes,
		MimeType: patch.ScoreMimeType,
	}, nil
}

func buildScoreCardArtifact(ctx context.Context, jsonBytes []byte) (*rpc.Artifact, error) {
	m := &rpc.ScoreCard{}
	if err := protojson.Unmarshal(jsonBytes, m); err != nil {
		return nil, err
	}
	artifactBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &rpc.Artifact{
		Contents: artifactBytes,
		MimeType: patch.ScoreCardMimeType,
	}, nil
}
