// Copyright 2023 Google LLC. All Rights Reserved.
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
	a := &Api{
		ProjectID: "p",
		ApiID:     "a",
	}
	err := a.Validate()
	if err != nil {
		t.Errorf("Validate() failed for name %s: %s", a, err)
	}
	if a.Project().String() != "projects/p" {
		t.Errorf("%s Project() returned incorrect value %s", a, a.Project())
	}
	if a.Version("v").String() != "projects/p/locations/global/apis/a/versions/v" {
		t.Errorf("%s Version() returned incorrect value %s", a, a.Version("v"))
	}
	if a.Deployment("d").String() != "projects/p/locations/global/apis/a/deployments/d" {
		t.Errorf("%s Deployment() returned incorrect value %s", a, a.Deployment("d"))
	}
	if a.Artifact("x").String() != "projects/p/locations/global/apis/a/artifacts/x" {
		t.Errorf("%s Artifact() returned incorrect value %s", a, a.Artifact("x"))
	}
	if a.Parent() != "projects/p/locations/global" {
		t.Errorf("%s Parent() returned incorrect value %s", a, a.Parent())
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
