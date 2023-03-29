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

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/apigee/registry/cmd/registry/plugins/linter"
	"github.com/apigee/registry/pkg/application/style"
)

type testLinterRunner struct{}

func (*testLinterRunner) Run(req *style.LinterRequest) (*style.LinterResponse, error) {
	lintFiles := make([]*style.LintFile, 0)
	err := filepath.WalkDir(req.SpecDirectory,
		func(p string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			} else if entry.IsDir() {
				return nil // Do nothing for the directory, but still walk its contents.
			}
			bytes, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			lines := strings.Split(string(bytes), "\n")
			filePath := strings.TrimPrefix(p, req.SpecDirectory+"/")
			lintFile := &style.LintFile{
				FilePath: filePath,
				Problems: []*style.LintProblem{{
					RuleId:     "size",
					Message:    fmt.Sprintf("%d", len(bytes)),
					RuleDocUri: "https://github.com/apigee/registry",
					Suggestion: fmt.Sprintf("This is the size of %s.", filePath),
					Location: &style.LintLocation{
						StartPosition: &style.LintPosition{
							LineNumber:   1,
							ColumnNumber: 1,
						},
						EndPosition: &style.LintPosition{
							LineNumber:   int32(len(lines) - 1),
							ColumnNumber: int32(len(lines[len(lines)-1]) + 1),
						},
					},
				}},
			}
			lintFiles = append(lintFiles, lintFile)
			return nil
		})
	if err != nil {
		return nil, fs.ErrClosed
	}
	return &style.LinterResponse{
		Lint: &style.Lint{
			Name:  "registry-lint-test",
			Files: lintFiles,
		},
	}, nil
}

func main() {
	linter.Main(&testLinterRunner{})
}
