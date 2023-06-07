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

package controller

import (
	"fmt"
	"testing"

	"github.com/apigee/registry/pkg/application/controller"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestGenerateCommand(t *testing.T) {
	tests := []struct {
		desc         string
		action       string
		resourceName string
		want         string
	}{
		{
			desc:         "api reference",
			action:       "compute lintstats $resource.api --linter=gnostic",
			resourceName: "projects/demo/locations/global/apis/petstore/artifacts/lintstats-gnostic",
			want:         "compute lintstats projects/demo/locations/global/apis/petstore --linter=gnostic",
		},
		{
			desc:         "version reference",
			action:       "compute lintstats $resource.version --linter=gnostic",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/artifacts/lintstats-gnostic",
			want:         "compute lintstats projects/demo/locations/global/apis/petstore/versions/1.0.0 --linter=gnostic",
		},
		{
			desc:         "spec reference",
			action:       "compute lint $resource.spec --linter=gnostic",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
			want:         "compute lint projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi --linter=gnostic",
		},
		{
			desc:         "multiple args",
			action:       "compute score $resource.spec/artifacts/complexity $resource.spec/artifacts/vocabulary",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score",
			want: fmt.Sprintf("compute score %s %s",
				"projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity",
				"projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/vocabulary"),
		},
		{
			desc:         "extended reference",
			action:       "compute score $resource.spec/artifacts/complexity",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			want:         "compute score projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := generateCommand(test.action, test.resourceName)
			if err != nil {
				t.Errorf("generateCommand returned unexpected error: %s", err)
			}
			if got != test.want {
				t.Errorf("generateCommand returned unexpected value want: %q got:%q", test.want, got)
			}
		})
	}
}

func TestGenerateCommandError(t *testing.T) {
	tests := []struct {
		desc         string
		action       string
		resourceName string
	}{
		{
			desc:         "incorrect reference",
			action:       "compute lintstats $resource.apispec",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
		{
			desc:         "incorrect format",
			action:       "compute lintstats $resourceversion",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
		{
			desc:         "invalid reference",
			action:       "compute lintstats $resource.artifact",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := generateCommand(test.action, test.resourceName)
			if err == nil {
				t.Errorf("expected generateCommand to return error, got: %q", got)
			}
		})
	}
}

func TestValidateGeneratedResourceEntry(t *testing.T) {
	tests := []struct {
		desc              string
		generatedResource *controller.GeneratedResource
	}{
		{
			desc: "single entity reference",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/artifacts/complexity",
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.spec",
					},
				},
				Action: "registry compute complexity $resource.spec",
			},
		},
		{
			desc: "multiple entity reference",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/compliance",
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.version",
					},
					{
						Pattern: "$resource.api/artifacts/recommended-version",
					},
				},
				Action: "registry compute compliance $resource.version $resource.api/artifacts/recommended-version",
			},
		},
		{
			desc: "present/absent entity reference",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/artifacts/conformance",
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.spec",
					},
					{
						Pattern: "artifacts/registry-styleguide",
					},
				},
				Action: "registry compute conformance $resource.spec",
			},
		},
		{
			desc: "present/absent entity and multiple reference",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/version-summary",
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.api/versions/-",
					},
					{
						Pattern: "$resource.api/artifacts/prod-version-metadata",
					},
					{
						Pattern: "artifacts/summary-config",
					},
				},
				Action: "registry generate summary $resource.version",
			},
		},
		{
			desc: "refresh field with missing dependencies",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/score-receipt",
				Receipt: true,
				Refresh: &durationpb.Duration{
					Seconds: 5,
				},
				Action: "registry generate summary $resource.version",
			},
		},
		{
			desc: "refresh field (in nanoseconds)",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/score-receipt",
				Receipt: true,
				Refresh: &durationpb.Duration{
					Nanos: 5,
				},
				Action: "registry generate summary $resource.version",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := validateGeneratedResourceEntry("projects/demo/locations/global", test.generatedResource)
			if len(gotErrs) > 0 {
				t.Errorf("ValidateGeneratedResourceEntry() returned unexpected errors: %s", gotErrs)
			}
		})
	}
}

func TestValidateGeneratedResourceEntryError(t *testing.T) {
	tests := []struct {
		desc              string
		generatedResource *controller.GeneratedResource
	}{
		{
			desc: "invalid target pattern",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/complexity", // Correct pattern: apis/-/versions/-/specs/-/artifacts/complexity
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.spec",
					},
				},
				Action: "registry compute complexity $resource.spec",
			},
		},
		{
			desc: "no target resource name",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-", // Correct pattern: apis/-/versions/-/specs/openapi
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.spec",
						Filter:  "mime_type.contains('openapi')",
					},
				},
				Action: "registry generate openapispec $resource.spec",
			},
		},
		{
			desc: "invalid reference in deps",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/version-summary",
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.version",
					},
					{
						Pattern: "$resource.spec", // Correct pattern: $resource.version/specs/-
					},
				},
				Action: "registry compute conformance $resource.spec",
			},
		},
		{
			desc: "invalid reference in action",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/version-summary",
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.api/versions/-",
					},
					{
						Pattern: "$resource.api/artifacts/prod-version-metadata",
					},
					{
						Pattern: "artifacts/summary-config",
					},
				},
				Action: "registry generate summary $resource.spec", // Correct pattern: registry generate summary $resource.version
			},
		},
		{
			desc: "refresh field equal to 0",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/score-receipt",
				Receipt: true,
				Refresh: &durationpb.Duration{
					Seconds: 0,
				},
				Action: "registry generate summary $resource.version",
			},
		},
		{
			desc: "missing dependencies and refresh",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/score-receipt",
				Receipt: true,
				Action:  "registry generate summary $resource.version",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErrs := validateGeneratedResourceEntry("projects/demo/locations/global", test.generatedResource)
			if len(gotErrs) == 0 {
				t.Errorf("Expected ValidateGeneratedResourceEntry() to return errors")
			}
		})
	}
}
