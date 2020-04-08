// Copyright 2020 Google Inc. All Rights Reserved.

package models

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	rpc "apigov.dev/flame/rpc"
	"cloud.google.com/go/datastore"
	ptypes "github.com/golang/protobuf/ptypes"
)

// SpecEntityName is used to represent specs in the datastore.
const SpecEntityName = "Spec"

// Spec ...
type Spec struct {
	ProjectID   string    // Uniquely identifies a project.
	ProductID   string    // Uniquely identifies a product within a project.
	VersionID   string    // Uniquely identifies a version within a product.
	SpecID      string    // Uniquely identifies a spec within a version.
	DisplayName string    // A human-friendly name.
	Description string    // A detailed description.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
	Style       string    // Specification format.
}

// ParseParentVersion ...
func ParseParentVersion(parent string) ([]string, error) {
	r := regexp.MustCompile("^projects/" + nameRegex +
		"/products/" + nameRegex +
		"/versions/" + nameRegex +
		"$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid version '%s'", parent)
	}
	return m[0], nil
}

// NewSpecFromParentAndSpecID returns an initialized spec for a specified parent and specID.
func NewSpecFromParentAndSpecID(parent string, specID string) (*Spec, error) {
	r := regexp.MustCompile("^projects/" + nameRegex +
		"/products/" + nameRegex +
		"/versions/" + nameRegex + "$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid parent '%s'", parent)
	}
	if err := validateID(specID); err != nil {
		return nil, err
	}
	spec := &Spec{}
	spec.ProjectID = m[0][1]
	spec.ProductID = m[0][2]
	spec.VersionID = m[0][3]
	spec.SpecID = specID
	return spec, nil
}

// NewSpecFromResourceName parses resource names and returns an initialized spec.
func NewSpecFromResourceName(name string) (*Spec, error) {
	spec := &Spec{}
	m := SpecRegexp().FindAllStringSubmatch(name, -1)
	if m == nil {
		return nil, errors.New("invalid spec name")
	}
	spec.ProjectID = m[0][1]
	spec.ProductID = m[0][2]
	spec.VersionID = m[0][3]
	spec.SpecID = m[0][4]
	return spec, nil
}

// NewSpecFromMessage returns an initialized spec from a message.
func NewSpecFromMessage(message *rpc.Spec) (*Spec, error) {
	spec, err := NewSpecFromResourceName(message.GetName())
	if err != nil {
		return nil, err
	}
	spec.DisplayName = message.GetDisplayName()
	spec.Description = message.GetDescription()
	//spec.Availability = message.GetAvailability()
	//spec.RecommendedVersion = message.GetRecommendedVersion()
	return spec, nil
}

// ResourceName generates the resource name of a spec.
func (spec *Spec) ResourceName() string {
	return fmt.Sprintf("projects/%s/products/%s/versions/%s/specs/%s", spec.ProjectID, spec.ProductID, spec.VersionID, spec.SpecID)
}

// Message returns a message representing a spec.
func (spec *Spec) Message() (message *rpc.Spec, err error) {
	message = &rpc.Spec{}
	message.Name = spec.ResourceName()
	message.DisplayName = spec.DisplayName
	message.Description = spec.Description
	message.CreateTime, err = ptypes.TimestampProto(spec.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(spec.UpdateTime)
	//message.Availability = spec.Availability
	//message.RecommendedVersion = spec.RecommendedVersion
	return message, err
}

// Update modifies a spec using the contents of a message.
func (spec *Spec) Update(message *rpc.Spec) error {
	spec.Style = message.GetStyle()
	spec.UpdateTime = spec.CreateTime
	return nil
}

// DeleteChildren deletes all the children of a spec.
func (spec *Spec) DeleteChildren(ctx context.Context, client *datastore.Client) error {
	for _, entityName := range []string{FileEntityName} {
		q := datastore.NewQuery(entityName)
		q = q.KeysOnly()
		q = q.Filter("ProjectID =", spec.ProjectID)
		q = q.Filter("ProductID =", spec.ProductID)
		q = q.Filter("VersionID =", spec.VersionID)
		q = q.Filter("SpecID =", spec.SpecID)
		err := deleteAllMatches(ctx, client, q)
		if err != nil {
			return err
		}
	}
	return nil
}
