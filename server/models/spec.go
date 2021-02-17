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
	"crypto/sha1"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	ptypes "github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// SpecEntityName is used to represent specs in storage.
const SpecEntityName = "Spec"

// SpecRevisionTagEntityName is used to represent tags in storage.
const SpecRevisionTagEntityName = "SpecRevisionTag"

// This was originally a boolean but gorm does not correctly update booleans from structs.
// https://stackoverflow.com/questions/56653423/gorm-doesnt-update-boolean-field-to-false
const (
	// NotCurrent indicates that a revision is NOT the current revision of a spec
	NotCurrent = 1
	// IsCurrent indicates that a revision is the current revision of a spec
	IsCurrent = 2
)

// Spec is the storage-side representation of a spec.
type Spec struct {
	Key                string    `datastore:"-" gorm:"primaryKey"`
	Currency           int32     // IsCurrent for the current revision of the spec.
	ProjectID          string    // Uniquely identifies a project.
	ApiID              string    // Uniquely identifies an api within a project.
	VersionID          string    // Uniquely identifies a version within a api.
	SpecID             string    // Uniquely identifies a spec within a version.
	RevisionID         string    // Uniquely identifies a revision of a spec.
	Description        string    // A detailed description.
	CreateTime         time.Time // Creation time.
	RevisionCreateTime time.Time // Revision creation time.
	RevisionUpdateTime time.Time // Time of last change.
	MimeType           string    // Spec format.
	SizeInBytes        int32     // Size of the spec.
	Hash               string    // A hash of the spec.
	FileName           string    // Name of spec file.
	SourceURI          string    // The original source URI of the spec.
}

// NewSpecFromParentAndSpecID returns an initialized spec for a specified parent and specID.
func NewSpecFromParentAndSpecID(parent string, specID string) (*Spec, error) {
	r := regexp.MustCompile("^projects/" + names.NameRegex +
		"/apis/" + names.NameRegex +
		"/versions/" + names.NameRegex + "$")
	m := r.FindStringSubmatch(parent)
	if m == nil {
		return nil, fmt.Errorf("invalid parent '%s'", parent)
	}
	if err := names.ValidateID(specID); err != nil {
		return nil, err
	}
	spec := &Spec{}
	spec.ProjectID = m[1]
	spec.ApiID = m[2]
	spec.VersionID = m[3]
	spec.SpecID = specID
	return spec, nil
}

// NewSpecFromResourceName parses resource names and returns an initialized spec.
func NewSpecFromResourceName(name string) (*Spec, error) {
	spec := &Spec{}
	m := names.SpecRegexp().FindStringSubmatch(name)
	if m == nil {
		return nil, errors.New("invalid spec name")
	}
	spec.ProjectID = m[1]
	spec.ApiID = m[2]
	spec.VersionID = m[3]
	spec.SpecID = m[4]
	if strings.HasPrefix(m[5], "@") {
		spec.RevisionID = m[5][1:]
	}
	return spec, nil
}

// NewSpecFromMessage returns an initialized spec from a message.
func NewSpecFromMessage(message *rpc.ApiSpec) (*Spec, error) {
	spec, err := NewSpecFromResourceName(message.GetName())
	if err != nil {
		return nil, err
	}
	spec.Description = message.GetDescription()
	spec.FileName = message.GetFilename()
	return spec, nil
}

// ResourceName generates the resource name of a spec.
func (spec *Spec) ResourceName() string {
	return fmt.Sprintf("projects/%s/apis/%s/versions/%s/specs/%s",
		spec.ProjectID, spec.ApiID, spec.VersionID, spec.SpecID)
}

// ResourceNameWithRevision generates the resource name of a spec which includes the revision id.
func (spec *Spec) ResourceNameWithRevision() string {
	return fmt.Sprintf("projects/%s/apis/%s/versions/%s/specs/%s@%s",
		spec.ProjectID, spec.ApiID, spec.VersionID, spec.SpecID, spec.RevisionID)
}

// ResourceNameWithSpecifiedRevision generates the resource name of a spec which includes the revision id.
func (spec *Spec) ResourceNameWithSpecifiedRevision(revision string) string {
	return fmt.Sprintf("projects/%s/apis/%s/versions/%s/specs/%s@%s",
		spec.ProjectID, spec.ApiID, spec.VersionID, spec.SpecID, revision)
}

// ParentResourceName generates the resource name of a spec's parent.
func (spec *Spec) ParentResourceName() string {
	return fmt.Sprintf("projects/%s/apis/%s/versions/%s", spec.ProjectID, spec.ApiID, spec.VersionID)
}

