// Copyright 2021 Google LLC. All Rights Reserved.
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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.1
// source: google/cloud/apigeeregistry/v1/admin_models.proto

package rpc

import (
	reflect "reflect"
	sync "sync"

	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// BuildInfo describes a server build.
type BuildInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Version of Go that produced this binary.
	GoVersion string `protobuf:"bytes,1,opt,name=go_version,json=goVersion,proto3" json:"go_version,omitempty"`
	// The main package path.
	Path string `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	// The main module.
	Main *BuildInfo_Module `protobuf:"bytes,3,opt,name=main,proto3" json:"main,omitempty"`
	// The dependencies of the build.
	Dependencies []*BuildInfo_Module `protobuf:"bytes,4,rep,name=dependencies,proto3" json:"dependencies,omitempty"`
	// Settings associated with the build.
	Settings map[string]string `protobuf:"bytes,5,rep,name=settings,proto3" json:"settings,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *BuildInfo) Reset() {
	*x = BuildInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildInfo) ProtoMessage() {}

func (x *BuildInfo) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildInfo.ProtoReflect.Descriptor instead.
func (*BuildInfo) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescGZIP(), []int{0}
}

func (x *BuildInfo) GetGoVersion() string {
	if x != nil {
		return x.GoVersion
	}
	return ""
}

func (x *BuildInfo) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *BuildInfo) GetMain() *BuildInfo_Module {
	if x != nil {
		return x.Main
	}
	return nil
}

func (x *BuildInfo) GetDependencies() []*BuildInfo_Module {
	if x != nil {
		return x.Dependencies
	}
	return nil
}

func (x *BuildInfo) GetSettings() map[string]string {
	if x != nil {
		return x.Settings
	}
	return nil
}

// Status represents the status of the service.
type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A string describing the status.
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	// Information about the build of the server.
	Buildinfo *BuildInfo `protobuf:"bytes,2,opt,name=buildinfo,proto3" json:"buildinfo,omitempty"`
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Status.ProtoReflect.Descriptor instead.
func (*Status) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescGZIP(), []int{1}
}

func (x *Status) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Status) GetBuildinfo() *BuildInfo {
	if x != nil {
		return x.Buildinfo
	}
	return nil
}

// Storage describes the data stored by the service.
type Storage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A string describing the storage backend.
	Description string `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"`
	// A list of collections in the storage backend.
	// Collections are listed in alphabetical order.
	Collections []*Storage_Collection `protobuf:"bytes,2,rep,name=collections,proto3" json:"collections,omitempty"`
}

func (x *Storage) Reset() {
	*x = Storage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Storage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Storage) ProtoMessage() {}

func (x *Storage) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Storage.ProtoReflect.Descriptor instead.
func (*Storage) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescGZIP(), []int{2}
}

func (x *Storage) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Storage) GetCollections() []*Storage_Collection {
	if x != nil {
		return x.Collections
	}
	return nil
}

// A Project is a top-level description of a collection of APIs.
// Typically there would be one project for an entire organization.
// Note: in a Google Cloud deployment, this resource and associated methods
// will be omitted and its children will instead be associated with Google
// Cloud projects.
type Project struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Resource name.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Human-meaningful name.
	DisplayName string `protobuf:"bytes,2,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	// A detailed description.
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	// Creation timestamp.
	CreateTime *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// Last update timestamp.
	UpdateTime *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
}

func (x *Project) Reset() {
	*x = Project{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Project) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Project) ProtoMessage() {}

func (x *Project) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Project.ProtoReflect.Descriptor instead.
func (*Project) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescGZIP(), []int{3}
}

func (x *Project) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Project) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *Project) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Project) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *Project) GetUpdateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdateTime
	}
	return nil
}

// A module used to create the build.
type BuildInfo_Module struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The module path.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// The module version.
	Version string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	// The module checksum.
	Sum string `protobuf:"bytes,3,opt,name=sum,proto3" json:"sum,omitempty"`
	// An optional replacement for the module
	Replacement *BuildInfo_Module `protobuf:"bytes,4,opt,name=replacement,proto3" json:"replacement,omitempty"`
}

func (x *BuildInfo_Module) Reset() {
	*x = BuildInfo_Module{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildInfo_Module) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildInfo_Module) ProtoMessage() {}

func (x *BuildInfo_Module) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildInfo_Module.ProtoReflect.Descriptor instead.
func (*BuildInfo_Module) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescGZIP(), []int{0, 0}
}

func (x *BuildInfo_Module) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *BuildInfo_Module) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *BuildInfo_Module) GetSum() string {
	if x != nil {
		return x.Sum
	}
	return ""
}

func (x *BuildInfo_Module) GetReplacement() *BuildInfo_Module {
	if x != nil {
		return x.Replacement
	}
	return nil
}

// A description of a collection in the backend database.
type Storage_Collection struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the collection.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The number of entries in the collection.
	Count int64 `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *Storage_Collection) Reset() {
	*x = Storage_Collection{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Storage_Collection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Storage_Collection) ProtoMessage() {}

func (x *Storage_Collection) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Storage_Collection.ProtoReflect.Descriptor instead.
func (*Storage_Collection) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescGZIP(), []int{2, 0}
}

func (x *Storage_Collection) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Storage_Collection) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

var File_google_cloud_apigeeregistry_v1_admin_models_proto protoreflect.FileDescriptor

