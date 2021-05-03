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

// Project is the storage-side representation of a project.
type Project struct {
	Key         string    `gorm:"primaryKey"`
	ProjectID   string    // Uniquely identifies a project.
	DisplayName string    // A human-friendly name.
	Description string    // A detailed description.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
}

// NewProject initializes a new resource.
func NewProject(name names.Project, body *rpc.Project) *Project {
	now := time.Now()
	return &Project{
		ProjectID:   name.ProjectID,
		Description: body.GetDescription(),
		DisplayName: body.GetDisplayName(),
		CreateTime:  now,
		UpdateTime:  now,
	}
}

// Name returns the resource name of the project.
func (p *Project) Name() string {
	return names.Project{
		ProjectID: p.ProjectID,
	}.String()
}

// Message returns a message representing a project.
func (p *Project) Message() (message *rpc.Project, err error) {
	message = &rpc.Project{
		Name:        p.Name(),
		DisplayName: p.DisplayName,
		Description: p.Description,
	}

	message.CreateTime, err = ptypes.TimestampProto(p.CreateTime)
	if err != nil {
		return nil, err
	}

	message.UpdateTime, err = ptypes.TimestampProto(p.UpdateTime)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// Update modifies a project using the contents of a message.
func (p *Project) Update(message *rpc.Project, mask *fieldmaskpb.FieldMask) {
	if activeUpdateMask(mask) {
		for _, field := range mask.Paths {
			switch field {
			case "display_name":
				p.DisplayName = message.GetDisplayName()
			case "description":
				p.Description = message.GetDescription()
			}
		}
	} else {
		p.DisplayName = message.GetDisplayName()
		p.Description = message.GetDescription()
	}
	p.UpdateTime = time.Now()
}
