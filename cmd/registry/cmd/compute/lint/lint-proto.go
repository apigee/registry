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

package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/pkg/application/style"
	"google.golang.org/protobuf/encoding/protojson"
)

// NewLintFromZippedProtos runs the API linter and returns the results.
func NewLintFromZippedProtos(name string, b []byte) (*style.Lint, error) {
	// create a tmp directory
	root, err := os.MkdirTemp("", "registry-protos-")
	if err != nil {
		return nil, err
	}
	// whenever we finish, delete the tmp directory
	defer os.RemoveAll(root)
	// unzip the protos to the temp directory
	_, err = compress.UnzipArchiveToPath(b, root+"/protos")
	if err != nil {
		return nil, err
	}
	// unpack api-common-protos in the temp directory
	cmd := exec.Command("git", "clone", "https://github.com/googleapis/api-common-protos")
	cmd.Dir = root
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	// run the api-linter on each proto file in the archive
	lint, err := lintDirectory(name, root)
	if err == nil {
		return lint, nil
	}
	// if we had errors, add googleapis to the temp directory and retry
	cmd = exec.Command("git", "clone", "https://github.com/googleapis/googleapis")
	cmd.Dir = root
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	// rerun the api-linter with the extra googleapis protos
	return lintDirectory(name, root)
}

func lintDirectory(name string, root string) (*style.Lint, error) {
	lint := &style.Lint{}
	lint.Name = name
	// run the api-linter on each proto file
	err := filepath.Walk(root+"/protos",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				lintFile, err := lintFileForProto(path, root)
				if err != nil {
					return err
				}
				lint.Files = append(lint.Files, lintFile)
			}
			return nil
		})
	return lint, err
}

func lintFileForProto(path string, root string) (*style.LintFile, error) {
	filename := strings.TrimPrefix(path, root+"/protos/")
	cmd := exec.Command("api-linter", filename, "-I", "protos", "-I", "api-common-protos", "-I", "googleapis", "--output-format", "json")
	cmd.Dir = root
	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var result style.Lint
	// The API linter returns a JSON array. Since the proto parser requires a top-level struct,
	// wrap the results so that they are in the form of an style.Lint JSON serialization.
	wrappedJSON := "{\"files\": " + string(data) + "}"
	err = protojson.Unmarshal([]byte(wrappedJSON), &result)
	if err != nil {
		return nil, err
	}
	if len(result.Files) > 0 {
		return result.Files[0], err
	}
	return nil, err
}
