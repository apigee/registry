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

func TestVersionNames(t *testing.T) {
	name := &Version{
		ProjectID: "p",
		ApiID:     "a",
		VersionID: "v",
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
	if name.Spec("s").String() != "projects/p/locations/global/apis/a/versions/v/specs/s" {
		t.Errorf("%s Spec() returned incorrect value %s", name, name.Spec("v"))
	}
	if name.Artifact("x").String() != "projects/p/locations/global/apis/a/versions/v/artifacts/x" {
		t.Errorf("%s Artifact() returned incorrect value %s", name, name.Artifact("x"))
	}
	if name.Parent() != "projects/p/locations/global/apis/a" {
		t.Errorf("%s Parent() returned incorrect value %s", name, name.Parent())
	}
}

func TestInvalidVersionNames(t *testing.T) {
	names := []Version{
		{
			ProjectID: "!!",
			ApiID:     "a",
			VersionID: "v",
		},
		{
			ProjectID: "p",
			ApiID:     "!!",
			VersionID: "v",
		},
		{
			ProjectID: "p",
			ApiID:     "a",
			VersionID: "!!",
		},
	}
	for _, name := range names {
		err := name.Validate()
		if err == nil {
			t.Errorf("Validate() succeeded for %s and should have failed", name)
		}
	}
}
