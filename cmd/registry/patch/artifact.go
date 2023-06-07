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

package patch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/encoding"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// NewArtifact allows an artifact to be individually exported as a YAML file.
func NewArtifact(ctx context.Context, client *gapic.RegistryClient, message *rpc.Artifact) (*encoding.Artifact, error) {
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
		m, err := mime.MessageForMimeType(message.MimeType)
		if err != nil {
			return nil, err
		}
		// Unmarshal the serialized protobuf containing the artifact content.
		if err = proto.Unmarshal(message.Contents, m); err != nil {
			return nil, err
		}
		if node, err = encoding.NodeForMessage(m); err != nil {
			return nil, err
		}
	}
	// Wrap the artifact for YAML export.
	return &encoding.Artifact{
		Header: encoding.Header{
			ApiVersion: encoding.RegistryV1,
			Kind:       mime.KindForMimeType(message.MimeType),
			Metadata: encoding.Metadata{
				Name:        artifactName.ArtifactID(),
				Parent:      names.ExportableName(artifactName.Parent(), artifactName.ProjectID()),
				Labels:      message.Labels,
				Annotations: message.Annotations,
			},
		},
		Data: *node,
	}, nil
}

func applyArtifactPatchBytes(ctx context.Context, client connection.RegistryClient, bytes []byte, project string, filename string) error {
	var artifact encoding.Artifact
	err := yaml.Unmarshal(bytes, &artifact)
	if err != nil {
		return err
	}
	return applyArtifactPatch(ctx, client, &artifact, project, filename)
}

func artifactName(parent string, metadata encoding.Metadata) (names.Artifact, error) {
	if metadata.Parent != "" {
		parent = parent + "/" + metadata.Parent
	}
	return names.ParseArtifact(parent + "/artifacts/" + metadata.Name)
}

// TODO: remove when default
var yamlArchiveKey = struct{}{}

func SetStoreArchivesAsYaml(ctx context.Context) context.Context {
	return context.WithValue(ctx, yamlArchiveKey, true)
}

func storeArchivesAsYaml(ctx context.Context) bool {
	return ctx.Value(yamlArchiveKey) != nil && ctx.Value(yamlArchiveKey).(bool)
}

func applyArtifactPatch(ctx context.Context, client connection.RegistryClient, content *encoding.Artifact, parent string, filename string) error {
	// Restyle the YAML representation so that yaml.Marshal will marshal it as JSON.
	encoding.StyleForJSON(&content.Data)
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
	var mimeType string
	var bytes []byte
	// Unmarshal the JSON serialization into the message struct.
	var m proto.Message
	m, err = mime.MessageForKind(content.Kind)
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
		if storeArchivesAsYaml(ctx) {
			mimeType = mime.YamlMimeTypeForKind(content.Kind)
			encoding.StyleForYAML(&content.Data)
			bytes, err = yaml.Marshal(content.Data)
			if err != nil {
				return err
			}
		} else {
			mimeType = mime.MimeTypeForKind(content.Kind)
			// Marshal the message struct to bytes.
			bytes, err = proto.Marshal(m)
			if err != nil {
				return err
			}
		}
	} else {
		// If there was no struct defined for the type, marshal it struct as YAML
		mimeType = mime.MimeTypeForKind(content.Kind)
		encoding.StyleForYAML(&content.Data)
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
		MimeType:    mimeType,
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
		if _, err := client.ReplaceArtifact(ctx, req); err != nil {
			return fmt.Errorf("ReplaceArtifact: %s", err)
		}
	}
	return nil
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

func UnmarshalContents(contents []byte, mimeType string, message proto.Message) error {
	if !mime.IsYamlKind(mimeType) {
		return proto.Unmarshal(contents, message)
	}
	var node yaml.Node
	if err := yaml.Unmarshal(contents, &node); err != nil {
		return err
	}
	encoding.StyleForJSON(&node)
	bytes, err := yaml.Marshal(&node)
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(bytes, message); err != nil {
		return err
	}

	return err
}
