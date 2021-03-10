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
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

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

// NewApi initializes a new resource.
func NewApi(name names.Api, body *rpc.Api) (api *Api, err error) {
	now := time.Now()
	api = &Api{
		ProjectID:          name.ProjectID,
		ApiID:              name.ApiID,
		Description:        body.GetDescription(),
		DisplayName:        body.GetDisplayName(),
		Availability:       body.GetAvailability(),
		RecommendedVersion: body.GetRecommendedVersion(),
		CreateTime:         now,
		UpdateTime:         now,
	}

	api.Labels, err = bytesForMap(body.GetLabels())
	if err != nil {
		return nil, err
	}

	api.Annotations, err = bytesForMap(body.GetAnnotations())
	if err != nil {
		return nil, err
	}

	return api, nil
}

// ResourceName generates the resource name of a api.
func (a *Api) ResourceName() string {
	return fmt.Sprintf("projects/%s/apis/%s", a.ProjectID, a.ApiID)
}

// Message returns a message representing an api.
func (a *Api) Message(view rpc.View) (message *rpc.Api, err error) {
	message = &rpc.Api{
		Name:               a.ResourceName(),
		DisplayName:        a.DisplayName,
		Description:        a.Description,
		Availability:       a.Availability,
		RecommendedVersion: a.RecommendedVersion,
	}

	message.CreateTime, err = ptypes.TimestampProto(a.CreateTime)
	if err != nil {
		return nil, err
	}

	message.UpdateTime, err = ptypes.TimestampProto(a.UpdateTime)
	if err != nil {
		return nil, err
	}

	message.Labels, err = a.LabelsMap()
	if err != nil {
		return nil, err
	}

	if view == rpc.View_FULL {
		message.Annotations, err = mapForBytes(a.Annotations)
		if err != nil {
			return nil, err
		}
	}

	return message, nil
}

// Update modifies a api using the contents of a message.
func (a *Api) Update(message *rpc.Api, mask *fieldmaskpb.FieldMask) error {
	if activeUpdateMask(mask) {
		for _, field := range mask.Paths {
			switch field {
			case "display_name":
				a.DisplayName = message.GetDisplayName()
			case "description":
				a.Description = message.GetDescription()
			case "availability":
				a.Availability = message.GetAvailability()
			case "recommended_version":
				a.RecommendedVersion = message.GetRecommendedVersion()
			case "labels":
				var err error
				if a.Labels, err = bytesForMap(message.GetLabels()); err != nil {
					return err
				}
			case "annotations":
				var err error
				if a.Annotations, err = bytesForMap(message.GetAnnotations()); err != nil {
					return err
				}
			}
		}
	} else {
		a.DisplayName = message.GetDisplayName()
		a.Description = message.GetDescription()
		a.Availability = message.GetAvailability()
		a.RecommendedVersion = message.GetRecommendedVersion()
		var err error
		if a.Labels, err = bytesForMap(message.GetLabels()); err != nil {
			return err
		}
		if a.Annotations, err = bytesForMap(message.GetAnnotations()); err != nil {
			return err
		}
	}
	a.UpdateTime = time.Now()
	return nil
}

// LabelsMap returns a map representation of stored labels.
func (a *Api) LabelsMap() (map[string]string, error) {
	return mapForBytes(a.Labels)
}
