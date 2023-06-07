// Copyright 2023 Google LLC.
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
	"testing"
)

func TestArtifactNames(t *testing.T) {
	tests := []struct {
		name         Artifact
		projectID    string
		apiID        string
		versionID    string
		specID       string
		deploymentID string
		revisionID   string
		artifactID   string
		parent       string
	}{
		{
			name: Artifact{
				name: projectArtifact{
					ProjectID:  "p",
					ArtifactID: "x",
				},
			},
			projectID:  "p",
			artifactID: "x",
			parent:     "projects/p/locations/global",
		},
		{
			name: Artifact{
				name: apiArtifact{
					ProjectID:  "p",
					ApiID:      "a",
					ArtifactID: "x",
				},
			},
			projectID:  "p",
			apiID:      "a",
			artifactID: "x",
			parent:     "projects/p/locations/global/apis/a",
		},
		{
			name: Artifact{
				name: versionArtifact{
					ProjectID:  "p",
					ApiID:      "a",
					VersionID:  "v",
					ArtifactID: "x",
				},
			},
			projectID:  "p",
			apiID:      "a",
			versionID:  "v",
			artifactID: "x",
			parent:     "projects/p/locations/global/apis/a/versions/v",
		},
		{
			name: Artifact{
				name: specArtifact{
					ProjectID:  "p",
					ApiID:      "a",
					VersionID:  "v",
					SpecID:     "s",
					RevisionID: "123",
					ArtifactID: "x",
				},
			},
			projectID:  "p",
			apiID:      "a",
			versionID:  "v",
			specID:     "s",
			revisionID: "123",
			artifactID: "x",
			parent:     "projects/p/locations/global/apis/a/versions/v/specs/s@123",
		},
		{
			name: Artifact{
				name: deploymentArtifact{
					ProjectID:    "p",
					ApiID:        "a",
					DeploymentID: "d",
					RevisionID:   "123",
					ArtifactID:   "x",
				},
			},
			projectID:    "p",
			apiID:        "a",
			deploymentID: "d",
			revisionID:   "123",
			artifactID:   "x",
			parent:       "projects/p/locations/global/apis/a/deployments/d@123",
		},
	}
	for _, test := range tests {
		err := test.name.Validate()
		if err != nil {
			t.Errorf("Validate() failed for name %s: %s", test.name, err)
		}
		if test.name.ProjectID() != test.projectID {
			t.Errorf("%s ProjectID() returned incorrect value %s", test.name, test.name.ProjectID())
		}
		if test.name.ApiID() != test.apiID {
			t.Errorf("%s ApiID() returned incorrect value %s", test.name, test.name.ApiID())
		}
		if test.name.VersionID() != test.versionID {
			t.Errorf("%s VersionID() returned incorrect value %s", test.name, test.name.VersionID())
		}
		if test.name.SpecID() != test.specID {
			t.Errorf("%s SpecID() returned incorrect value %s", test.name, test.name.SpecID())
		}
		if test.name.DeploymentID() != test.deploymentID {
			t.Errorf("%s DeploymentID() returned incorrect value %s", test.name, test.name.DeploymentID())
		}
		if test.name.RevisionID() != test.revisionID {
			t.Errorf("%s RevisionID() returned incorrect value %s", test.name, test.name.RevisionID())
		}
		if test.name.ArtifactID() != test.artifactID {
			t.Errorf("%s ArtifactID() returned incorrect value %s", test.name, test.name.ArtifactID())
		}
		if test.name.Parent() != test.parent {
			t.Errorf("%s Parent() returned incorrect value %s", test.name, test.name.Parent())
		}
	}
}

func TestInvalidArtifactNames(t *testing.T) {
	names := []Artifact{
		{
			name: projectArtifact{
				ProjectID:  "!!",
				ArtifactID: "x",
			},
		},
		{
			name: projectArtifact{
				ProjectID:  "p",
				ArtifactID: "!!",
			},
		},
		{
			name: apiArtifact{
				ProjectID:  "p",
				ApiID:      "!!",
				ArtifactID: "x",
			},
		},
		{
			name: apiArtifact{
				ProjectID:  "p",
				ApiID:      "a",
				ArtifactID: "!!",
			},
		},
		{
			name: versionArtifact{
				ProjectID:  "p",
				ApiID:      "a",
				VersionID:  "!!",
				ArtifactID: "x",
			},
		},
		{
			name: versionArtifact{
				ProjectID:  "p",
				ApiID:      "a",
				VersionID:  "v",
				ArtifactID: "!!",
			},
		},
		{
			name: specArtifact{
				ProjectID:  "p",
				ApiID:      "a",
				VersionID:  "v",
				SpecID:     "!!",
				RevisionID: "123",
				ArtifactID: "x",
			},
		},
		{
			name: specArtifact{
				ProjectID:  "p",
				ApiID:      "a",
				VersionID:  "v",
				SpecID:     "s",
				RevisionID: "!!!",
				ArtifactID: "x",
			},
		},
		{
			name: specArtifact{
				ProjectID:  "p",
				ApiID:      "a",
				VersionID:  "v",
				SpecID:     "s",
				RevisionID: "123",
				ArtifactID: "!!",
			},
		},
		{
			name: deploymentArtifact{
				ProjectID:    "p",
				ApiID:        "a",
				DeploymentID: "!!",
				RevisionID:   "123",
				ArtifactID:   "x",
			},
		},
		{
			name: deploymentArtifact{
				ProjectID:    "p",
				ApiID:        "a",
				DeploymentID: "d",
				RevisionID:   "!!!",
				ArtifactID:   "x",
			},
		},
		{
			name: deploymentArtifact{
				ProjectID:    "p",
				ApiID:        "a",
				DeploymentID: "d",
				RevisionID:   "123",
				ArtifactID:   "!!",
			},
		},
	}
	for _, name := range names {
		err := name.Validate()
		if err == nil {
			t.Errorf("Validate() succeeded for %s and should have failed", name)
		}
	}
}

func TestNullArtifactName(t *testing.T) {
	name := Artifact{}
	if name.ProjectID() != "" {
		t.Errorf("%s has incorrect project id %s", name, name.ProjectID())
	}
	if name.ApiID() != "" {
		t.Errorf("%s has incorrect api id %s", name, name.ApiID())
	}
	if name.VersionID() != "" {
		t.Errorf("%s has incorrect version id %s", name, name.VersionID())
	}
	if name.SpecID() != "" {
		t.Errorf("%s has incorrect spec id %s", name, name.SpecID())
	}
	if name.DeploymentID() != "" {
		t.Errorf("%s has incorrect deployment id %s", name, name.DeploymentID())
	}
	if name.RevisionID() != "" {
		t.Errorf("%s has incorrect revision id %s", name, name.RevisionID())
	}
	if name.ArtifactID() != "" {
		t.Errorf("%s has incorrect artifact id %s", name, name.ArtifactID())
	}
	if name.Parent() != "" {
		t.Errorf("%s has incorrect parent %s", name, name.Parent())
	}
}
