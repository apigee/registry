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

package lint

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/apigee/registry/pkg/application/style"
)

// NewLintFromOpenAPI runs the API linter and returns the results.
func NewLintFromOpenAPI(name string, spec []byte, linter string) (*style.Lint, error) {
	// create a tmp directory
	root, err := os.MkdirTemp("", "registry-openapi-")
	if err != nil {
		return nil, err
	}
	name = filepath.Base(name)
	// whenever we finish, delete the tmp directory
	defer os.RemoveAll(root)
	// write the file to the temp directory
	err = os.WriteFile(filepath.Join(root, name), spec, 0644)
	if err != nil {
		return nil, err
	}
	// run the linter on the spec
	var lintFile *style.LintFile
	switch linter {
	case "":
		err = errors.New("unspecified linter")
	case "gnostic":
		lintFile, err = lintFileForOpenAPIWithGnostic(name, root)
	case "spectral":
		lintFile, err = lintFileForOpenAPIWithSpectral(name, root)
	default:
		err = errors.New("unknown linter: " + linter)
	}
	if err != nil {
		return nil, err
	}
	lint := &style.Lint{
		Name:  name,
		Files: []*style.LintFile{lintFile},
	}
	return lint, nil
}
