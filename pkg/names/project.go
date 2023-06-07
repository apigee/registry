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

// Project represents a resource name for a project.
type Project struct {
	ProjectID string
}

// Validate returns an error if the resource name is invalid.
// For backward compatibility, names should only be validated at creation time.
func (p Project) Validate() error {
	// Project names are simple enough that we can be sure that a name with a
	// valid ProjectID that is generated with the String() function (below) will
	// satisfy the regular expression returned by projectRegexp() (also below).
	// So unlike other names, this Validate() includes no regular expression check.
	return validateID(p.ProjectID)
}

// Project returns this resource
func (p Project) Project() Project {
	return p
}

// Api returns an API with the provided ID and this resource as its parent.
func (p Project) Api(id string) Api {
	return Api{
		ProjectID: p.ProjectID,
		ApiID:     id,
	}
}

// Artifact returns an artifact with the provided ID and this resource as its parent.
func (p Project) Artifact(id string) Artifact {
	return Artifact{
		name: projectArtifact{
			ProjectID:  p.ProjectID,
			ArtifactID: id,
		},
	}
}

func (p Project) String() string {
	return normalize(fmt.Sprintf("projects/%s", p.ProjectID))
}

// projectCollectionRegexp returns a regular expression that matches collection of projects.
func projectCollectionRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects$")
}

// projectRegexp returns a regular expression that matches a project resource name.
func projectRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^projects/%s$", identifier))
}

// projectWithLocationRegexp returns a regular expression that matches a project resource name followed by a location.
func projectWithLocationRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s$", identifier, Location))
}

// ParseProject parses the name of a project.
func ParseProject(name string) (Project, error) {
	r := projectRegexp()
	if !r.MatchString(name) {
		return Project{}, fmt.Errorf("invalid project name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	return Project{
		ProjectID: m[1],
	}, nil
}

// ParseProjectCollection parses the name of a project collection.
func ParseProjectCollection(name string) (Project, error) {
	r := projectCollectionRegexp()
	if !r.MatchString(name) {
		return Project{}, fmt.Errorf("invalid project collection name %q: must match %q", name, r)
	}

	return Project{
		ProjectID: "",
	}, nil
}

// ParseProjectWithLocation parses the name of a project.
func ParseProjectWithLocation(name string) (Project, error) {
	r := projectWithLocationRegexp()
	if !r.MatchString(name) {
		return Project{}, fmt.Errorf("invalid project name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	return Project{
		ProjectID: m[1],
	}, nil
}
