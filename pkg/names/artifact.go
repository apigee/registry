// Copyright 2021 Google LLC.
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
	projectArtifactCollectionRegexp    = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/artifacts$", identifier, Location))
	apiArtifactCollectionRegexp        = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/artifacts$", identifier, Location, identifier))
	versionArtifactCollectionRegexp    = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/artifacts$", identifier, Location, identifier, identifier))
	specArtifactCollectionRegexp       = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/specs/%s(?:@%s)?/artifacts$", identifier, Location, identifier, identifier, identifier, revisionTag))
	deploymentArtifactCollectionRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/deployments/%s(?:@%s)?/artifacts$", identifier, Location, identifier, identifier, revisionTag))

	projectArtifactRegexp    = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/artifacts/%s$", identifier, Location, identifier))
	apiArtifactRegexp        = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/artifacts/%s$", identifier, Location, identifier, identifier))
	versionArtifactRegexp    = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/artifacts/%s$", identifier, Location, identifier, identifier, identifier))
	specArtifactRegexp       = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/versions/%s/specs/%s(?:@%s)?/artifacts/%s$", identifier, Location, identifier, identifier, identifier, revisionTag, identifier))
	deploymentArtifactRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/deployments/%s(?:@%s)?/artifacts/%s$", identifier, Location, identifier, identifier, revisionTag, identifier))
)

// Artifact represents a resource name for an artifact.
type Artifact struct {
	name interface {
		String() string
		Validate() error
	}
}

