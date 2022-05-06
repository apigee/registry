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
	"fmt"
	"strings"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

type Artifact struct {
	Header `yaml:",inline"`
	Data   yaml.Node `yaml:"data"`
}

// ExportArtifact allows an artifact to be individually exported as a YAML file.
func ExportArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) ([]byte, *Header, error) {
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

func styleForYAML(node *yaml.Node) {
	node.Style = 0
	for _, n := range node.Content {
		styleForYAML(n)
	}
}

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
			if k.Value == "id" || k.Value == "kind" {
				// skip
			} else {
				content = append(content, node.Content[i])
				content = append(content, node.Content[i+1])
			}
		}
		node.Content = content
	}
	return node
}

func newArtifact(message *rpc.Artifact) (*Artifact, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	// unmarshal the serialized protobuf containing the artifact content
	m, err := protoMessageForMimeType(message.MimeType)
	if err != nil {
		return nil, err
	}
	if err = proto.Unmarshal(message.Contents, m); err != nil {
		return nil, err
	}
	// marshal the artifact content as JSON
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
	// read the JSON with yaml.v3
	var doc yaml.Node
	err = yaml.Unmarshal([]byte(s), &doc)
	if err != nil {
		return nil, err
	}
	// the top-level node is a "document" node that we want to unpack
	node := doc.Content[0]
	// restyle the doc so that it will be serialized with YAML defaults
	styleForYAML(node)
	node = removeIdAndKind(node)
	// wrap the artifact for YAML export
	artifact := &Artifact{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       kindForMimeType(message.MimeType),
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
		Data: *node,
	}
	if err != nil {
		return nil, err
	}
	return artifact, nil
}

func applyArtifactPatchBytes(ctx context.Context, client connection.Client, bytes []byte, parent string) error {
	var artifact Artifact
	err := yaml.Unmarshal(bytes, &artifact)
	if err != nil {
		return err
	}
	return applyArtifactPatch(ctx, client, &artifact, parent)
}

func applyArtifactPatch(ctx context.Context, client connection.Client, content *Artifact, parent string) error {
	styleForJSON(&content.Data)
	// get json version of artifact
	j, err := yaml.Marshal(content.Data)
	if err != nil {
		return err
	}
	// read serialized proto from json
	var m proto.Message
	m, err = protoMessageForKind(content.Kind)
	if err != nil {
		return err
	}
	err = protojson.Unmarshal(j, m)
	if err != nil {
		return err
	}
	bytes, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	artifact := &rpc.Artifact{
		Name:     fmt.Sprintf("%s/artifacts/%s", parent, content.Header.Metadata.Name),
		MimeType: MimeTypeForKind(content.Kind),
		Contents: bytes,
	}
	req := &rpc.CreateArtifactRequest{
		Parent:     parent,
		ArtifactId: content.Header.Metadata.Name,
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

func kindForMimeType(mimeType string) string {
	parts := strings.Split(".", mimeType)
	return parts[len(parts)-1]
}

func protoMessageForMimeType(mimeType string) (proto.Message, error) {
	messageType := strings.TrimPrefix(mimeType, "application/octet-stream;type=")
	for k, v := range artifactMessageTypes() {
		if k == messageType {
			return v(), nil
		}
	}
	return nil, fmt.Errorf("unsupported message type %s", messageType)
}

func protoMessageForKind(kind string) (proto.Message, error) {
	for k, v := range artifactMessageTypes() {
		if strings.HasSuffix(k, "."+kind) {
			return v(), nil
		}
	}
	return nil, fmt.Errorf("unsupported kind %s", kind)
}

func messageTypeForKind(kind string) string {
	for k, _ := range artifactMessageTypes() {
		if strings.HasSuffix(k, "."+kind) {
			return k
		}
	}
	return ""
}

func MimeTypeForKind(kind string) string {
	return fmt.Sprintf("application/octet-stream;type=%s", messageTypeForKind(kind))
}

type MessageFactory func() proto.Message

// This is the source of truth for supported artifact types.
func artifactMessageTypes() map[string]MessageFactory {
	return map[string]MessageFactory{
		"google.cloud.apigeeregistry.applications.v1alpha1.StyleGuide": func() proto.Message { return new(rpc.StyleGuide) },
		"google.cloud.apigeeregistry.v1.apihub.DisplaySettings":        func() proto.Message { return new(rpc.DisplaySettings) },
		"google.cloud.apigeeregistry.v1.apihub.Lifecycle":              func() proto.Message { return new(rpc.Lifecycle) },
		"google.cloud.apigeeregistry.v1.apihub.ReferenceList":          func() proto.Message { return new(rpc.ReferenceList) },
		"google.cloud.apigeeregistry.v1.apihub.TaxonomyList":           func() proto.Message { return new(rpc.TaxonomyList) },
		"google.cloud.apigeeregistry.v1.controller.Manifest":           func() proto.Message { return new(rpc.Manifest) },
		"google.cloud.apigeeregistry.v1.scoring.ScoreDefinition":       func() proto.Message { return new(rpc.ScoreDefinition) },
		"google.cloud.apigeeregistry.v1.scoring.Score":                 func() proto.Message { return new(rpc.Score) },
		"google.cloud.apigeeregistry.v1.scoring.ScoreCardDefinition":   func() proto.Message { return new(rpc.ScoreCardDefinition) },
		"google.cloud.apigeeregistry.v1.scoring.ScoreCard":             func() proto.Message { return new(rpc.ScoreCard) },
	}
}
