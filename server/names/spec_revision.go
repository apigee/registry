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

package names

import (
	"fmt"
	"regexp"
)

var specRevisionRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/apis/%s/versions/%s/specs/%s@%s$", identifier, identifier, identifier, identifier, revisionTag))

// SpecRevision represents a resource name for an API spec revision.
type SpecRevision struct {
	ProjectID  string
	ApiID      string
	VersionID  string
	SpecID     string
	RevisionID string
}

// Spec returns the parent spec for this resource.
func (s SpecRevision) Spec() Spec {
	return Spec{
		ProjectID: s.ProjectID,
		ApiID:     s.ApiID,
		VersionID: s.VersionID,
		SpecID:    s.SpecID,
	}
}

func (s SpecRevision) String() string {
	return normalize(fmt.Sprintf("projects/%s/apis/%s/versions/%s/specs/%s@%s", s.ProjectID, s.ApiID, s.VersionID, s.SpecID, s.RevisionID))
}

// ParseSpecRevision parses the name of a spec.
func ParseSpecRevision(name string) (SpecRevision, error) {
	if !specRevisionRegexp.MatchString(name) {
		return SpecRevision{}, fmt.Errorf("invalid spec revision name %q: must match %q", name, specRevisionRegexp)
	}

	m := specRevisionRegexp.FindStringSubmatch(normalize(name))
	revision := SpecRevision{
		ProjectID:  m[1],
		ApiID:      m[2],
		VersionID:  m[3],
		SpecID:     m[4],
		RevisionID: m[5],
	}

	return revision, nil
}
