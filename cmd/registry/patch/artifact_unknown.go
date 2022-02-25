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

const UnknownArtifactMimeType = "application/octet-stream"

type UnknownArtifact struct {
	Header `yaml:",inline"`
	Data   struct {
		MimeType string `yaml:"mimeType,omitempty"`
	} `yaml:"data"`
}

func (a *UnknownArtifact) GetMimeType() string {
	return UnknownArtifactMimeType
}

func (a *UnknownArtifact) GetHeader() *Header {
	return &a.Header
}

func (a *UnknownArtifact) GetMessage() proto.Message {
	return nil
}

func newUnknownArtifact(message *rpc.Artifact) (*UnknownArtifact, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	artifact := &UnknownArtifact{
		Header: Header{
			APIVersion: RegistryV1,
			Kind:       "Artifact",
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
	}
	return artifact, nil
}
