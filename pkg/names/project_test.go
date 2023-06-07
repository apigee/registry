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

func TestProjectNames(t *testing.T) {
	name := &Project{
		ProjectID: "p",
	}
	err := name.Validate()
	if err != nil {
		t.Errorf("Validate() failed for name %s: %s", name, err)
	}
	if name.Api("a").String() != "projects/p/locations/global/apis/a" {
		t.Errorf("%s Api() returned incorrect value %s", name, name.Api("a"))
	}
	if name.Artifact("x").String() != "projects/p/locations/global/artifacts/x" {
		t.Errorf("%s Artifact() returned incorrect value %s", name, name.Artifact("x"))
	}
}

func TestInvalidProjectNames(t *testing.T) {
	names := []Project{
		{
			ProjectID: "!!",
		},
	}
	for _, name := range names {
		err := name.Validate()
		if err == nil {
			t.Errorf("Validate() succeeded for %s and should have failed", name)
		}
	}
}
