// Copyright 2021 Google LLC. All Rights Reserved.
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

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
)

func generateSpec(t *testing.T, specName string) names.Spec {
	t.Helper()
	spec, err := names.ParseSpec(specName)
	if err != nil {
		t.Fatalf("Failed generateSpec(%s): %s", specName, err.Error())
	}
	return spec
}

func generateArtifact(t *testing.T, artifactName string) names.Artifact {
	t.Helper()
	artifact, err := names.ParseArtifact(artifactName)
	if err != nil {
		t.Fatalf("Failed generateSpec(%s): %s", artifactName, err.Error())
	}
	return artifact
}

func TestExtendDependencyPattern(t *testing.T) {
	tests := []struct {
		desc              string
		resourcePattern   string
		dependencyPattern string
		want              string
	}{
		{
			desc:              "artifact reference",
			resourcePattern:   "projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/lint-gnostic",
			dependencyPattern: "$resource.artifact",
			want:              "projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/lint-gnostic",
		},
		{
			desc:              "spec reference",
			resourcePattern:   "projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-",
			dependencyPattern: "$resource.spec",
			want:              "projects/demo/locations/global/apis/-/versions/-/specs/-",
		},
		{
			desc:              "version reference",
			resourcePattern:   "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/-",
			dependencyPattern: "$resource.version/artifacts/lintstats",
			want:              "projects/demo/locations/global/apis/petstore/versions/1.0.0/artifacts/lintstats",
		},
		{
			desc:              "api reference",
			resourcePattern:   "projects/demo/locations/global/apis/-/versions/-/specs/-",
			dependencyPattern: "$resource.version/artifacts/lintstats",
			want:              "projects/demo/locations/global/apis/-/versions/-/artifacts/lintstats",
		},
		{
			desc:              "no reference",
			resourcePattern:   "projects/demo/locations/global/apis/-/artifacts/lintstats",
			dependencyPattern: "apis/-/versions/-",
			want:              "projects/demo/locations/global/apis/-/versions/-",
		},
	}

	const projectID = "demo"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := extendDependencyPattern(test.resourcePattern, test.dependencyPattern, projectID)
			if err != nil {
				t.Errorf("extendDependencyPattern returned unexpected error: %s", err)
			}
			if got != test.want {
				t.Errorf("extendDependencyPattern returned unexpected value want: %q got:%q", test.want, got)
			}
		})
	}
}

func TestExtendDependencyPatternError(t *testing.T) {
	tests := []struct {
		desc              string
		resourcePattern   string
		dependencyPattern string
	}{
		{
			desc:              "non-existent reference",
			resourcePattern:   "projects/demo/locations/global/apis/-/versions/-/specs/-",
			dependencyPattern: "$resource.artifact",
		},
		{
			desc:              "incorrect reference keyword",
			resourcePattern:   "projects/demo/locations/global/apis/-/versions/-/specs/-",
			dependencyPattern: "$resource.aip",
		},
		{
			desc:              "incorrect resourcePattern",
			resourcePattern:   "projects/demo/locations/global/-/versions/-/specs/-",
			dependencyPattern: "$resource.api/artifacts/lintstats",
		},
	}

	const projectID = "demo"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := extendDependencyPattern(test.resourcePattern, test.dependencyPattern, projectID)
			if err == nil {
				t.Errorf("expected extendDependencyPattern to return error, got: %q", got)
			}
		})
	}
}

