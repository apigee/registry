// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package patterns

import (
	"testing"

	"github.com/apigee/registry/pkg/names"
)

func generateSpecName(t *testing.T, specName string) SpecName {
	t.Helper()
	rev, err := names.ParseSpecRevision(specName)
	if err != nil {
		t.Fatalf("Failed generateSpec(%s): %s", specName, err.Error())
	}
	return SpecName{
		Name:       rev.Spec(),
		RevisionID: rev.RevisionID,
	}
}

func generateVersion(t *testing.T, versionName string) names.Version {
	t.Helper()
	version, err := names.ParseVersion(versionName)
	if err != nil {
		t.Fatalf("Failed generateVersion(%s): %s", versionName, err.Error())
	}
	return version
}

func generateArtifact(t *testing.T, artifactName string) names.Artifact {
	t.Helper()
	artifact, err := names.ParseArtifact(artifactName)
	if err != nil {
		t.Fatalf("Failed generateArtifact(%s): %s", artifactName, err.Error())
	}
	return artifact
}

func TestSubstituteReferenceEntity(t *testing.T) {
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
			desc:              "spec revision reference",
			resourcePattern:   "projects/demo/locations/global/apis/-/versions/-/specs/-@-/artifacts/-",
			dependencyPattern: "$resource.spec",
			want:              "projects/demo/locations/global/apis/-/versions/-/specs/-@-",
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

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			resourceName, err := ParseResourcePattern(test.resourcePattern)
			if err != nil {
				t.Fatalf("Error in parsing, %s", err)
			}
			got, err := SubstituteReferenceEntity(test.dependencyPattern, resourceName)
			if err != nil {
				t.Errorf("SubstituteReferenceEntity returned unexpected error: %s", err)
			}
			if got.String() != test.want {
				t.Errorf("SubstituteReferenceEntity returned unexpected value want: %q got:%q", test.want, got)
			}
		})
	}
}

func TestSubstituteReferenceEntityError(t *testing.T) {
	tests := []struct {
		desc              string
		resourcePattern   string
		dependencyPattern string
	}{
		{
			desc:              "nonexistent reference",
			resourcePattern:   "projects/demo/locations/global/apis/-/versions/-/specs/-",
			dependencyPattern: "$resource.artifact",
		},
		{
			desc:              "nonexistent specrev reference",
			resourcePattern:   "projects/demo/locations/global/apis/-/versions/-/specs/-@-",
			dependencyPattern: "$resource.artifact",
		},
		{
			desc:              "incorrect reference keyword",
			resourcePattern:   "projects/demo/locations/global/apis/-/versions/-/specs/-",
			dependencyPattern: "$resource.aip",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			resourceName, err := ParseResourcePattern(test.resourcePattern)
			if err != nil {
				t.Fatalf("Error in parsing, %s", err)
			}
			got, err := SubstituteReferenceEntity(test.dependencyPattern, resourceName)
			if err == nil {
				t.Errorf("expected SubstituteReferenceEntity to return error, got: %q", got.String())
			}
		})
	}
}

func TestFullResourceNameFromParent(t *testing.T) {
	tests := []struct {
		desc            string
		resourcePattern string
		parent          string
		want            ResourceName
	}{
		{
			desc:            "version pattern",
			resourcePattern: "projects/demo/locations/global/apis/-/versions/1.0.0",
			parent:          "projects/demo/locations/global/apis/petstore",
			want: VersionName{
				Name: generateVersion(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0"),
			},
		},
		{
			desc:            "spec pattern",
			resourcePattern: "projects/demo/locations/global/apis/-/versions/-/specs/openapi",
			parent:          "projects/demo/locations/global/apis/petstore/versions/1.0.0",
			want:            generateSpecName(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi"),
		},
		{
			desc:            "specrev pattern",
			resourcePattern: "projects/demo/locations/global/apis/-/versions/-/specs/openapi@rev",
			parent:          "projects/demo/locations/global/apis/petstore/versions/1.0.0",
			want:            generateSpecName(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi@rev"),
		},
		{
			desc:            "artifact pattern",
			resourcePattern: "projects/demo/locations/global/apis/-/versions/-/specs/-/artifacts/complexity",
			parent:          "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			want: ArtifactName{
				Name: generateArtifact(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := FullResourceNameFromParent(test.resourcePattern, test.parent)
			if err != nil {
				t.Errorf("FullResourceNameFromParent returned unexpected error: %s", err)
			}
			if got != test.want {
				t.Errorf("FullResourceNameFromParent returned unexpected value want: %q got:%q", test.want, got)
			}
		})
	}
}

func TestFullResourceNameFromParentError(t *testing.T) {
	tests := []struct {
		desc            string
		resourcePattern string
		parent          string
	}{
		{
			desc:            "incorrect keywords",
			resourcePattern: "projects/demo/locations/global/apis/-/versions/-/apispecs/-",
			parent:          "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
		{
			desc:            "incorrect pattern",
			resourcePattern: "projects/demo/locations/global/apis/-/specs/-",
			parent:          "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := FullResourceNameFromParent(test.resourcePattern, test.parent)
			if err == nil {
				t.Errorf("expected FullResourceNameFromParent to return error, got: %q", got)
			}
		})
	}
}

func TestGetReferenceEntityValue(t *testing.T) {
	tests := []struct {
		desc            string
		resourcePattern string
		referred        ResourceName
		want            string
	}{
		{
			desc:            "api group",
			resourcePattern: "$resource.api/versions/-/specs/-",
			referred:        generateSpecName(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi"),
			want:            "projects/demo/locations/global/apis/petstore",
		},
		{
			desc:            "version group",
			resourcePattern: "$resource.version/specs/-",
			referred:        generateSpecName(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi"),
			want:            "projects/demo/locations/global/apis/petstore/versions/1.0.0",
		},
		{
			desc:            "spec group",
			resourcePattern: "$resource.spec",
			referred:        generateSpecName(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi"),
			want:            "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
		{
			desc:            "spec revision group",
			resourcePattern: "$resource.spec",
			referred:        generateSpecName(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi@rev"),
			want:            "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi@rev",
		},
		{
			desc:            "artifact group",
			resourcePattern: "$resource.artifact",
			referred:        ArtifactName{Name: generateArtifact(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic")},
			want:            "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
		},
		{
			desc:            "no group",
			resourcePattern: "apis/-/versions/-/specs/-",
			referred:        ArtifactName{Name: generateArtifact(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic")},
			want:            "default",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := GetReferenceEntityValue(test.resourcePattern, test.referred)
			if err != nil {
				t.Errorf("GetReferenceEntityValue returned unexpected error: %s", err)
			}
			if got != test.want {
				t.Errorf("GetReferenceEntityValue returned unexpected value want: %q got:%q", test.want, got)
			}
		})
	}
}

func TestGetReferenceEntityValueError(t *testing.T) {
	tests := []struct {
		desc            string
		resourcePattern string
		referred        ResourceName
	}{
		{
			desc:            "typo",
			resourcePattern: "$resource.apis/versions/-/specs/-",
			referred:        generateSpecName(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi"),
		},
		{
			desc:            "incorrect reference",
			resourcePattern: "$resource.name/versions/-/specs/-",
			referred:        generateSpecName(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi"),
		},
		{
			desc:            "incorrect resourceKW",
			resourcePattern: "$resources.api/versions/-/specs/-",
			referred:        generateSpecName(t, "projects/demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi"),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := GetReferenceEntityValue(test.resourcePattern, test.referred)
			if err == nil {
				t.Errorf("expected GetReferenceEntityValue to return error, got: %q", got)
			}
		})
	}
}
