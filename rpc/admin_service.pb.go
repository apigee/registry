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
// 	protoc        v3.17.3
// source: google/cloud/apigeeregistry/v1/admin_service.proto

package rpc

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Response message for GetStatus.
// GetStatus is not included in hosted versions of the API.
type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A string describing the status.
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[0]
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
	return file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescGZIP(), []int{0}
}

func (x *Status) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

// Request message for ListProjects.
// (-- api-linter: core::0132::request-parent-required=disabled
//     aip.dev/not-precedent: the parent of Project is implicit. --)
type ListProjectsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The maximum number of projects to return.
	// The service may return fewer than this value.
	// If unspecified, at most 50 values will be returned.
	// The maximum is 1000; values above 1000 will be coerced to 1000.
	PageSize int32 `protobuf:"varint,1,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	// A page token, received from a previous `ListProjects` call.
	// Provide this to retrieve the subsequent page.
	//
	// When paginating, all other parameters provided to `ListProjects` must match
	// the call that provided the page token.
	PageToken string `protobuf:"bytes,2,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
	// An expression that can be used to filter the list. Filters use the Common
	// Expression Language and can refer to all message fields.
	Filter string `protobuf:"bytes,3,opt,name=filter,proto3" json:"filter,omitempty"`
}

func (x *ListProjectsRequest) Reset() {
	*x = ListProjectsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListProjectsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListProjectsRequest) ProtoMessage() {}

func (x *ListProjectsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListProjectsRequest.ProtoReflect.Descriptor instead.
func (*ListProjectsRequest) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescGZIP(), []int{1}
}

func (x *ListProjectsRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ListProjectsRequest) GetPageToken() string {
	if x != nil {
		return x.PageToken
	}
	return ""
}

func (x *ListProjectsRequest) GetFilter() string {
	if x != nil {
		return x.Filter
	}
	return ""
}

// Response message for ListProjects.
type ListProjectsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The projects from the specified publisher.
	Projects []*Project `protobuf:"bytes,1,rep,name=projects,proto3" json:"projects,omitempty"`
	// A token, which can be sent as `page_token` to retrieve the next page.
	// If this field is omitted, there are no subsequent pages.
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
}

func (x *ListProjectsResponse) Reset() {
	*x = ListProjectsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListProjectsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListProjectsResponse) ProtoMessage() {}

func (x *ListProjectsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListProjectsResponse.ProtoReflect.Descriptor instead.
func (*ListProjectsResponse) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescGZIP(), []int{2}
}

func (x *ListProjectsResponse) GetProjects() []*Project {
	if x != nil {
		return x.Projects
	}
	return nil
}

func (x *ListProjectsResponse) GetNextPageToken() string {
	if x != nil {
		return x.NextPageToken
	}
	return ""
}

// Request message for GetProject.
type GetProjectRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the project to retrieve.
	// Format: projects/*
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *GetProjectRequest) Reset() {
	*x = GetProjectRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetProjectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetProjectRequest) ProtoMessage() {}

func (x *GetProjectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetProjectRequest.ProtoReflect.Descriptor instead.
func (*GetProjectRequest) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescGZIP(), []int{3}
}

func (x *GetProjectRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// Request message for CreateProject.
// (-- api-linter: core::0133::request-parent-required=disabled
//     aip.dev/not-precedent: the parent of Project is implicit. --)
type CreateProjectRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The project to create.
	Project *Project `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// The ID to use for the project, which will become the final component of
	// the project's resource name.
	//
	// This value should be at most 80 characters, and valid characters
	// are /[a-z][0-9]-./.
	ProjectId string `protobuf:"bytes,2,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
}

func (x *CreateProjectRequest) Reset() {
	*x = CreateProjectRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateProjectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateProjectRequest) ProtoMessage() {}

func (x *CreateProjectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateProjectRequest.ProtoReflect.Descriptor instead.
func (*CreateProjectRequest) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescGZIP(), []int{4}
}

func (x *CreateProjectRequest) GetProject() *Project {
	if x != nil {
		return x.Project
	}
	return nil
}

func (x *CreateProjectRequest) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

