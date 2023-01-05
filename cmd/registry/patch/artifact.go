// Copyright 2022 Google LLC. All Rights Reserved.
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

package patch

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/apigee/registry/cmd/registry/types"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/models"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// styleForYAML sets the style field on a tree of yaml.Nodes for YAML export.
func styleForYAML(node *yaml.Node) {
	node.Style = 0
	for _, n := range node.Content {
		styleForYAML(n)
	}
}

// styleForYAML sets the style field on a tree of yaml.Nodes for JSON export.
func styleForJSON(node *yaml.Node) {
	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode, yaml.MappingNode:
		node.Style = yaml.FlowStyle
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str":
			node.Style = yaml.DoubleQuotedStyle
		default:
			node.Style = 0
		}
	case yaml.AliasNode:
	default:
	}
	for _, n := range node.Content {
		styleForJSON(n)
	}
}

func removeIdAndKind(node *yaml.Node) *yaml.Node {
	if node.Kind == yaml.MappingNode {
		content := make([]*yaml.Node, 0)
		for i := 0; i < len(node.Content); i += 2 {
			k := node.Content[i]
			if k.Value != "id" && k.Value != "kind" {
				content = append(content, node.Content[i])
				content = append(content, node.Content[i+1])
			}
		}
		node.Content = content
	}
	return node
}

// PatchForArtifact allows an artifact to be individually exported as a YAML file.
func PatchForArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) (*models.Artifact, error) {
	if message.Contents == nil {
		req := &rpc.GetArtifactContentsRequest{
			Name: message.Name,
		}
		body, err := client.GetArtifactContents(ctx, req)
		if err != nil {
			return nil, err
		}
		message.Contents = body.Data
	}
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	var node *yaml.Node
	if strings.HasPrefix(message.MimeType, "application/yaml") {
		var doc yaml.Node
		err = yaml.Unmarshal(message.Contents, &doc)
		if err != nil {
			return nil, err
		}
		// The top-level node is a "document" node. We need to marshal the node below it.
		node = doc.Content[0]
	} else {
		m, err := types.MessageForMimeType(message.MimeType)
		if err != nil {
			return nil, err
		}
		// Unmarshal the serialized protobuf containing the artifact content.
		if err = proto.Unmarshal(message.Contents, m); err != nil {
			return nil, err
		}
		// Marshal the artifact content as JSON using the protobuf marshaller.
		var s []byte
		s, err = protojson.MarshalOptions{
			UseEnumNumbers:  false,
			EmitUnpopulated: true,
			Indent:          "  ",
			UseProtoNames:   false,
		}.Marshal(m)
		if err != nil {
			return nil, err
		}
		// Unmarshal the JSON with yaml.v3 so that we can re-marshal it as YAML.
		var doc yaml.Node
		err = yaml.Unmarshal([]byte(s), &doc)
		if err != nil {
			return nil, err
		}
		// The top-level node is a "document" node. We need to marshal the node below it.
		node = doc.Content[0]
		// Restyle the YAML representation so that it will be serialized with YAML defaults.
		styleForYAML(node)
		// We exclude the id and kind fields from YAML serializations.
		node = removeIdAndKind(node)
	}
	// Wrap the artifact for YAML export.
	return &models.Artifact{
		Header: models.Header{
			ApiVersion: RegistryV1,
			Kind:       types.KindForMimeType(message.MimeType),
			Metadata: models.Metadata{
				Name:        artifactName.ArtifactID(),
				Parent:      names.ExportableName(artifactName.Parent(), artifactName.ProjectID()),
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: *node,
	}, nil
}

func applyArtifactPatchBytes(ctx context.Context, client connection.RegistryClient, bytes []byte, parent string) error {
	var artifact models.Artifact
	err := yaml.Unmarshal(bytes, &artifact)
	if err != nil {
		return err
	}
	return applyArtifactPatch(ctx, client, &artifact, parent)
}

func artifactName(parent string, metadata models.Metadata) (names.Artifact, error) {
	if metadata.Parent != "" {
		parent = parent + "/" + metadata.Parent
	}
	return names.ParseArtifact(parent + "/artifacts/" + metadata.Name)
}

func applyArtifactPatch(ctx context.Context, client connection.RegistryClient, content *models.Artifact, parent string) error {
	// Restyle the YAML representation so that yaml.Marshal will marshal it as JSON.
	styleForJSON(&content.Data)
	// Marshal the YAML representation into the JSON serialization.
	j, err := yaml.Marshal(content.Data)
	if err != nil {
		return err
	}
	// Populate Id and Kind fields in the contents of the artifact
	jWithIdAndKind, err := populateIdAndKind(j, content.Kind, content.Metadata.Name)
	if err != nil {
		return err
	}
	var bytes []byte
	// Unmarshal the JSON serialization into the message struct.
	var m proto.Message
	m, err = types.MessageForKind(content.Kind)
	if err == nil {
		err = protojson.Unmarshal(jWithIdAndKind, m)
		if err != nil {
			if strings.Contains(err.Error(), "unknown field") {
				// Try unmarshaling the original YAML (without the additional Id and Kind fields).
				err = protojson.Unmarshal(j, m)
				if err != nil {
					return err
				}
			}
		}
		// Marshal the message struct to bytes.
		bytes, err = proto.Marshal(m)
		if err != nil {
			return err
		}
	} else {
		// If there was no struct defined for the type, marshal it struct as YAML
		styleForYAML(&content.Data)
		bytes, err = yaml.Marshal(content.Data)
		if err != nil {
			return err
		}
	}
	name, err := artifactName(parent, content.Header.Metadata)
	if err != nil {
		return err
	}
	artifact := &rpc.Artifact{
		Name:        name.String(),
		MimeType:    types.MimeTypeForKind(content.Kind),
		Contents:    bytes,
		Labels:      content.Metadata.Labels,
		Annotations: content.Metadata.Annotations,
	}
	req := &rpc.CreateArtifactRequest{
		Parent:     name.Parent(),
		ArtifactId: name.ArtifactID(),
		Artifact:   artifact,
	}
	_, err = client.CreateArtifact(ctx, req)
	if err != nil {
		req := &rpc.ReplaceArtifactRequest{
			Artifact: artifact,
		}
		_, err = client.ReplaceArtifact(ctx, req)
	}
	return err
}

// populateIdAndKind inserts the "id" and "kind" fields in the supplied json bytes.
func populateIdAndKind(bytes []byte, kind, id string) ([]byte, error) {
	var jsonData map[string]interface{}
	err := json.Unmarshal(bytes, &jsonData)
	if err != nil {
		return nil, err
	}
	if jsonData == nil {
		return nil, errors.New("missing data")
	}
	jsonData["id"] = id
	jsonData["kind"] = kind

	rBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	return rBytes, nil
}
