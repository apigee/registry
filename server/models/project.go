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

// ProjectEntityName is used to represent projrcts in storage.
const ProjectEntityName = "Project"

// Project is the storage-side representation of a project.
type Project struct {
	Key         string    `datastore:"-" gorm:"primaryKey"`
	ProjectID   string    // Uniquely identifies a project.
	DisplayName string    // A human-friendly name.
	Description string    // A detailed description.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
}

// NewProjectFromProjectID returns an initialized project for a specified ID.
func NewProjectFromProjectID(id string) (*Project, error) {
	if err := names.ValidateCustomID(id); err != nil {
		return nil, err
	}

	return &Project{
		ProjectID: id,
	}, nil
}

// NewProjectFromResourceName parses resource names and returns an initialized project.
func NewProjectFromResourceName(name string) (*Project, error) {
	m, err := names.ParseProject(name)
	if err != nil {
		return nil, err
	}

	return &Project{
		ProjectID: m[1],
	}, nil
}

// ResourceName generates the resource name of a project.
func (project *Project) ResourceName() string {
	return fmt.Sprintf("projects/%s", project.ProjectID)
}

// Message returns a message representing a project.
func (project *Project) Message() (message *rpc.Project, err error) {
	message = &rpc.Project{}
	message.Name = project.ResourceName()
	message.DisplayName = project.DisplayName
	message.Description = project.Description
	message.CreateTime, err = ptypes.TimestampProto(project.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(project.UpdateTime)
	return message, err
}

// Update modifies a project using the contents of a message.
func (project *Project) Update(message *rpc.Project, mask *fieldmaskpb.FieldMask) error {
	if activeUpdateMask(mask) {
		for _, field := range mask.Paths {
			switch field {
			case "display_name":
				project.DisplayName = message.GetDisplayName()
			case "description":
				project.Description = message.GetDescription()
			}
		}
	} else {
		project.DisplayName = message.GetDisplayName()
		project.Description = message.GetDescription()
	}
	project.UpdateTime = time.Now()
	return nil
}
