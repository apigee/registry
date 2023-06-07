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
	now := time.Now().Round(time.Microsecond)
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
func (p *Project) Message() *rpc.Project {
	return &rpc.Project{
		Name:        p.Name(),
		DisplayName: p.DisplayName,
		Description: p.Description,
		CreateTime:  timestamppb.New(p.CreateTime),
		UpdateTime:  timestamppb.New(p.UpdateTime),
	}
}

// Update modifies a project using the contents of a message.
func (p *Project) Update(message *rpc.Project, mask *fieldmaskpb.FieldMask) {
	p.UpdateTime = time.Now().Round(time.Microsecond)
	for _, field := range mask.GetPaths() {
		switch field {
		case "display_name":
			p.DisplayName = message.GetDisplayName()
		case "description":
			p.Description = message.GetDescription()
		}
	}
}