// Project returns the name of this resource's parent project.
func (a Artifact) Project() Project {
	return Project{
		ProjectID: a.ProjectID(),
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
	case deploymentArtifact:
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
	case deploymentArtifact:
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

// RevisionID returns the artifact's revision ID, or empty string if it doesn't have one.
func (a Artifact) RevisionID() string {
	switch name := a.name.(type) {
	case specArtifact:
		return name.RevisionID
	case deploymentArtifact:
		return name.RevisionID
	default:
		return ""
	}
}

// DeploymentID returns the artifact's deployment ID, or empty string if it doesn't have one.
func (a Artifact) DeploymentID() string {
	switch name := a.name.(type) {
	case deploymentArtifact:
		return name.DeploymentID
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
	case deploymentArtifact:
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
	case deploymentArtifact:
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
	} else if n, err := parseDeploymentArtifact(name); err == nil {
		return Artifact{name: n}, nil
	}

	return Artifact{}, fmt.Errorf("invalid artifact name %q, must match one of: %v", name, []string{
		projectArtifactRegexp.String(),
		apiArtifactRegexp.String(),
		versionArtifactRegexp.String(),
		specArtifactRegexp.String(),
		deploymentArtifactRegexp.String(),
	})
}

// ParseArtifactCollection parses the name of an artifact collection.
func ParseArtifactCollection(name string) (Artifact, error) {
	if n, err := parseSpecArtifactCollection(name); err == nil {
		return Artifact{name: n}, nil
	} else if n, err := parseVersionArtifactCollection(name); err == nil {
		return Artifact{name: n}, nil
	} else if n, err := parseApiArtifactCollection(name); err == nil {
		return Artifact{name: n}, nil
	} else if n, err := parseProjectArtifactCollection(name); err == nil {
		return Artifact{name: n}, nil
	} else if n, err := parseDeploymentArtifactCollection(name); err == nil {
		return Artifact{name: n}, nil
	}

	return Artifact{}, fmt.Errorf("invalid artifact collection name %q, must match one of: %v", name, []string{
		projectArtifactCollectionRegexp.String(),
		apiArtifactCollectionRegexp.String(),
		versionArtifactCollectionRegexp.String(),
		specArtifactCollectionRegexp.String(),
		deploymentArtifactCollectionRegexp.String(),
	})
}

type projectArtifact struct {
	ProjectID  string
	ArtifactID string
}

func (a projectArtifact) Validate() error {
	if err := validateID(a.ArtifactID); err != nil {
		return err
	}

	if name := a.String(); !projectArtifactRegexp.MatchString(name) {
		return fmt.Errorf("invalid project artifact name %q: must match %q", name, projectArtifactRegexp)
	}

	return nil
}

func (a projectArtifact) Parent() string {
	return fmt.Sprintf("projects/%s/locations/%s", a.ProjectID, Location)
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

func parseProjectArtifactCollection(name string) (projectArtifact, error) {
	if !projectArtifactCollectionRegexp.MatchString(name) {
		return projectArtifact{}, fmt.Errorf("invalid project artifact name %q: must match %q", name, projectArtifactCollectionRegexp)
	}

	m := projectArtifactCollectionRegexp.FindStringSubmatch(name)
	artifact := projectArtifact{
		ProjectID:  m[1],
		ArtifactID: "",
	}

	return artifact, nil
}

type apiArtifact struct {
	ProjectID  string
	ApiID      string
	ArtifactID string
}

func (a apiArtifact) Validate() error {
	if err := validateID(a.ArtifactID); err != nil {
		return err
	}

	if name := a.String(); !apiArtifactRegexp.MatchString(name) {
		return fmt.Errorf("invalid api artifact name %q: must match %q", name, apiArtifactRegexp)
	}

	return nil
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

func parseApiArtifactCollection(name string) (apiArtifact, error) {
	if !apiArtifactCollectionRegexp.MatchString(name) {
		return apiArtifact{}, fmt.Errorf("invalid api artifact name %q: must match %q", name, apiArtifactCollectionRegexp)
	}

	m := apiArtifactCollectionRegexp.FindStringSubmatch(name)
	artifact := apiArtifact{
		ProjectID:  m[1],
		ApiID:      m[2],
		ArtifactID: "",
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
	if err := validateID(a.ArtifactID); err != nil {
		return err
	}

	if name := a.String(); !versionArtifactRegexp.MatchString(name) {
		return fmt.Errorf("invalid version artifact name %q: must match %q", name, versionArtifactRegexp)
	}

	return nil
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

func parseVersionArtifactCollection(name string) (versionArtifact, error) {
	if !versionArtifactCollectionRegexp.MatchString(name) {
		return versionArtifact{}, fmt.Errorf("invalid version artifact name %q: must match %q", name, versionArtifactCollectionRegexp)
	}

	m := versionArtifactCollectionRegexp.FindStringSubmatch(name)
	artifact := versionArtifact{
		ProjectID:  m[1],
		ApiID:      m[2],
		VersionID:  m[3],
		ArtifactID: "",
	}

	return artifact, nil
}

type specArtifact struct {
	ProjectID  string
	ApiID      string
	VersionID  string
	SpecID     string
	RevisionID string
	ArtifactID string
}

func (a specArtifact) Validate() error {
	if err := validateID(a.ArtifactID); err != nil {
		return err
	}

	if name := a.String(); !specArtifactRegexp.MatchString(name) {
		return fmt.Errorf("invalid spec artifact name %q: must match %q", name, specArtifactRegexp)
	}

	return nil
}

func (a specArtifact) Parent() string {
	return SpecRevision{
		ProjectID:  a.ProjectID,
		ApiID:      a.ApiID,
		VersionID:  a.VersionID,
		SpecID:     a.SpecID,
		RevisionID: a.RevisionID,
	}.String()
}

func (a specArtifact) String() string {
	if a.RevisionID == "" { // use latest revision
		return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/specs/%s/artifacts/%s",
			a.ProjectID, Location, a.ApiID, a.VersionID, a.SpecID, a.ArtifactID))
	} else {
		return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/versions/%s/specs/%s@%s/artifacts/%s",
			a.ProjectID, Location, a.ApiID, a.VersionID, a.SpecID, a.RevisionID, a.ArtifactID))
	}
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
		RevisionID: m[5],
		ArtifactID: m[6],
	}

	return artifact, nil
}

func parseSpecArtifactCollection(name string) (specArtifact, error) {
	if !specArtifactCollectionRegexp.MatchString(name) {
		return specArtifact{}, fmt.Errorf("invalid spec artifact name %q: must match %q", name, specArtifactCollectionRegexp)
	}

	m := specArtifactCollectionRegexp.FindStringSubmatch(name)
	artifact := specArtifact{
		ProjectID:  m[1],
		ApiID:      m[2],
		VersionID:  m[3],
		SpecID:     m[4],
		RevisionID: m[5],
		ArtifactID: "",
	}

	return artifact, nil
}

type deploymentArtifact struct {
	ProjectID    string
	ApiID        string
	DeploymentID string
	RevisionID   string
	ArtifactID   string
}

func (a deploymentArtifact) Validate() error {
	if err := validateID(a.ArtifactID); err != nil {
		return err
	}

	if name := a.String(); !deploymentArtifactRegexp.MatchString(name) {
		return fmt.Errorf("invalid version artifact name %q: must match %q", name, deploymentArtifactRegexp)
	}

	return nil
}

func (a deploymentArtifact) Parent() string {
	return DeploymentRevision{
		ProjectID:    a.ProjectID,
		ApiID:        a.ApiID,
		DeploymentID: a.DeploymentID,
		RevisionID:   a.RevisionID,
	}.String()
}

func (a deploymentArtifact) String() string {
	if a.RevisionID == "" { // use latest revision
		return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s/artifacts/%s",
			a.ProjectID, Location, a.ApiID, a.DeploymentID, a.ArtifactID))
	} else {
		return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s@%s/artifacts/%s",
			a.ProjectID, Location, a.ApiID, a.DeploymentID, a.RevisionID, a.ArtifactID))
	}
}

func parseDeploymentArtifact(name string) (deploymentArtifact, error) {
	if !deploymentArtifactRegexp.MatchString(name) {
		return deploymentArtifact{}, fmt.Errorf("invalid deployment artifact name %q: must match %q", name, deploymentArtifactRegexp)
	}

	m := deploymentArtifactRegexp.FindStringSubmatch(name)
	artifact := deploymentArtifact{
		ProjectID:    m[1],
		ApiID:        m[2],
		DeploymentID: m[3],
		RevisionID:   m[4],
		ArtifactID:   m[5],
	}

	return artifact, nil
}

func parseDeploymentArtifactCollection(name string) (deploymentArtifact, error) {
	if !deploymentArtifactCollectionRegexp.MatchString(name) {
		return deploymentArtifact{}, fmt.Errorf("invalid deployment artifact name %q: must match %q", name, deploymentArtifactCollectionRegexp)
	}

	m := deploymentArtifactCollectionRegexp.FindStringSubmatch(name)
	artifact := deploymentArtifact{
		ProjectID:    m[1],
		ApiID:        m[2],
		DeploymentID: m[3],
		RevisionID:   m[4],
		ArtifactID:   "",
	}

	return artifact, nil
}
