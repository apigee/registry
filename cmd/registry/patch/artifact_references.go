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
	"google.golang.org/protobuf/proto"
)

const ReferenceListMimeType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.ReferenceList"

type ReferenceListData struct {
	DisplayName string       `yaml:"displayName,omitempty"`
	Description string       `yaml:"description,omitempty"`
	References  []*Reference `yaml:"references"`
}

func (d *ReferenceListData) mimeType() string {
	return ReferenceListMimeType
}

func (d *ReferenceListData) buildMessage() proto.Message {
	return &rpc.ReferenceList{
		DisplayName: d.DisplayName,
		Description: d.Description,
		References:  buildReferenceListReferencesProto(d),
	}
}

type Reference struct {
	ID          string `yaml:"id"`
	DisplayName string `yaml:"displayName,omitempty"`
	Category    string `yaml:"category,omitempty"`
	Resource    string `yaml:"resource,omitempty"`
	URI         string `uri:"displayName,omitempty"`
}

func buildReferenceListReferencesProto(d *ReferenceListData) []*rpc.ReferenceList_Reference {
	a := make([]*rpc.ReferenceList_Reference, len(d.References))
	for i, v := range d.References {
		a[i] = &rpc.ReferenceList_Reference{
			Id:          v.ID,
			DisplayName: v.DisplayName,
			Category:    v.Category,
			Resource:    v.Resource,
			Uri:         v.URI,
		}
	}
	return a
}
