// Copyright 2021 Google LLC. All Rights Reserved.
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

var (
	projectArtifactRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/artifacts/%s", identifier, Location, identifier))
	apiArtifactRegexp     = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/artifacts/%s", identifier, Location, identifier, identifier))
	versionArtifactRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/artifacts/%s", identifier, Location, identifier, identifier, identifier))
	specArtifactRegexp    = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/specs/%s/artifacts/%s", identifier, Location, identifier, identifier, identifier, identifier))
)

// Artifact represents a resource name for an artifact.
type Artifact struct {
	name interface {
		String() string
		Validate() error
	}
}

// ProjectID returns the artifact's project ID, or empty string if it doesn't have one.
func (a Artifact) ProjectID() string {
	switch name := a.name.(type) {
	case projectArtifact:
		return name.ProjectID
	case apiArtifact:
		return name.ProjectID
	case versionArtifact:
		return name.ProjectID
	case specArtifact:
		return name.ProjectID
	default:
		return ""
	}
}

// ApiID returns the artifact's API ID, or empty string if it doesn't have one.
func (a Artifact) ApiID() string {
	switch name := a.name.(type) {
	case apiArtifact:
		return name.ApiID
	case versionArtifact:
		return name.ApiID
	case specArtifact:
		return name.ApiID
	default:
		return ""
	}
}

// VersionID returns the artifact's version ID, or empty string if it doesn't have one.
func (a Artifact) VersionID() string {
	switch name := a.name.(type) {
	case versionArtifact:
		return name.VersionID
	case specArtifact:
		return name.VersionID
	default:
		return ""
	}
}

// SpecID returns the artifact's spec ID, or empty string if it doesn't have one.
func (a Artifact) SpecID() string {
	switch name := a.name.(type) {
	case specArtifact:
		return name.SpecID
	default:
		return ""
	}
}

// ArtifactID returns the artifact's ID.
func (a Artifact) ArtifactID() string {
	switch name := a.name.(type) {
	case projectArtifact:
		return name.ArtifactID
	case apiArtifact:
		return name.ArtifactID
	case versionArtifact:
		return name.ArtifactID
	case specArtifact:
		return name.ArtifactID
	default:
		return ""
	}
}

// Validate returns an error if the resource name is invalid.
// For backward compatibility, names should only be validated at creation time.
func (a Artifact) Validate() error {
	return a.name.Validate()
}

// Parent returns the resource name of the artifact's parent.
func (a Artifact) Parent() string {
	switch name := a.name.(type) {
	case projectArtifact:
		return name.Parent()
	case apiArtifact:
		return name.Parent()
	case versionArtifact:
		return name.Parent()
	case specArtifact:
		return name.Parent()
	default:
		return ""
	}
}

func (a Artifact) String() string {
	return normalize(a.name.String())
}

// ParseArtifact parses the name of an artifact.
func ParseArtifact(name string) (Artifact, error) {
	if n, err := parseSpecArtifact(name); err == nil {
		return Artifact{name: n}, nil
	} else if n, err := parseVersionArtifact(name); err == nil {
		return Artifact{name: n}, nil
	} else if n, err := parseApiArtifact(name); err == nil {
		return Artifact{name: n}, nil
	} else if n, err := parseProjectArtifact(name); err == nil {
		return Artifact{name: n}, nil
	}

	return Artifact{}, fmt.Errorf("invalid artifact name %q, must match one of: %v", name, []string{
		projectArtifactRegexp.String(),
		apiArtifactRegexp.String(),
		versionArtifactRegexp.String(),
		specArtifactRegexp.String(),
	})
}

type projectArtifact struct {
	ProjectID  string
	ArtifactID string
}

func (a projectArtifact) Validate() error {
	if name := a.String(); !projectArtifactRegexp.MatchString(name) {
		return fmt.Errorf("invalid project artifact name %q: must match %q", name, projectArtifactRegexp)
	}

	return validateID(a.ArtifactID)
}

func (a projectArtifact) Parent() string {
	return Project{
		ProjectID: a.ProjectID,
	}.String()
}

func (a projectArtifact) String() string {
	return normalize(fmt.Sprintf("projects/%s/locations/%s/artifacts/%s",
		a.ProjectID, Location, a.ArtifactID))
}

func parseProjectArtifact(name string) (projectArtifact, error) {
	if !projectArtifactRegexp.MatchString(name) {
		return projectArtifact{}, fmt.Errorf("invalid project artifact name %q: must match %q", name, projectArtifactRegexp)
	}

	m := projectArtifactRegexp.FindStringSubmatch(name)
	artifact := projectArtifact{
		ProjectID:  m[1],
		ArtifactID: m[2],
	}

	return artifact, nil
}

