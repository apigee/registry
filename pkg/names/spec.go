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
)

// simpleSpecRegexp is the regex pattern for spec resource names.
// Notably, this differs from SpecRegexp() by not accepting spec revision IDs in the resource name.
var simpleSpecRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/specs/%s$",
	identifier, Location, identifier, identifier, identifier))

// Spec represents a resource name for an API spec.
type Spec struct {
	ProjectID string
	ApiID     string
	VersionID string
	SpecID    string
}

// Validate returns an error if the resource name is invalid.
// For backward compatibility, names should only be validated at creation time.
func (s Spec) Validate() error {
	if err := validateID(s.SpecID); err != nil {
		return err
	}

	r := specRegexp()
	if name := s.String(); !r.MatchString(name) {
		return fmt.Errorf("invalid spec name %q: must match %q", name, r)
	}

	return nil
}

// Project returns the parent project for this resource.
func (s Spec) Project() Project {
	return Project{
		ProjectID: s.ProjectID,
	}
}

// Api returns the parent API for this resource.
func (s Spec) Api() Api {
	return Api{
		ProjectID: s.ProjectID,
		ApiID:     s.ApiID,
	}
}

// Version returns the parent API version for this resource.
func (s Spec) Version() Version {
	return Version{
		ProjectID: s.ProjectID,
		ApiID:     s.ApiID,
		VersionID: s.VersionID,
	}
}

// Revision returns an API spec revision with the provided ID and this resource as its parent.
func (s Spec) Revision(id string) SpecRevision {
	return SpecRevision{
		ProjectID:  s.ProjectID,
		ApiID:      s.ApiID,
		VersionID:  s.VersionID,
		SpecID:     s.SpecID,
		RevisionID: id,
	}
}

// Artifact returns an artifact with the provided ID and this resource as its parent.
func (s Spec) Artifact(id string) Artifact {
	return Artifact{
		name: specArtifact{
			ProjectID:  s.ProjectID,
			ApiID:      s.ApiID,
			VersionID:  s.VersionID,
			SpecID:     s.SpecID,
			ArtifactID: id,
		},
	}
}

// Normal returns the resource name with normalized identifiers.
func (s Spec) Normal() Spec {
	return Spec{
		ProjectID: normalize(s.ProjectID),
		ApiID:     normalize(s.ApiID),
		VersionID: normalize(s.VersionID),
		SpecID:    normalize(s.SpecID),
	}
}

// Parent returns this resource's parent version resource name.
func (s Spec) Parent() string {
	return s.Version().String()
}

func (s Spec) String() string {
	return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/specs/%s",
		s.ProjectID, Location, s.ApiID, s.VersionID, s.SpecID))
}

// specCollectionRegexp returns a regular expression that matches a collection of specs.
func specCollectionRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/specs$",
		identifier, Location, identifier, identifier))
}

// specRegexp returns a regular expression that matches a spec resource name with an optional revision identifier.
func specRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/specs/%s(@%s)?$",
		identifier, Location, identifier, identifier, identifier, revisionTag))
}

// ParseSpec parses the name of a spec.
func ParseSpec(name string) (Spec, error) {
	if !simpleSpecRegexp.MatchString(name) {
		return Spec{}, fmt.Errorf("invalid spec name %q: must match %q", name, simpleSpecRegexp)
	}

	m := simpleSpecRegexp.FindStringSubmatch(name)
	spec := Spec{
		ProjectID: m[1],
		ApiID:     m[2],
		VersionID: m[3],
		SpecID:    m[4],
	}

	return spec, nil
}

// ParseSpecCollection parses the name of a spec collection.
func ParseSpecCollection(name string) (Spec, error) {
	r := specCollectionRegexp()
	if !r.MatchString(name) {
		return Spec{}, fmt.Errorf("invalid spec collection name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	spec := Spec{
		ProjectID: m[1],
		ApiID:     m[2],
		VersionID: m[3],
		SpecID:    "",
	}

	return spec, nil
}
