// Copyright 2020 Google Inc. All Rights Reserved.

package models

// Spec ...
type Spec struct {
	ProjectID   string    // Uniquely identifies a project.
	ProductID   string    // Uniquely identifies a product within a project.
	VersionID   string    // Uniquely identifies a version within a product.
	SpecID      string    // Uniquely identifies a spec within a version.
	DisplayName string    // A human-friendly name.
	Description string    // A detailed description.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
	Style       string    // Specification format.
}
