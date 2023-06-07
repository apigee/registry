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

func TestDeploymentNames(t *testing.T) {
	name := &Deployment{
		ProjectID:    "p",
		ApiID:        "a",
		DeploymentID: "d",
	}
	err := name.Validate()
	if err != nil {
		t.Errorf("Validate() failed for name %s: %s", name, err)
	}
	if name.Project().String() != "projects/p" {
		t.Errorf("%s Project() returned incorrect value %s", name, name.Project())
	}
	if name.Api().String() != "projects/p/locations/global/apis/a" {
		t.Errorf("%s Api() returned incorrect value %s", name, name.Api())
	}
	if name.Artifact("x").String() != "projects/p/locations/global/apis/a/deployments/d/artifacts/x" {
		t.Errorf("%s Artifact() returned incorrect value %s", name, name.Artifact("x"))
	}
	if name.Revision("123").String() != "projects/p/locations/global/apis/a/deployments/d@123" {
		t.Errorf("%s Revision() returned incorrect value %s", name, name.Revision("123"))
	}
	if name.Parent() != "projects/p/locations/global/apis/a" {
		t.Errorf("%s Parent() returned incorrect value %s", name, name.Parent())
	}
	if name.Normal().String() != "projects/p/locations/global/apis/a/deployments/d" {
		t.Errorf("%s Normal() returned incorrect value %s", name, name.Normal())
	}
}

func TestInvalidDeploymentNames(t *testing.T) {
	names := []Deployment{
		{
			ProjectID:    "!!",
			ApiID:        "a",
			DeploymentID: "d",
		},
		{
			ProjectID:    "p",
			ApiID:        "!!",
			DeploymentID: "d",
		},
		{
			ProjectID:    "p",
			ApiID:        "a",
			DeploymentID: "!!",
		},
	}
	for _, name := range names {
		err := name.Validate()
		if err == nil {
			t.Errorf("Validate() succeeded for %s and should have failed", name)
		}
	}
}
