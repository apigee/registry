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
	"time"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	ptypes "github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ApiEntityName is used to represent apis in storage.
const ApiEntityName = "Api"

// Api is the storage-side representation of an API.
type Api struct {
	Key                string    `datastore:"-" gorm:"primaryKey"`
	ProjectID          string    // Uniquely identifies a project.
	ApiID              string    // Uniquely identifies an api within a project.
	DisplayName        string    // A human-friendly name.
	Description        string    // A detailed description.
	CreateTime         time.Time // Creation time.
	UpdateTime         time.Time // Time of last change.
	Availability       string    // Availability of the API.
	RecommendedVersion string    // Recommended API version.
	Labels             []byte    `datastore:",noindex"` // Serialized labels.
	Annotations        []byte    `datastore:",noindex"` // Serialized annotations.
}

// NewApiFromParentAndApiID returns an initialized api for a specified parent and ID.
func NewApiFromParentAndApiID(parent string, id string) (*Api, error) {
	m, err := names.ParseProject(parent)
	if err != nil {
		return nil, err
	} else if err := names.ValidateCustomID(id); err != nil {
		return nil, err
	}

	return &Api{
		ProjectID: m[1],
		ApiID:     id,
	}, nil
}

// NewApiFromResourceName parses resource names and returns an initialized api.
func NewApiFromResourceName(name string) (*Api, error) {
	m, err := names.ParseApi(name)
	if err != nil {
		return nil, err
	}

	return &Api{
		ProjectID: m[1],
		ApiID:     m[2],
	}, nil
}

// ResourceName generates the resource name of a api.
func (api *Api) ResourceName() string {
	return fmt.Sprintf("projects/%s/apis/%s", api.ProjectID, api.ApiID)
}

// Message returns a message representing an api.
func (api *Api) Message(view rpc.View) (message *rpc.Api, err error) {
	message = &rpc.Api{}
	message.Name = api.ResourceName()
	message.DisplayName = api.DisplayName
	message.Description = api.Description
	message.CreateTime, err = ptypes.TimestampProto(api.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(api.UpdateTime)
	message.Availability = api.Availability
	message.RecommendedVersion = api.RecommendedVersion
	if message.Labels, err = mapForBytes(api.Labels); err != nil {
		return nil, err
	}
	if view == rpc.View_FULL {
		if message.Annotations, err = mapForBytes(api.Annotations); err != nil {
			return nil, err
		}
	}
	return message, nil
}

// Update modifies a api using the contents of a message.
func (api *Api) Update(message *rpc.Api, mask *fieldmaskpb.FieldMask) error {
	if activeUpdateMask(mask) {
		for _, field := range mask.Paths {
			switch field {
			case "display_name":
				api.DisplayName = message.GetDisplayName()
			case "description":
				api.Description = message.GetDescription()
			case "availability":
				api.Availability = message.GetAvailability()
			case "recommended_version":
				api.RecommendedVersion = message.GetRecommendedVersion()
			case "labels":
				var err error
				if api.Labels, err = bytesForMap(message.GetLabels()); err != nil {
					return err
				}
			case "annotations":
				var err error
				if api.Annotations, err = bytesForMap(message.GetAnnotations()); err != nil {
					return err
				}
			}
		}
	} else {
		api.DisplayName = message.GetDisplayName()
		api.Description = message.GetDescription()
		api.Availability = message.GetAvailability()
		api.RecommendedVersion = message.GetRecommendedVersion()
		var err error
		if api.Labels, err = bytesForMap(message.GetLabels()); err != nil {
			return err
		}
		if api.Annotations, err = bytesForMap(message.GetAnnotations()); err != nil {
			return err
		}
	}
	api.UpdateTime = time.Now()
	return nil
}

// LabelsMap returns a map representation of stored labels.
func (api *Api) LabelsMap() (map[string]string, error) {
	return mapForBytes(api.Labels)
}
