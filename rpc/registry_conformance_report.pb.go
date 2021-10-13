// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: google/cloud/apigeeregistry/applications/v1alpha1/registry_conformance_report.proto

package rpc

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

// ConformanceReport describes how well an API Spec or a series of
// API Specs conform to a specific API Style guide.
type ConformanceReport struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Resource name of the conformance report.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Name of the style guide that this report pertains to.
	StyleguideName string `protobuf:"bytes,2,opt,name=styleguide_name,json=styleguideName,proto3" json:"styleguide_name,omitempty"`
	// A list of guideline report groups.
	GuidelineReportGroups []*GuidelineReportGroup `protobuf:"bytes,3,rep,name=guideline_report_groups,json=guidelineReportGroups,proto3" json:"guideline_report_groups,omitempty"`
}

func (x *ConformanceReport) Reset() {
	*x = ConformanceReport{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConformanceReport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConformanceReport) ProtoMessage() {}

func (x *ConformanceReport) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConformanceReport.ProtoReflect.Descriptor instead.
func (*ConformanceReport) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescGZIP(), []int{0}
}

func (x *ConformanceReport) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ConformanceReport) GetStyleguideName() string {
	if x != nil {
		return x.StyleguideName
	}
	return ""
}

func (x *ConformanceReport) GetGuidelineReportGroups() []*GuidelineReportGroup {
	if x != nil {
		return x.GuidelineReportGroups
	}
	return nil
}

// GuidelineReport describes how well an API Spec or a series of
// API Specs conform to a guideline within an API Style Guide.
type GuidelineReport struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the guideline that this report pertains to.
	GuidelineName string `protobuf:"bytes,1,opt,name=guideline_name,json=guidelineName,proto3" json:"guideline_name,omitempty"`
	// A list of rule report groups.
	RuleReportGroups []*RuleReportGroup `protobuf:"bytes,2,rep,name=rule_report_groups,json=ruleReportGroups,proto3" json:"rule_report_groups,omitempty"`
}

func (x *GuidelineReport) Reset() {
	*x = GuidelineReport{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GuidelineReport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GuidelineReport) ProtoMessage() {}

func (x *GuidelineReport) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GuidelineReport.ProtoReflect.Descriptor instead.
func (*GuidelineReport) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescGZIP(), []int{1}
}

func (x *GuidelineReport) GetGuidelineName() string {
	if x != nil {
		return x.GuidelineName
	}
	return ""
}

func (x *GuidelineReport) GetRuleReportGroups() []*RuleReportGroup {
	if x != nil {
		return x.RuleReportGroups
	}
	return nil
}

// RuleReport provides information and feedback on a rule that
// a spec breaches within a guideline on an API Style Guide.
type RuleReport struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the rule that the spec breaches.
	RuleName string `protobuf:"bytes,1,opt,name=rule_name,json=ruleName,proto3" json:"rule_name,omitempty"`
	// Resource name of the spec that the rule was breached on.
	SpecName string `protobuf:"bytes,2,opt,name=spec_name,json=specName,proto3" json:"spec_name,omitempty"`
	// A suggestion for resolving the problem.
	Suggestion string `protobuf:"bytes,3,opt,name=suggestion,proto3" json:"suggestion,omitempty"`
	// The location of the problem in the spec file.
	Location *LintLocation `protobuf:"bytes,4,opt,name=location,proto3" json:"location,omitempty"`
}

func (x *RuleReport) Reset() {
	*x = RuleReport{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RuleReport) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RuleReport) ProtoMessage() {}

func (x *RuleReport) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RuleReport.ProtoReflect.Descriptor instead.
func (*RuleReport) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescGZIP(), []int{2}
}

func (x *RuleReport) GetRuleName() string {
	if x != nil {
		return x.RuleName
	}
	return ""
}

func (x *RuleReport) GetSpecName() string {
	if x != nil {
		return x.SpecName
	}
	return ""
}

func (x *RuleReport) GetSuggestion() string {
	if x != nil {
		return x.Suggestion
	}
	return ""
}

func (x *RuleReport) GetLocation() *LintLocation {
	if x != nil {
		return x.Location
	}
	return nil
}

// GuidelineReportGroup is an abstraction that maps status
// (PROPOSED, ACTIVE, DEPRECATED, DISABLED) to a list of
// guideline reports for guidelines of that status.
type GuidelineReportGroup struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Status of the guidelines in this report group.
	Status Guideline_Status `protobuf:"varint,1,opt,name=status,proto3,enum=google.cloud.apigeeregistry.applications.v1alpha1.Guideline_Status" json:"status,omitempty"`
	// A list of guideline reports.
	GuidelineReports []*GuidelineReport `protobuf:"bytes,2,rep,name=guideline_reports,json=guidelineReports,proto3" json:"guideline_reports,omitempty"`
}

