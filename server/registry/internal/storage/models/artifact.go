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
	"fmt"
	"strings"
	"time"

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Artifact is the storage-side representation of an artifact.
type Artifact struct {
	Key          string    `gorm:"primaryKey"`
	ProjectID    string    // Project associated with artifact (required).
	ApiID        string    `gorm:"index"` // Api associated with artifact (if appropriate).
	VersionID    string    // Version associated with artifact (if appropriate).
	SpecID       string    // Spec associated with artifact (if appropriate).
	RevisionID   string    // Revision associated with parent (if appropriate).
	DeploymentID string    // Deployment associated with artifact (if appropriate).
	ArtifactID   string    // Artifact identifier (required).
	CreateTime   time.Time // Creation time.
	UpdateTime   time.Time // Time of last change.
	MimeType     string    // MIME type of artifact
	SizeInBytes  int32     // Size of the spec.
	Hash         string    // A hash of the spec.
	Labels       []byte    // Serialized labels.
	Annotations  []byte    // Serialized annotations.
}

// NewArtifact initializes a new resource.
func NewArtifact(name names.Artifact, body *rpc.Artifact) (artifact *Artifact, err error) {
	now := time.Now().Round(time.Microsecond)
	artifact = &Artifact{
		ProjectID:    name.ProjectID(),
		ApiID:        name.ApiID(),
		VersionID:    name.VersionID(),
		SpecID:       name.SpecID(),
		RevisionID:   name.RevisionID(),
		DeploymentID: name.DeploymentID(),
		ArtifactID:   name.ArtifactID(),
		CreateTime:   now,
		UpdateTime:   now,
		MimeType:     body.GetMimeType(),
	}

	if body.GetContents() != nil {
		contents := body.GetContents()
		// if contents are gzipped, uncompress before computing size and hash.
		if strings.Contains(artifact.MimeType, "+gzip") && len(contents) > 0 {
			contents, err = GUnzippedBytes(contents)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
		artifact.SizeInBytes = int32(len(contents))
		artifact.Hash = hashForBytes(contents)
	}

	artifact.Labels, err = bytesForMap(body.GetLabels())
	if err != nil {
		return nil, err
	}

	artifact.Annotations, err = bytesForMap(body.GetAnnotations())
	if err != nil {
		return nil, err
	}

	return artifact, nil
}

// Name returns the resource name of the artifact.
func (artifact *Artifact) Name() string {
	switch {
	case artifact.SpecID != "":
		return fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/specs/%s@%s/artifacts/%s",
			artifact.ProjectID, names.Location, artifact.ApiID, artifact.VersionID, artifact.SpecID, artifact.RevisionID, artifact.ArtifactID)
	case artifact.VersionID != "":
		return fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/artifacts/%s",
			artifact.ProjectID, names.Location, artifact.ApiID, artifact.VersionID, artifact.ArtifactID)
	case artifact.DeploymentID != "":
		return fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s@%s/artifacts/%s",
			artifact.ProjectID, names.Location, artifact.ApiID, artifact.DeploymentID, artifact.RevisionID, artifact.ArtifactID)
	case artifact.ApiID != "":
		return fmt.Sprintf("projects/%s/locations/%s/apis/%s/artifacts/%s",
			artifact.ProjectID, names.Location, artifact.ApiID, artifact.ArtifactID)
	case artifact.ProjectID != "":
		return fmt.Sprintf("projects/%s/locations/%s/artifacts/%s",
			artifact.ProjectID, names.Location, artifact.ArtifactID)
	default:
		return "UNKNOWN"
	}
}

// Message returns an RPC message representing the artifact.
func (artifact *Artifact) Message() (message *rpc.Artifact, err error) {
	message = &rpc.Artifact{
		Name:       artifact.Name(),
		MimeType:   artifact.MimeType,
		SizeBytes:  artifact.SizeInBytes,
		Hash:       artifact.Hash,
		CreateTime: timestamppb.New(artifact.CreateTime),
		UpdateTime: timestamppb.New(artifact.UpdateTime),
	}

	message.Labels, err = artifact.LabelsMap()
	if err != nil {
		return nil, err
	}

	message.Annotations, err = mapForBytes(artifact.Annotations)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// LabelsMap returns a map representation of stored labels.
func (artifact *Artifact) LabelsMap() (map[string]string, error) {
	return mapForBytes(artifact.Labels)
}
