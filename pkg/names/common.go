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

package names

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const (
	// The format of a resource identifier.
	// This may be extended to include all characters that do not require escaping.
	// See https://aip.dev/122#resource-id-segments.
	identifier = `([a-z0-9-.]+)`

	// The format of a custom revision tag.
	revisionTag = `([a-z0-9-]+)`
)

// The format of a custom resource identifier.
// User provided identifiers should be validated according to this format.
var customIdentifier = regexp.MustCompile(`^[a-z0-9-.]+$`)

// validateID returns an error if the provided ID is invalid.
func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("invalid identifier %q: identifier must be nonempty", id)
	} else if !customIdentifier.MatchString(id) {
		return fmt.Errorf("invalid identifier %q: must match %q", id, customIdentifier)
	} else if _, err := uuid.Parse(id); err == nil {
		return fmt.Errorf("invalid identifier %q: must not match UUID format", id)
	} else if len(id) > 80 {
		return fmt.Errorf("invalid identifier %q: must be 80 characters or less", id)
	} else if strings.HasPrefix(id, "-") || strings.HasPrefix(id, ".") {
		return fmt.Errorf("invalid identifier %q: must begin with a number or letter", id)
	} else if strings.HasSuffix(id, "-") || strings.HasSuffix(id, ".") {
		return fmt.Errorf("invalid identifier %q: must end with a number or letter", id)
	}

	return nil
}

// ValidateRevisionTag returns an error if the provided revision tag is invalid.
func ValidateRevisionTag(tag string) error {
	r := regexp.MustCompile("^" + revisionTag + "$")
	if !r.MatchString(tag) {
		return fmt.Errorf("invalid revision tag %q: must contain only lowercase letters, digits, and dashes", tag)
	} else if tag == "-" {
		return fmt.Errorf("invalid revision tag %q: must not be a single dash", tag)
	}

	return nil
}

// Normalize is an idempotent operation for normalizing resource names and identifiers.
// Identifiers `a` and `b` should be considered equal if and only if normalize(a) == normalize(b).
func normalize(identifier string) string {
	return strings.ToLower(identifier)
}

// Location is included in resource names immediately following the project_id.
const Location = "global"

// Name is an interface that represents resource names.
type Name interface {
	String() string   // all names have a string representation.
	Project() Project // all names are associated with a project
}

// ExportableName returns a name suitable for export.
// Its project and location are removed, making it project-local,
// and any revision ID is removed since revision ids cannot be imported.
//
// Note that currently revision ids are only removed from trailing segments.
// That's all that we need now, but support for removing internal revision ids
// could be added if needed.
//
// Also note: this would be a nice function to add to the Name interface
// but that might require improving the regularity of access to ProjectIDs
// for different types of Names.
func ExportableName(name string, projectID string) string {
	// first remove the located project
	name = strings.TrimPrefix(name, "projects/"+projectID+"/locations/global")
	// if there's anything left, trim the leading slash
	name = strings.TrimPrefix(name, "/")
	// if there's a revision id, remove it (we only export the current revisions)
	parts := strings.Split(name, "@")
	if len(parts) > 1 {
		name = parts[0]
	}
	return name
}
