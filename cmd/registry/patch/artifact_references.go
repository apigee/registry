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

const ReferenceListMimeType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.ReferenceList"

type ReferenceListData struct {
	DisplayName string      `yaml:"displayName,omitempty"`
	Description string      `yaml:"description,omitempty"`
	References  []Reference `yaml:"references"`
}

func (l *ReferenceListData) GetMimeType() string {
	return ReferenceListMimeType
}

type Reference struct {
	ID          string `yaml:"id"`
	DisplayName string `yaml:"displayName,omitempty"`
	Category    string `yaml:"category,omitempty"`
	Resource    string `yaml:"resource,omitempty"`
	URI         string `uri:"displayName,omitempty"`
}

func (l *ReferenceListData) GetMessage() proto.Message {
	return &rpc.ReferenceList{
		DisplayName: l.DisplayName,
		Description: l.Description,
		References:  l.references(),
	}
}

func (l *ReferenceListData) references() []*rpc.ReferenceList_Reference {
	references := make([]*rpc.ReferenceList_Reference, 0)
	for _, t := range l.References {
		references = append(references,
			&rpc.ReferenceList_Reference{
				Id:          t.ID,
				DisplayName: t.DisplayName,
				Category:    t.Category,
				Resource:    t.Resource,
				Uri:         t.URI,
			},
		)
	}
	return references
}

func newReferenceList(message *rpc.Artifact) (*Artifact, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	value := &rpc.ReferenceList{}
	err = proto.Unmarshal(message.Contents, value)
	if err != nil {
		return nil, err
	}
	referenceListData := &ReferenceListData{
		DisplayName: value.DisplayName,
		Description: value.Description,
	}
	referenceList := &Artifact{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "ReferenceList",
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
		Data: referenceListData,
	}
	for _, t := range value.References {
		reference := Reference{
			ID:          t.Id,
			DisplayName: t.DisplayName,
			Category:    t.Category,
			Resource:    t.Resource,
			URI:         t.Uri,
		}
		referenceListData.References = append(referenceListData.References, reference)
	}
	return referenceList, nil
}
