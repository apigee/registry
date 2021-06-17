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
	"time"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Api is the storage-side representation of an API.
type Api struct {
	Key                string    `gorm:"primaryKey"`
	ProjectID          string    // Uniquely identifies a project.
	ApiID              string    // Uniquely identifies an api within a project.
	DisplayName        string    // A human-friendly name.
	Description        string    // A detailed description.
	CreateTime         time.Time // Creation time.
	UpdateTime         time.Time // Time of last change.
	Availability       string    // Availability of the API.
	RecommendedVersion string    // Recommended API version.
	Labels             []byte    // Serialized labels.
	Annotations        []byte    // Serialized annotations.
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

// Name returns the resource name of the api.
func (api *Api) Name() string {
	return names.Api{
		ProjectID: api.ProjectID,
		ApiID:     api.ApiID,
	}.String()
}

// Message returns a message representing an api.
func (api *Api) Message() (message *rpc.Api, err error) {
	message = &rpc.Api{
		Name:               api.Name(),
		DisplayName:        api.DisplayName,
		Description:        api.Description,
		Availability:       api.Availability,
		RecommendedVersion: api.RecommendedVersion,
	}

	message.CreateTime, err = ptypes.TimestampProto(api.CreateTime)
	if err != nil {
		return nil, err
	}

	message.UpdateTime, err = ptypes.TimestampProto(api.UpdateTime)
	if err != nil {
		return nil, err
	}

	message.Labels, err = api.LabelsMap()
	if err != nil {
		return nil, err
	}

	message.Annotations, err = mapForBytes(api.Annotations)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// Update modifies a api using the contents of a message.
func (api *Api) Update(message *rpc.Api, mask *fieldmaskpb.FieldMask) error {
	api.UpdateTime = time.Now()
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

	return nil
}

// LabelsMap returns a map representation of stored labels.
func (api *Api) LabelsMap() (map[string]string, error) {
	return mapForBytes(api.Labels)
}
