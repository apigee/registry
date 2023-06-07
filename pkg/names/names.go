// Copyright 2022 Google LLC.
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

import "fmt"

// Parse attempts to parse a string into an resource name.
func Parse(name string) (Name, error) {
	if collection, err := ParseResourceCollection(name); err == nil {
		return collection, err
	}
	return ParseResourceEntity(name)
}

// Parse attempts to parse a string into a resource collection name.
func ParseResourceCollection(name string) (Name, error) {
	if project, err := ParseProjectCollection(name); err == nil {
		return project, err
	} else if api, err := ParseApiCollection(name); err == nil {
		return api, err
	} else if deployment, err := ParseDeploymentCollection(name); err == nil {
		return deployment, err
	} else if rev, err := ParseDeploymentRevisionCollection(name); err == nil {
		return rev, err
	} else if version, err := ParseVersionCollection(name); err == nil {
		return version, err
	} else if spec, err := ParseSpecCollection(name); err == nil {
		return spec, err
	} else if rev, err := ParseSpecRevisionCollection(name); err == nil {
		return rev, err
	} else if artifact, err := ParseArtifactCollection(name); err == nil {
		return artifact, err
	}
	return nil, fmt.Errorf("invalid name: %s", name)
}

// Parse attempts to parse a string into a single resource name.
func ParseResourceEntity(name string) (Name, error) {
	if project, err := ParseProject(name); err == nil {
		return project, err
	} else if api, err := ParseApi(name); err == nil {
		return api, err
	} else if deployment, err := ParseDeployment(name); err == nil {
		return deployment, err
	} else if rev, err := ParseDeploymentRevision(name); err == nil {
		return rev, err
	} else if version, err := ParseVersion(name); err == nil {
		return version, err
	} else if spec, err := ParseSpec(name); err == nil {
		return spec, err
	} else if rev, err := ParseSpecRevision(name); err == nil {
		return rev, err
	} else if artifact, err := ParseArtifact(name); err == nil {
		return artifact, err
	}
	return nil, fmt.Errorf("invalid name: %s", name)
}