func (x *GuidelineReportGroup) Reset() {
	*x = GuidelineReportGroup{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GuidelineReportGroup) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GuidelineReportGroup) ProtoMessage() {}

func (x *GuidelineReportGroup) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GuidelineReportGroup.ProtoReflect.Descriptor instead.
func (*GuidelineReportGroup) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescGZIP(), []int{3}
}

func (x *GuidelineReportGroup) GetStatus() Guideline_Status {
	if x != nil {
		return x.Status
	}
	return Guideline_STATUS_UNSPECIFIED
}

func (x *GuidelineReportGroup) GetGuidelineReports() []*GuidelineReport {
	if x != nil {
		return x.GuidelineReports
	}
	return nil
}

// RuleReportGroup is an abstraction that maps severity
// (ERROR WARNING, INFO, HINT) to a list of rule reports for
// rules of that severity.
type RuleReportGroup struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Severity of the rules in this report group.
	Severity Rule_Severity `protobuf:"varint,1,opt,name=severity,proto3,enum=google.cloud.apigeeregistry.applications.v1alpha1.Rule_Severity" json:"severity,omitempty"`
	// A list of rule reports.
	RuleReports []*RuleReport `protobuf:"bytes,2,rep,name=rule_reports,json=ruleReports,proto3" json:"rule_reports,omitempty"`
}

func (x *RuleReportGroup) Reset() {
	*x = RuleReportGroup{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RuleReportGroup) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RuleReportGroup) ProtoMessage() {}

func (x *RuleReportGroup) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RuleReportGroup.ProtoReflect.Descriptor instead.
func (*RuleReportGroup) Descriptor() ([]byte, []int) {
	return file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescGZIP(), []int{4}
}

func (x *RuleReportGroup) GetSeverity() Rule_Severity {
	if x != nil {
		return x.Severity
	}
	return Rule_SEVERITY_UNSPECIFIED
}

func (x *RuleReportGroup) GetRuleReports() []*RuleReport {
	if x != nil {
		return x.RuleReports
	}
	return nil
}

var File_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto protoreflect.FileDescriptor

var file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDesc = []byte{
	0x0a, 0x53, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61,
	0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x61, 0x70,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x5f, 0x63, 0x6f, 0x6e,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x31, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76,
	0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x4b, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x5f, 0x73, 0x74, 0x79, 0x6c, 0x65, 0x67, 0x75, 0x69, 0x64, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x45, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x72, 0x79, 0x5f, 0x6c, 0x69, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xdb, 0x01,
	0x0a, 0x11, 0x43, 0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x70,
	0x6f, 0x72, 0x74, 0x12, 0x17, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x2c, 0x0a, 0x0f,
	0x73, 0x74, 0x79, 0x6c, 0x65, 0x67, 0x75, 0x69, 0x64, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0e, 0x73, 0x74, 0x79, 0x6c,
	0x65, 0x67, 0x75, 0x69, 0x64, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x7f, 0x0a, 0x17, 0x67, 0x75,
	0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x67,
	0x72, 0x6f, 0x75, 0x70, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x47, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65,
	0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x47, 0x75, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x47,
	0x72, 0x6f, 0x75, 0x70, 0x52, 0x15, 0x67, 0x75, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x22, 0xaf, 0x01, 0x0a, 0x0f,
	0x47, 0x75, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x12,
	0x2a, 0x0a, 0x0e, 0x67, 0x75, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0d, 0x67, 0x75,
	0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x70, 0x0a, 0x12, 0x72,
	0x75, 0x6c, 0x65, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x42, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52, 0x75, 0x6c, 0x65,
	0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x10, 0x72, 0x75, 0x6c,
	0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x22, 0xcd, 0x01,
	0x0a, 0x0a, 0x52, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x20, 0x0a, 0x09,
	0x72, 0x75, 0x6c, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x03, 0xe0, 0x41, 0x02, 0x52, 0x08, 0x72, 0x75, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20,
	0x0a, 0x09, 0x73, 0x70, 0x65, 0x63, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x08, 0x73, 0x70, 0x65, 0x63, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x75, 0x67, 0x67, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x75, 0x67, 0x67, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x5b, 0x0a, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x3f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4c, 0x69, 0x6e, 0x74, 0x4c, 0x6f, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xee, 0x01,
	0x0a, 0x14, 0x47, 0x75, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72,
	0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x60, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x43, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x75, 0x69, 0x64, 0x65,
	0x6c, 0x69, 0x6e, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x03, 0xe0, 0x41, 0x02,
	0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x74, 0x0a, 0x11, 0x67, 0x75, 0x69, 0x64,
	0x65, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x72, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x42, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x75, 0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x10, 0x67, 0x75,
	0x69, 0x64, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x22, 0xdb,
	0x01, 0x0a, 0x0f, 0x52, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x12, 0x61, 0x0a, 0x08, 0x73, 0x65, 0x76, 0x65, 0x72, 0x69, 0x74, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x40, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x52, 0x75, 0x6c, 0x65, 0x2e, 0x53, 0x65,
	0x76, 0x65, 0x72, 0x69, 0x74, 0x79, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x08, 0x73, 0x65, 0x76,
	0x65, 0x72, 0x69, 0x74, 0x79, 0x12, 0x65, 0x0a, 0x0c, 0x72, 0x75, 0x6c, 0x65, 0x5f, 0x72, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3d, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x67, 0x65,
	0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x52, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52,
	0x0b, 0x72, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x42, 0x7d, 0x0a, 0x35,
	0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x2e, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e,
	0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x1e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x43,
	0x6f, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x22, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x67, 0x65, 0x65, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2f, 0x72, 0x70, 0x63, 0x3b, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescOnce sync.Once
	file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescData = file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDesc
)

