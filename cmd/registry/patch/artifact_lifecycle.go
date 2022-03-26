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

type LifecycleStage struct {
	ID           string `yaml:"id"`
	DisplayName  string `yaml:"displayName,omitempty"`
	Description  string `yaml:"description,omitempty"`
	URL          string `yaml:"url,omitempty"`
	DisplayOrder int    `yaml:"displayOrder"`
}

type LifecycleData struct {
	DisplayName string            `yaml:"displayName,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Stages      []*LifecycleStage `yaml:"stages"`
}

func (a *LifecycleData) GetMimeType() string {
	return LifecycleMimeType
}

func (l *LifecycleData) GetMessage() proto.Message {
	return &rpc.Lifecycle{
		DisplayName: l.DisplayName,
		Description: l.Description,
		Stages:      l.stages(),
	}
}

func (l *LifecycleData) stages() []*rpc.Lifecycle_Stage {
	stages := make([]*rpc.Lifecycle_Stage, 0)
	for _, s := range l.Stages {
		stages = append(stages, &rpc.Lifecycle_Stage{
			Id:           s.ID,
			DisplayName:  s.DisplayName,
			Description:  s.Description,
			Url:          s.URL,
			DisplayOrder: int32(s.DisplayOrder),
		})
	}
	return stages
}

func newLifecycle(message *rpc.Artifact) (*Artifact, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	value := &rpc.Lifecycle{}
	err = proto.Unmarshal(message.Contents, value)
	if err != nil {
		return nil, err
	}
	lifecycleData := &LifecycleData{
		DisplayName: value.DisplayName,
		Description: value.Description,
	}
	for _, s := range value.Stages {
		lifecycleData.Stages = append(
			lifecycleData.Stages,
			&LifecycleStage{
				ID:           s.Id,
				DisplayName:  s.DisplayName,
				Description:  s.Description,
				URL:          s.Url,
				DisplayOrder: int(s.DisplayOrder),
			})
	}
	lifecycle := &Artifact{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "Lifecycle",
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
		Data: lifecycleData,
	}
	return lifecycle, nil
}
