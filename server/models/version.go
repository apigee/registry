// Copyright 2020 Google LLC. All Rights Reserved.

package models

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"apigov.dev/registry/rpc"
	"apigov.dev/registry/server/names"
	"cloud.google.com/go/datastore"
	ptypes "github.com/golang/protobuf/ptypes"
)

// VersionEntityName is used to represent versions in the datastore.
const VersionEntityName = "Version"

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

// ParseParentProduct ...
func ParseParentProduct(parent string) ([]string, error) {
	r := regexp.MustCompile("^projects/" + names.NameRegex +
		"/products/" + names.NameRegex +
		"$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid parent '%s'", parent)
	}
	return m[0], nil
}

// NewVersionFromParentAndVersionID returns an initialized product for a specified parent and productID.
func NewVersionFromParentAndVersionID(parent string, versionID string) (*Version, error) {
	r := regexp.MustCompile("^projects/" + names.NameRegex + "/products/" + names.NameRegex + "$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid product '%s'", parent)
	}
	if err := names.ValidateID(versionID); err != nil {
		return nil, err
	}
	version := &Version{}
	version.ProjectID = m[0][1]
	version.ProductID = m[0][2]
	version.VersionID = versionID
	return version, nil
}

// NewVersionFromResourceName parses resource names and returns an initialized version.
func NewVersionFromResourceName(name string) (*Version, error) {
	version := &Version{}
	m := names.VersionRegexp().FindAllStringSubmatch(name, -1)
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
	return fmt.Sprintf("projects/%s/products/%s/versions/%s", version.ProjectID, version.ProductID, version.VersionID)
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
	version.DisplayName = message.GetDisplayName()
	version.Description = message.GetDescription()
	version.State = message.GetState()
	version.UpdateTime = time.Now()
	return nil
}

// DeleteChildren deletes all the children of a version.
func (version *Version) DeleteChildren(ctx context.Context, client *datastore.Client) error {
	for _, entityName := range []string{SpecEntityName} {
		q := datastore.NewQuery(entityName)
		q = q.KeysOnly()
		q = q.Filter("ProjectID =", version.ProjectID)
		q = q.Filter("ProductID =", version.ProductID)
		q = q.Filter("VersionID =", version.VersionID)
		err := DeleteAllMatches(ctx, client, q)
		if err != nil {
			return err
		}
	}
	return nil
}
