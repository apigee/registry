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

func TestApiNames(t *testing.T) {
	name := &Api{
		ProjectID: "p",
		ApiID:     "a",
	}
	err := name.Validate()
	if err != nil {
		t.Errorf("Validate() failed for name %s: %s", name, err)
	}
	if name.Project().String() != "projects/p" {
		t.Errorf("%s Project() returned incorrect value %s", name, name.Project())
	}
	if name.Version("v").String() != "projects/p/locations/global/apis/a/versions/v" {
		t.Errorf("%s Version() returned incorrect value %s", name, name.Version("v"))
	}
	if name.Deployment("d").String() != "projects/p/locations/global/apis/a/deployments/d" {
		t.Errorf("%s Deployment() returned incorrect value %s", name, name.Deployment("d"))
	}
	if name.Artifact("x").String() != "projects/p/locations/global/apis/a/artifacts/x" {
		t.Errorf("%s Artifact() returned incorrect value %s", name, name.Artifact("x"))
	}
	if name.Parent() != "projects/p/locations/global" {
		t.Errorf("%s Parent() returned incorrect value %s", name, name.Parent())
	}
}

func TestInvalidApiNames(t *testing.T) {
	names := []Api{
		{
			ProjectID: "!!",
			ApiID:     "a",
		},
		{
			ProjectID: "p",
			ApiID:     "!!",
		},
	}
	for _, name := range names {
		err := name.Validate()
		if err == nil {
			t.Errorf("Validate() succeeded for %s and should have failed", name)
		}
	}
}
