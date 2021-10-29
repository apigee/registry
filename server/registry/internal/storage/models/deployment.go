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
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Deployment is the storage-side representation of a deployment.
type Deployment struct {
	Key                string    `gorm:"primaryKey"`
	ProjectID          string    // Uniquely identifies a project.
	ApiID              string    // Uniquely identifies an api within a project.
	DeploymentID       string    // Uniquely identifies a deployment within an api.
	RevisionID         string    // Uniquely identifies a revision of a deployment.
	Description        string    // A detailed description.
	CreateTime         time.Time // Creation time.
	RevisionCreateTime time.Time // Revision creation time.
	RevisionUpdateTime time.Time // Time of last change.
	Labels             []byte    // Serialized labels.
	Annotations        []byte    // Serialized annotations.
}

// NewDeployment initializes a new resource.
func NewDeployment(name names.Deployment, body *rpc.ApiDeployment) (deployment *Deployment, err error) {
	now := time.Now().Round(time.Microsecond)
	deployment = &Deployment{
		ProjectID:          name.ProjectID,
		ApiID:              name.ApiID,
		DeploymentID:       name.DeploymentID,
		Description:        body.GetDescription(),
		CreateTime:         now,
		RevisionCreateTime: now,
		RevisionUpdateTime: now,
		RevisionID:         newRevisionID(),
	}

	deployment.Labels, err = bytesForMap(body.GetLabels())
	if err != nil {
		return nil, err
	}

	deployment.Annotations, err = bytesForMap(body.GetAnnotations())
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

// NewRevision returns a new revision based on the deployment.
func (s *Deployment) NewRevision() *Deployment {
	now := time.Now().Round(time.Microsecond)
	return &Deployment{
		ProjectID:          s.ProjectID,
		ApiID:              s.ApiID,
		DeploymentID:       s.DeploymentID,
		Description:        s.Description,
		CreateTime:         s.CreateTime,
		RevisionCreateTime: now,
		RevisionUpdateTime: now,
		RevisionID:         newRevisionID(),
	}
}

// Name returns the resource name of the deployment.
func (s *Deployment) Name() string {
	return names.Deployment{
		ProjectID:    s.ProjectID,
		ApiID:        s.ApiID,
		DeploymentID: s.DeploymentID,
	}.String()
}

// RevisionName generates the resource name of the deployment revision.
func (s *Deployment) RevisionName() string {
	return fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s@%s",
		s.ProjectID, names.Location, s.ApiID, s.DeploymentID, s.RevisionID)
}

// BasicMessage returns the basic view of the deployment resource as an RPC message.
func (s *Deployment) BasicMessage(name string, tags []string) (message *rpc.ApiDeployment, err error) {
	message = &rpc.ApiDeployment{
		Name:               name,
		Description:        s.Description,
		RevisionId:         s.RevisionID,
		RevisionTags:       tags,
		CreateTime:         timestamppb.New(s.CreateTime),
		RevisionCreateTime: timestamppb.New(s.RevisionCreateTime),
		RevisionUpdateTime: timestamppb.New(s.RevisionUpdateTime),
	}

	message.Labels, err = mapForBytes(s.Labels)
	if err != nil {
		return nil, err
	}

	message.Annotations, err = mapForBytes(s.Annotations)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// Update modifies a deployment using the contents of a message.
func (s *Deployment) Update(message *rpc.ApiDeployment, mask *fieldmaskpb.FieldMask) error {
	s.RevisionUpdateTime = time.Now().Round(time.Microsecond)
	for _, field := range mask.Paths {
		switch field {
		case "description":
			s.Description = message.GetDescription()
		case "labels":
			var err error
			if s.Labels, err = bytesForMap(message.GetLabels()); err != nil {
				return err
			}
		case "annotations":
			var err error
			if s.Annotations, err = bytesForMap(message.GetAnnotations()); err != nil {
				return err
			}
		}
	}

	return nil
}

// LabelsMap returns a map representation of stored labels.
func (s *Deployment) LabelsMap() (map[string]string, error) {
	return mapForBytes(s.Labels)
}

// DeploymentRevisionTag is the storage-side representation of a deployment revision tag.
type DeploymentRevisionTag struct {
	Key          string    `gorm:"primaryKey"`
	ProjectID    string    // Uniquely identifies a project.
	ApiID        string    // Uniquely identifies an api within a project.
	DeploymentID string    // Uniquely identifies a deployment within an api.
	RevisionID   string    // Uniquely identifies a revision of a deployment.
	Tag          string    // The tag to use for the revision.
	CreateTime   time.Time // Creation time.
	UpdateTime   time.Time // Time of last change.
}

// NewDeploymentRevisionTag initializes a new revision tag from a given revision name and tag string.
func NewDeploymentRevisionTag(name names.DeploymentRevision, tag string) *DeploymentRevisionTag {
	now := time.Now().Round(time.Microsecond)
	return &DeploymentRevisionTag{
		ProjectID:    name.ProjectID,
		ApiID:        name.ApiID,
		DeploymentID: name.DeploymentID,
		RevisionID:   name.RevisionID,
		Tag:          tag,
		CreateTime:   now,
		UpdateTime:   now,
	}
}

func (t *DeploymentRevisionTag) String() string {
	return fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s@%s",
		t.ProjectID, names.Location, t.ApiID, t.DeploymentID, t.Tag)
}
