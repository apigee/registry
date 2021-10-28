// Copyright 2021 Google LLC
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
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.1
// source: google/cloud/apigeeregistry/v1/controller/manifest.proto

package rpc

import (
	reflect "reflect"
	sync "sync"

	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// A Manifest represents a list of resources in a registry that should be
// automatically generated and updated in response to changes to their
// dependencies.
type Manifest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Artifact identifier. May be used in YAML representations to indicate the id
	// to be used to attach the artifact.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Artifact kind. May be used in YAML representations to identify the type of
	// this artifact.
	Kind string `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	// A human-friendly name for the manifest.
	DisplayName string `protobuf:"bytes,3,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	// A more detailed description of the manifest.
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	// List of Generated resources.
	GeneratedResources []*GeneratedResource `protobuf:"bytes,5,rep,name=generated_resources,json=generatedResources,proto3" json:"generated_resources,omitempty"`
}

func (x *Manifest) Reset() {
	*x = Manifest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Manifest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Manifest) ProtoMessage() {}

func (x *Manifest) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Manifest.ProtoReflect.Descriptor instead.
func (*Manifest) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescGZIP(), []int{0}
}

func (x *Manifest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Manifest) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *Manifest) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *Manifest) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Manifest) GetGeneratedResources() []*GeneratedResource {
	if x != nil {
		return x.GeneratedResources
	}
	return nil
}

// A GeneratedResource describes a resource that is stored in the
// registry and generated automatically using a specified action.
// Actions include invocations of the registry tool and other tools
// available to a deployed instance of the controller.
type GeneratedResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A pattern that specifies a generated resource.
	// This can specify one particular resource or a group of resources.
	// Format:
	//   apis/{api}/versions/{version}/specs/{spec}/artifacts/{artifact}
	//   apis/-/versions/-/specs/-/artifacts/-
	Pattern string `protobuf:"bytes,1,opt,name=pattern,proto3" json:"pattern,omitempty"`
	// A filter expression that limits the resources that match the pattern.
	Filter string `protobuf:"bytes,2,opt,name=filter,proto3" json:"filter,omitempty"`
	// The receipt field should be set to true if the action is not going to
	// store any artifact and instead requires a receipt artifact as its
	// generated resource.
	Receipt bool `protobuf:"varint,3,opt,name=receipt,proto3" json:"receipt,omitempty"`
	// The dependencies of the resource.
	Dependencies []*Dependency `protobuf:"bytes,4,rep,name=dependencies,proto3" json:"dependencies,omitempty"`
	// The action used to generate the resource.
	// An action can contain references to both the resource and dependencies
	// Example: "registry compute complexity $dependency0 $resource"
	Action string `protobuf:"bytes,5,opt,name=action,proto3" json:"action,omitempty"`
}

func (x *GeneratedResource) Reset() {
	*x = GeneratedResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeneratedResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeneratedResource) ProtoMessage() {}

func (x *GeneratedResource) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeneratedResource.ProtoReflect.Descriptor instead.
func (*GeneratedResource) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescGZIP(), []int{1}
}

func (x *GeneratedResource) GetPattern() string {
	if x != nil {
		return x.Pattern
	}
	return ""
}

func (x *GeneratedResource) GetFilter() string {
	if x != nil {
		return x.Filter
	}
	return ""
}

func (x *GeneratedResource) GetReceipt() bool {
	if x != nil {
		return x.Receipt
	}
	return false
}

func (x *GeneratedResource) GetDependencies() []*Dependency {
	if x != nil {
		return x.Dependencies
	}
	return nil
}

func (x *GeneratedResource) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

// A dependency of a generated resource is another resource in the registry
// which should always be older than the generated resource. When dependencies
// are updated, the generated resource that depends on them should be
// regenerated.
type Dependency struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A pattern that specifies a dependency.
	// This can specify one particular resource or a group of resources.
	// A pattern in a dependency can contain references to the original resource.
	// Format:
	//   $resource.api/versions/-/specs/-
	//   $resource.version/specs/-/artifacts/-
	Pattern string `protobuf:"bytes,1,opt,name=pattern,proto3" json:"pattern,omitempty"`
	// A filter expression that limits the resources that match the pattern.
	Filter string `protobuf:"bytes,2,opt,name=filter,proto3" json:"filter,omitempty"`
}

