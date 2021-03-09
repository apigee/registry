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

// Project represents a resource name for a project.
type Project struct {
	ProjectID string
}

// Validate returns an error if the resource name is invalid.
// For backward compatibility, names should only be validated at creation time.
func (p Project) Validate() error {
	r := ProjectRegexp()
	if name := p.String(); !r.MatchString(name) {
		return fmt.Errorf("invalid project name %q: must match %q", name, r)
	}

	return nil
}

func (p Project) String() string {
	return fmt.Sprintf("projects/%s", p.ProjectID)
}

// ProjectsRegexp returns a regular expression that matches collection of projects.
func ProjectsRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects$")
}

// ProjectRegexp returns a regular expression that matches a project resource name.
func ProjectRegexp() *regexp.Regexp {
	return regexp.MustCompile("^projects/" + identifier + "$")
}

// ParseProject parses the name of a project.
func ParseProject(name string) (Project, error) {
	r := ProjectRegexp()
	if !r.MatchString(name) {
		return Project{}, fmt.Errorf("invalid project name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	return Project{
		ProjectID: m[1],
	}, nil
}
