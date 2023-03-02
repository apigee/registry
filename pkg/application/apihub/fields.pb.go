// Copyright 2023 Google LLC
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
// source: google/cloud/apigeeregistry/v1/apihub/fields.proto

// (-- api-linter: core::0215::versioned-packages=disabled
//     aip.dev/not-precedent: Support protos for the apigeeregistry.v1 API. --)

package apihub

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// Defines a structure for general field storage.
type FieldSet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Artifact identifier. May be used in YAML representations to indicate the id
	// to be used to attach the artifact.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Artifact kind. May be used in YAML representations to identify the type of
	// this artifact.
	Kind string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	// Full resource name of an FieldSetDefinition artifact that
	// describes this field set.
	DefinitionName string `protobuf:"bytes,3,opt,name=definition_name,json=definitionName,proto3" json:"definition_name,omitempty"`
	// The field values, stored using field ids as keys.
	Values map[string]string `protobuf:"bytes,4,rep,name=values,proto3" json:"values,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *FieldSet) Reset() {
	*x = FieldSet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FieldSet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldSet) ProtoMessage() {}

func (x *FieldSet) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldSet.ProtoReflect.Descriptor instead.
func (*FieldSet) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescGZIP(), []int{0}
}

func (x *FieldSet) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *FieldSet) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *FieldSet) GetDefinitionName() string {
	if x != nil {
		return x.DefinitionName
	}
	return ""
}

func (x *FieldSet) GetValues() map[string]string {
	if x != nil {
		return x.Values
	}
	return nil
}

type FieldSetDefinition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Artifact identifier. May be used in YAML representations to indicate the id
	// to be used to attach the artifact.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Artifact kind. May be used in YAML representations to identify the type of
	// this artifact.
	Kind string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	// A short displayable name for the field set definition.
	DisplayName string `protobuf:"bytes,3,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	// A description of the field set being defined.
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	// The field definitions
	Fields []*FieldDefinition `protobuf:"bytes,5,rep,name=fields,proto3" json:"fields,omitempty"`
}

func (x *FieldSetDefinition) Reset() {
	*x = FieldSetDefinition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FieldSetDefinition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldSetDefinition) ProtoMessage() {}

func (x *FieldSetDefinition) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldSetDefinition.ProtoReflect.Descriptor instead.
func (*FieldSetDefinition) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescGZIP(), []int{1}
}

func (x *FieldSetDefinition) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *FieldSetDefinition) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *FieldSetDefinition) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *FieldSetDefinition) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *FieldSetDefinition) GetFields() []*FieldDefinition {
	if x != nil {
		return x.Fields
	}
	return nil
}

