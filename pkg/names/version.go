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

// Version represents a resource name for an API version.
type Version struct {
	ProjectID string
	ApiID     string
	VersionID string
}

// Validate returns an error if the resource name is invalid.
// For backward compatibility, names should only be validated at creation time.
func (v Version) Validate() error {
	if err := validateID(v.VersionID); err != nil {
		return err
	}

	r := versionRegexp()
	if name := v.String(); !r.MatchString(name) {
		return fmt.Errorf("invalid version name %q: must match %q", name, r)
	}

	return nil
}

// Project returns the parent project for this resource.
func (v Version) Project() Project {
	return v.Api().Project()
}

// Api returns the parent API for this resource.
func (v Version) Api() Api {
	return Api{
		ProjectID: v.ProjectID,
		ApiID:     v.ApiID,
	}
}

// Artifact returns an artifact with the provided ID and this resource as its parent.
func (v Version) Artifact(id string) Artifact {
	return Artifact{
		name: versionArtifact{
			ProjectID:  v.ProjectID,
			ApiID:      v.ApiID,
			VersionID:  v.VersionID,
			ArtifactID: id,
		},
	}
}

// Spec returns an API spec with the provided ID and this resource as its parent.
func (v Version) Spec(id string) Spec {
	return Spec{
		ProjectID: v.ProjectID,
		ApiID:     v.ApiID,
		VersionID: v.VersionID,
		SpecID:    id,
	}
}

// Parent returns this resource's parent API resource name.
func (v Version) Parent() string {
	return v.Api().String()
}

func (v Version) String() string {
	return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s",
		v.ProjectID, Location, v.ApiID, v.VersionID))
}

// versionCollectionRegexp returns a regular expression that matches a collection of versions.
func versionCollectionRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions$",
		identifier, Location, identifier))
}

// versionRegexp returns a regular expression that matches a version resource name.
func versionRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s$",
		identifier, Location, identifier, identifier))
}

// ParseVersion parses the name of a version.
func ParseVersion(name string) (Version, error) {
	r := versionRegexp()
	if !r.MatchString(name) {
		return Version{}, fmt.Errorf("invalid version name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	return Version{
		ProjectID: m[1],
		ApiID:     m[2],
		VersionID: m[3],
	}, nil
}

// ParseVersionCollection parses the name of a version collection.
func ParseVersionCollection(name string) (Version, error) {
	r := versionCollectionRegexp()
	if !r.MatchString(name) {
		return Version{}, fmt.Errorf("invalid version collection name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	return Version{
		ProjectID: m[1],
		ApiID:     m[2],
		VersionID: "",
	}, nil
}
