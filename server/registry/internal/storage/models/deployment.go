// Copyright 2021 Google LLC.
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

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Deployment is the storage-side representation of a deployment.
type Deployment struct {
	Key                string    `gorm:"primaryKey"`
	ProjectID          string    // Uniquely identifies a project.
	ApiID              string    `gorm:"index:idx_latest"` // Uniquely identifies an api within a project.
	DeploymentID       string    `gorm:"index:idx_latest"` // Uniquely identifies a deployment within an api.
	RevisionID         string    // Uniquely identifies a revision of a deployment.
	DisplayName        string    // A human-friendly name.
	Description        string    // A detailed description.
	CreateTime         time.Time // Creation time.
	RevisionCreateTime time.Time `gorm:"index:idx_latest,sort:desc"` // Revision creation time.
	RevisionUpdateTime time.Time // Time of last change.
	ApiSpecRevision    string    // The spec being served by the deployment.
	EndpointURI        string    // The address where the deployment is serving.
	ExternalChannelURI string    // The address of the external channel of the API.
	IntendedAudience   string    // The intended audience of the API.
	AccessGuidance     string    // A brief description of how to access the endpoint.
	Labels             []byte    // Serialized labels.
	Annotations        []byte    // Serialized annotations.
	ParentApiKey       string
	ParentApi          *Api `gorm:"foreignKey:ParentApiKey;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// NewDeployment initializes a new resource.
func NewDeployment(name names.Deployment, body *rpc.ApiDeployment) (deployment *Deployment, err error) {
	now := time.Now().Round(time.Microsecond)
	deployment = &Deployment{
		ProjectID:          name.ProjectID,
		ApiID:              name.ApiID,
		DeploymentID:       name.DeploymentID,
		RevisionID:         newRevisionID(),
		DisplayName:        body.GetDisplayName(),
		Description:        body.GetDescription(),
		CreateTime:         now,
		RevisionCreateTime: now,
		RevisionUpdateTime: now,
		ApiSpecRevision:    body.GetApiSpecRevision(),
		EndpointURI:        body.GetEndpointUri(),
		ExternalChannelURI: body.GetExternalChannelUri(),
		IntendedAudience:   body.GetIntendedAudience(),
		AccessGuidance:     body.GetAccessGuidance(),
		ParentApiKey:       name.Api().String(),
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
		RevisionID:         newRevisionID(),
		DisplayName:        s.DisplayName,
		Description:        s.Description,
		CreateTime:         s.CreateTime,
		RevisionCreateTime: now,
		RevisionUpdateTime: now,
		ApiSpecRevision:    s.ApiSpecRevision,
		EndpointURI:        s.EndpointURI,
		ExternalChannelURI: s.ExternalChannelURI,
		IntendedAudience:   s.IntendedAudience,
		AccessGuidance:     s.AccessGuidance,
		ParentApiKey:       s.ParentApiKey,
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
func (s *Deployment) BasicMessage(name string) (message *rpc.ApiDeployment, err error) {
	message = &rpc.ApiDeployment{
		Name:               name,
		DisplayName:        s.DisplayName,
		Description:        s.Description,
		RevisionId:         s.RevisionID,
		CreateTime:         timestamppb.New(s.CreateTime),
		RevisionCreateTime: timestamppb.New(s.RevisionCreateTime),
		RevisionUpdateTime: timestamppb.New(s.RevisionUpdateTime),
		ApiSpecRevision:    s.ApiSpecRevision,
		EndpointUri:        s.EndpointURI,
		ExternalChannelUri: s.ExternalChannelURI,
		IntendedAudience:   s.IntendedAudience,
		AccessGuidance:     s.AccessGuidance,
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
	needsNewRevision := false // we update the revision if certain fields change
	s.RevisionUpdateTime = time.Now().Round(time.Microsecond)
	for _, field := range mask.Paths {
		switch field {
		case "display_name":
			s.DisplayName = message.GetDisplayName()
		case "description":
			s.Description = message.GetDescription()
		case "api_spec_revision":
			if s.ApiSpecRevision != message.GetApiSpecRevision() {
				needsNewRevision = true
			}
			s.ApiSpecRevision = message.GetApiSpecRevision()
		case "endpoint_uri":
			if s.EndpointURI != message.GetEndpointUri() {
				needsNewRevision = true
			}
			s.EndpointURI = message.GetEndpointUri()
		case "external_channel_uri":
			s.ExternalChannelURI = message.GetExternalChannelUri()
		case "intended_audience":
			s.IntendedAudience = message.GetIntendedAudience()
		case "access_guidance":
			s.AccessGuidance = message.GetAccessGuidance()
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
		if needsNewRevision {
			s.RevisionID = newRevisionID()
			now := time.Now().Round(time.Microsecond)
			s.RevisionCreateTime = now
			s.RevisionUpdateTime = now
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
	Key                 string    `gorm:"primaryKey"`
	ProjectID           string    // Uniquely identifies a project.
	ApiID               string    // Uniquely identifies an api within a project.
	DeploymentID        string    // Uniquely identifies a deployment within an api.
	RevisionID          string    // Uniquely identifies a revision of a deployment.
	Tag                 string    // The tag to use for the revision.
	CreateTime          time.Time // Creation time.
	UpdateTime          time.Time // Time of last change.
	ParentDeploymentKey string
	ParentDeployment    *Deployment `gorm:"foreignKey:ParentDeploymentKey;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// NewDeploymentRevisionTag initializes a new revision tag from a given revision name and tag string.
func NewDeploymentRevisionTag(name names.DeploymentRevision, tag string) *DeploymentRevisionTag {
	now := time.Now().Round(time.Microsecond)
	return &DeploymentRevisionTag{
		ProjectID:           name.ProjectID,
		ApiID:               name.ApiID,
		DeploymentID:        name.DeploymentID,
		RevisionID:          name.RevisionID,
		Tag:                 tag,
		CreateTime:          now,
		UpdateTime:          now,
		ParentDeploymentKey: name.String(),
	}
}

func (t *DeploymentRevisionTag) String() string {
	return fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s@%s",
		t.ProjectID, names.Location, t.ApiID, t.DeploymentID, t.Tag)
}
