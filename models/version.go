// Copyright 2020 Google Inc. All Rights Reserved.

package models

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	rpc "apigov.dev/flame/rpc"
	ptypes "github.com/golang/protobuf/ptypes"
)

// Version ...
type Version struct {
	ProjectID   string    // Uniquely identifies a project.
	ProductID   string    // Uniquely identifies a product within a project.
	VersionID   string    // Uniquely identifies a version wihtin a product.
	DisplayName string    // A human-friendly name.
	Description string    // A detailed description.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
	State       string    // Lifecycle stage.
}

// NewVersionFromResourceName parses resource names and returns an initialized version.
func NewVersionFromResourceName(name string) (*Version, error) {
	version := &Version{}
	r := regexp.MustCompile("^/projects/" + nameRegex + "/products/" + nameRegex + "/versions/" + nameRegex + "$")
	m := r.FindAllStringSubmatch(name, -1)
	if m == nil {
		return nil, errors.New("invalid version name")
	}
	version.ProjectID = m[0][1]
	version.ProductID = m[0][2]
	version.VersionID = m[0][3]
	return version, nil
}

// NewVersionFromMessage returns an initialized version from a message.
func NewVersionFromMessage(message *rpc.Version) (*Version, error) {
	version, err := NewVersionFromResourceName(message.GetName())
	if err != nil {
		return nil, err
	}
	version.DisplayName = message.GetDisplayName()
	version.Description = message.GetDescription()
	//version.Availability = message.GetAvailability()
	//version.RecommendedVersion = message.GetRecommendedVersion()
	return version, nil
}

// ResourceName generates the resource name of a version.
func (version *Version) ResourceName() string {
	return fmt.Sprintf("/projects/%s/products/%s/versions/%s", version.ProjectID, version.ProductID, version.VersionID)
}

// Message returns a message representing a version.
func (version *Version) Message() (message *rpc.Version, err error) {
	message = &rpc.Version{}
	message.Name = version.ResourceName()
	message.DisplayName = version.DisplayName
	message.Description = version.Description
	message.CreateTime, err = ptypes.TimestampProto(version.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(version.UpdateTime)
	//message.Availability = version.Availability
	//message.RecommendedVersion = version.RecommendedVersion
	return message, err
}

// Update modifies a version using the contents of a message.
func (version *Version) Update(message *rpc.Version) error {
	version.UpdateTime = version.CreateTime
	return nil
}
