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
	"regexp"
	"strings"
	"testing"
)

func TestResourceNames(t *testing.T) {
	groups := []struct {
		name   string
		regexp *regexp.Regexp
		pass   []string
		fail   []string
	}{
		{
			name:   "projects",
			regexp: ProjectsRegexp(),
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
			name:   "project",
			regexp: ProjectRegexp(),
			pass: []string{
				"projects/google",
				"projects/-",
			},
			fail: []string{
				"-",
			},
		},
		{
			name:   "apis",
			regexp: ApisRegexp(),
			pass: []string{
				"projects/google/locations/LOCATION/apis",
				"projects/-/locations/LOCATION/apis",
			},
			fail: []string{
				"-",
			},
		},
		{
			name:   "api",
			regexp: ApiRegexp(),
			pass: []string{
				"projects/google/locations/LOCATION/apis/sample",
				"projects/-/locations/LOCATION/apis/-",
				"projects/123/locations/LOCATION/apis/abc",
				"projects/1-2-3/locations/LOCATION/apis/abc",
			},
			fail: []string{
				"-",
				"invalid",
				"projects//locations/LOCATION/apis/123",
				"projects/123/locations/LOCATION/apis/",
				"projects/123/locations/LOCATION/invalid/123",
				"projects/123/locations/LOCATION/apis/ 123",
			},
		},
		{
			name:   "versions",
			regexp: VersionsRegexp(),
			pass: []string{
				"projects/google/locations/LOCATION/apis/sample/versions",
				"projects/-/locations/LOCATION/apis/-/versions",
			},
			fail: []string{
				"-",
			},
		},
		{
			name:   "version",
			regexp: VersionRegexp(),
			pass: []string{
				"projects/google/locations/LOCATION/apis/sample/versions/v1",
				"projects/-/locations/LOCATION/apis/-/versions/-",
				"projects/123/locations/LOCATION/apis/abc/versions/123",
				"projects/1-2-3/locations/LOCATION/apis/abc/versions/123",
			},
			fail: []string{
				"-",
				"invalid",
				"projects//locations/LOCATION/apis/123",
				"projects/123/locations/LOCATION/apis/",
				"projects/123/locations/LOCATION/invalid/123",
				"projects/123/locations/LOCATION/apis/ 123",
			},
		},
		{
			name:   "specs",
			regexp: SpecsRegexp(),
			pass: []string{
				"projects/google/locations/LOCATION/apis/sample/versions/v1/specs",
				"projects/-/locations/LOCATION/apis/-/versions/-/specs",
			},
			fail: []string{
				"-",
			},
		},
		{
			name:   "spec",
			regexp: SpecRegexp(),
			pass: []string{
				"projects/google/locations/LOCATION/apis/sample/versions/v1/specs/openapi.yaml@1234567890abcdef",
				"projects/google/locations/LOCATION/apis/sample/versions/v1/specs/openapi.yaml",
				"projects/-/locations/LOCATION/apis/-/versions/-/specs/-",
				"projects/123/locations/LOCATION/apis/abc/versions/123/specs/abc",
				"projects/1-2-3/locations/LOCATION/apis/abc/versions/123/specs/abc",
			},
			fail: []string{
				"-",
				"invalid",
				"projects//locations/LOCATION/apis/123",
				"projects/123/locations/LOCATION/apis/",
				"projects/123/locations/LOCATION/invalid/123",
				"projects/123/locations/LOCATION/apis/ 123",
			},
		},
		{
			name:   "artifacts",
			regexp: ArtifactsRegexp(),
			pass: []string{
				"projects/google/locations/LOCATION/apis/sample/versions/v1/specs/openapi.yaml/artifacts",
				"projects/google/locations/LOCATION/apis/sample/versions/v1/artifacts",
				"projects/google/locations/LOCATION/apis/sample/artifacts",
				"projects/google/locations/LOCATION/artifacts",
			},
			fail: []string{
				"-",
			},
		},
		{
			name:   "artifact",
			regexp: ArtifactRegexp(),
			pass: []string{
				"projects/google/locations/LOCATION/apis/sample/versions/v1/specs/openapi.yaml/artifacts/test-artifact",
				"projects/google/locations/LOCATION/apis/sample/versions/v1/artifacts/test-artifact",
				"projects/google/locations/LOCATION/apis/sample/artifacts/test-artifact",
				"projects/google/locations/LOCATION/artifacts/test-artifact",
			},
			fail: []string{
				"-",
			},
		},
	}
	for _, g := range groups {
		for _, path := range g.pass {
			path = strings.ReplaceAll(path, "LOCATION", Location)
			m := g.regexp.FindStringSubmatch(path)
			if m == nil {
				t.Errorf("failed to match %s: %s", g.name, path)
			}
		}
		for _, path := range g.fail {
			path = strings.ReplaceAll(path, "LOCATION", Location)
			m := g.regexp.FindStringSubmatch(path)
			if m != nil {
				t.Errorf("false match %s: %s", g.name, path)
			}
		}
	}
}