type FieldDefinition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The id of the field, used as a key in the fields map.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The display_name of the field, used when the field is displayed.
	DisplayName string `protobuf:"bytes,2,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	// A description of the field, possibly displayable as a tooltip.
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	// Optional string specifying the field format.
	// Currently applications are free to define values,
	// but we expect to formalize values for this in the future.
	Format string `protobuf:"bytes,4,opt,name=format,proto3" json:"format,omitempty"`
	// Optional list of allowed values for the field.
	AllowedValues []string `protobuf:"bytes,5,rep,name=allowed_values,json=allowedValues,proto3" json:"allowed_values,omitempty"`
}

func (x *FieldDefinition) Reset() {
	*x = FieldDefinition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FieldDefinition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldDefinition) ProtoMessage() {}

func (x *FieldDefinition) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldDefinition.ProtoReflect.Descriptor instead.
func (*FieldDefinition) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescGZIP(), []int{2}
}

func (x *FieldDefinition) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *FieldDefinition) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *FieldDefinition) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *FieldDefinition) GetFormat() string {
	if x != nil {
		return x.Format
	}
	return ""
}

func (x *FieldDefinition) GetAllowedValues() []string {
	if x != nil {
		return x.AllowedValues
	}
	return nil
}

var File_google_cloud_apigeeregistry_v1_apihub_fields_proto protoreflect.FileDescriptor

var file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDesc = []byte{
	0x0a, 0x32, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31,
	0x2f, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x25, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x1a, 0x1f, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65,
	0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xec, 0x01, 0x0a,
	0x08, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x53, 0x65, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x2c, 0x0a,
	0x0f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0e, 0x64, 0x65, 0x66,
	0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x53, 0x0a, 0x06, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3b, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65,
	0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x70, 0x69,
	0x68, 0x75, 0x62, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x53, 0x65, 0x74, 0x2e, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73,
	0x1a, 0x39, 0x0a, 0x0b, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xcd, 0x01, 0x0a, 0x12,
	0x46, 0x69, 0x65, 0x6c, 0x64, 0x53, 0x65, 0x74, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61,
	0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x69,
	0x73, 0x70, 0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x4e, 0x0a, 0x06, 0x66,
	0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x36, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65,
	0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x70, 0x69,
	0x68, 0x75, 0x62, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x22, 0xa5, 0x01, 0x0a, 0x0f,
	0x46, 0x69, 0x65, 0x6c, 0x64, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x25, 0x0a, 0x0e,
	0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x05,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x0d, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x42, 0x76, 0x0a, 0x29, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62,
	0x42, 0x0d, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x53, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50,
	0x01, 0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70,
	0x69, 0x67, 0x65, 0x65, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x61, 0x70,
	0x69, 0x68, 0x75, 0x62, 0x3b, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescOnce sync.Once
	file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescData = file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDesc
)

func file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescGZIP() []byte {
	file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescOnce.Do(func() {
		file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescData)
	})
	return file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDescData
}

var file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_google_cloud_apigeeregistry_v1_apihub_fields_proto_goTypes = []interface{}{
	(*FieldSet)(nil),           // 0: google.cloud.apigeeregistry.v1.apihub.FieldSet
	(*FieldSetDefinition)(nil), // 1: google.cloud.apigeeregistry.v1.apihub.FieldSetDefinition
	(*FieldDefinition)(nil),    // 2: google.cloud.apigeeregistry.v1.apihub.FieldDefinition
	nil,                        // 3: google.cloud.apigeeregistry.v1.apihub.FieldSet.ValuesEntry
}
var file_google_cloud_apigeeregistry_v1_apihub_fields_proto_depIdxs = []int32{
	3, // 0: google.cloud.apigeeregistry.v1.apihub.FieldSet.values:type_name -> google.cloud.apigeeregistry.v1.apihub.FieldSet.ValuesEntry
	2, // 1: google.cloud.apigeeregistry.v1.apihub.FieldSetDefinition.fields:type_name -> google.cloud.apigeeregistry.v1.apihub.FieldDefinition
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_google_cloud_apigeeregistry_v1_apihub_fields_proto_init() }
func file_google_cloud_apigeeregistry_v1_apihub_fields_proto_init() {
	if File_google_cloud_apigeeregistry_v1_apihub_fields_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FieldSet); i {
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
		file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FieldSetDefinition); i {
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
		file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FieldDefinition); i {
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
			RawDescriptor: file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_apigeeregistry_v1_apihub_fields_proto_goTypes,
		DependencyIndexes: file_google_cloud_apigeeregistry_v1_apihub_fields_proto_depIdxs,
		MessageInfos:      file_google_cloud_apigeeregistry_v1_apihub_fields_proto_msgTypes,
	}.Build()
	File_google_cloud_apigeeregistry_v1_apihub_fields_proto = out.File
	file_google_cloud_apigeeregistry_v1_apihub_fields_proto_rawDesc = nil
	file_google_cloud_apigeeregistry_v1_apihub_fields_proto_goTypes = nil
	file_google_cloud_apigeeregistry_v1_apihub_fields_proto_depIdxs = nil
}
