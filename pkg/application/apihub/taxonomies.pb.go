// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: google/cloud/apigeeregistry/v1/apihub/taxonomies.proto

// (-- api-linter: core::0215::versioned-packages=disabled
//     aip.dev/not-precedent: Support protos for the apigeeregistry.v1 API. --)

package apihub

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// A TaxonomyList message contains a list of taxonomies that can be used to
// classify resources in the registry. Typically all of the system-managed
// taxonomies would be stored in a registry as a single TaxonomyList artifact.
type TaxonomyList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Artifact identifier. May be used in YAML representations to indicate the id
	// to be used to attach the artifact.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Artifact kind. May be used in YAML representations to identify the type of
	// this artifact.
	Kind string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	// A human-friendly name for the taxonomy list.
	DisplayName string `protobuf:"bytes,3,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	// A more detailed description of the taxonomy list.
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	// The taxonomies in the list.
	Taxonomies []*TaxonomyList_Taxonomy `protobuf:"bytes,5,rep,name=taxonomies,proto3" json:"taxonomies,omitempty"`
}

func (x *TaxonomyList) Reset() {
	*x = TaxonomyList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaxonomyList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaxonomyList) ProtoMessage() {}

func (x *TaxonomyList) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaxonomyList.ProtoReflect.Descriptor instead.
func (*TaxonomyList) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescGZIP(), []int{0}
}

func (x *TaxonomyList) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TaxonomyList) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *TaxonomyList) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *TaxonomyList) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *TaxonomyList) GetTaxonomies() []*TaxonomyList_Taxonomy {
	if x != nil {
		return x.Taxonomies
	}
	return nil
}

// A Taxonomy specifies a list of values that can be associated with an item
// in a registry, typically an API. There may be multiple taxonomies, each
// representing a different aspect or dimension of the item being labelled.
type TaxonomyList_Taxonomy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Taxonomy identifier.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// A human-friendly name of the taxonomy.
	DisplayName string `protobuf:"bytes,2,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	// A detailed description of the taxonomy.
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	// If true, this taxonomy is applied by admins only.
	AdminApplied bool `protobuf:"varint,4,opt,name=admin_applied,json=adminApplied,proto3" json:"admin_applied,omitempty"`
	// If true, this taxonomy only allows one of its members to be associated
	// with an item (multiple selection is disallowed).
	SingleSelection bool `protobuf:"varint,5,opt,name=single_selection,json=singleSelection,proto3" json:"single_selection,omitempty"`
	// If true, this taxonomy is not included in search indexes.
	SearchExcluded bool `protobuf:"varint,6,opt,name=search_excluded,json=searchExcluded,proto3" json:"search_excluded,omitempty"`
	// If true, this taxonomy is a system-managed taxonomy.
	SystemManaged bool `protobuf:"varint,7,opt,name=system_managed,json=systemManaged,proto3" json:"system_managed,omitempty"`
	// An ordering value used to configure display of the taxonomy.
	DisplayOrder int32 `protobuf:"varint,8,opt,name=display_order,json=displayOrder,proto3" json:"display_order,omitempty"`
	// The elements of the taxonomy.
	Elements []*TaxonomyList_Taxonomy_Element `protobuf:"bytes,9,rep,name=elements,proto3" json:"elements,omitempty"`
}

func (x *TaxonomyList_Taxonomy) Reset() {
	*x = TaxonomyList_Taxonomy{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaxonomyList_Taxonomy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaxonomyList_Taxonomy) ProtoMessage() {}

func (x *TaxonomyList_Taxonomy) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaxonomyList_Taxonomy.ProtoReflect.Descriptor instead.
func (*TaxonomyList_Taxonomy) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescGZIP(), []int{0, 0}
}

func (x *TaxonomyList_Taxonomy) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TaxonomyList_Taxonomy) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *TaxonomyList_Taxonomy) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *TaxonomyList_Taxonomy) GetAdminApplied() bool {
	if x != nil {
		return x.AdminApplied
	}
	return false
}

func (x *TaxonomyList_Taxonomy) GetSingleSelection() bool {
	if x != nil {
		return x.SingleSelection
	}
	return false
}

func (x *TaxonomyList_Taxonomy) GetSearchExcluded() bool {
	if x != nil {
		return x.SearchExcluded
	}
	return false
}

func (x *TaxonomyList_Taxonomy) GetSystemManaged() bool {
	if x != nil {
		return x.SystemManaged
	}
	return false
}

func (x *TaxonomyList_Taxonomy) GetDisplayOrder() int32 {
	if x != nil {
		return x.DisplayOrder
	}
	return 0
}

func (x *TaxonomyList_Taxonomy) GetElements() []*TaxonomyList_Taxonomy_Element {
	if x != nil {
		return x.Elements
	}
	return nil
}