type apiArtifact struct {
	ProjectID  string
	ApiID      string
	ArtifactID string
}

func (a apiArtifact) Validate() error {
	if name := a.String(); !apiArtifactRegexp.MatchString(name) {
		return fmt.Errorf("invalid api artifact name %q: must match %q", name, apiArtifactRegexp)
	}

	return validateID(a.ArtifactID)
}

func (a apiArtifact) Parent() string {
	return Api{
		ProjectID: a.ProjectID,
		ApiID:     a.ApiID,
	}.String()
}

func (a apiArtifact) String() string {
	return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/artifacts/%s",
		a.ProjectID, Location, a.ApiID, a.ArtifactID))
}

func parseApiArtifact(name string) (apiArtifact, error) {
	if !apiArtifactRegexp.MatchString(name) {
		return apiArtifact{}, fmt.Errorf("invalid api artifact name %q: must match %q", name, apiArtifactRegexp)
	}

	m := apiArtifactRegexp.FindStringSubmatch(name)
	artifact := apiArtifact{
		ProjectID:  m[1],
		ApiID:      m[2],
		ArtifactID: m[3],
	}

	return artifact, nil
}

type versionArtifact struct {
	ProjectID  string
	ApiID      string
	VersionID  string
	ArtifactID string
}

func (a versionArtifact) Validate() error {
	if name := a.String(); !versionArtifactRegexp.MatchString(name) {
		return fmt.Errorf("invalid version artifact name %q: must match %q", name, versionArtifactRegexp)
	}

	return validateID(a.ArtifactID)
}

func (a versionArtifact) Parent() string {
	return Version{
		ProjectID: a.ProjectID,
		ApiID:     a.ApiID,
		VersionID: a.VersionID,
	}.String()
}

func (a versionArtifact) String() string {
	return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/artifacts/%s",
		a.ProjectID, Location, a.ApiID, a.VersionID, a.ArtifactID))
}

func parseVersionArtifact(name string) (versionArtifact, error) {
	if !versionArtifactRegexp.MatchString(name) {
		return versionArtifact{}, fmt.Errorf("invalid version artifact name %q: must match %q", name, versionArtifactRegexp)
	}

	m := versionArtifactRegexp.FindStringSubmatch(name)
	artifact := versionArtifact{
		ProjectID:  m[1],
		ApiID:      m[2],
		VersionID:  m[3],
		ArtifactID: m[4],
	}

	return artifact, nil
}

type specArtifact struct {
	ProjectID  string
	ApiID      string
	VersionID  string
	SpecID     string
	ArtifactID string
}

func (a specArtifact) Validate() error {
	if name := a.String(); !specArtifactRegexp.MatchString(name) {
		return fmt.Errorf("invalid spec artifact name %q: must match %q", name, specArtifactRegexp)
	}

	return validateID(a.ArtifactID)
}

func (a specArtifact) Parent() string {
	return Spec{
		ProjectID: a.ProjectID,
		ApiID:     a.ApiID,
		VersionID: a.VersionID,
		SpecID:    a.SpecID,
	}.String()
}

func (a specArtifact) String() string {
	return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/specs/%s/artifacts/%s",
		a.ProjectID, Location, a.ApiID, a.VersionID, a.SpecID, a.ArtifactID))
}

func parseSpecArtifact(name string) (specArtifact, error) {
	if !specArtifactRegexp.MatchString(name) {
		return specArtifact{}, fmt.Errorf("invalid spec artifact name %q: must match %q", name, specArtifactRegexp)
	}

	m := specArtifactRegexp.FindStringSubmatch(name)
	artifact := specArtifact{
		ProjectID:  m[1],
		ApiID:      m[2],
		VersionID:  m[3],
		SpecID:     m[4],
		ArtifactID: m[5],
	}

	return artifact, nil
}

// ArtifactsRegexp returns a regular expression that matches collection of artifacts.
func ArtifactsRegexp() *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("^projects/%s/locations/%s(/apis/%s(/versions/%s(/specs/%s)?)?)?/artifacts$",
			identifier, Location, identifier, identifier, identifier))
}

// ArtifactRegexp returns a regular expression that matches an artifact resource name.
func ArtifactRegexp() *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("^projects/%s/locations/%s(/apis/%s(/versions/%s(/specs/%s)?)?)?/artifacts/%s$",
			identifier, Location, identifier, identifier, identifier, identifier))
}
