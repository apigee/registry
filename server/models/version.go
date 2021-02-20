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
	"regexp"
	"time"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	ptypes "github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// VersionEntityName is used to represent versions in storage.
const VersionEntityName = "Version"

// Version is the storage-side representation of a version.
type Version struct {
	Key         string    `datastore:"-" gorm:"primaryKey"`
	ProjectID   string    // Uniquely identifies a project.
	ApiID       string    // Uniquely identifies an api within a project.
	VersionID   string    // Uniquely identifies a version wihtin a api.
	DisplayName string    // A human-friendly name.
	Description string    // A detailed description.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
	State       string    // Lifecycle stage.
	Labels      []byte    `datastore:",noindex"` // Serialized labels.
	Annotations []byte    `datastore:",noindex"` // Serialized annotations.
}

// ParseParentApi parses the name of an API that is the parent of a version.
func ParseParentApi(parent string) ([]string, error) {
	r := regexp.MustCompile("^projects/" + names.NameRegex +
		"/apis/" + names.NameRegex +
		"$")
	m := r.FindStringSubmatch(parent)
	if m == nil {
		return nil, fmt.Errorf("invalid parent '%s'", parent)
	}
	return m, nil
}

// NewVersionFromParentAndVersionID returns an initialized api for a specified parent and apiID.
func NewVersionFromParentAndVersionID(parent string, versionID string) (*Version, error) {
	r := regexp.MustCompile("^projects/" + names.NameRegex + "/apis/" + names.NameRegex + "$")
	m := r.FindStringSubmatch(parent)
	if m == nil {
		return nil, fmt.Errorf("invalid api '%s'", parent)
	}
	if err := names.ValidateID(versionID); err != nil {
		return nil, err
	}
	version := &Version{}
	version.ProjectID = m[1]
	version.ApiID = m[2]
	version.VersionID = versionID
	return version, nil
}

// NewVersionFromResourceName parses resource names and returns an initialized version.
func NewVersionFromResourceName(name string) (*Version, error) {
	version := &Version{}
	m := names.VersionRegexp().FindStringSubmatch(name)
	if m == nil {
		return nil, fmt.Errorf("invalid version name (%s)", name)
	}
	version.ProjectID = m[1]
	version.ApiID = m[2]
	version.VersionID = m[3]
	return version, nil
}

// NewVersionFromMessage returns an initialized version from a message.
func NewVersionFromMessage(message *rpc.ApiVersion) (*Version, error) {
	version, err := NewVersionFromResourceName(message.GetName())
	if err != nil {
		return nil, err
	}
	version.DisplayName = message.GetDisplayName()
	version.Description = message.GetDescription()
	version.State = message.GetState()
	return version, nil
}

// ResourceName generates the resource name of a version.
func (version *Version) ResourceName() string {
	return fmt.Sprintf("projects/%s/apis/%s/versions/%s", version.ProjectID, version.ApiID, version.VersionID)
}

// Message returns a message representing a version.
func (version *Version) Message(view rpc.View) (message *rpc.ApiVersion, err error) {
	message = &rpc.ApiVersion{}
	message.Name = version.ResourceName()
	message.DisplayName = version.DisplayName
	message.Description = version.Description
	message.CreateTime, err = ptypes.TimestampProto(version.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(version.UpdateTime)
	message.State = version.State
	if message.Labels, err = mapForBytes(version.Labels); err != nil {
		return nil, err
	}
	if view == rpc.View_FULL {
		if message.Annotations, err = mapForBytes(version.Annotations); err != nil {
			return nil, err
		}
	}
	return message, err
}

// Update modifies a version using the contents of a message.
func (version *Version) Update(message *rpc.ApiVersion, mask *fieldmaskpb.FieldMask) error {
	if activeUpdateMask(mask) {
		for _, field := range mask.Paths {
			switch field {
			case "display_name":
				version.DisplayName = message.GetDisplayName()
			case "description":
				version.Description = message.GetDescription()
			case "state":
				version.State = message.GetState()
			case "labels":
				var err error
				if version.Labels, err = bytesForMap(message.GetLabels()); err != nil {
					return err
				}
			case "annotations":
				var err error
				if version.Annotations, err = bytesForMap(message.GetAnnotations()); err != nil {
					return err
				}
			}
		}
	} else {
		version.DisplayName = message.GetDisplayName()
		version.Description = message.GetDescription()
		version.State = message.GetState()
		var err error
		if version.Labels, err = bytesForMap(message.GetLabels()); err != nil {
			return err
		}
		if version.Annotations, err = bytesForMap(message.GetAnnotations()); err != nil {
			return err
		}
	}
	version.UpdateTime = time.Now()
	return nil
}