// An element in a taxonomy represents one of the values that can be used
// to label an item.
type TaxonomyList_Taxonomy_Element struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Element identifier.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// A human-friendly name of the element.
	DisplayName string `protobuf:"bytes,2,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	// A detailed description of the element.
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
}

func (x *TaxonomyList_Taxonomy_Element) Reset() {
	*x = TaxonomyList_Taxonomy_Element{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaxonomyList_Taxonomy_Element) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaxonomyList_Taxonomy_Element) ProtoMessage() {}

func (x *TaxonomyList_Taxonomy_Element) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaxonomyList_Taxonomy_Element.ProtoReflect.Descriptor instead.
func (*TaxonomyList_Taxonomy_Element) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescGZIP(), []int{0, 0, 0}
}

func (x *TaxonomyList_Taxonomy_Element) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TaxonomyList_Taxonomy_Element) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *TaxonomyList_Taxonomy_Element) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

var File_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto protoreflect.FileDescriptor

var file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDesc = []byte{
	0x0a, 0x36, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31,
	0x2f, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x2f, 0x74, 0x61, 0x78, 0x6f, 0x6e, 0x6f, 0x6d, 0x69,
	0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x25, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x22,
	0xbe, 0x05, 0x0a, 0x0c, 0x54, 0x61, 0x78, 0x6f, 0x6e, 0x6f, 0x6d, 0x79, 0x4c, 0x69, 0x73, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6b, 0x69, 0x6e, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x69, 0x73, 0x70,
	0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x5c, 0x0a, 0x0a, 0x74, 0x61, 0x78,
	0x6f, 0x6e, 0x6f, 0x6d, 0x69, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3c, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x61,
	0x70, 0x69, 0x68, 0x75, 0x62, 0x2e, 0x54, 0x61, 0x78, 0x6f, 0x6e, 0x6f, 0x6d, 0x79, 0x4c, 0x69,
	0x73, 0x74, 0x2e, 0x54, 0x61, 0x78, 0x6f, 0x6e, 0x6f, 0x6d, 0x79, 0x52, 0x0a, 0x74, 0x61, 0x78,
	0x6f, 0x6e, 0x6f, 0x6d, 0x69, 0x65, 0x73, 0x1a, 0xe6, 0x03, 0x0a, 0x08, 0x54, 0x61, 0x78, 0x6f,
	0x6e, 0x6f, 0x6d, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x69, 0x73, 0x70,
	0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x5f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x0c, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x65, 0x64, 0x12, 0x29,
	0x0a, 0x10, 0x73, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x5f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x73, 0x69, 0x6e, 0x67, 0x6c, 0x65,
	0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x27, 0x0a, 0x0f, 0x73, 0x65, 0x61,
	0x72, 0x63, 0x68, 0x5f, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x45, 0x78, 0x63, 0x6c, 0x75, 0x64,
	0x65, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x5f, 0x6d, 0x61, 0x6e,
	0x61, 0x67, 0x65, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x73, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x64, 0x69, 0x73,
	0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x60,
	0x0a, 0x08, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x44, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76,
	0x31, 0x2e, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x2e, 0x54, 0x61, 0x78, 0x6f, 0x6e, 0x6f, 0x6d,
	0x79, 0x4c, 0x69, 0x73, 0x74, 0x2e, 0x54, 0x61, 0x78, 0x6f, 0x6e, 0x6f, 0x6d, 0x79, 0x2e, 0x45,
	0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x08, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73,
	0x1a, 0x5e, 0x0a, 0x07, 0x45, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x64,
	0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20,
	0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x42, 0x82, 0x01, 0x0a, 0x29, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x42, 0x19,
	0x41, 0x70, 0x69, 0x48, 0x75, 0x62, 0x54, 0x61, 0x78, 0x6f, 0x6e, 0x6f, 0x6d, 0x79, 0x4d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x38, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x2f, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x3b, 0x61,
	0x70, 0x69, 0x68, 0x75, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescOnce sync.Once
	file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescData = file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDesc
)

func file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescGZIP() []byte {
	file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescOnce.Do(func() {
		file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescData)
	})
	return file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDescData
}

var file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_goTypes = []interface{}{
	(*TaxonomyList)(nil),                  // 0: google.cloud.apigeeregistry.v1.apihub.TaxonomyList
	(*TaxonomyList_Taxonomy)(nil),         // 1: google.cloud.apigeeregistry.v1.apihub.TaxonomyList.Taxonomy
	(*TaxonomyList_Taxonomy_Element)(nil), // 2: google.cloud.apigeeregistry.v1.apihub.TaxonomyList.Taxonomy.Element
}
var file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_depIdxs = []int32{
	1, // 0: google.cloud.apigeeregistry.v1.apihub.TaxonomyList.taxonomies:type_name -> google.cloud.apigeeregistry.v1.apihub.TaxonomyList.Taxonomy
	2, // 1: google.cloud.apigeeregistry.v1.apihub.TaxonomyList.Taxonomy.elements:type_name -> google.cloud.apigeeregistry.v1.apihub.TaxonomyList.Taxonomy.Element
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_init() }
func file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_init() {
	if File_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaxonomyList); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaxonomyList_Taxonomy); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaxonomyList_Taxonomy_Element); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_goTypes,
		DependencyIndexes: file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_depIdxs,
		MessageInfos:      file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_msgTypes,
	}.Build()
	File_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto = out.File
	file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_rawDesc = nil
	file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_goTypes = nil
	file_google_cloud_apigeeregistry_v1_apihub_taxonomies_proto_depIdxs = nil
}