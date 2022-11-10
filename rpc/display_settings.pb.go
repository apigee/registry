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
// source: google/cloud/apigeeregistry/v1/apihub/display_settings.proto

// (-- api-linter: core::0215::versioned-packages=disabled
//     aip.dev/not-precedent: Support protos for the apigeeregistry.v1 API. --)

package rpc

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

// Defines display settings for the API hub UI.
type DisplaySettings struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Artifact identifier. May be used in YAML representations to indicate the id
	// to be used to attach the artifact.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Artifact kind. May be used in YAML representations to identify the type of
	// this artifact.
	Kind string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	// A more detailed description of the display settings.
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	// The organization name. Displayed in areas throughout the API hub UI to
	// identify APIs owned by the user's organization.
	Organization string `protobuf:"bytes,4,opt,name=organization,proto3" json:"organization,omitempty"`
	// If true the API guide tab will be displayed in the API detail page; if
	// false, it will be hidden
	ApiGuideEnabled bool `protobuf:"varint,5,opt,name=api_guide_enabled,json=apiGuideEnabled,proto3" json:"api_guide_enabled,omitempty"`
	// If true the API scores will be displayed on the API list page and API
	// detail page; if false, they will be hidden
	ApiScoreEnabled bool `protobuf:"varint,6,opt,name=api_score_enabled,json=apiScoreEnabled,proto3" json:"api_score_enabled,omitempty"`
}

func (x *DisplaySettings) Reset() {
	*x = DisplaySettings{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DisplaySettings) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DisplaySettings) ProtoMessage() {}

func (x *DisplaySettings) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DisplaySettings.ProtoReflect.Descriptor instead.
func (*DisplaySettings) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDescGZIP(), []int{0}
}

func (x *DisplaySettings) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *DisplaySettings) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *DisplaySettings) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *DisplaySettings) GetOrganization() string {
	if x != nil {
		return x.Organization
	}
	return ""
}

func (x *DisplaySettings) GetApiGuideEnabled() bool {
	if x != nil {
		return x.ApiGuideEnabled
	}
	return false
}

func (x *DisplaySettings) GetApiScoreEnabled() bool {
	if x != nil {
		return x.ApiScoreEnabled
	}
	return false
}

var File_google_cloud_apigeeregistry_v1_apihub_display_settings_proto protoreflect.FileDescriptor

var file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDesc = []byte{
	0x0a, 0x3c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31,
	0x2f, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x2f, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f,
	0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x25,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x61,
	0x70, 0x69, 0x68, 0x75, 0x62, 0x22, 0xd3, 0x01, 0x0a, 0x0f, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61,
	0x79, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x20, 0x0a,
	0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x22, 0x0a, 0x0c, 0x6f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x2a, 0x0a, 0x11, 0x61, 0x70, 0x69, 0x5f, 0x67, 0x75, 0x69, 0x64, 0x65,
	0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f,
	0x61, 0x70, 0x69, 0x47, 0x75, 0x69, 0x64, 0x65, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12,
	0x2a, 0x0a, 0x11, 0x61, 0x70, 0x69, 0x5f, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x5f, 0x65, 0x6e, 0x61,
	0x62, 0x6c, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x61, 0x70, 0x69, 0x53,
	0x63, 0x6f, 0x72, 0x65, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x42, 0x67, 0x0a, 0x29, 0x63,
	0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76,
	0x31, 0x2e, 0x61, 0x70, 0x69, 0x68, 0x75, 0x62, 0x42, 0x14, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61,
	0x79, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x22, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x72, 0x70, 0x63,
	0x3b, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDescOnce sync.Once
	file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDescData = file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDesc
)

func file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDescGZIP() []byte {
	file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDescOnce.Do(func() {
		file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDescData)
	})
	return file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDescData
}

var file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_goTypes = []interface{}{
	(*DisplaySettings)(nil), // 0: google.cloud.apigeeregistry.v1.apihub.DisplaySettings
}
var file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_init() }
func file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_init() {
	if File_google_cloud_apigeeregistry_v1_apihub_display_settings_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DisplaySettings); i {
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
			RawDescriptor: file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_goTypes,
		DependencyIndexes: file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_depIdxs,
		MessageInfos:      file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_msgTypes,
	}.Build()
	File_google_cloud_apigeeregistry_v1_apihub_display_settings_proto = out.File
	file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_rawDesc = nil
	file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_goTypes = nil
	file_google_cloud_apigeeregistry_v1_apihub_display_settings_proto_depIdxs = nil
}
