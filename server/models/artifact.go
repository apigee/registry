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
)

// Artifact is the storage-side representation of an artifact.
type Artifact struct {
	Key         string    `gorm:"primaryKey"`
	ProjectID   string    // Project associated with artifact (required).
	ApiID       string    // Api associated with artifact (if appropriate).
	VersionID   string    // Version associated with artifact (if appropriate).
	SpecID      string    // Spec associated with artifact (if appropriate).
	ArtifactID  string    // Artifact identifier (required).
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
	MimeType    string    // MIME type of artifact
	SizeInBytes int32     // Size of the spec.
	Hash        string    // A hash of the spec.
}

// NewArtifact initializes a new resource.
func NewArtifact(name names.Artifact, body *rpc.Artifact) *Artifact {
	now := time.Now()
	artifact := &Artifact{
		ProjectID:  name.ProjectID(),
		ApiID:      name.ApiID(),
		VersionID:  name.VersionID(),
		SpecID:     name.SpecID(),
		ArtifactID: name.ArtifactID(),
		CreateTime: now,
		UpdateTime: now,
		MimeType:   body.GetMimeType(),
	}

	if body.GetContents() != nil {
		artifact.SizeInBytes = int32(len(body.GetContents()))
		artifact.Hash = hashForBytes(body.GetContents())
	}

	return artifact
}

// Name returns the resource name of the artifact.
func (artifact *Artifact) Name() string {
	switch {
	case artifact.SpecID != "":
		return fmt.Sprintf("projects/%s%s/apis/%s/versions/%s/specs/%s/artifacts/%s",
			artifact.ProjectID, names.LocationSegment, artifact.ApiID, artifact.VersionID, artifact.SpecID, artifact.ArtifactID)
	case artifact.VersionID != "":
		return fmt.Sprintf("projects/%s%s/apis/%s/versions/%s/artifacts/%s",
			artifact.ProjectID, names.LocationSegment, artifact.ApiID, artifact.VersionID, artifact.ArtifactID)
	case artifact.ApiID != "":
		return fmt.Sprintf("projects/%s%s/apis/%s/artifacts/%s",
			artifact.ProjectID, names.LocationSegment, artifact.ApiID, artifact.ArtifactID)
	case artifact.ProjectID != "":
		return fmt.Sprintf("projects/%s%s/artifacts/%s",
			artifact.ProjectID, names.LocationSegment, artifact.ArtifactID)
	default:
		return "UNKNOWN"
	}
}

// BasicMessage returns the basic view of the artifact resource as an RPC message.
func (artifact *Artifact) BasicMessage() (message *rpc.Artifact, err error) {
	message = &rpc.Artifact{
		Name:      artifact.Name(),
		MimeType:  artifact.MimeType,
		SizeBytes: artifact.SizeInBytes,
		Hash:      artifact.Hash,
	}

	message.CreateTime, err = ptypes.TimestampProto(artifact.CreateTime)
	if err != nil {
		return nil, err
	}

	message.UpdateTime, err = ptypes.TimestampProto(artifact.UpdateTime)
	if err != nil {
		return nil, err
	}

	return message, nil
}
