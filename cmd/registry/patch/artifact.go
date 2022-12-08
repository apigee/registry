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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/models"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ExportArtifact allows an artifact to be individually exported as a YAML file.
func ExportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) ([]byte, *models.Header, error) {
	if message.Contents == nil {
		req := &rpc.GetArtifactContentsRequest{
			Name: message.Name,
		}
		body, err := client.GetArtifactContents(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		message.Contents = body.Data
	}
	artifact, err := newArtifact(message)
	if err != nil {
		return nil, nil, err
	}
	var b bytes.Buffer
	err = yamlEncoder(&b).Encode(artifact)
	if err != nil {
		return nil, nil, err
	}
	return b.Bytes(), &artifact.Header, nil
}

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

func newArtifact(message *rpc.Artifact) (*models.Artifact, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	// Unmarshal the serialized protobuf containing the artifact content.
	m, err := protoMessageForMimeType(message.MimeType)
	if err != nil {
		return nil, err
	}
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
	// The top-level node is a "document" node. We need to remove this before marshalling.
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 {
		return nil, errors.New("failed to unmarshal artifact")
	}
	node := doc.Content[0]
	// Restyle the YAML representation so that it will be serialized with YAML defaults.
	styleForYAML(node)
	// We exclude the id and kind fields from YAML serializations.
	node = removeIdAndKind(node)
	// Wrap the artifact for YAML export.
	return &models.Artifact{
		Header: models.Header{
			ApiVersion: RegistryV1,
			Kind:       kindForMimeType(message.MimeType),
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

func artifactName(parent, artifactID string) (names.Artifact, error) {
	return names.ParseArtifact(parent + "/artifacts/" + artifactID)
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
	j, err = populateIdAndKind(j, content.Kind, content.Metadata.Name)
	if err != nil {
		return err
	}
	// Unmarshal the JSON serialization into the message struct.
	var m proto.Message
	m, err = protoMessageForKind(content.Kind)
	if err != nil {
		return err
	}
	err = protojson.Unmarshal(j, m)
	if err != nil {
		return err
	}
	// Marshal the message struct to bytes.
	bytes, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	name, err := artifactName(parent, content.Header.Metadata.Name)
	if err != nil {
		return err
	}
	artifact := &rpc.Artifact{
		Name:        name.String(),
		MimeType:    MimeTypeForKind(content.Kind),
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
	jsonData["id"] = id
	jsonData["kind"] = kind

	rBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	return rBytes, nil
}

// kindForMimeType returns the message name to be used as the "kind" of the artifact.
func kindForMimeType(mimeType string) string {
	parts := strings.Split(mimeType, ".")
	return parts[len(parts)-1]
}

// protoMessageForMimeType returns an instance of the message that represents the specified type.
func protoMessageForMimeType(mimeType string) (proto.Message, error) {
	messageType := strings.TrimPrefix(mimeType, "application/octet-stream;type=")
	for k, v := range artifactMessageTypes {
		if k == messageType {
			return v(), nil
		}
	}
	return nil, fmt.Errorf("unsupported message type %s", messageType)
}

// protoMessageForMimeType returns an instance of the message that represents the specified kind.
func protoMessageForKind(kind string) (proto.Message, error) {
	for k, v := range artifactMessageTypes {
		if strings.HasSuffix(k, "."+kind) {
			return v(), nil
		}
	}
	return nil, fmt.Errorf("unsupported kind %s", kind)
}

// MimeTypeForKind returns the mime type that corresponds to a kind.
func MimeTypeForKind(kind string) string {
	for k := range artifactMessageTypes {
		if strings.HasSuffix(k, "."+kind) {
			return fmt.Sprintf("application/octet-stream;type=%s", k)
		}
	}
	return "application/octet-stream"
}

// messageFactory represents functions that construct message structs.
type messageFactory func() proto.Message

// artifactMessageTypes is the single source of truth for artifact types that can be represented in YAML.
var artifactMessageTypes map[string]messageFactory = map[string]messageFactory{
	"google.cloud.apigeeregistry.v1.apihub.ApiSpecExtensionList": func() proto.Message { return new(rpc.ApiSpecExtensionList) },
	"google.cloud.apigeeregistry.v1.apihub.DisplaySettings":      func() proto.Message { return new(rpc.DisplaySettings) },
	"google.cloud.apigeeregistry.v1.apihub.Lifecycle":            func() proto.Message { return new(rpc.Lifecycle) },
	"google.cloud.apigeeregistry.v1.apihub.ReferenceList":        func() proto.Message { return new(rpc.ReferenceList) },
	"google.cloud.apigeeregistry.v1.apihub.TaxonomyList":         func() proto.Message { return new(rpc.TaxonomyList) },
	"google.cloud.apigeeregistry.v1.controller.Manifest":         func() proto.Message { return new(rpc.Manifest) },
	"google.cloud.apigeeregistry.v1.scoring.Score":               func() proto.Message { return new(rpc.Score) },
	"google.cloud.apigeeregistry.v1.scoring.ScoreDefinition":     func() proto.Message { return new(rpc.ScoreDefinition) },
	"google.cloud.apigeeregistry.v1.scoring.ScoreCard":           func() proto.Message { return new(rpc.ScoreCard) },
	"google.cloud.apigeeregistry.v1.scoring.ScoreCardDefinition": func() proto.Message { return new(rpc.ScoreCardDefinition) },
	"google.cloud.apigeeregistry.v1.style.StyleGuide":            func() proto.Message { return new(rpc.StyleGuide) },
	"google.cloud.apigeeregistry.v1.style.ConformanceReport":     func() proto.Message { return new(rpc.ConformanceReport) },
	"google.cloud.apigeeregistry.v1.style.Lint":                  func() proto.Message { return new(rpc.Lint) },
}
