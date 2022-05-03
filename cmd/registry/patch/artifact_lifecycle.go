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

const LifecycleMimeType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.Lifecycle"

type LifecycleData struct {
	DisplayName string            `yaml:"displayName,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Stages      []*LifecycleStage `yaml:"stages"`
}

func (d *LifecycleData) mimeType() string {
	return LifecycleMimeType
}

func (d *LifecycleData) buildMessage() proto.Message {
	return &rpc.Lifecycle{
		DisplayName: d.DisplayName,
		Description: d.Description,
		Stages:      buildLifecycleStagesProto(d),
	}
}

func buildLifecycleArtifact(a *rpc.Artifact) (*Artifact, error) {
	artifactName, err := names.ParseArtifact(a.Name)
	if err != nil {
		return nil, err
	}
	m := &rpc.Lifecycle{}
	if err = proto.Unmarshal(a.Contents, m); err != nil {
		return nil, err
	}
	return &Artifact{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "Lifecycle",
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
		Data: &LifecycleData{
			DisplayName: m.DisplayName,
			Description: m.Description,
			Stages:      buildLifecycleStagesData(m),
		},
	}, nil
}

type LifecycleStage struct {
	ID           string `yaml:"id"`
	DisplayName  string `yaml:"displayName,omitempty"`
	Description  string `yaml:"description,omitempty"`
	URL          string `yaml:"url,omitempty"`
	DisplayOrder int    `yaml:"displayOrder"`
}

func buildLifecycleStagesProto(d *LifecycleData) []*rpc.Lifecycle_Stage {
	a := make([]*rpc.Lifecycle_Stage, len(d.Stages))
	for i, v := range d.Stages {
		a[i] = &rpc.Lifecycle_Stage{
			Id:           v.ID,
			DisplayName:  v.DisplayName,
			Description:  v.Description,
			Url:          v.URL,
			DisplayOrder: int32(v.DisplayOrder),
		}
	}
	return a
}

func buildLifecycleStagesData(m *rpc.Lifecycle) []*LifecycleStage {
	a := make([]*LifecycleStage, len(m.Stages))
	for i, v := range m.Stages {
		a[i] = &LifecycleStage{
			ID:           v.Id,
			DisplayName:  v.DisplayName,
			Description:  v.Description,
			URL:          v.Url,
			DisplayOrder: int(v.DisplayOrder),
		}
	}
	return a
}
