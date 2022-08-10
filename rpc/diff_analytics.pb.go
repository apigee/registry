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
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.3
// source: google/cloud/apigeeregistry/applications/v1alpha1/diff_analytics.proto

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

//Diff contains the diff of a spec and its revision.
type Diff struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// additions holds every addition change in the diff.
	//The string will hold the entire field path of one addition change in the
	//format foo.bar.x .
	Additions []string `protobuf:"bytes,1,rep,name=additions,proto3" json:"additions,omitempty"`
	// deletions holds every deletion change in the diff.
	//The string will hold the entire field path of one deletion change in the
	//format foo.bar.x .
	Deletions []string `protobuf:"bytes,2,rep,name=deletions,proto3" json:"deletions,omitempty"`
	// modifications holds every modification change in the diff.
	//The string key will hold the field path of one modification change in the
	//format foo.bar.x.
	//The value of the key will repersent the element that was modified in the
	//field.
	Modifications map[string]*Diff_ValueChange `protobuf:"bytes,3,rep,name=modifications,proto3" json:"modifications,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Diff) Reset() {
	*x = Diff{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Diff) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Diff) ProtoMessage() {}

func (x *Diff) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Diff.ProtoReflect.Descriptor instead.
func (*Diff) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescGZIP(), []int{0}
}

func (x *Diff) GetAdditions() []string {
	if x != nil {
		return x.Additions
	}
	return nil
}

func (x *Diff) GetDeletions() []string {
	if x != nil {
		return x.Deletions
	}
	return nil
}

func (x *Diff) GetModifications() map[string]*Diff_ValueChange {
	if x != nil {
		return x.Modifications
	}
	return nil
}

// ChangeDetails classifies changes from diff in to seperate categories.
type ChangeDetails struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// breakingChanges is a Diff proto that only contians the breaking changes
	//of a diff.
	BreakingChanges *Diff `protobuf:"bytes,1,opt,name=breaking_changes,json=breakingChanges,proto3" json:"breaking_changes,omitempty"`
	// nonBreakingChanges is a Diff proto that only contians the non-breaking
	//changes of a diff.
	NonBreakingChanges *Diff `protobuf:"bytes,2,opt,name=non_breaking_changes,json=nonBreakingChanges,proto3" json:"non_breaking_changes,omitempty"`
	// unknownChanges is a Diff proto that contians all the changes that could not
	//be classifed in the other categories.
	UnknownChanges *Diff `protobuf:"bytes,3,opt,name=unknown_changes,json=unknownChanges,proto3" json:"unknown_changes,omitempty"`
}

func (x *ChangeDetails) Reset() {
	*x = ChangeDetails{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChangeDetails) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeDetails) ProtoMessage() {}

func (x *ChangeDetails) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeDetails.ProtoReflect.Descriptor instead.
func (*ChangeDetails) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescGZIP(), []int{1}
}

func (x *ChangeDetails) GetBreakingChanges() *Diff {
	if x != nil {
		return x.BreakingChanges
	}
	return nil
}

func (x *ChangeDetails) GetNonBreakingChanges() *Diff {
	if x != nil {
		return x.NonBreakingChanges
	}
	return nil
}

func (x *ChangeDetails) GetUnknownChanges() *Diff {
	if x != nil {
		return x.UnknownChanges
	}
	return nil
}

// ChangeStats holds information relating to a list of diffs
type ChangeStats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// breaking_change_count represents the total number of breaking changes.
	BreakingChangeCount int64 `protobuf:"varint,1,opt,name=breaking_change_count,json=breakingChangeCount,proto3" json:"breaking_change_count,omitempty"`
	// nonbreaking_change_count represents the total number of non-breaking changes.
	NonbreakingChangeCount int64 `protobuf:"varint,2,opt,name=nonbreaking_change_count,json=nonbreakingChangeCount,proto3" json:"nonbreaking_change_count,omitempty"`
	// diff_count represents the number of diffs used in this stats
	DiffCount int64 `protobuf:"varint,3,opt,name=diff_count,json=diffCount,proto3" json:"diff_count,omitempty"`
}

func (x *ChangeStats) Reset() {
	*x = ChangeStats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChangeStats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeStats) ProtoMessage() {}

func (x *ChangeStats) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeStats.ProtoReflect.Descriptor instead.
func (*ChangeStats) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescGZIP(), []int{2}
}

func (x *ChangeStats) GetBreakingChangeCount() int64 {
	if x != nil {
		return x.BreakingChangeCount
	}
	return 0
}

func (x *ChangeStats) GetNonbreakingChangeCount() int64 {
	if x != nil {
		return x.NonbreakingChangeCount
	}
	return 0
}

func (x *ChangeStats) GetDiffCount() int64 {
	if x != nil {
		return x.DiffCount
	}
	return 0
}

// ChangeMetrics holds metrics about a list of diffs. Each metric is computed from
//two or more stats.
type ChangeMetrics struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// breaking_change_percentage is the precentage of changes that are breaking.
	//It is computed by the equation
	//(breaking_change_count / (nonbreaking_change_count + breaking_change_count))
	BreakingChangePercentage float64 `protobuf:"fixed64,1,opt,name=breaking_change_percentage,json=breakingChangePercentage,proto3" json:"breaking_change_percentage,omitempty"`
	// breaking_change_rate is the average number of breaking changes that are
	//introduced per Diff.
	//It is computed by the equation
	//((nonbreaking_change_count + breaking_change_count) / diff_count)
	BreakingChangeRate float64 `protobuf:"fixed64,2,opt,name=breaking_change_rate,json=breakingChangeRate,proto3" json:"breaking_change_rate,omitempty"`
}

func (x *ChangeMetrics) Reset() {
	*x = ChangeMetrics{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChangeMetrics) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeMetrics) ProtoMessage() {}

func (x *ChangeMetrics) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeMetrics.ProtoReflect.Descriptor instead.
func (*ChangeMetrics) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescGZIP(), []int{3}
}

func (x *ChangeMetrics) GetBreakingChangePercentage() float64 {
	if x != nil {
		return x.BreakingChangePercentage
	}
	return 0
}

func (x *ChangeMetrics) GetBreakingChangeRate() float64 {
	if x != nil {
		return x.BreakingChangeRate
	}
	return 0
}

// ValueChange hold the values of the elements that changed in one diff change.
type Diff_ValueChange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// from represents the previous value of the element.
	From string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	// to represents the current value of the element.
	To string `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
}