func file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescGZIP() []byte {
	file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescOnce.Do(func() {
		file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescData)
	})
	return file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDescData
}

var file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_goTypes = []interface{}{
	(*ConformanceReport)(nil),    // 0: google.cloud.apigeeregistry.applications.v1alpha1.ConformanceReport
	(*GuidelineReport)(nil),      // 1: google.cloud.apigeeregistry.applications.v1alpha1.GuidelineReport
	(*RuleReport)(nil),           // 2: google.cloud.apigeeregistry.applications.v1alpha1.RuleReport
	(*GuidelineReportGroup)(nil), // 3: google.cloud.apigeeregistry.applications.v1alpha1.GuidelineReportGroup
	(*RuleReportGroup)(nil),      // 4: google.cloud.apigeeregistry.applications.v1alpha1.RuleReportGroup
	(*LintLocation)(nil),         // 5: google.cloud.apigeeregistry.applications.v1alpha1.LintLocation
	(Guideline_Status)(0),        // 6: google.cloud.apigeeregistry.applications.v1alpha1.Guideline.Status
	(Rule_Severity)(0),           // 7: google.cloud.apigeeregistry.applications.v1alpha1.Rule.Severity
}
var file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_depIdxs = []int32{
	3, // 0: google.cloud.apigeeregistry.applications.v1alpha1.ConformanceReport.guideline_report_groups:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.GuidelineReportGroup
	4, // 1: google.cloud.apigeeregistry.applications.v1alpha1.GuidelineReport.rule_report_groups:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.RuleReportGroup
	5, // 2: google.cloud.apigeeregistry.applications.v1alpha1.RuleReport.location:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.LintLocation
	6, // 3: google.cloud.apigeeregistry.applications.v1alpha1.GuidelineReportGroup.status:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.Guideline.Status
	1, // 4: google.cloud.apigeeregistry.applications.v1alpha1.GuidelineReportGroup.guideline_reports:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.GuidelineReport
	7, // 5: google.cloud.apigeeregistry.applications.v1alpha1.RuleReportGroup.severity:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.Rule.Severity
	2, // 6: google.cloud.apigeeregistry.applications.v1alpha1.RuleReportGroup.rule_reports:type_name -> google.cloud.apigeeregistry.applications.v1alpha1.RuleReport
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() {
	file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_init()
}
func file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_init() {
	if File_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto != nil {
		return
	}
	file_google_cloud_apigeeregistry_applications_v1alpha1_registry_styleguide_proto_init()
	file_google_cloud_apigeeregistry_applications_v1alpha1_registry_lint_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConformanceReport); i {
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
		file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GuidelineReport); i {
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
		file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RuleReport); i {
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
		file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GuidelineReportGroup); i {
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
		file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RuleReportGroup); i {
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
			RawDescriptor: file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_goTypes,
		DependencyIndexes: file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_depIdxs,
		MessageInfos:      file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_msgTypes,
	}.Build()
	File_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto = out.File
	file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_rawDesc = nil
	file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_goTypes = nil
	file_google_cloud_apigeeregistry_applications_v1alpha1_registry_conformance_report_proto_depIdxs = nil
}
