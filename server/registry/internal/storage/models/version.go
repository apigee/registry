// Copyright 2020 Google LLC.
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
	"time"

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Version is the storage-side representation of a version.
type Version struct {
	Key          string    `gorm:"primaryKey"`
	ProjectID    string    // Uniquely identifies a project.
	ApiID        string    // Uniquely identifies an api within a project.
	VersionID    string    // Uniquely identifies a version within an api.
	DisplayName  string    // A human-friendly name.
	Description  string    // A detailed description.
	CreateTime   time.Time // Creation time.
	UpdateTime   time.Time // Time of last change.
	State        string    // Lifecycle stage.
	Labels       []byte    // Serialized labels.
	Annotations  []byte    // Serialized annotations.
	PrimarySpec  string    // Primary Spec for this version.
	ParentApiKey string
	ParentApi    *Api `gorm:"foreignKey:ParentApiKey;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// NewVersion initializes a new resource.
func NewVersion(name names.Version, body *rpc.ApiVersion) (version *Version, err error) {
	now := time.Now().Round(time.Microsecond)
	version = &Version{
		ProjectID:    name.ProjectID,
		ApiID:        name.ApiID,
		VersionID:    name.VersionID,
		Description:  body.GetDescription(),
		DisplayName:  body.GetDisplayName(),
		State:        body.GetState(),
		PrimarySpec:  body.GetPrimarySpec(),
		CreateTime:   now,
		UpdateTime:   now,
		ParentApiKey: name.Api().String(),
	}

	version.Labels, err = bytesForMap(body.GetLabels())
	if err != nil {
		return nil, err
	}

	version.Annotations, err = bytesForMap(body.GetAnnotations())
	if err != nil {
		return nil, err
	}

	return version, nil
}

// Name returns the resource name of the version.
func (v *Version) Name() string {
	return names.Version{
		ProjectID: v.ProjectID,
		ApiID:     v.ApiID,
		VersionID: v.VersionID,
	}.String()
}

// Message returns a message representing a version.
func (v *Version) Message() (message *rpc.ApiVersion, err error) {
	message = &rpc.ApiVersion{
		Name:        v.Name(),
		DisplayName: v.DisplayName,
		Description: v.Description,
		State:       v.State,
		PrimarySpec: v.PrimarySpec,
		CreateTime:  timestamppb.New(v.CreateTime),
		UpdateTime:  timestamppb.New(v.UpdateTime),
	}

	message.Labels, err = v.LabelsMap()
	if err != nil {
		return nil, err
	}

	message.Annotations, err = mapForBytes(v.Annotations)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// Update modifies a version using the contents of a message.
func (v *Version) Update(message *rpc.ApiVersion, mask *fieldmaskpb.FieldMask) error {
	v.UpdateTime = time.Now().Round(time.Microsecond)
	for _, field := range mask.Paths {
		switch field {
		case "display_name":
			v.DisplayName = message.GetDisplayName()
		case "description":
			v.Description = message.GetDescription()
		case "state":
			v.State = message.GetState()
		case "primary_spec":
			v.PrimarySpec = message.GetPrimarySpec()
		case "labels":
			var err error
			if v.Labels, err = bytesForMap(message.GetLabels()); err != nil {
				return err
			}
		case "annotations":
			var err error
			if v.Annotations, err = bytesForMap(message.GetAnnotations()); err != nil {
				return err
			}
		}
	}

	return nil
}

// LabelsMap returns a map representation of stored labels.
func (v *Version) LabelsMap() (map[string]string, error) {
	return mapForBytes(v.Labels)
}