func TestResourceNameFromEntityKey(t *testing.T) {
	tests := []struct {
		desc            string
		resourcePattern string
		groupKey        string
		want            string
	}{
		{
			desc:            "api pattern",
			resourcePattern: "projects/demo/locations/global/apis/-",
			groupKey:        "projects/demo/locations/global/apis/petstore",
			want:            "projects/demo/locations/global/apis/petstore",
		},
		{
			desc:            "version pattern",
			resourcePattern: "projects/demo/locations/global/apis/-/versions/-",
			groupKey:        "projects/demo/locations/global/apis/petstore/versions/1.0.0",
			want:            "projects/demo/locations/global/apis/petstore/versions/1.0.0",
		},
		{
			desc:            "spec pattern",
			resourcePattern: "projects/demo/locations/global/apis/petstore/versions/-/specs/-/artifacts/complexity",
			groupKey:        "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			want:            "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
		},
		{
			desc:            "artifact pattern",
			resourcePattern: "projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/-",
			groupKey:        "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
			want:            "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := resourceNameFromEntityKey(test.resourcePattern, test.groupKey)
			if err != nil {
				t.Errorf("resourceNameFromEntityKey returned unexpected error: %s", err)
			}
			if got != test.want {
				t.Errorf("resourceNameFromEntityKey returned unexpected value want: %q got:%q", test.want, got)
			}
		})
	}

}

func TestResourceNameFromEntityKeyError(t *testing.T) {
	tests := []struct {
		desc            string
		resourcePattern string
		groupKey        string
	}{
		{
			desc:            "incorrect keywords",
			resourcePattern: "projects/demo/locations/global/apis/-/versions/-/apispecs/-",
			groupKey:        "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		},
		{
			desc:            "incorrect pattern",
			resourcePattern: "projects/demo/locations/global/apis/-/specs/-",
			groupKey:        "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := resourceNameFromEntityKey(test.resourcePattern, test.groupKey)
			if err == nil {
				t.Errorf("expected resourceNameFromEntityKey to return error, got: %q", got)
			}
		})
	}

}