func (x *Dependency) Reset() {
	*x = Dependency{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Dependency) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Dependency) ProtoMessage() {}

func (x *Dependency) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Dependency.ProtoReflect.Descriptor instead.
func (*Dependency) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescGZIP(), []int{2}
}

func (x *Dependency) GetPattern() string {
	if x != nil {
		return x.Pattern
	}
	return ""
}

func (x *Dependency) GetFilter() string {
	if x != nil {
		return x.Filter
	}
	return ""
}

var File_google_cloud_apigeeregistry_v1_controller_manifest_proto protoreflect.FileDescriptor

var file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDesc = []byte{
	0x0a, 0x38, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31,
	0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x6d, 0x61, 0x6e, 0x69,
	0x66, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x29, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe7, 0x01, 0x0a, 0x08, 0x4d, 0x61, 0x6e, 0x69, 0x66,
	0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c,
	0x61, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64,
	0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x72, 0x0a, 0x13,
	0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3c, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x12, 0x67, 0x65,
	0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x22, 0xdc, 0x01, 0x0a, 0x11, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x1d, 0x0a, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x07, 0x70, 0x61,
	0x74, 0x74, 0x65, 0x72, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x18, 0x0a,
	0x07, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07,
	0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x12, 0x59, 0x0a, 0x0c, 0x64, 0x65, 0x70, 0x65, 0x6e,
	0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x35, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x63,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x44, 0x65, 0x70, 0x65, 0x6e, 0x64,
	0x65, 0x6e, 0x63, 0x79, 0x52, 0x0c, 0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69,
	0x65, 0x73, 0x12, 0x1b, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22,
	0x43, 0x0a, 0x0a, 0x44, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x79, 0x12, 0x1d, 0x0a,
	0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03,
	0xe0, 0x41, 0x02, 0x52, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x12, 0x16, 0x0a, 0x06,
	0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x66, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x42, 0x6e, 0x0a, 0x2d, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x42, 0x17, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65,
	0x72, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x22, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x72, 0x70, 0x63,
	0x3b, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescOnce sync.Once
	file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescData = file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDesc
)

func file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescGZIP() []byte {
	file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescOnce.Do(func() {
		file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescData)
	})
	return file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDescData
}

var file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_google_cloud_apigeeregistry_v1_controller_manifest_proto_goTypes = []interface{}{
	(*Manifest)(nil),          // 0: google.cloud.apigeeregistry.v1.controller.Manifest
	(*GeneratedResource)(nil), // 1: google.cloud.apigeeregistry.v1.controller.GeneratedResource
	(*Dependency)(nil),        // 2: google.cloud.apigeeregistry.v1.controller.Dependency
}
var file_google_cloud_apigeeregistry_v1_controller_manifest_proto_depIdxs = []int32{
	1, // 0: google.cloud.apigeeregistry.v1.controller.Manifest.generated_resources:type_name -> google.cloud.apigeeregistry.v1.controller.GeneratedResource
	2, // 1: google.cloud.apigeeregistry.v1.controller.GeneratedResource.dependencies:type_name -> google.cloud.apigeeregistry.v1.controller.Dependency
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_google_cloud_apigeeregistry_v1_controller_manifest_proto_init() }
func file_google_cloud_apigeeregistry_v1_controller_manifest_proto_init() {
	if File_google_cloud_apigeeregistry_v1_controller_manifest_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Manifest); i {
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
		file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GeneratedResource); i {
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
		file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Dependency); i {
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
			RawDescriptor: file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_apigeeregistry_v1_controller_manifest_proto_goTypes,
		DependencyIndexes: file_google_cloud_apigeeregistry_v1_controller_manifest_proto_depIdxs,
		MessageInfos:      file_google_cloud_apigeeregistry_v1_controller_manifest_proto_msgTypes,
	}.Build()
	File_google_cloud_apigeeregistry_v1_controller_manifest_proto = out.File
	file_google_cloud_apigeeregistry_v1_controller_manifest_proto_rawDesc = nil
	file_google_cloud_apigeeregistry_v1_controller_manifest_proto_goTypes = nil
	file_google_cloud_apigeeregistry_v1_controller_manifest_proto_depIdxs = nil
}