// Request message for UpdateProject.
type UpdateProjectRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The project to update.
	//
	// The `name` field is used to identify the project to update.
	// Format: projects/*
	Project *Project `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// The list of fields to be updated. If omitted, all fields are updated that
	// are set in the request message (fields set to default values are ignored).
	// If a "*" is specified, all fields are updated, including fields that are
	// unspecified/default in the request.
	UpdateMask *fieldmaskpb.FieldMask `protobuf:"bytes,2,opt,name=update_mask,json=updateMask,proto3" json:"update_mask,omitempty"`
}

func (x *UpdateProjectRequest) Reset() {
	*x = UpdateProjectRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateProjectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateProjectRequest) ProtoMessage() {}

func (x *UpdateProjectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateProjectRequest.ProtoReflect.Descriptor instead.
func (*UpdateProjectRequest) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescGZIP(), []int{5}
}

func (x *UpdateProjectRequest) GetProject() *Project {
	if x != nil {
		return x.Project
	}
	return nil
}

func (x *UpdateProjectRequest) GetUpdateMask() *fieldmaskpb.FieldMask {
	if x != nil {
		return x.UpdateMask
	}
	return nil
}

// Request message for DeleteProject.
type DeleteProjectRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the project to delete.
	// Format: projects/*
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *DeleteProjectRequest) Reset() {
	*x = DeleteProjectRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteProjectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteProjectRequest) ProtoMessage() {}

func (x *DeleteProjectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteProjectRequest.ProtoReflect.Descriptor instead.
func (*DeleteProjectRequest) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescGZIP(), []int{6}
}

func (x *DeleteProjectRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

var File_google_cloud_apigeeregistry_v1_admin_service_proto protoreflect.FileDescriptor

var file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDesc = []byte{
	0x0a, 0x32, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31,
	0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x17, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65,
	0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x31, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x5f, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74,
	0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x6d,
	0x61, 0x73, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x22, 0x0a, 0x06, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x69, 0x0a,
	0x13, 0x4c, 0x69, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a,
	0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e,
	0x12, 0x16, 0x0a, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x22, 0x83, 0x01, 0x0a, 0x14, 0x4c, 0x69, 0x73,
	0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x43, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x08, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x65, 0x78, 0x74, 0x5f, 0x70,
	0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x56,
	0x0a, 0x11, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x41, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x2d, 0xe0, 0x41, 0x02, 0xfa, 0x41, 0x27, 0x0a, 0x25, 0x61, 0x70, 0x69, 0x67, 0x65,
	0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x7d, 0x0a, 0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x46,
	0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x27, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x07, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x49, 0x64, 0x22, 0x9b, 0x01, 0x0a, 0x14, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x46,
	0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x27, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x07, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x3b, 0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x5f, 0x6d, 0x61, 0x73, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x4d, 0x61, 0x73, 0x6b, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4d,
	0x61, 0x73, 0x6b, 0x22, 0x59, 0x0a, 0x14, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x50, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x41, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x2d, 0xe0, 0x41, 0x02, 0xfa, 0x41,
	0x27, 0x0a, 0x25, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x32, 0x8f,
	0x07, 0x0a, 0x05, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x5f, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x26, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x12, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0c, 0x12, 0x0a, 0x2f,
	0x76, 0x31, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x8f, 0x01, 0x0a, 0x0c, 0x4c, 0x69,
	0x73, 0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x12, 0x33, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65,
	0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74,
	0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x34, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x14, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0e, 0x12, 0x0c, 0x2f,
	0x76, 0x31, 0x2f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x12, 0x8e, 0x01, 0x0a, 0x0a,
	0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x31, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65,
	0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x50,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x50,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x22, 0x24, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17, 0x12, 0x15,
	0x2f, 0x76, 0x31, 0x2f, 0x7b, 0x6e, 0x61, 0x6d, 0x65, 0x3d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x73, 0x2f, 0x2a, 0x7d, 0xda, 0x41, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0xa2, 0x01, 0x0a,
	0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x34,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70,
	0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x22, 0x32, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x17, 0x22, 0x0c, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x73, 0x3a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0xda, 0x41, 0x12, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2c, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69,
	0x64, 0x12, 0xb4, 0x01, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x12, 0x34, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x22, 0x44, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x28, 0x32, 0x1d, 0x2f, 0x76, 0x31, 0x2f,
	0x7b, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2e, 0x6e, 0x61, 0x6d, 0x65, 0x3d, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x2f, 0x2a, 0x7d, 0x3a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0xda, 0x41, 0x13, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2c, 0x75, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x5f, 0x6d, 0x61, 0x73, 0x6b, 0x12, 0x83, 0x01, 0x0a, 0x0d, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x34, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65,
	0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x24, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17,
	0x2a, 0x15, 0x2f, 0x76, 0x31, 0x2f, 0x7b, 0x6e, 0x61, 0x6d, 0x65, 0x3d, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x73, 0x2f, 0x2a, 0x7d, 0xda, 0x41, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x1a, 0x20,
	0xca, 0x41, 0x1d, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x6d,
	0x42, 0x5d, 0x0a, 0x22, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x42, 0x11, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x22, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x2f, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x72, 0x70, 0x63, 0x3b, 0x72, 0x70, 0x63, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescOnce sync.Once
	file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescData = file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDesc
)

func file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescGZIP() []byte {
	file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescOnce.Do(func() {
		file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescData)
	})
	return file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDescData
}

var file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_google_cloud_apigeeregistry_v1_admin_service_proto_goTypes = []interface{}{
	(*Status)(nil),                // 0: google.cloud.apigeeregistry.v1.Status
	(*ListProjectsRequest)(nil),   // 1: google.cloud.apigeeregistry.v1.ListProjectsRequest
	(*ListProjectsResponse)(nil),  // 2: google.cloud.apigeeregistry.v1.ListProjectsResponse
	(*GetProjectRequest)(nil),     // 3: google.cloud.apigeeregistry.v1.GetProjectRequest
	(*CreateProjectRequest)(nil),  // 4: google.cloud.apigeeregistry.v1.CreateProjectRequest
	(*UpdateProjectRequest)(nil),  // 5: google.cloud.apigeeregistry.v1.UpdateProjectRequest
	(*DeleteProjectRequest)(nil),  // 6: google.cloud.apigeeregistry.v1.DeleteProjectRequest
	(*Project)(nil),               // 7: google.cloud.apigeeregistry.v1.Project
	(*fieldmaskpb.FieldMask)(nil), // 8: google.protobuf.FieldMask
	(*emptypb.Empty)(nil),         // 9: google.protobuf.Empty
}
var file_google_cloud_apigeeregistry_v1_admin_service_proto_depIdxs = []int32{
	7,  // 0: google.cloud.apigeeregistry.v1.ListProjectsResponse.projects:type_name -> google.cloud.apigeeregistry.v1.Project
	7,  // 1: google.cloud.apigeeregistry.v1.CreateProjectRequest.project:type_name -> google.cloud.apigeeregistry.v1.Project
	7,  // 2: google.cloud.apigeeregistry.v1.UpdateProjectRequest.project:type_name -> google.cloud.apigeeregistry.v1.Project
	8,  // 3: google.cloud.apigeeregistry.v1.UpdateProjectRequest.update_mask:type_name -> google.protobuf.FieldMask
	9,  // 4: google.cloud.apigeeregistry.v1.Admin.GetStatus:input_type -> google.protobuf.Empty
	1,  // 5: google.cloud.apigeeregistry.v1.Admin.ListProjects:input_type -> google.cloud.apigeeregistry.v1.ListProjectsRequest
	3,  // 6: google.cloud.apigeeregistry.v1.Admin.GetProject:input_type -> google.cloud.apigeeregistry.v1.GetProjectRequest
	4,  // 7: google.cloud.apigeeregistry.v1.Admin.CreateProject:input_type -> google.cloud.apigeeregistry.v1.CreateProjectRequest
	5,  // 8: google.cloud.apigeeregistry.v1.Admin.UpdateProject:input_type -> google.cloud.apigeeregistry.v1.UpdateProjectRequest
	6,  // 9: google.cloud.apigeeregistry.v1.Admin.DeleteProject:input_type -> google.cloud.apigeeregistry.v1.DeleteProjectRequest
	0,  // 10: google.cloud.apigeeregistry.v1.Admin.GetStatus:output_type -> google.cloud.apigeeregistry.v1.Status
	2,  // 11: google.cloud.apigeeregistry.v1.Admin.ListProjects:output_type -> google.cloud.apigeeregistry.v1.ListProjectsResponse
	7,  // 12: google.cloud.apigeeregistry.v1.Admin.GetProject:output_type -> google.cloud.apigeeregistry.v1.Project
	7,  // 13: google.cloud.apigeeregistry.v1.Admin.CreateProject:output_type -> google.cloud.apigeeregistry.v1.Project
	7,  // 14: google.cloud.apigeeregistry.v1.Admin.UpdateProject:output_type -> google.cloud.apigeeregistry.v1.Project
	9,  // 15: google.cloud.apigeeregistry.v1.Admin.DeleteProject:output_type -> google.protobuf.Empty
	10, // [10:16] is the sub-list for method output_type
	4,  // [4:10] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_google_cloud_apigeeregistry_v1_admin_service_proto_init() }
func file_google_cloud_apigeeregistry_v1_admin_service_proto_init() {
	if File_google_cloud_apigeeregistry_v1_admin_service_proto != nil {
		return
	}
	file_google_cloud_apigeeregistry_v1_admin_models_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListProjectsRequest); i {
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
		file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListProjectsResponse); i {
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
		file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetProjectRequest); i {
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
		file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateProjectRequest); i {
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
		file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateProjectRequest); i {
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
		file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteProjectRequest); i {
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
			RawDescriptor: file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_google_cloud_apigeeregistry_v1_admin_service_proto_goTypes,
		DependencyIndexes: file_google_cloud_apigeeregistry_v1_admin_service_proto_depIdxs,
		MessageInfos:      file_google_cloud_apigeeregistry_v1_admin_service_proto_msgTypes,
	}.Build()
	File_google_cloud_apigeeregistry_v1_admin_service_proto = out.File
	file_google_cloud_apigeeregistry_v1_admin_service_proto_rawDesc = nil
	file_google_cloud_apigeeregistry_v1_admin_service_proto_goTypes = nil
	file_google_cloud_apigeeregistry_v1_admin_service_proto_depIdxs = nil
}
