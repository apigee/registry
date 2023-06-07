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

// deploymentRegexp is a regular expression that matches a deployment resource name.
var deploymentRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/deployments/%s(@%s)?$",
	identifier, Location, identifier, identifier, revisionTag))

// deploymentCollectionRegexp is a regular expression that matches a collection of deployments.
var deploymentCollectionRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/deployments$",
	identifier, Location, identifier))

// simpleDeploymentRegexp is the regex pattern for deployment resource names.
// Notably, this differs from deploymentRegexp by not accepting revision IDs in the resource name.
var simpleDeploymentRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/deployments/%s$",
	identifier, Location, identifier, identifier))

// Deployment represents a resource name for an API deployment.
type Deployment struct {
	ProjectID    string
	ApiID        string
	DeploymentID string
}

// Validate returns an error if the resource name is invalid.
// For backward compatibility, names should only be validated at creation time.
func (d Deployment) Validate() error {
	if err := validateID(d.DeploymentID); err != nil {
		return err
	}

	r := deploymentRegexp
	if name := d.String(); !r.MatchString(name) {
		return fmt.Errorf("invalid deployment name %q: must match %q", name, r)
	}

	return nil
}

// Project returns the parent project for this resource.
func (d Deployment) Project() Project {
	return d.Api().Project()
}

// Api returns the parent API for this resource.
func (d Deployment) Api() Api {
	return Api{
		ProjectID: d.ProjectID,
		ApiID:     d.ApiID,
	}
}

// Revision returns an API deployment revision with the provided ID and this resource as its parent.
func (d Deployment) Revision(id string) DeploymentRevision {
	return DeploymentRevision{
		ProjectID:    d.ProjectID,
		ApiID:        d.ApiID,
		DeploymentID: d.DeploymentID,
		RevisionID:   id,
	}
}

// Artifact returns an artifact with the provided ID and this resource as its parent.
func (d Deployment) Artifact(id string) Artifact {
	return Artifact{
		name: deploymentArtifact{
			ProjectID:    d.ProjectID,
			ApiID:        d.ApiID,
			DeploymentID: d.DeploymentID,
			ArtifactID:   id,
		},
	}
}

// Normal returns the resource name with normalized identifiers.
func (d Deployment) Normal() Deployment {
	return Deployment{
		ProjectID:    normalize(d.ProjectID),
		ApiID:        normalize(d.ApiID),
		DeploymentID: normalize(d.DeploymentID),
	}
}

// Parent returns this resource's parent API resource name.
func (d Deployment) Parent() string {
	return d.Api().String()
}

func (d Deployment) String() string {
	return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s",
		d.ProjectID, Location, d.ApiID, d.DeploymentID))
}

// ParseDeployment parses the name of a deployment.
func ParseDeployment(name string) (Deployment, error) {
	r := simpleDeploymentRegexp
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
	r := deploymentCollectionRegexp
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