var file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDesc = []byte{
	0x0a, 0x31, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31,
	0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x5f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x8b, 0x04, 0x0a, 0x09, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1d,
	0x0a, 0x0a, 0x67, 0x6f, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x67, 0x6f, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74,
	0x68, 0x12, 0x44, 0x0a, 0x04, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x30, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c,
	0x65, 0x52, 0x04, 0x6d, 0x61, 0x69, 0x6e, 0x12, 0x54, 0x0a, 0x0c, 0x64, 0x65, 0x70, 0x65, 0x6e,
	0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x30, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x42,
	0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x52,
	0x0c, 0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x12, 0x53, 0x0a,
	0x08, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x37, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x53, 0x65, 0x74, 0x74, 0x69,
	0x6e, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e,
	0x67, 0x73, 0x1a, 0x9c, 0x01, 0x0a, 0x06, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74,
	0x68, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x73,
	0x75, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x75, 0x6d, 0x12, 0x52, 0x0a,
	0x0b, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x30, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2e, 0x76, 0x31, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x4d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x52, 0x0b, 0x72, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x6d, 0x65, 0x6e,
	0x74, 0x1a, 0x3b, 0x0a, 0x0d, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x6b,
	0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x12, 0x47, 0x0a, 0x09, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x69, 0x6e, 0x66, 0x6f, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x09, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x69, 0x6e, 0x66, 0x6f, 0x22, 0xb9, 0x01, 0x0a, 0x07,
	0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x54, 0x0a, 0x0b, 0x63, 0x6f, 0x6c,
	0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70,
	0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e,
	0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x0b, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x1a,
	0x36, 0x0a, 0x0a, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0xa6, 0x02, 0x0a, 0x07, 0x50, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c,
	0x61, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64,
	0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x40, 0x0a, 0x0b,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x03, 0xe0,
	0x41, 0x03, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x40,
	0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42,
	0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65,
	0x3a, 0x3e, 0xea, 0x41, 0x3b, 0x0a, 0x25, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x12, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x2f, 0x7b, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x7d,
	0x42, 0x5c, 0x0a, 0x22, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x42, 0x10, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x4d, 0x6f, 0x64,
	0x65, 0x6c, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x22, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x2f, 0x72, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x72, 0x70, 0x63, 0x3b, 0x72, 0x70, 0x63, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescOnce sync.Once
	file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescData = file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDesc
)

func file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescGZIP() []byte {
	file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescOnce.Do(func() {
		file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescData)
	})
	return file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDescData
}

var file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_google_cloud_apigeeregistry_v1_admin_models_proto_goTypes = []interface{}{
	(*BuildInfo)(nil),             // 0: google.cloud.apigeeregistry.v1.BuildInfo
	(*Status)(nil),                // 1: google.cloud.apigeeregistry.v1.Status
	(*Storage)(nil),               // 2: google.cloud.apigeeregistry.v1.Storage
	(*Project)(nil),               // 3: google.cloud.apigeeregistry.v1.Project
	(*BuildInfo_Module)(nil),      // 4: google.cloud.apigeeregistry.v1.BuildInfo.Module
	nil,                           // 5: google.cloud.apigeeregistry.v1.BuildInfo.SettingsEntry
	(*Storage_Collection)(nil),    // 6: google.cloud.apigeeregistry.v1.Storage.Collection
	(*timestamppb.Timestamp)(nil), // 7: google.protobuf.Timestamp
}
var file_google_cloud_apigeeregistry_v1_admin_models_proto_depIdxs = []int32{
	4, // 0: google.cloud.apigeeregistry.v1.BuildInfo.main:type_name -> google.cloud.apigeeregistry.v1.BuildInfo.Module
	4, // 1: google.cloud.apigeeregistry.v1.BuildInfo.dependencies:type_name -> google.cloud.apigeeregistry.v1.BuildInfo.Module
	5, // 2: google.cloud.apigeeregistry.v1.BuildInfo.settings:type_name -> google.cloud.apigeeregistry.v1.BuildInfo.SettingsEntry
	0, // 3: google.cloud.apigeeregistry.v1.Status.buildinfo:type_name -> google.cloud.apigeeregistry.v1.BuildInfo
	6, // 4: google.cloud.apigeeregistry.v1.Storage.collections:type_name -> google.cloud.apigeeregistry.v1.Storage.Collection
	7, // 5: google.cloud.apigeeregistry.v1.Project.create_time:type_name -> google.protobuf.Timestamp
	7, // 6: google.cloud.apigeeregistry.v1.Project.update_time:type_name -> google.protobuf.Timestamp
	4, // 7: google.cloud.apigeeregistry.v1.BuildInfo.Module.replacement:type_name -> google.cloud.apigeeregistry.v1.BuildInfo.Module
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_google_cloud_apigeeregistry_v1_admin_models_proto_init() }
func file_google_cloud_apigeeregistry_v1_admin_models_proto_init() {
	if File_google_cloud_apigeeregistry_v1_admin_models_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildInfo); i {
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
		file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Status); i {
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
		file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Storage); i {
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
		file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Project); i {
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
		file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildInfo_Module); i {
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
		file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Storage_Collection); i {
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
			RawDescriptor: file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_apigeeregistry_v1_admin_models_proto_goTypes,
		DependencyIndexes: file_google_cloud_apigeeregistry_v1_admin_models_proto_depIdxs,
		MessageInfos:      file_google_cloud_apigeeregistry_v1_admin_models_proto_msgTypes,
	}.Build()
	File_google_cloud_apigeeregistry_v1_admin_models_proto = out.File
	file_google_cloud_apigeeregistry_v1_admin_models_proto_rawDesc = nil
	file_google_cloud_apigeeregistry_v1_admin_models_proto_goTypes = nil
	file_google_cloud_apigeeregistry_v1_admin_models_proto_depIdxs = nil
}
