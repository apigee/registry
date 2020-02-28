// Copyright 2020 Google Inc. All Rights Reserved.

package models

import (
	"time"
)

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
	Contents    []byte    // The contents of the file.
}
