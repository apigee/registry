// Copyright 2020 Google Inc. All Rights Reserved.

package models

import "time"

// Version ...
type Version struct {
	ProjectID   string    // Uniquely identifies a project.
	ProductID   string    // Uniquely identifies a product within a project.
	VersionID   string    // Uniquely identifies a version wihtin a product.
	DisplayName string    // A human-friendly name.
	Description string    // A detailed description.
	CreateTime  time.Time // Creation time.
	UpdateTime  time.Time // Time of last change.
	State       string    // Lifecycle stage.
}
