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
	"context"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

const TaxonomyListMimeType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.TaxonomyList"

type TaxonomyListData struct {
	DisplayName string     `yaml:"displayName,omitempty"`
	Description string     `yaml:"description,omitempty"`
	Taxonomies  []Taxonomy `yaml:"taxonomies"`
}

type TaxonomyList struct {
	Header `yaml:",inline"`
	Data   TaxonomyListData `yaml:"data"`
}

func (a *TaxonomyList) GetMimeType() string {
	return TaxonomyListMimeType
}

func (a *TaxonomyList) GetHeader() *Header {
	return &a.Header
}

type Taxonomy struct {
	ID              string            `yaml:"id"`
	DisplayName     string            `yaml:"displayName,omitempty"`
	Description     string            `yaml:"description,omitempty"`
	AdminApplied    bool              `yaml:"adminApplied,omitempty"`
	SingleSelection bool              `yaml:"singleSelection,omitempty"`
	SearchExcluded  bool              `yaml:"searchExcluded,omitempty"`
	SystemManaged   bool              `yaml:"systemManaged,omitempty"`
	DisplayOrder    int               `yaml:"displayOrder"`
	Elements        []TaxonomyElement `yaml:"elements"`
}

type TaxonomyElement struct {
	ID          string `yaml:"id"`
	DisplayName string `yaml:"displayName,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Message returns the rpc representation of the taxonomies.
func (l *TaxonomyList) GetMessage() proto.Message {
	return &rpc.TaxonomyList{
		Id:          l.Header.Metadata.Name,
		Kind:        TaxonomyListMimeType,
		DisplayName: l.Data.DisplayName,
		Description: l.Data.Description,
		Taxonomies:  l.taxonomies(),
	}
}

func (l *TaxonomyList) taxonomies() []*rpc.TaxonomyList_Taxonomy {
	taxonomies := make([]*rpc.TaxonomyList_Taxonomy, 0)
	for _, t := range l.Data.Taxonomies {
		taxonomies = append(taxonomies,
			&rpc.TaxonomyList_Taxonomy{
				Id:              t.ID,
				DisplayName:     t.DisplayName,
				Description:     t.Description,
				AdminApplied:    t.AdminApplied,
				SingleSelection: t.SingleSelection,
				SearchExcluded:  t.SearchExcluded,
				SystemManaged:   t.SystemManaged,
				DisplayOrder:    int32(t.DisplayOrder),
				Elements:        t.elements(),
			},
		)
	}
	return taxonomies
}

func (t *Taxonomy) elements() []*rpc.TaxonomyList_Taxonomy_Element {
	elements := make([]*rpc.TaxonomyList_Taxonomy_Element, 0)
	for _, e := range t.Elements {
		elements = append(elements, &rpc.TaxonomyList_Taxonomy_Element{
			Id:          e.ID,
			DisplayName: e.DisplayName,
			Description: e.Description,
		})
	}
	return elements
}

// newTaxonomyList creates a TaxonomyList object from an rpc representation.
func newTaxonomyList(message *rpc.Artifact) (*TaxonomyList, error) {
	artifactName, err := names.ParseArtifact(message.Name)
	if err != nil {
		return nil, err
	}
	value := &rpc.TaxonomyList{}
	err = proto.Unmarshal(message.Contents, value)
	if err != nil {
		return nil, err
	}
	taxonomyList := &TaxonomyList{
		Header: Header{
			ApiVersion: RegistryV1,
			Kind:       "TaxonomyList",
			Metadata: Metadata{
				Name: artifactName.ArtifactID(),
			},
		},
		Data: TaxonomyListData{
			DisplayName: value.DisplayName,
			Description: value.Description,
		},
	}
	for _, t := range value.Taxonomies {
		elements := make([]TaxonomyElement, 0)
		for _, e := range t.Elements {
			elements = append(elements,
				TaxonomyElement{
					ID:          e.Id,
					DisplayName: e.DisplayName,
					Description: e.Description,
				})
		}
		taxonomy := Taxonomy{
			ID:              t.Id,
			DisplayName:     t.DisplayName,
			Description:     t.Description,
			AdminApplied:    t.AdminApplied,
			SingleSelection: t.SingleSelection,
			SearchExcluded:  t.SearchExcluded,
			SystemManaged:   t.SystemManaged,
			DisplayOrder:    int(t.DisplayOrder),
			Elements:        elements,
		}
		taxonomyList.Data.Taxonomies = append(taxonomyList.Data.Taxonomies, taxonomy)
	}
	return taxonomyList, nil
}

func applyTaxonomyListArtifactPatch(ctx context.Context, client connection.Client, bytes []byte, parent string) error {
	var taxonomyList TaxonomyList
	err := yaml.Unmarshal(bytes, &taxonomyList)
	if err != nil {
		return err
	}
	return applyArtifactPatch(ctx, client, &taxonomyList, parent)
}
