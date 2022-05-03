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

type ManifestData struct {
	DisplayName        string                       `yaml:"displayName,omitempty"`
	Description        string                       `yaml:"description,omitempty"`
	GeneratedResources []*ManifestGeneratedResource `yaml:"generatedResources"`
}

func (d *ManifestData) mimeType() string {
	return ManifestMimeType
}

func (d *ManifestData) buildMessage() proto.Message {
	return &rpc.Manifest{
		DisplayName:        d.DisplayName,
		Description:        d.Description,
		GeneratedResources: buildManifestGeneratedResourcesProto(d),
	}
}

func buildManifestArtifact(a *rpc.Artifact) (*Artifact, error) {
	artifactName, err := names.ParseArtifact(a.Name)
	if err != nil {
		return nil, err
	}
	m := &rpc.Manifest{}
	if err = proto.Unmarshal(a.Contents, m); err != nil {
		return nil, err
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
			DisplayName:        m.DisplayName,
			Description:        m.Description,
			GeneratedResources: buildManifestGeneratedResourcesData(m),
		},
	}, nil
}

type ManifestGeneratedResource struct {
	Pattern      string                `yaml:"pattern"`
	Filter       string                `yaml:"filter,omitempty"`
	Receipt      bool                  `yaml:"receipt,omitempty"`
	Dependencies []*ManifestDependency `yaml:"dependencies"`
	Action       string                `yaml:"action"`
}

func buildManifestGeneratedResourcesProto(d *ManifestData) []*rpc.GeneratedResource {
	a := make([]*rpc.GeneratedResource, len(d.GeneratedResources))
	for i, v := range d.GeneratedResources {
		a[i] = &rpc.GeneratedResource{
			Pattern:      v.Pattern,
			Filter:       v.Filter,
			Receipt:      v.Receipt,
			Dependencies: buildDependenciesProto(v),
			Action:       v.Action,
		}
	}
	return a
}

func buildManifestGeneratedResourcesData(m *rpc.Manifest) []*ManifestGeneratedResource {
	a := make([]*ManifestGeneratedResource, len(m.GeneratedResources))
	for i, v := range m.GeneratedResources {
		a[i] = &ManifestGeneratedResource{
			Pattern:      v.Pattern,
			Filter:       v.Filter,
			Receipt:      v.Receipt,
			Dependencies: buildDependenciesData(v),
			Action:       v.Action,
		}
	}
	return a
}

type ManifestDependency struct {
	Pattern string `yaml:"pattern"`
	Filter  string `yaml:"filter,omitempty"`
}

func buildDependenciesProto(d *ManifestGeneratedResource) []*rpc.Dependency {
	a := make([]*rpc.Dependency, len(d.Dependencies))
	for i, v := range d.Dependencies {
		a[i] = &rpc.Dependency{
			Pattern: v.Pattern,
			Filter:  v.Filter,
		}
	}
	return a
}

func buildDependenciesData(m *rpc.GeneratedResource) []*ManifestDependency {
	a := make([]*ManifestDependency, len(m.Dependencies))
	for i, v := range m.Dependencies {
		a[i] = &ManifestDependency{
			Pattern: v.Pattern,
			Filter:  v.Filter,
		}
	}
	return a
}
