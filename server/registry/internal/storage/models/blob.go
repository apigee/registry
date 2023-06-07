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

import "time"

// Blob is the storage-side representation of a blob.
type Blob struct {
	Key          string    `gorm:"primaryKey"`
	ProjectID    string    // Uniquely identifies a project.
	ApiID        string    // Uniquely identifies an API within a project.
	VersionID    string    // Uniquely identifies a version of an API.
	SpecID       string    // Uniquely identifies a spec of a version.
	RevisionID   string    // Uniquely identifies a revision of a spec.
	DeploymentID string    // Uniquely identifies a deployment of an API.
	ArtifactID   string    // Uniquely identifies an artifact on a resource.
	Hash         string    // Hash of the blob contents.
	SizeInBytes  int32     // Size of the blob contents.
	Contents     []byte    // The contents of the blob.
	CreateTime   time.Time // Creation time.
	UpdateTime   time.Time // Time of last change.
}

// NewBlobForSpec creates a new Blob object to store spec contents.
func NewBlobForSpec(spec *Spec, contents []byte) *Blob {
	now := time.Now().Round(time.Microsecond)
	return &Blob{
		ProjectID:   spec.ProjectID,
		ApiID:       spec.ApiID,
		VersionID:   spec.VersionID,
		SpecID:      spec.SpecID,
		RevisionID:  spec.RevisionID,
		Hash:        spec.Hash,
		SizeInBytes: spec.SizeInBytes,
		Contents:    contents,
		CreateTime:  now,
		UpdateTime:  now,
	}
}

// NewBlobForArtifact creates a new Blob object to store artifact contents.
func NewBlobForArtifact(artifact *Artifact, contents []byte) *Blob {
	now := time.Now().Round(time.Microsecond)
	return &Blob{
		ProjectID:    artifact.ProjectID,
		ApiID:        artifact.ApiID,
		VersionID:    artifact.VersionID,
		SpecID:       artifact.SpecID,
		RevisionID:   artifact.RevisionID,
		DeploymentID: artifact.DeploymentID,
		ArtifactID:   artifact.ArtifactID,
		Hash:         artifact.Hash,
		SizeInBytes:  artifact.SizeInBytes,
		Contents:     contents,
		CreateTime:   now,
		UpdateTime:   now,
	}
}
