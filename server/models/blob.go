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

import "time"

// BlobEntityName is used to represent blobs in storage.
const BlobEntityName = "Blob"

// Blob ...
type Blob struct {
	Key         string    `datastore:"-", gorm:"PRIMARY_KEY"`
	ProjectID   string    // Uniquely identifies a project.
	ApiID       string    // Uniquely identifies an api within a project.
	VersionID   string    // Uniquely identifies a version within a api.
	SpecID      string    // Uniquely identifies a spec within a version.
	RevisionID  string    // Uniquely identifies a revision of a spec.
	Hash        string    // Hash of the blob contents.
	SizeInBytes int32     // Size of the blob contents.
	Contents    []byte    `datastore:",noindex"` // The contents of the blob.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
}

// NewBlob creates a new Blob object.
func NewBlob(spec *Spec, contents []byte) *Blob {
	now := time.Now()
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
