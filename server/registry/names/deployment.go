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

// Deployment represents a resource name for an API deployment.
type Deployment struct {
	ProjectID    string
	ApiID        string
	DeploymentID string
}

// Validate returns an error if the resource name is invalid.
// For backward compatibility, names should only be validated at creation time.
func (v Deployment) Validate() error {
	r := deploymentRegexp()
	if name := v.String(); !r.MatchString(name) {
		return fmt.Errorf("invalid deployment name %q: must match %q", name, r)
	}

	return validateID(v.DeploymentID)
}

// Project returns the parent project for this resource.
func (v Deployment) Project() Project {
	return v.Api().Project()
}

// Api returns the parent API for this resource.
func (v Deployment) Api() Api {
	return Api{
		ProjectID: v.ProjectID,
		ApiID:     v.ApiID,
	}
}

// Revision returns an API deployment revision with the provided ID and this resource as its parent.
func (s Deployment) Revision(id string) DeploymentRevision {
	return DeploymentRevision{
		ProjectID:    s.ProjectID,
		ApiID:        s.ApiID,
		DeploymentID: s.DeploymentID,
		RevisionID:   id,
	}
}

// Artifact returns an artifact with the provided ID and this resource as its parent.
func (v Deployment) Artifact(id string) Artifact {
	return Artifact{
		name: deploymentArtifact{
			ProjectID:    v.ProjectID,
			ApiID:        v.ApiID,
			DeploymentID: v.DeploymentID,
			ArtifactID:   id,
		},
	}
}

// Parent returns this resource's parent API resource name.
func (v Deployment) Parent() string {
	return v.Api().String()
}

func (v Deployment) String() string {
	return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s",
		v.ProjectID, Location, v.ApiID, v.DeploymentID))
}

// deploymentCollectionRegexp returns a regular expression that matches a collection of deployments.
func deploymentCollectionRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/deployments$",
		identifier, Location, identifier))
}

// deploymentRegexp returns a regular expression that matches a deployment resource name.
func deploymentRegexp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/deployments/%s$",
		identifier, Location, identifier, identifier))
}

// ParseDeployment parses the name of a deployment.
func ParseDeployment(name string) (Deployment, error) {
	r := deploymentRegexp()
	if !r.MatchString(name) {
		return Deployment{}, fmt.Errorf("invalid deployment name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	return Deployment{
		ProjectID:    m[1],
		ApiID:        m[2],
		DeploymentID: m[3],
	}, nil
}

func ParseDeploymentCollection(name string) (Deployment, error) {
	r := deploymentCollectionRegexp()
	if !r.MatchString(name) {
		return Deployment{}, fmt.Errorf("invalid deployment collection name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	return Deployment{
		ProjectID:    m[1],
		ApiID:        m[2],
		DeploymentID: "",
	}, nil
}
