// Copyright 2023 Google LLC
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

package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSpectralFilenameWithGlobs(t *testing.T) {
	path, err := exec.LookPath("spectral")
	if path == "" || err != nil {
		t.Skip("spectral not present, skipping check")
	}

	tests := []struct {
		testName string
		fileName string
	}{
		{"noglobs", "name.yaml"},
		{"replacement", "name (1|2).yaml"},
		{"sequence", "name [[:digit:]].yaml"},
		{"expansion", "name {1..3}.yaml"},
		{"dollar", "name$.yaml"},
		{"question", "name?.yaml"},
		{"asterik", "name*.yaml"},
		{"negation", "!name.yaml"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			root, err := os.MkdirTemp("", "registry-openapi-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(root)
			if err := os.WriteFile(filepath.Join(root, ".spectral.yaml"), []byte(`extends: ["spectral:oas", "spectral:asyncapi"]`), 0644); err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(filepath.Join(root, test.fileName), []byte("not openapi"), 0644); err != nil {
				t.Fatal(err)
			}

			lint, err := lintFileForOpenAPIWithSpectral(test.fileName, root)
			if err != nil {
				t.Fatal(err)
			}
			if len(lint.Problems) != 1 {
				t.Error("should've gotten an error")
			}
		})
	}
}
