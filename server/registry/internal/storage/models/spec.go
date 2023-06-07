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
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Spec is the storage-side representation of a spec.
type Spec struct {
	Key                string    `gorm:"primaryKey"`
	ProjectID          string    // Uniquely identifies a project.
	ApiID              string    `gorm:"index"` // Uniquely identifies an api within a project.
	VersionID          string    // Uniquely identifies a version within an api.
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
	Labels             []byte    // Serialized labels.
	Annotations        []byte    // Serialized annotations.
	ParentVersionKey   string
	ParentVersion      *Version `gorm:"foreignKey:ParentVersionKey;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// NewSpec initializes a new resource.
func NewSpec(name names.Spec, body *rpc.ApiSpec) (spec *Spec, err error) {
	now := time.Now().Round(time.Microsecond)
	spec = &Spec{
		ProjectID:          name.ProjectID,
		ApiID:              name.ApiID,
		VersionID:          name.VersionID,
		SpecID:             name.SpecID,
		Description:        body.GetDescription(),
		FileName:           body.GetFilename(),
		MimeType:           body.GetMimeType(),
		SourceURI:          body.GetSourceUri(),
		CreateTime:         now,
		RevisionCreateTime: now,
		RevisionUpdateTime: now,
		RevisionID:         newRevisionID(),
		ParentVersionKey:   name.Version().String(),
	}

	spec.Labels, err = bytesForMap(body.GetLabels())
	if err != nil {
		return nil, err
	}

	spec.Annotations, err = bytesForMap(body.GetAnnotations())
	if err != nil {
		return nil, err
	}

	if body.GetContents() != nil {
		contents := body.GetContents()
		// if contents are gzipped, uncompress before computing size and hash.
		if strings.Contains(spec.MimeType, "+gzip") && len(contents) > 0 {
			contents, err = GUnzippedBytes(contents)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
		spec.SizeInBytes = int32(len(contents))
		spec.Hash = hashForBytes(contents)
	}

	return spec, nil
}

// NewRevision returns a new revision based on the spec.
func (s *Spec) NewRevision() *Spec {
	now := time.Now().Round(time.Microsecond)
	return &Spec{
		ProjectID:          s.ProjectID,
		ApiID:              s.ApiID,
		VersionID:          s.VersionID,
		SpecID:             s.SpecID,
		Description:        s.Description,
		FileName:           s.FileName,
		MimeType:           s.MimeType,
		SizeInBytes:        s.SizeInBytes,
		Hash:               s.Hash,
		SourceURI:          s.SourceURI,
		CreateTime:         s.CreateTime,
		RevisionCreateTime: now,
		RevisionUpdateTime: now,
		RevisionID:         newRevisionID(),
		ParentVersionKey:   s.ParentVersionKey,
	}
}

// Name returns the resource name of the spec.
func (s *Spec) Name() string {
	return names.Spec{
		ProjectID: s.ProjectID,
		ApiID:     s.ApiID,
		VersionID: s.VersionID,
		SpecID:    s.SpecID,
	}.String()
}

// RevisionName generates the resource name of the spec revision.
func (s *Spec) RevisionName() string {
	return fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/specs/%s@%s",
		s.ProjectID, names.Location, s.ApiID, s.VersionID, s.SpecID, s.RevisionID)
}

// BasicMessage returns the basic view of the spec resource as an RPC message.
func (s *Spec) BasicMessage(name string) (message *rpc.ApiSpec, err error) {
	message = &rpc.ApiSpec{
		Name:               name,
		Filename:           s.FileName,
		Description:        s.Description,
		Hash:               s.Hash,
		SizeBytes:          s.SizeInBytes,
		MimeType:           s.MimeType,
		SourceUri:          s.SourceURI,
		RevisionId:         s.RevisionID,
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

// Update modifies a spec using the contents of a message.
func (s *Spec) Update(message *rpc.ApiSpec, mask *fieldmaskpb.FieldMask) error {
	s.RevisionUpdateTime = time.Now().Round(time.Microsecond)
	var hasContents bool
	for _, field := range mask.Paths {
		switch field {
		case "filename":
			s.FileName = message.GetFilename()
		case "description":
			s.Description = message.GetDescription()
		case "contents":
			hasContents = true
		case "mime_type":
			s.MimeType = message.GetMimeType()
		case "source_uri":
			s.SourceURI = message.GetSourceUri()
		case "labels":
			var err error
			if s.Labels, err = bytesForMap(message.GetLabels()); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		case "annotations":
			var err error
			if s.Annotations, err = bytesForMap(message.GetAnnotations()); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}

	// Content updates depend on the current MIME type of the spec, so they
	// should only happen after we update the MIME type field.
	if hasContents {
		return s.updateContents(message.GetContents())
	}

	return nil
}

func (s *Spec) updateContents(contents []byte) error {
	// Compute size and hash using uncompressed spec.
	if strings.Contains(s.MimeType, "+gzip") && len(contents) > 0 {
		var err error
		contents, err = GUnzippedBytes(contents)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}

	if hash := hashForBytes(contents); hash != s.Hash {
		s.Hash = hash
		s.RevisionID = newRevisionID()
		s.SizeInBytes = int32(len(contents))

		now := time.Now().Round(time.Microsecond)
		s.RevisionCreateTime = now
		s.RevisionUpdateTime = now
	}

	return nil
}

// LabelsMap returns a map representation of stored labels.
func (s *Spec) LabelsMap() (map[string]string, error) {
	return mapForBytes(s.Labels)
}

func newRevisionID() string {
	s := uuid.New().String()
	return s[len(s)-8:]
}

func hashForBytes(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	h := sha256.New()
	_, _ = h.Write(b)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// SpecRevisionTag is the storage-side representation of a spec revision tag.
type SpecRevisionTag struct {
	Key           string    `gorm:"primaryKey"`
	ProjectID     string    // Uniquely identifies a project.
	ApiID         string    // Uniquely identifies an api within a project.
	VersionID     string    // Uniquely identifies a version within an api.
	SpecID        string    // Uniquely identifies a spec within a version.
	RevisionID    string    // Uniquely identifies a revision of a spec.
	Tag           string    // The tag to use for the revision.
	CreateTime    time.Time // Creation time.
	UpdateTime    time.Time // Time of last change.
	ParentSpecKey string
	ParentSpec    *Spec `gorm:"foreignKey:ParentSpecKey;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// NewSpecRevisionTag initializes a new revision tag from a given revision name and tag string.
func NewSpecRevisionTag(name names.SpecRevision, tag string) *SpecRevisionTag {
	now := time.Now().Round(time.Microsecond)
	return &SpecRevisionTag{
		ProjectID:     name.ProjectID,
		ApiID:         name.ApiID,
		VersionID:     name.VersionID,
		SpecID:        name.SpecID,
		RevisionID:    name.RevisionID,
		Tag:           tag,
		CreateTime:    now,
		UpdateTime:    now,
		ParentSpecKey: name.String(),
	}
}

func (t *SpecRevisionTag) String() string {
	return fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/specs/%s@%s",
		t.ProjectID, names.Location, t.ApiID, t.VersionID, t.SpecID, t.Tag)
}