// Message returns a message representing a spec.
func (spec *Spec) Message(blob *Blob, revision string) (message *rpc.ApiSpec, err error) {
	message = &rpc.ApiSpec{}
	if revision != "" {
		message.Name = spec.ResourceNameWithSpecifiedRevision(revision)
	} else {
		message.Name = spec.ResourceName()
	}
	message.Filename = spec.FileName
	message.Description = spec.Description
	if blob != nil {
		message.Contents = blob.Contents
	}
	message.Hash = spec.Hash
	message.SizeBytes = spec.SizeInBytes
	message.MimeType = spec.MimeType
	message.SourceUri = spec.SourceURI
	message.CreateTime, err = ptypes.TimestampProto(spec.CreateTime)
	message.RevisionCreateTime, err = ptypes.TimestampProto(spec.RevisionCreateTime)
	message.RevisionUpdateTime, err = ptypes.TimestampProto(spec.RevisionUpdateTime)
	message.RevisionId = spec.RevisionID
	return message, err
}

// Update modifies a spec using the contents of a message.
func (spec *Spec) Update(message *rpc.ApiSpec, mask *fieldmaskpb.FieldMask) error {
	now := time.Now()
	if activeUpdateMask(mask) {
		for _, field := range mask.Paths {
			switch field {
			case "filename":
				spec.FileName = message.GetFilename()
			case "description":
				spec.Description = message.GetDescription()
			case "contents":
				contents := message.GetContents()
				// Save some properties of the spec contents.
				// The bytes of the contents are stored in a Blob.
				hash := hashForBytes(contents)
				if spec.Hash != hash {
					spec.Hash = hash
					spec.RevisionID = newRevisionID()
					spec.CreateTime = now
				}
				spec.SizeInBytes = int32(len(contents))
			case "mime_type":
				spec.MimeType = message.GetMimeType()
			case "source_uri":
				spec.SourceURI = message.GetSourceUri()
			}
		}
	} else {
		filename := message.GetFilename()
		if filename != "" {
			spec.FileName = filename
		}
		description := message.GetDescription()
		if description != "" {
			spec.Description = description
		}
		contents := message.GetContents()
		if contents != nil {
			// Save some properties of the spec contents.
			// The bytes of the contents are stored in a Blob.
			hash := hashForBytes(contents)
			if spec.Hash != hash {
				spec.Hash = hash
				spec.RevisionID = newRevisionID()
				spec.CreateTime = now
			}
			spec.SizeInBytes = int32(len(contents))
		}
		mimeType := message.GetMimeType()
		if mimeType != "" {
			spec.MimeType = mimeType
		}
		sourceURI := message.GetSourceUri()
		if sourceURI != "" {
			spec.SourceURI = sourceURI
		}
	}
	spec.RevisionUpdateTime = now
	return nil
}

// BumpRevision updates the revision id for a spec and makes no other changes.
func (spec *Spec) BumpRevision() error {
	spec.RevisionID = newRevisionID()
	spec.RevisionCreateTime = time.Now()
	spec.RevisionUpdateTime = time.Now()
	return nil
}

func newRevisionID() string {
	s := uuid.New().String()
	return s[len(s)-8:]
}

func hashForBytes(b []byte) string {
	h := sha1.New()
	h.Write(b)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// SpecRevisionTag is the storage-side representation of a spec revision tag.
type SpecRevisionTag struct {
	Key        string    `datastore:"-" gorm:"primaryKey"`
	ProjectID  string    // Uniquely identifies a project.
	ApiID      string    // Uniquely identifies an api within a project.
	VersionID  string    // Uniquely identifies a version within a api.
	SpecID     string    // Uniquely identifies a spec within a version.
	RevisionID string    // Uniquely identifies a revision of a spec.
	Tag        string    // The tag to use for the revision.
	CreateTime time.Time // Creation time.
	UpdateTime time.Time // Time of last change.
}

// ResourceNameWithTag generates a resource name which includes the tag.
func (tag *SpecRevisionTag) ResourceNameWithTag() string {
	return fmt.Sprintf("projects/%s/apis/%s/versions/%s/specs/%s@%s",
		tag.ProjectID, tag.ApiID, tag.VersionID, tag.SpecID, tag.Tag)
}

// ResourceNameWithRevision generates a resource name which includes the revision id.
func (tag *SpecRevisionTag) ResourceNameWithRevision() string {
	return fmt.Sprintf("projects/%s/apis/%s/versions/%s/specs/%s@%s",
		tag.ProjectID, tag.ApiID, tag.VersionID, tag.SpecID, tag.RevisionID)
}
