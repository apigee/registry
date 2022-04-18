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

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

type Artifact struct {
	Header `yaml:",inline"`
	Data   ArtifactData `yaml:"-"`
}

type ArtifactData interface {
	GetMessage() proto.Message
	GetMimeType() string
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

func newArtifact(message *rpc.Artifact) (*Artifact, error) {
	var artifact *Artifact
	var err error
	switch message.GetMimeType() {
	case DisplaySettingsMimeType:
		artifact, err = newDisplaySettings(message)
	case LifecycleMimeType:
		artifact, err = newLifecycle(message)
	case ManifestMimeType:
		artifact, err = newManifest(message)
	case ReferenceListMimeType:
		artifact, err = newReferenceList(message)
	case TaxonomyListMimeType:
		artifact, err = newTaxonomyList(message)
	default:
		artifact, err = newUnknownArtifact(message)
	}
	return artifact, err
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
	default:
		return fmt.Errorf("unable to unmarshal %s", a.Kind)
	}
	return wrapper.Data.Decode(a.Data)
}

func (a *Artifact) MarshalYAML() (interface{}, error) {
	// This struct provides a temporary equivalent Artifact representation with
	// a YAML struct tag that allows the Data field to be directly exported.
	// The primary definition (above) has a dash ("-") for this field to defer
	// unmarshalling until its type is known. But that causes YAML marshalling
	// to skip the field, and rather than try to tweak the struct tags at
	// runtime, we instead copy the artifact data into this equivalent
	// structure that is tagged for direct export.
	type exportable struct {
		Header `yaml:",inline"`
		Data   ArtifactData `yaml:"data"`
	}
	return &exportable{
		Header: a.Header,
		Data:   a.Data,
	}, nil
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
	bytes, err := proto.Marshal(content.Data.GetMessage())
	if err != nil {
		return err
	}
	artifact := &rpc.Artifact{
		Name:     fmt.Sprintf("%s/artifacts/%s", parent, content.Header.Metadata.Name),
		MimeType: content.Data.GetMimeType(),
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