func TestGetEntityKey(t *testing.T) {
	tests := []struct {
		desc     string
		pattern  string
		resource ResourceInstance
		want     string
	}{
		{
			desc:    "api group",
			pattern: "$resource.api/versions/-/specs/-",
			resource: SpecResource{
				SpecName: SpecName{Spec: generateSpec(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml")},
			},
			want: "projects/demo/locations/global/apis/petstore",
		},
		{
			desc:    "version group",
			pattern: "$resource.version/specs/-",
			resource: SpecResource{
				SpecName: SpecName{Spec: generateSpec(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml")},
			},
			want: "projects/demo/locations/global/apis/petstore/versions/1.0.0",
		},
		{
			desc:    "spec group",
			pattern: "$resource.spec",
			resource: SpecResource{
				SpecName: SpecName{Spec: generateSpec(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml")},
			},
			want: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		},
		{
			desc:    "artifact group",
			pattern: "$resource.artifact",
			resource: ArtifactResource{
				ArtifactName: ArtifactName{Artifact: generateArtifact(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")},
			},
			want: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
		},
		{
			desc:    "no group",
			pattern: "apis/-/versions/-/specs/-",
			resource: ArtifactResource{
				ArtifactName: ArtifactName{Artifact: generateArtifact(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")},
			},
			want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := getEntityKey(test.pattern, test.resource)
			if err != nil {
				t.Errorf("getEntityKey returned unexpected error: %s", err)
			}
			if got != test.want {
				t.Errorf("getEntityKey returned unexpected value want: %q got:%q", test.want, got)
			}
		})
	}
}

func TestGetEntityKeyError(t *testing.T) {
	tests := []struct {
		desc     string
		pattern  string
		resource ResourceInstance
	}{
		{
			desc:    "typo",
			pattern: "$resource.apis/versions/-/specs/-",
			resource: SpecResource{
				SpecName: SpecName{Spec: generateSpec(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml")},
			},
		},
		{
			desc:    "incorrect reference",
			pattern: "$resource.name/versions/-/specs/-",
			resource: SpecResource{
				SpecName: SpecName{Spec: generateSpec(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml")},
			},
		},
		{
			desc:    "incorrect resourceKW",
			pattern: "$resources.api/versions/-/specs/-",
			resource: SpecResource{
				SpecName: SpecName{Spec: generateSpec(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := getEntityKey(test.pattern, test.resource)
			if err == nil {
				t.Errorf("expected getEntityKey to return error, got: %q", got)
			}
		})
	}
}

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
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
			want:         "compute lint projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml --linter=gnostic",
		},
		{
			desc:         "multiple args",
			action:       "compute score $resource.spec/artifacts/complexity $resource.spec/artifacts/vocabulary",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score",
			want: fmt.Sprintf("compute score %s %s",
				"projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
				"projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/vocabulary"),
		},
		{
			desc:         "extended reference",
			action:       "compute score $resource.spec/artifacts/complexity",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			want:         "compute score projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
		},
		{
			desc:         "multiple reference entities",
			action:       "compute compliance $resource.spec --policy=$resource.api/artifacts/design-policy",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			want:         "compute compliance projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml --policy=projects/demo/locations/global/apis/petstore/artifacts/design-policy",
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
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		},
		{
			desc:         "incorrect format",
			action:       "compute lintstats $resourceversion",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		},
		{
			desc:         "invalid reference",
			action:       "compute lintstats $resource.artifact",
			resourceName: "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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

func TestValidateResourceEntry(t *testing.T) {
	tests := []struct {
		desc              string
		generatedResource *rpc.GeneratedResource
	}{
		{
			desc: "single entity reference",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/artifacts/complexity",
				Dependencies: []*rpc.Dependency{
					&rpc.Dependency{
						Pattern: "$resource.spec",
					},
				},
				Action: "registry compute complexity $resource.spec",
			},
		},
		{
			desc: "no entity reference",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "artifacts/summary",
				Dependencies: []*rpc.Dependency{
					&rpc.Dependency{
						Pattern: "apis/-/versions/-/specs/-",
					},
				},
				Action: "registry compute summary apis/-/versions/-/specs/-",
			},
		},
		{
			desc: "multiple entity references",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/artifacts/conformance-apihub-styleguide",
				Dependencies: []*rpc.Dependency{
					&rpc.Dependency{
						Pattern: "$resource.spec",
					},
					&rpc.Dependency{
						Pattern: "artifacts/apihub-styleguide",
					},
				},
				Action: "registry compute conformance $resource.spec",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := ValidateResourceEntry(test.generatedResource)
			if err != nil {
				t.Errorf("ValidateResourceEntry() generated unexpected error: %s", err)
			}
		})
	}
}

func TestValidateResourceEntryError(t *testing.T) {
	tests := []struct {
		desc              string
		generatedResource *rpc.GeneratedResource
	}{
		{
			desc: "invalid entity reference",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/version-summary",
				Dependencies: []*rpc.Dependency{
					&rpc.Dependency{
						Pattern: "$resource.version",
					},
					&rpc.Dependency{
						Pattern: "$resource.spec", // Correct pattern: $resource.version/specs/-
					},
				},
				Action: "registry generate summary $resource.version",
			},
		},
		{
			desc: "lower entity in action",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/version-summary",
				Dependencies: []*rpc.Dependency{
					&rpc.Dependency{
						Pattern: "$resource.api/versions/-", // Correct pattern: $resource.version
					},
					&rpc.Dependency{
						Pattern: "$resource.api/artifacts/prod-version-metadata",
					},
					&rpc.Dependency{
						Pattern: "artifacts/summary-config",
					},
				},
				Action: "registry generate summary $resource.version",
			},
		},
		{
			desc: "lower enity in deps",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/artifacts/conformance-apihub-styleguide",
				Dependencies: []*rpc.Dependency{
					&rpc.Dependency{
						Pattern: "$resource.spec",
					},
					&rpc.Dependency{
						Pattern: "$resource.artifact", // Invalid dependency
					},
					&rpc.Dependency{
						Pattern: "artifacts/apihub-styleguide",
					},
				},
				Action: "registry compute conformance $resource.spec",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := ValidateResourceEntry(test.generatedResource)
			if err == nil {
				t.Errorf("Expected ValidateResourceEntry() to fail, unexpected success")
			}
		})
	}
}
