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
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/proto"
)

const ManifestMimeType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.controller.Manifest"

type ManifestDependency struct {
	Pattern string `yaml:"pattern"`
	Filter  string `yaml:"filter,omitempty"`
}

type ManifestGeneratedResource struct {
	Pattern      string                `yaml:"pattern"`
	Filter       string                `yaml:"filter,omitempty"`
	Receipt      bool                  `yaml:"receipt,omitempty"`
	Dependencies []*ManifestDependency `yaml:"dependencies"`
	Action       string                `yaml:"action"`
}

type ManifestData struct {
	DisplayName        string                       `yaml:"displayName,omitempty"`
	Description        string                       `yaml:"description,omitempty"`
	GeneratedResources []*ManifestGeneratedResource `yaml:"generatedResources"`
}

func (a *ManifestData) GetMimeType() string {
	return ManifestMimeType
}

func (m *ManifestData) GetMessage() proto.Message {
	return &rpc.Manifest{
		DisplayName:        m.DisplayName,
		Description:        m.Description,
		GeneratedResources: m.generatedResources(),
	}
}

func (m *ManifestData) generatedResources() []*rpc.GeneratedResource {
	v := make([]*rpc.GeneratedResource, 0)
	for _, g := range m.GeneratedResources {
		v = append(v, &rpc.GeneratedResource{
			Pattern:      g.Pattern,
			Filter:       g.Filter,
			Receipt:      g.Receipt,
			Dependencies: g.dependencies(),
			Action:       g.Action,
		})
	}
	return v
}

func (g *ManifestGeneratedResource) dependencies() []*rpc.Dependency {
	v := make([]*rpc.Dependency, 0)
	for _, d := range g.Dependencies {
		v = append(v, &rpc.Dependency{
			Pattern: d.Pattern,
			Filter:  d.Filter,
		})
	}
	return v
}

func newManifest(message *rpc.Artifact) (*Artifact, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	value := &rpc.Manifest{}
	err = proto.Unmarshal(message.Contents, value)
	if err != nil {
		return nil, err
	}
	generatedResources := make([]*ManifestGeneratedResource, len(value.GeneratedResources))
	for i, g := range value.GeneratedResources {
		dependencies := make([]*ManifestDependency, len(g.Dependencies))
		for j, d := range g.Dependencies {
			dependencies[j] = &ManifestDependency{
				Pattern: d.Pattern,
				Filter:  d.Filter,
			}
		}
		generatedResources[i] = &ManifestGeneratedResource{
			Pattern:      g.Pattern,
			Filter:       g.Filter,
			Receipt:      g.Receipt,
			Dependencies: dependencies,
			Action:       g.Action,
		}
	}
	return &Artifact{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "Manifest",
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
		Data: &ManifestData{
			DisplayName:        value.DisplayName,
			Description:        value.Description,
			GeneratedResources: generatedResources,
		},
	}, nil
}