func (x *Diff_ValueChange) Reset() {
	*x = Diff_ValueChange{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Diff_ValueChange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Diff_ValueChange) ProtoMessage() {}

func (x *Diff_ValueChange) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Diff_ValueChange.ProtoReflect.Descriptor instead.
func (*Diff_ValueChange) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Diff_ValueChange) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *Diff_ValueChange) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

var File_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto protoreflect.FileDescriptor

var file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDesc = []byte{
	0x0a, 0x46, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x61, 0x70,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2f, 0x64, 0x69, 0x66, 0x66, 0x5f, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69,
	0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x31, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x22, 0xef, 0x02, 0x0a, 0x04,
	0x44, 0x69, 0x66, 0x66, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x64, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x61, 0x64, 0x64, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0x70, 0x0a, 0x0d, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x4a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69, 0x66, 0x66,
	0x2e, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x0d, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x1a, 0x31, 0x0a, 0x0b, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x74, 0x6f, 0x1a, 0x85, 0x01, 0x0a, 0x12, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x59,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x43, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x44, 0x69, 0x66, 0x66, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x43, 0x68, 0x61, 0x6e,
	0x67, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xc0, 0x02,
	0x0a, 0x0d, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12,
	0x62, 0x0a, 0x10, 0x62, 0x72, 0x65, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x5f, 0x63, 0x68, 0x61, 0x6e,
	0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x37, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69,
	0x66, 0x66, 0x52, 0x0f, 0x62, 0x72, 0x65, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x43, 0x68, 0x61, 0x6e,
	0x67, 0x65, 0x73, 0x12, 0x69, 0x0a, 0x14, 0x6e, 0x6f, 0x6e, 0x5f, 0x62, 0x72, 0x65, 0x61, 0x6b,
	0x69, 0x6e, 0x67, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x37, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e,
	0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69, 0x66, 0x66, 0x52, 0x12, 0x6e, 0x6f, 0x6e, 0x42,
	0x72, 0x65, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x12, 0x60,
	0x0a, 0x0f, 0x75, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x37, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x69, 0x66, 0x66,
	0x52, 0x0e, 0x75, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73,
	0x22, 0x9a, 0x01, 0x0a, 0x0b, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x53, 0x74, 0x61, 0x74, 0x73,
	0x12, 0x32, 0x0a, 0x15, 0x62, 0x72, 0x65, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x5f, 0x63, 0x68, 0x61,
	0x6e, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x13, 0x62, 0x72, 0x65, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x12, 0x38, 0x0a, 0x18, 0x6e, 0x6f, 0x6e, 0x62, 0x72, 0x65, 0x61, 0x6b,
	0x69, 0x6e, 0x67, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x16, 0x6e, 0x6f, 0x6e, 0x62, 0x72, 0x65, 0x61, 0x6b,
	0x69, 0x6e, 0x67, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1d,
	0x0a, 0x0a, 0x64, 0x69, 0x66, 0x66, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x09, 0x64, 0x69, 0x66, 0x66, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x7f, 0x0a,
	0x0d, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x3c,
	0x0a, 0x1a, 0x62, 0x72, 0x65, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x5f, 0x70, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x18, 0x62, 0x72, 0x65, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x43, 0x68, 0x61, 0x6e,
	0x67, 0x65, 0x50, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x12, 0x30, 0x0a, 0x14,
	0x62, 0x72, 0x65, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x5f,
	0x72, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x12, 0x62, 0x72, 0x65, 0x61,
	0x6b, 0x69, 0x6e, 0x67, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x61, 0x74, 0x65, 0x42, 0x24,
	0x5a, 0x22, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x69,
	0x67, 0x65, 0x65, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x72, 0x70, 0x63,
	0x3b, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescOnce sync.Once
	file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescData = file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDesc
)

