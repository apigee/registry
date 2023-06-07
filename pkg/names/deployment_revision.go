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

var deploymentRevisionRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/deployments/%s(?:@%s)?$",
	identifier, Location, identifier, identifier, revisionTag))

var deploymentRevisionCollectionRegexp = regexp.MustCompile(fmt.Sprintf("^projects/%s/locations/%s/apis/%s/deployments/%s@$",
	identifier, Location, identifier, identifier))

// DeploymentRevision represents a resource name for an API deployment revision.
type DeploymentRevision struct {
	ProjectID    string
	ApiID        string
	DeploymentID string
	RevisionID   string
}

// Deployment returns the parent deployment for this resource.
func (s DeploymentRevision) Deployment() Deployment {
	return Deployment{
		ProjectID:    s.ProjectID,
		ApiID:        s.ApiID,
		DeploymentID: s.DeploymentID,
	}
}

// Project returns the parent project for this resource.
func (d DeploymentRevision) Project() Project {
	return Project{
		ProjectID: d.ProjectID,
	}
}

// Api returns the parent API for this resource.
func (d DeploymentRevision) Api() Api {
	return Api{
		ProjectID: d.ProjectID,
		ApiID:     d.ApiID,
	}
}

// Artifact returns an artifact with the provided ID and this resource as its parent.
func (s DeploymentRevision) Artifact(id string) Artifact {
	return Artifact{
		name: deploymentArtifact{
			ProjectID:    s.ProjectID,
			ApiID:        s.ApiID,
			DeploymentID: s.DeploymentID,
			RevisionID:   s.RevisionID,
			ArtifactID:   id,
		},
	}
}

// Parent returns this resource's parent API resource name.
func (s DeploymentRevision) Parent() string {
	return s.Deployment().Parent()
}

func (s DeploymentRevision) String() string {
	if s.RevisionID == "" { // use latest revision
		return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s",
			s.ProjectID, Location, s.ApiID, s.DeploymentID))
	} else {
		return normalize(fmt.Sprintf("projects/%s/locations/%s/apis/%s/deployments/%s@%s",
			s.ProjectID, Location, s.ApiID, s.DeploymentID, s.RevisionID))
	}
}

// ParseDeploymentRevision parses the name of a deployment.
func ParseDeploymentRevision(name string) (DeploymentRevision, error) {
	if !deploymentRevisionRegexp.MatchString(name) {
		return DeploymentRevision{}, fmt.Errorf("invalid deployment revision name %q: must match %q", name, deploymentRevisionRegexp)
	}

	m := deploymentRevisionRegexp.FindStringSubmatch(name)
	revision := DeploymentRevision{
		ProjectID:    m[1],
		ApiID:        m[2],
		DeploymentID: m[3],
		RevisionID:   m[4],
	}

	return revision, nil
}

func ParseDeploymentRevisionCollection(name string) (DeploymentRevision, error) {
	r := deploymentRevisionCollectionRegexp
	if !r.MatchString(name) {
		return DeploymentRevision{}, fmt.Errorf("invalid deployment revision collection name %q: must match %q", name, r)
	}

	m := r.FindStringSubmatch(name)
	return DeploymentRevision{
		ProjectID:    m[1],
		ApiID:        m[2],
		DeploymentID: m[3],
		RevisionID:   "-",
	}, nil
}
