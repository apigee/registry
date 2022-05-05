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

const TaxonomyListMimeType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.TaxonomyList"

type TaxonomyListData struct {
	DisplayName string      `yaml:"displayName,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Taxonomies  []*Taxonomy `yaml:"taxonomies"`
}

func (d *TaxonomyListData) mimeType() string {
	return TaxonomyListMimeType
}

func (d *TaxonomyListData) buildMessage() proto.Message {
	return &rpc.TaxonomyList{
		DisplayName: d.DisplayName,
		Description: d.Description,
		Taxonomies:  buildTaxonomiesProto(d),
	}
}

type Taxonomy struct {
	ID              string             `yaml:"id"`
	DisplayName     string             `yaml:"displayName,omitempty"`
	Description     string             `yaml:"description,omitempty"`
	AdminApplied    bool               `yaml:"adminApplied,omitempty"`
	SingleSelection bool               `yaml:"singleSelection,omitempty"`
	SearchExcluded  bool               `yaml:"searchExcluded,omitempty"`
	SystemManaged   bool               `yaml:"systemManaged,omitempty"`
	DisplayOrder    int                `yaml:"displayOrder"`
	Elements        []*TaxonomyElement `yaml:"elements"`
}

func buildTaxonomiesProto(d *TaxonomyListData) []*rpc.TaxonomyList_Taxonomy {
	a := make([]*rpc.TaxonomyList_Taxonomy, len(d.Taxonomies))
	for i, v := range d.Taxonomies {
		a[i] = &rpc.TaxonomyList_Taxonomy{
			Id:              v.ID,
			DisplayName:     v.DisplayName,
			Description:     v.Description,
			AdminApplied:    v.AdminApplied,
			SingleSelection: v.SingleSelection,
			SearchExcluded:  v.SearchExcluded,
			SystemManaged:   v.SystemManaged,
			DisplayOrder:    int32(v.DisplayOrder),
			Elements:        buildTaxonomyElementsProto(v),
		}
	}
	return a
}

type TaxonomyElement struct {
	ID          string `yaml:"id"`
	DisplayName string `yaml:"displayName,omitempty"`
	Description string `yaml:"description,omitempty"`
}

func buildTaxonomyElementsProto(t *Taxonomy) []*rpc.TaxonomyList_Taxonomy_Element {
	a := make([]*rpc.TaxonomyList_Taxonomy_Element, len(t.Elements))
	for i, v := range t.Elements {
		a[i] = &rpc.TaxonomyList_Taxonomy_Element{
			Id:          v.ID,
			DisplayName: v.DisplayName,
			Description: v.Description,
		}
	}
	return a
}