func file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescGZIP() []byte {
	file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescOnce.Do(func() {
		file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescData)
	})
	return file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDescData
}

var file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_goTypes = []interface{}{
	(*Diff)(nil),             // 0: google.cloud.apigeeregistry.applications.v1alpha1.Diff
	(*ChangeDetails)(nil),    // 1: google.cloud.apigeeregistry.applications.v1alpha1.ChangeDetails
	(*ChangeStats)(nil),      // 2: google.cloud.apigeeregistry.applications.v1alpha1.ChangeStats
	(*ChangeMetrics)(nil),    // 3: google.cloud.apigeeregistry.applications.v1alpha1.ChangeMetrics
	(*Diff_ValueChange)(nil), // 4: google.cloud.apigeeregistry.applications.v1alpha1.Diff.ValueChange
	nil,                      // 5: google.cloud.apigeeregistry.applications.v1alpha1.Diff.ModificationsEntry
}
var file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_depIdxs = []int32{
	5, // 0: google.cloud.apigeeregistry.applications.v1alpha1.Diff.modifications:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.Diff.ModificationsEntry
	0, // 1: google.cloud.apigeeregistry.applications.v1alpha1.ChangeDetails.breaking_changes:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.Diff
	0, // 2: google.cloud.apigeeregistry.applications.v1alpha1.ChangeDetails.non_breaking_changes:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.Diff
	0, // 3: google.cloud.apigeeregistry.applications.v1alpha1.ChangeDetails.unknown_changes:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.Diff
	4, // 4: google.cloud.apigeeregistry.applications.v1alpha1.Diff.ModificationsEntry.value:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.Diff.ValueChange
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_init() }
func file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_init() {
	if File_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Diff); i {
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
		file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChangeDetails); i {
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
		file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChangeStats); i {
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
		file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChangeMetrics); i {
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
		file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Diff_ValueChange); i {
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
			RawDescriptor: file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_goTypes,
		DependencyIndexes: file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_depIdxs,
		MessageInfos:      file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_msgTypes,
	}.Build()
	File_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto = out.File
	file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_rawDesc = nil
	file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_goTypes = nil
	file_google_cloud_apigeeregistry_applications_v1alpha1_diff_analytics_proto_depIdxs = nil
}
