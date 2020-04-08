// Copyright 2020 Google Inc. All Rights Reserved.

package models

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"regexp"
	"time"

	rpc "apigov.dev/flame/rpc"
	ptypes "github.com/golang/protobuf/ptypes"
)

// FileEntityName is used to represent files in the datastore.
const FileEntityName = "File"

// File ...
type File struct {
	ProjectID   string    // Uniquely identifies a project.
	ProductID   string    // Uniquely identifies a product within a project.
	VersionID   string    // Uniquely identifies a version within a product.
	SpecID      string    // Uniquely identifies a spec within a version.
	FileID      string    // Uniquely identifies a file within a spec.
	DisplayName string    // A human-friendly name.
	Description string    // A detailed description.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
	FileName    string    // Name of file.
	SizeInBytes int32     // Size of the file.
	Hash        string    // A hash of the file.
	SourceURI   string    // The original source URI of the file.
	Contents    []byte    `datastore:",noindex"` // The contents of the file.
}

// ParseParentSpec ...
func ParseParentSpec(parent string) ([]string, error) {
	r := regexp.MustCompile("^projects/" + nameRegex +
		"/products/" + nameRegex +
		"/versions/" + nameRegex +
		"/specs/" + nameRegex +
		"$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid spec '%s'", parent)
	}
	return m[0], nil
}

// NewFileFromParentAndFileID returns an initialized file for a specified parent and fileID.
func NewFileFromParentAndFileID(parent string, fileID string) (*File, error) {
	r := regexp.MustCompile("^projects/" + nameRegex +
		"/products/" + nameRegex +
		"/versions/" + nameRegex +
		"/specs/" + nameRegex + "$")
	m := r.FindAllStringSubmatch(parent, -1)
	if m == nil {
		return nil, fmt.Errorf("invalid parent '%s'", parent)
	}
	if err := validateFileID(fileID); err != nil {
		return nil, err
	}
	file := &File{}
	file.ProjectID = m[0][1]
	file.ProductID = m[0][2]
	file.VersionID = m[0][3]
	file.SpecID = m[0][4]
	file.FileID = fileID
	return file, nil
}

// NewFileFromResourceName parses resource names and returns an initialized file.
func NewFileFromResourceName(name string) (*File, error) {
	file := &File{}
	m := FileRegexp().FindAllStringSubmatch(name, -1)
	if m == nil {
		return nil, errors.New("invalid file name")
	}
	file.ProjectID = m[0][1]
	file.ProductID = m[0][2]
	file.VersionID = m[0][3]
	file.SpecID = m[0][4]
	file.FileID = m[0][5]
	return file, nil
}

// NewFileFromMessage returns an initialized file from a message.
func NewFileFromMessage(message *rpc.File) (*File, error) {
	file, err := NewFileFromResourceName(message.GetName())
	if err != nil {
		return nil, err
	}
	file.DisplayName = message.GetDisplayName()
	file.Description = message.GetDescription()
	return file, nil
}

// ResourceName generates the resource name of a file.
func (file *File) ResourceName() string {
	return fmt.Sprintf("projects/%s/products/%s/versions/%s/specs/%s/files/%s",
		file.ProjectID, file.ProductID, file.VersionID, file.SpecID, file.FileID)
}

// Message returns a message representing a file.
func (file *File) Message(view rpc.FileView) (message *rpc.File, err error) {
	message = &rpc.File{}
	message.Name = file.ResourceName()
	message.DisplayName = file.DisplayName
	message.Description = file.Description
	message.SizeBytes = file.SizeInBytes
	message.Hash = file.Hash
	message.CreateTime, err = ptypes.TimestampProto(file.CreateTime)
	message.UpdateTime, err = ptypes.TimestampProto(file.UpdateTime)
	if view == rpc.FileView_FULL {
		message.Contents = file.Contents
	}
	return message, err
}

// Update modifies a file using the contents of a message.
func (file *File) Update(message *rpc.File) error {
	contents := message.GetContents()
	file.Contents = contents
	file.SizeInBytes = int32(len(contents))
	h := sha1.New()
	h.Write(contents)
	bs := h.Sum(nil)
	file.Hash = fmt.Sprintf("%x", bs)
	file.UpdateTime = file.CreateTime
	return nil
}
