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

	"github.com/apigee/registry/cmd/registry/core"
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
	Data   *yaml.Node `yaml:"data"`
}

type ArtifactData interface {
	buildMessage() proto.Message
	mimeType() string
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

func restyle(node *yaml.Node, style yaml.Style) {
	node.Style = style
	for _, n := range node.Content {
		restyle(n, style)
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

	m, err := protoMessageForMimeType(message.MimeType)
	if err != nil {
		return nil, err
	}
	if err = proto.Unmarshal(message.Contents, m); err != nil {
		return nil, err
	}
	// marshal the artifact contents as JSON
	marshaler := protojson.MarshalOptions{
		UseEnumNumbers:  false,
		EmitUnpopulated: false,
		Indent:          "  ",
		UseProtoNames:   false,
	}
	var s []byte
	s, err = marshaler.Marshal(m)
	if err != nil {
		return nil, err
	}
	// read the JSON with yaml.v3
	var node yaml.Node
	err = yaml.Unmarshal([]byte(s), &node)
	if err != nil {
		return nil, err
	}
	doc := node.Content[0]
	// restyle the doc to be serialized with YAML defaults
	restyle(doc, 0)
	doc = removeIdAndKind(doc)
	// wrap the artifact for YAML export
	artifact := &Artifact{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       kindForMimeType(message.MimeType),
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
		Data: doc,
	}
	if err != nil {
		return nil, err
	}
	return artifact, nil
}

func (a *Artifact) UnmarshalYAML(node *yaml.Node) error {
	// https://stackoverflow.com/questions/66709979/dynamically-parse-yaml-field-to-one-of-a-finite-set-of-structs-in-go
	type Alias Artifact
	type Wrapper struct {
		*Alias `yaml:",inline"`
		Data   yaml.Node `yaml:"data"`
	}
	wrapper := &Wrapper{Alias: (*Alias)(a)}
	if err := node.Decode(wrapper); err != nil {
		return err
	}
	switch a.Kind {
	/*
		case "DisplaySettings":
			a.Data = new(DisplaySettingsData)
		case "Lifecycle":
			a.Data = new(LifecycleData)
		case "Manifest":
			a.Data = new(ManifestData)
		case "ReferenceList":
			a.Data = new(ReferenceListData)
		case "TaxonomyList":
			a.Data = new(TaxonomyListData)
	*/
	default:
		return fmt.Errorf("unable to unmarshal %s", a.Kind)
	}
	return wrapper.Data.Decode(a.Data)
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
	//bytes, err := proto.Marshal(content.Data.buildMessage())
	//if err != nil {
	//	return err
	//}
	bytes := []byte{}
	var err error
	artifact := &rpc.Artifact{
		Name:     fmt.Sprintf("%s/artifacts/%s", parent, content.Header.Metadata.Name),
		MimeType: "", // content.Data.mimeType(),
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

func protoMessageForMimeType(mimeType string) (proto.Message, error) {
	var m proto.Message
	switch mimeType {
	case DisplaySettingsMimeType:
		m = &rpc.DisplaySettings{}
	case LifecycleMimeType:
		m = &rpc.Lifecycle{}
	case ManifestMimeType:
		m = &rpc.Manifest{}
	case ReferenceListMimeType:
		m = &rpc.ReferenceList{}
	case core.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.StyleGuide"):
		m = &rpc.StyleGuide{}
	case TaxonomyListMimeType:
		m = &rpc.TaxonomyList{}
	default:
		return nil, fmt.Errorf("unsupported type %s", mimeType)
	}
	return m, nil
}

func kindForMimeType(mimeType string) string {
	switch mimeType {
	case DisplaySettingsMimeType:
		return "DisplaySettings"
	case LifecycleMimeType:
		return "Lifecycle"
	case ManifestMimeType:
		return "Manifest"
	case ReferenceListMimeType:
		return "ReferenceList"
	case core.MimeTypeForMessageType("google.cloud.apigeeregistry.applications.v1alpha1.StyleGuide"):
		return "StyleGuide"
	case TaxonomyListMimeType:
		return "TaxonomyList"
	default:
		return ""
	}
}
