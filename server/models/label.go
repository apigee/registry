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
)

// LabelEntityName is used to represent labels in storage.
const LabelEntityName = "Label"

// Label is the storage-side representation of a label.
type Label struct {
	Key        string    `datastore:"-", gorm:"primaryKey"`
	ProjectID  string    // Project associated with label (required).
	ApiID      string    // Api associated with label (if appropriate).
	VersionID  string    // Version associated with label (if appropriate).
	SpecID     string    // Spec associated with label (if appropriate).
	LabelID    string    // Label identifier (required).
	CreateTime time.Time // Creation time.
	UpdateTime time.Time // Time of last change.
	Subject    string    // Subject of the label.
}

// NewLabelFromParentAndLabelID returns an initialized label for a specified parent and labelID.
func NewLabelFromParentAndLabelID(parent string, labelID string) (*Label, error) {
	// Return an error if the labelID is invalid.
	if err := names.ValidateID(labelID); err != nil {
		return nil, err
	}
	// Match regular expressions to identify the parent of this label.
	var m []string
	// Is the parent a project?
	m = names.ProjectRegexp().FindStringSubmatch(parent)
	if m != nil {
		return &Label{
			ProjectID: m[1],
			LabelID:   labelID,
			Subject:   parent,
		}, nil
	}
	// Is the parent an api?
	m = names.ApiRegexp().FindStringSubmatch(parent)
	if m != nil {
		return &Label{
			ProjectID: m[1],
			ApiID:     m[2],
			LabelID:   labelID,
			Subject:   parent,
		}, nil
	}
	// Is the parent a version?
	m = names.VersionRegexp().FindStringSubmatch(parent)
	if m != nil {
		return &Label{
			ProjectID: m[1],
			ApiID:     m[2],
			VersionID: m[3],
			LabelID:   labelID,
			Subject:   parent,
		}, nil
	}
	// Is the parent a spec?
	m = names.SpecRegexp().FindStringSubmatch(parent)
	if m != nil {
		return &Label{
			ProjectID: m[1],
			ApiID:     m[2],
			VersionID: m[3],
			SpecID:    m[4],
			LabelID:   labelID,
			Subject:   parent,
		}, nil
	}
	// Return an error for an unrecognized parent.
	return nil, fmt.Errorf("invalid parent '%s'", parent)
}

// NewLabelFromResourceName parses resource names and returns an initialized label.
func NewLabelFromResourceName(name string) (*Label, error) {
	// split name into parts
	parts := strings.Split(name, "/")
	if parts[len(parts)-2] != "labels" {
		return nil, fmt.Errorf("invalid label name '%s'", name)
	}
	// build label from parent and labelID
	parent := strings.Join(parts[0:len(parts)-2], "/")
	labelID := parts[len(parts)-1]
	return NewLabelFromParentAndLabelID(parent, labelID)
}

// ResourceName generates the resource name of a label.
func (label *Label) ResourceName() string {
	switch {
	case label.SpecID != "":
		return fmt.Sprintf("projects/%s/apis/%s/versions/%s/specs/%s/labels/%s",
			label.ProjectID, label.ApiID, label.VersionID, label.SpecID, label.LabelID)
	case label.VersionID != "":
		return fmt.Sprintf("projects/%s/apis/%s/versions/%s/labels/%s",
			label.ProjectID, label.ApiID, label.VersionID, label.LabelID)
	case label.ApiID != "":
		return fmt.Sprintf("projects/%s/apis/%s/labels/%s",
			label.ProjectID, label.ApiID, label.LabelID)
	case label.ProjectID != "":
		return fmt.Sprintf("projects/%s/labels/%s",
			label.ProjectID, label.LabelID)
	default:
		return "UNKNOWN"
	}
}

// Message returns a message representing a label.
func (label *Label) Message() (message *rpc.Label, err error) {
	message = &rpc.Label{}
	message.Name = label.ResourceName()
	message.Subject = label.Subject
	message.CreateTime, err = ptypes.TimestampProto(label.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(label.UpdateTime)
	return message, err
}

// Update modifies a label using the contents of a message.
func (label *Label) Update(message *rpc.Label) error {
	label.Subject = message.GetSubject()
	label.UpdateTime = time.Now()
	return nil
}
