// Copyright 2020 Google LLC. All Rights Reserved.
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

package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	ptypes "github.com/golang/protobuf/ptypes"
	any "github.com/golang/protobuf/ptypes/any"
)

// PropertyEntityName is used to represent properties in storage.
const PropertyEntityName = "Property"

// PropertyValueType is an enum representing the types of values stored in properties.
type PropertyValueType int

const (
	// StringType indicates that the stored property is a string.
	StringType PropertyValueType = iota
	// Int64Type indicates that the stored property is an integer.
	Int64Type
	// DoubleType indicates that the stored property is a double
	DoubleType
	// BoolType indicates that the stored property is a boolean.
	BoolType
	// BytesType indicates that the stored property is a range of bytes.
	BytesType
	// AnyType indicates that the stored property is a protobuf "Any" type.
	AnyType
)

// Property ...
type Property struct {
	ProjectID   string            // Project associated with property (required).
	ApiID       string            // Api associated with property (if appropriate).
	VersionID   string            // Version associated with property (if appropriate).
	SpecID      string            // Spec associated with property (if appropriate).
	PropertyID  string            // Property identifier (required).
	CreateTime  time.Time         // Creation time.
	UpdateTime  time.Time         // Time of last change.
	Subject     string            // Subject of the property.
	ValueType   PropertyValueType // Type of the property value.
	StringValue string            // Property value (if string).
	Int64Value  int64             // Property value (if int64).
	DoubleValue float64           // Property value (if double).
	BoolValue   bool              // Property value (if bool).
	BytesValue  []byte            `datastore:",noindex"` // Property value (if bytes).
}

// messageValue returns an Any object corresponding to the stored value (assuming one exists).
func (property *Property) messageValue() *any.Any {
	if property.ValueType == AnyType {
		return &any.Any{
			TypeUrl: property.StringValue,
			Value:   property.BytesValue,
		}
	}
	return nil
}

// NewPropertyFromParentAndPropertyID returns an initialized property for a specified parent and propertyID.
func NewPropertyFromParentAndPropertyID(parent string, propertyID string) (*Property, error) {
	// Return an error if the propertyID is invalid.
	if err := names.ValidateID(propertyID); err != nil {
		return nil, err
	}
	// Match regular expressions to identify the parent of this property.
	var m [][]string
	// Is the parent a project?
	m = names.ProjectRegexp().FindAllStringSubmatch(parent, -1)
	if m != nil {
		return &Property{
			ProjectID:  m[0][1],
			PropertyID: propertyID,
			Subject:    parent,
		}, nil
	}
	// Is the parent a api?
	m = names.ApiRegexp().FindAllStringSubmatch(parent, -1)
	if m != nil {
		return &Property{
			ProjectID:  m[0][1],
			ApiID:      m[0][2],
			PropertyID: propertyID,
			Subject:    parent,
		}, nil
	}
	// Is the parent a version?
	m = names.VersionRegexp().FindAllStringSubmatch(parent, -1)
	if m != nil {
		return &Property{
			ProjectID:  m[0][1],
			ApiID:      m[0][2],
			VersionID:  m[0][3],
			PropertyID: propertyID,
			Subject:    parent,
		}, nil
	}
	// Is the parent a spec?
	m = names.SpecRegexp().FindAllStringSubmatch(parent, -1)
	if m != nil {
		return &Property{
			ProjectID:  m[0][1],
			ApiID:      m[0][2],
			VersionID:  m[0][3],
			SpecID:     m[0][4],
			PropertyID: propertyID,
			Subject:    parent,
		}, nil
	}
	// Return an error for an unrecognized parent.
	return nil, fmt.Errorf("invalid parent '%s'", parent)
}

// NewPropertyFromResourceName parses resource names and returns an initialized property.
func NewPropertyFromResourceName(name string) (*Property, error) {
	// split name into parts
	parts := strings.Split(name, "/")
	if len(parts) < 2 || parts[len(parts)-2] != "properties" {
		return nil, fmt.Errorf("invalid property name '%s'", name)
	}
	// build property from parent and propertyID
	parent := strings.Join(parts[0:len(parts)-2], "/")
	propertyID := parts[len(parts)-1]
	return NewPropertyFromParentAndPropertyID(parent, propertyID)
}

// ResourceName generates the resource name of a property.
func (property *Property) ResourceName() string {
	switch {
	case property.SpecID != "":
		return fmt.Sprintf("projects/%s/apis/%s/versions/%s/specs/%s/properties/%s",
			property.ProjectID, property.ApiID, property.VersionID, property.SpecID, property.PropertyID)
	case property.VersionID != "":
		return fmt.Sprintf("projects/%s/apis/%s/versions/%s/properties/%s",
			property.ProjectID, property.ApiID, property.VersionID, property.PropertyID)
	case property.ApiID != "":
		return fmt.Sprintf("projects/%s/apis/%s/properties/%s",
			property.ProjectID, property.ApiID, property.PropertyID)
	case property.ProjectID != "":
		return fmt.Sprintf("projects/%s/properties/%s",
			property.ProjectID, property.PropertyID)
	default:
		return "UNKNOWN"
	}
}

// Message returns a message representing a property.
func (property *Property) Message() (message *rpc.Property, err error) {
	message = &rpc.Property{}
	message.Name = property.ResourceName()
	message.Subject = property.Subject
	message.Relation = property.PropertyID
	message.CreateTime, err = ptypes.TimestampProto(property.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(property.UpdateTime)
	switch property.ValueType {
	case StringType:
		message.Value = &rpc.Property_StringValue{StringValue: property.StringValue}
	case Int64Type:
		message.Value = &rpc.Property_Int64Value{Int64Value: property.Int64Value}
	case DoubleType:
		message.Value = &rpc.Property_DoubleValue{DoubleValue: property.DoubleValue}
	case BoolType:
		message.Value = &rpc.Property_BoolValue{BoolValue: property.BoolValue}
	case BytesType:
		message.Value = &rpc.Property_BytesValue{BytesValue: property.BytesValue}
	case AnyType:
		message.Value = &rpc.Property_MessageValue{MessageValue: property.messageValue()}
	}
	return message, err
}

// Update modifies a property using the contents of a message.
func (property *Property) Update(message *rpc.Property) error {
	property.Subject = message.GetSubject()
	property.UpdateTime = time.Now()
	switch message.GetValue().(type) {
	case *rpc.Property_StringValue:
		property.ValueType = StringType
		property.StringValue = message.GetStringValue()
	case *rpc.Property_Int64Value:
		property.ValueType = Int64Type
		property.Int64Value = message.GetInt64Value()
	case *rpc.Property_DoubleValue:
		property.ValueType = DoubleType
		property.DoubleValue = message.GetDoubleValue()
	case *rpc.Property_BoolValue:
		property.ValueType = BoolType
		property.BoolValue = message.GetBoolValue()
	case *rpc.Property_BytesValue:
		property.ValueType = BytesType
		property.BytesValue = message.GetBytesValue()
	case *rpc.Property_MessageValue:
		property.ValueType = AnyType
		property.StringValue = message.GetMessageValue().GetTypeUrl()
		property.BytesValue = message.GetMessageValue().GetValue()
	default:
	}
	return nil
}
