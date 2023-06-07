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

var specRevisionRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/specs/%s(?:@%s)?$",
	identifier, Location, identifier, identifier, identifier, revisionTag))

var specRevisionCollectionRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/specs/%s@$",
	identifier, Location, identifier, identifier, identifier))

// SpecRevision represents a resource name for an API spec revision.
type SpecRevision struct {
	ProjectID  string
	ApiID      string
	VersionID  string
	SpecID     string
	RevisionID string
}

// Project returns the parent project for this resource.
func (s SpecRevision) Project() Project {
	return Project{
		ProjectID: s.ProjectID,
	}
}

// Api returns the parent API for this resource.
func (s SpecRevision) Api() Api {
	return Api{
		ProjectID: s.ProjectID,
		ApiID:     s.ApiID,
	}
}

// Version returns the parent API version for this resource.
func (s SpecRevision) Version() Version {
	return Version{
		ProjectID: s.ProjectID,
		ApiID:     s.ApiID,
		VersionID: s.VersionID,
	}
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

// Artifact returns an artifact with the provided ID and this resource as its parent.
func (s SpecRevision) Artifact(id string) Artifact {
	return Artifact{
		name: specArtifact{
			ProjectID:  s.ProjectID,
			ApiID:      s.ApiID,
			VersionID:  s.VersionID,
			SpecID:     s.SpecID,
			RevisionID: s.RevisionID,
			ArtifactID: id,
		},
	}
}

// Parent returns this resource's parent version resource name.
func (s SpecRevision) Parent() string {
	return s.Spec().Parent()
}

func (s SpecRevision) String() string {
	if s.RevisionID == "" { // use latest revision
		return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/specs/%s",
			s.ProjectID, Location, s.ApiID, s.VersionID, s.SpecID))
	} else {
		return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/specs/%s@%s",
			s.ProjectID, Location, s.ApiID, s.VersionID, s.SpecID, s.RevisionID))
	}
}

// ParseSpecRevision parses the name of a spec.
func ParseSpecRevision(name string) (SpecRevision, error) {
	if !specRevisionRegexp.MatchString(name) {
		return SpecRevision{}, fmt.Errorf("invalid spec revision name %q: must match %q", name, specRevisionRegexp)
	}

	m := specRevisionRegexp.FindStringSubmatch(name)
	revision := SpecRevision{
		ProjectID:  m[1],
		ApiID:      m[2],
		VersionID:  m[3],
		SpecID:     m[4],
		RevisionID: m[5],
	}

	return revision, nil
}

// ParseSpecRevisionCollection parses the name of a spec revision collection.
func ParseSpecRevisionCollection(name string) (SpecRevision, error) {
	r := specRevisionCollectionRegexp
	if !r.MatchString(name) {
		return SpecRevision{}, fmt.Errorf("invalid spec revision collection name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	rev := SpecRevision{
		ProjectID:  m[1],
		ApiID:      m[2],
		VersionID:  m[3],
		SpecID:     m[4],
		RevisionID: "-",
	}

	return rev, nil
}
