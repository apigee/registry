// Copyright 2020 Google LLC. All Rights Reserved.

package models

import (
	"fmt"
	"regexp"
	"time"

	rpc "apigov.dev/flame/rpc"
	ptypes "github.com/golang/protobuf/ptypes"
)

// PropertyEntityName is used to represent properties in the datastore.
const PropertyEntityName = "Property"

// PropertiesRegexp returns a regular expression that matches collection of properties.
func PropertiesRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/properties$")
}

// PropertyRegexp returns a regular expression that matches a property resource name.
func PropertyRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + nameRegex + "/properties/" + nameRegex + "$")
}

type ValueType int

const (
	StringType ValueType = iota
	Int64Type
	DoubleType
	BoolType
	BytesType
)

// Property ...
type Property struct {
	ProjectID   string    // Uniquely identifies a project.
	PropertyID  string    // Uniquely identifies a property within a project.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
	Subject     string    // Subject of the property.
	Relation    string    // Relation of the property.
	ValueType   ValueType // Type of the property value.
	StringValue string    // Property value (if string).
	Int64Value  int64     // Property value (if int64).
	DoubleValue float64   // Property value (if double).
	BoolValue   bool      // Property value (if bool).
	BytesValue  []byte    `datastore:",noindex"` // Property value (if bytes).
}

// NewPropertyFromParentAndPropertyID returns an initialized property for a specified parent and propertyID.
func NewPropertyFromParentAndPropertyID(parent string, propertyID string) (*Property, error) {
	r := regexp.MustCompile("^projects/" + nameRegex + "$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid parent '%s'", parent)
	}
	if err := validateID(propertyID); err != nil {
		return nil, err
	}
	property := &Property{}
	property.ProjectID = m[0][1]
	property.PropertyID = propertyID
	return property, nil
}

// NewPropertyFromResourceName parses resource names and returns an initialized property.
func NewPropertyFromResourceName(name string) (*Property, error) {
	property := &Property{}
	m := PropertyRegexp().FindAllStringSubmatch(name, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid property name (%s)", name)
	}
	property.ProjectID = m[0][1]
	property.PropertyID = m[0][2]
	return property, nil
}

// ResourceName generates the resource name of a property.
func (property *Property) ResourceName() string {
	return fmt.Sprintf("projects/%s/properties/%s", property.ProjectID, property.PropertyID)
}

// Message returns a message representing a property.
func (property *Property) Message() (message *rpc.Property, err error) {
	message = &rpc.Property{}
	message.Name = property.ResourceName()
	message.Subject = property.Subject
	message.Relation = property.Relation
	message.CreateTime, err = ptypes.TimestampProto(property.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(property.UpdateTime)
	// TODO values
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
	}
	return message, err
}

// Update modifies a property using the contents of a message.
func (property *Property) Update(message *rpc.Property) error {
	property.Subject = message.GetSubject()
	property.Relation = message.GetRelation()
	property.UpdateTime = property.CreateTime
	// TODO values
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
	default:
	}
	return nil
}
