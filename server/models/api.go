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
)

// ApiEntityName is used to represent apis in storage.
const ApiEntityName = "Api"

// Api ...
type Api struct {
	Key                string    `datastore:"-", gorm:"PRIMARY_KEY"`
	ProjectID          string    // Uniquely identifies a project.
	ApiID              string    // Uniquely identifies an api within a project.
	DisplayName        string    // A human-friendly name.
	Description        string    // A detailed description.
	CreateTime         time.Time // Creation time.
	UpdateTime         time.Time // Time of last change.
	Availability       string    // Availability of the API.
	RecommendedVersion string    // Recommended API version.
}

// NewApiFromParentAndApiID returns an initialized api for a specified parent and apiID.
func NewApiFromParentAndApiID(parent string, apiID string) (*Api, error) {
	r := regexp.MustCompile("^projects/" + names.NameRegex + "$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid parent '%s'", parent)
	}
	if err := names.ValidateID(apiID); err != nil {
		return nil, err
	}
	api := &Api{}
	api.ProjectID = m[0][1]
	api.ApiID = apiID
	return api, nil
}

// NewApiFromResourceName parses resource names and returns an initialized api.
func NewApiFromResourceName(name string) (*Api, error) {
	api := &Api{}
	m := names.ApiRegexp().FindAllStringSubmatch(name, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid api name (%s)", name)
	}
	api.ProjectID = m[0][1]
	api.ApiID = m[0][2]
	return api, nil
}

// ResourceName generates the resource name of a api.
func (api *Api) ResourceName() string {
	return fmt.Sprintf("projects/%s/apis/%s", api.ProjectID, api.ApiID)
}

// Message returns a message representing a api.
func (api *Api) Message() (message *rpc.Api, err error) {
	message = &rpc.Api{}
	message.Name = api.ResourceName()
	message.DisplayName = api.DisplayName
	message.Description = api.Description
	message.CreateTime, err = ptypes.TimestampProto(api.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(api.UpdateTime)
	message.Availability = api.Availability
	message.RecommendedVersion = api.RecommendedVersion
	return message, err
}

// Update modifies a api using the contents of a message.
func (api *Api) Update(message *rpc.Api) error {
	api.DisplayName = message.GetDisplayName()
	api.Description = message.GetDescription()
	api.Availability = message.GetAvailability()
	api.RecommendedVersion = message.GetRecommendedVersion()
	api.UpdateTime = time.Now()
	return nil
}
