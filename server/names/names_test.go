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
	"testing"
)

type group struct {
	name   string
	regexp *regexp.Regexp
	pass   []string
	fail   []string
}

func TestResourceNames(t *testing.T) {
	{
		groups := []group{
			group{
				name:   "projects",
				regexp: ProjectsRegexp(),
				pass: []string{
					"projects",
				},
				fail: []string{
					"-",
					"project",
					"organizations",
					"products",
				},
			},
			group{
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
			group{
				name:   "products",
				regexp: ProductsRegexp(),
				pass: []string{
					"projects/google/products",
					"projects/-/products",
				},
				fail: []string{
					"-",
				},
			},
			group{
				name:   "product",
				regexp: ProductRegexp(),
				pass: []string{
					"projects/google/products/sample",
					"projects/-/products/-",
					"projects/123/products/abc",
					"projects/1-2_3/products/abc",
				},
				fail: []string{
					"-",
					"invalid",
					"projects//products/123",
					"projects/123/products/",
					"projects/123/invalid/123",
					"projects/123/products/ 123",
				},
			},
			group{
				name:   "versions",
				regexp: VersionsRegexp(),
				pass: []string{
					"projects/google/products/sample/versions",
					"projects/-/products/-/versions",
				},
				fail: []string{
					"-",
				},
			},
			group{
				name:   "version",
				regexp: VersionRegexp(),
				pass: []string{
					"projects/google/products/sample/versions/v1",
					"projects/-/products/-/versions/-",
					"projects/123/products/abc/versions/123",
					"projects/1-2_3/products/abc/versions/123",
				},
				fail: []string{
					"-",
					"invalid",
					"projects//products/123",
					"projects/123/products/",
					"projects/123/invalid/123",
					"projects/123/products/ 123",
				},
			},
			group{
				name:   "specs",
				regexp: SpecsRegexp(),
				pass: []string{
					"projects/google/products/sample/versions/v1/specs",
					"projects/-/products/-/versions/-/specs",
				},
				fail: []string{
					"-",
				},
			},
			group{
				name:   "spec",
				regexp: SpecRegexp(),
				pass: []string{
					"projects/google/products/sample/versions/v1/specs/openapi.yaml@1234567890ABCDEFabcdef",
					"projects/google/products/sample/versions/v1/specs/openapi.yaml",
					"projects/-/products/-/versions/-/specs/-",
					"projects/123/products/abc/versions/123/specs/abc",
					"projects/1-2_3/products/abc/versions/123/specs/abc",
				},
				fail: []string{
					"-",
					"invalid",
					"projects//products/123",
					"projects/123/products/",
					"projects/123/invalid/123",
					"projects/123/products/ 123",
				},
			},
			group{
				name:   "properties",
				regexp: PropertiesRegexp(),
				pass: []string{
					"projects/google/products/sample/versions/v1/specs/openapi.yaml/properties",
					"projects/google/products/sample/versions/v1/properties",
					"projects/google/products/sample/properties",
					"projects/google/properties",
				},
				fail: []string{
					"-",
				},
			},
			group{
				name:   "property",
				regexp: PropertyRegexp(),
				pass: []string{
					"projects/google/products/sample/versions/v1/specs/openapi.yaml/properties/test-property",
					"projects/google/products/sample/versions/v1/properties/test-property",
					"projects/google/products/sample/properties/test-property",
					"projects/google/properties/test-property",
				},
				fail: []string{
					"-",
				},
			},
			group{
				name:   "labels",
				regexp: LabelsRegexp(),
				pass: []string{
					"projects/google/products/sample/versions/v1/specs/openapi.yaml/labels",
					"projects/google/products/sample/versions/v1/labels",
					"projects/google/products/sample/labels",
					"projects/google/labels",
				},
				fail: []string{
					"-",
				},
			},
			group{
				name:   "label",
				regexp: LabelRegexp(),
				pass: []string{
					"projects/google/products/sample/versions/v1/specs/openapi.yaml/labels/test-label",
					"projects/google/products/sample/versions/v1/labels/test-label",
					"projects/google/products/sample/labels/test-label",
					"projects/google/labels/test-label",
				},
				fail: []string{
					"-",
				},
			},
		}
		for _, g := range groups {
			for _, path := range g.pass {
				m := g.regexp.FindAllStringSubmatch(path, -1)
				if m == nil {
					t.Logf("failed to match %s: %s", g.name, path)
					t.Fail()
				}
			}
			for _, path := range g.fail {
				m := g.regexp.FindAllStringSubmatch(path, -1)
				if m != nil {
					t.Logf("false match %s: %s", g.name, path)
					t.Fail()
				}
			}
		}
	}
}
