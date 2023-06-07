// Copyright 2022 Google LLC.
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

package encoding

import (
	"bytes"
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// Prefer this encoder because it uses tighter 2-space indentation.
func yamlEncoder(dst io.Writer) *yaml.Encoder {
	enc := yaml.NewEncoder(dst)
	enc.SetIndent(2)
	return enc
}

// Encode a model as YAML.
func EncodeYAML(model interface{}) ([]byte, error) {
	var b bytes.Buffer
	err := yamlEncoder(&b).Encode(model)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// NodeForMessage converts a proto message for export as a YAML node.
func NodeForMessage(m proto.Message) (*yaml.Node, error) {
	// Marshal the artifact content as JSON using the protobuf marshaller.
	var s []byte
	s, err := protojson.MarshalOptions{
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
	node := doc.Content[0]
	// Restyle the YAML representation so that it will be serialized with YAML defaults.
	StyleForYAML(node)
	// We exclude the id and kind fields from YAML serializations.
	node = removeIdAndKind(node)

	return node, nil
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

// StyleForYAML sets the style field on a tree of yaml.Nodes for YAML export.
func StyleForYAML(node *yaml.Node) {
	node.Style = 0
	for _, n := range node.Content {
		StyleForYAML(n)
	}
}

// StyleForJSON sets the style field on a tree of yaml.Nodes for JSON export.
func StyleForJSON(node *yaml.Node) {
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
		StyleForJSON(n)
	}
}
