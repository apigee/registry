// Copyright 2020 Google LLC.
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

func TestResourceNames(t *testing.T) {
	groups := []struct {
		name  string
		check func(name string) bool
		pass  []string
		fail  []string
	}{
		{
			name: "project collections",
			check: func(name string) bool {
				_, err := ParseProjectCollection(name)
				return err == nil
			},
			pass: []string{
				"projects",
			},
			fail: []string{
				"-",
				"project",
				"organizations",
				"apis",
			},
		},
		{
			name: "project",
			check: func(name string) bool {
				_, err := ParseProject(name)
				return err == nil
			},
			pass: []string{
				"projects/google",
				"projects/-",
			},
			fail: []string{
				"-",
			},
		},
		{
			name: "project with location",
			check: func(name string) bool {
				_, err := ParseProjectWithLocation(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global",
				"projects/-/locations/global",
			},
			fail: []string{
				"-",
				"projects/google",
				"projects/-",
				"projects/google/locations/us-central1",
				"projects/-/locations/us-central1",
			},
		},
		{
			name: "api collections",
			check: func(name string) bool {
				_, err := ParseApiCollection(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis",
				"projects/-/locations/global/apis",
			},
			fail: []string{
				"-",
			},
		},
		{
			name: "api",
			check: func(name string) bool {
				_, err := ParseApi(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis/sample",
				"projects/-/locations/global/apis/-",
				"projects/123/locations/global/apis/abc",
				"projects/1-2-3/locations/global/apis/abc",
			},
			fail: []string{
				"-",
				"invalid",
				"projects//locations/global/apis/123",
				"projects/123/locations/global/apis/",
				"projects/123/locations/global/invalid/123",
				"projects/123/locations/global/apis/ 123",
			},
		},
		{
			name: "version collections",
			check: func(name string) bool {
				_, err := ParseVersionCollection(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis/sample/versions",
				"projects/-/locations/global/apis/-/versions",
			},
			fail: []string{
				"-",
			},
		},
		{
			name: "version",
			check: func(name string) bool {
				_, err := ParseVersion(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis/sample/versions/v1",
				"projects/-/locations/global/apis/-/versions/-",
				"projects/123/locations/global/apis/abc/versions/123",
				"projects/1-2-3/locations/global/apis/abc/versions/123",
			},
			fail: []string{
				"-",
				"invalid",
				"projects//locations/global/apis/123",
				"projects/123/locations/global/apis/",
				"projects/123/locations/global/invalid/123",
				"projects/123/locations/global/apis/ 123",
			},
		},
		{
			name: "spec collections",
			check: func(name string) bool {
				_, err := ParseSpecCollection(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis/sample/versions/v1/specs",
				"projects/-/locations/global/apis/-/versions/-/specs",
			},
			fail: []string{
				"-",
			},
		},
		{
			name: "spec",
			check: func(name string) bool {
				_, err := ParseSpec(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis/sample/versions/v1/specs/openapi",
				"projects/-/locations/global/apis/-/versions/-/specs/-",
				"projects/123/locations/global/apis/abc/versions/123/specs/abc",
				"projects/1-2-3/locations/global/apis/abc/versions/123/specs/abc",
			},
			fail: []string{
				"-",
				"invalid",
				"projects//locations/global/apis/123",
				"projects/123/locations/global/apis/",
				"projects/123/locations/global/invalid/123",
				"projects/123/locations/global/apis/ 123",
				"projects/google/locations/global/apis/sample/versions/v1/specs/openapi@1234567890abcdef",
			},
		},
		{
			name: "spec revision collections",
			check: func(name string) bool {
				_, err := ParseSpecRevisionCollection(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis/sample/versions/v1/specs/s@",
				"projects/-/locations/global/apis/-/versions/-/specs/-@",
			},
			fail: []string{
				"-",
				"projects/google/locations/global/apis/sample/versions/v1/specs/s@123",
				"projects/-/locations/global/apis/-/versions/-/specs/-",
			},
		},
		{
			name: "spec revision",
			check: func(name string) bool {
				_, err := ParseSpecRevision(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis/sample/versions/v1/specs/openapi@1234567890abcdef",
				"projects/google/locations/global/apis/sample/versions/v1/specs/openapi",
				"projects/-/locations/global/apis/-/versions/-/specs/-",
				"projects/123/locations/global/apis/abc/versions/123/specs/abc",
				"projects/1-2-3/locations/global/apis/abc/versions/123/specs/abc",
			},
			fail: []string{
				"-",
				"invalid",
				"projects//locations/global/apis/123",
				"projects/123/locations/global/apis/",
				"projects/123/locations/global/invalid/123",
				"projects/123/locations/global/apis/ 123",
			},
		},
		{
			name: "deployment collections",
			check: func(name string) bool {
				_, err := ParseDeploymentCollection(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis/sample/deployments",
				"projects/-/locations/global/apis/-/deployments",
			},
			fail: []string{
				"-",
			},
		},
		{
			name: "deployment",
			check: func(name string) bool {
				_, err := ParseDeployment(name)
				return err == nil
			}, pass: []string{
				"projects/google/locations/global/apis/sample/deployments/v1",
				"projects/-/locations/global/apis/-/deployments/-",
				"projects/123/locations/global/apis/abc/deployments/123",
				"projects/1-2-3/locations/global/apis/abc/deployments/123",
			},
			fail: []string{
				"-",
				"invalid",
				"projects//locations/global/apis/123",
				"projects/123/locations/global/apis/",
				"projects/123/locations/global/invalid/123",
				"projects/123/locations/global/apis/ 123",
				"projects/google/locations/global/apis/sample/deployments/v1@1234567890abcdef",
			},
		},
		{
			name: "deployment revision collections",
			check: func(name string) bool {
				_, err := ParseDeploymentRevisionCollection(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/apis/sample/deployments/d@",
				"projects/-/locations/global/apis/-/deployments/d@",
			},
			fail: []string{
				"-",
				"projects/google/locations/global/apis/sample/deployments/d@123",
				"projects/-/locations/global/apis/-/deployments/d",
			},
		},
		{
			name: "deployment revision",
			check: func(name string) bool {
				_, err := ParseDeploymentRevision(name)
				return err == nil
			}, pass: []string{
				"projects/google/locations/global/apis/sample/deployments/v1@1234567890abcdef",
				"projects/google/locations/global/apis/sample/deployments/v1",
				"projects/-/locations/global/apis/-/deployments/-",
				"projects/123/locations/global/apis/abc/deployments/123",
				"projects/1-2-3/locations/global/apis/abc/deployments/123",
			},
			fail: []string{
				"-",
				"invalid",
				"projects//locations/global/apis/123",
				"projects/123/locations/global/apis/",
				"projects/123/locations/global/invalid/123",
				"projects/123/locations/global/apis/ 123",
			},
		},
		{
			name: "artifact collections",
			check: func(name string) bool {
				_, err := ParseArtifactCollection(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/artifacts",
				"projects/google/locations/global/apis/sample/artifacts",
				"projects/google/locations/global/apis/sample/versions/v1/artifacts",
				"projects/google/locations/global/apis/sample/versions/v1/specs/openapi/artifacts",
				"projects/google/locations/global/apis/sample/deployments/prod/artifacts",
			},
			fail: []string{
				"-",
			},
		},
		{
			name: "artifact",
			check: func(name string) bool {
				_, err := ParseArtifact(name)
				return err == nil
			},
			pass: []string{
				"projects/google/locations/global/artifacts/test-artifact",
				"projects/google/locations/global/apis/sample/artifacts/test-artifact",
				"projects/google/locations/global/apis/sample/versions/v1/artifacts/test-artifact",
				"projects/google/locations/global/apis/sample/versions/v1/specs/openapi/artifacts/test-artifact",
				"projects/google/locations/global/apis/sample/deployments/prod/artifacts/test-artifact",
			},
			fail: []string{
				"-",
			},
		},
	}
	for _, g := range groups {
		t.Run(g.name, func(t *testing.T) {
			for _, path := range g.pass {
				if !g.check(path) {
					t.Errorf("failed to match: %s", path)
				}
			}
			for _, path := range g.fail {
				if g.check(path) {
					t.Errorf("false match: %s", path)
				}
			}
		})
	}
}

func TestRevisionTags(t *testing.T) {
	tests := []struct {
		desc  string
		tag   string
		valid bool
	}{
		{
			desc:  "all letters",
			tag:   "sample",
			valid: true,
		},
		{
			desc:  "letters numbers and dashes",
			tag:   "alpha-1-beta-2-gamma-3",
			valid: true,
		},
		{
			desc:  "mixed case",
			tag:   "MixedCase",
			valid: false,
		},
		{
			desc:  "one letter",
			tag:   "x",
			valid: true,
		},
		{
			desc:  "one digit",
			tag:   "9",
			valid: true,
		},
		{
			desc:  "one dash",
			tag:   "-",
			valid: false,
		},
		{
			desc:  "two dashes",
			tag:   "--",
			valid: true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := ValidateRevisionTag(test.tag)
			if test.valid {
				if err != nil {
					t.Errorf("%s should be valid but was rejected with error %s", test.tag, err)
				}
			} else {
				if err == nil {
					t.Errorf("%s should be invalid but was accepted", test.tag)
				}
			}
		})
	}
}

func TestExportableName(t *testing.T) {
	tests := []struct {
		name       string
		projectID  string
		exportable string
	}{
		{
			name:       "projects/my-project/locations/global/apis/a",
			projectID:  "my-project",
			exportable: "apis/a",
		},
		{
			name:       "projects/my-project/locations/global/apis/a/versions/v",
			projectID:  "my-project",
			exportable: "apis/a/versions/v",
		},
		{
			name:       "projects/my-project/locations/global/apis/a/versions/v/specs/s",
			projectID:  "my-project",
			exportable: "apis/a/versions/v/specs/s",
		},
		{
			name:       "projects/my-project/locations/global/apis/a/versions/v/specs/s@123",
			projectID:  "my-project",
			exportable: "apis/a/versions/v/specs/s",
		},
		{
			name:       "projects/my-project/locations/global/apis/a/deployments/d",
			projectID:  "my-project",
			exportable: "apis/a/deployments/d",
		},
		{
			name:       "projects/my-project/locations/global/apis/a/deployments/d@123",
			projectID:  "my-project",
			exportable: "apis/a/deployments/d",
		},
		{
			name:       "projects/another-project/locations/global/artifacts/x",
			projectID:  "another-project",
			exportable: "artifacts/x",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exportable := ExportableName(test.name, test.projectID)
			if exportable != test.exportable {
				t.Errorf("exportable name of %s should be %s but was %s", test.name, test.exportable, exportable)
			}
		})
	}
}

func TestInvalidIDs(t *testing.T) {
	ids := []string{
		"",
		"!!",
		"3d46969a-d232-4fcc-88e3-0aa51c849e4b",
		"012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		"-invalid",
		".invalid",
		"invalid-",
		"invalid.",
	}
	for _, id := range ids {
		if validateID(id) == nil {
			t.Errorf("id %q should be invalid, but validated", id)
		}
	}
}

func TestParseValidNames(t *testing.T) {
	names := []string{
		"projects",
		"projects/-",
		"projects/p",
		"projects/p/locations/global/apis",
		"projects/p/locations/global/apis/-",
		"projects/p/locations/global/apis/a",
		"projects/p/locations/global/apis/a/versions",
		"projects/p/locations/global/apis/a/versions/-",
		"projects/p/locations/global/apis/a/versions/v",
		"projects/p/locations/global/apis/a/versions/v/specs",
		"projects/p/locations/global/apis/a/versions/v/specs/-",
		"projects/p/locations/global/apis/a/versions/v/specs/s",
		"projects/p/locations/global/apis/a/versions/v/specs/s@",
		"projects/p/locations/global/apis/a/versions/v/specs/s@-",
		"projects/p/locations/global/apis/a/versions/v/specs/s@123",
		"projects/p/locations/global/apis/a/deployments",
		"projects/p/locations/global/apis/a/deployments/-",
		"projects/p/locations/global/apis/a/deployments/d",
		"projects/p/locations/global/apis/a/deployments/d@",
		"projects/p/locations/global/apis/a/deployments/d@-",
		"projects/p/locations/global/apis/a/deployments/d@123",
		"projects/p/locations/global/artifacts",
		"projects/p/locations/global/artifacts/-",
		"projects/p/locations/global/artifacts/x",
	}
	for _, name := range names {
		_, err := Parse(name)
		if err != nil {
			t.Errorf("failed to parse name %s", name)
		}
	}
}

func TestParseInvalidNames(t *testing.T) {
	names := []string{
		"invalid",
	}
	for _, name := range names {
		_, err := Parse(name)
		if err == nil {
			t.Errorf("incorrectly parsed invalid name %s", name)
		}
	}
}
