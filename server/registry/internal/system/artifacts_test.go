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

package system

import (
	"testing"

	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestReservedIDs(t *testing.T) {
	tests := []struct {
		name     string
		reserved bool
	}{
		{
			name:     "projects/my-project/locations/global/artifacts/system-buildinfo",
			reserved: true,
		},
		{
			name:     "projects/my-project/locations/global/artifacts/system-other",
			reserved: true,
		},
		{
			name:     "projects/my-project/locations/global/artifacts/apihub-manifest",
			reserved: false,
		},
		{
			name:     "projects/my-project/locations/global/artifacts/systemstuff",
			reserved: false,
		},
	}
	for _, test := range tests {
		name, err := names.ParseArtifact(test.name)
		if err != nil {
			t.Fatalf("Failed to parse artifact name %s", test.name)
		}
		r := Reserved(name)
		if r != test.reserved {
			t.Errorf("Unexpected value for Reserved(%s): %t", test.name, r)
		}
	}
}

func TestMissingArtifacts(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "projects/my-project/locations/global/artifacts/system-undefined",
		},
		{
			name: "projects/my-project/locations/global/artifacts/undefined",
		},
	}
	for _, test := range tests {
		name, err := names.ParseArtifact(test.name)
		if err != nil {
			t.Fatalf("Failed to parse artifact name %s", test.name)
		}
		_, err = Artifact(name)
		if err == nil {
			t.Errorf("Creating artifact %q succeeded and should have failed", test.name)
		} else if status.Code(err) != codes.NotFound {
			t.Errorf("Creating artifact %q returned unexpected error code: %s", test.name, err)
		}
		_, err = ArtifactContents(name)
		if err == nil {
			t.Errorf("Creating contents of artifact %q succeeded and should have failed", test.name)
		} else if status.Code(err) != codes.NotFound {
			t.Errorf("Creating contents of artifact %q returned unexpected error code: %s", test.name, err)
		}
	}
}

func TestBuildInfo(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
	}{
		{
			name:     "projects/my-project/locations/global/artifacts/system-buildinfo",
			mimeType: "application/yaml;type=BuildInfo",
		},
	}
	for _, test := range tests {
		name, err := names.ParseArtifact(test.name)
		if err != nil {
			t.Fatalf("Failed to parse artifact name %s", test.name)
		}
		artifact, err := Artifact(name)
		if err != nil {
			t.Fatalf("Failed to build artifact %s", name)
		}
		if artifact.GetMimeType() != test.mimeType {
			t.Errorf("Incorrect mime type for %s: %q expected %q", name, artifact.GetMimeType(), test.mimeType)
		}
		contents, err := ArtifactContents(name)
		if err != nil {
			t.Fatalf("Failed to build artifact contents %s", name)
		}
		if contents.GetContentType() != test.mimeType {
			t.Errorf("Incorrect mime type for contents of %s: %q expected %q", name, contents.GetContentType(), test.mimeType)
		}
		if len(contents.GetData()) != int(artifact.GetSizeBytes()) {
			t.Errorf("Inconsistent content size for artifact %s", name)
		}
	}
}
