// Copyright 2021 Google LLC
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
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	lint "github.com/apigee/registry/cmd/registry/plugins/linter"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/log"
)

// Runs the API linter with a provided spec path
type runLinter func(specPath, specDir string) ([]*style.LintProblem, error)

// apiLinterRunner implements the LinterRunner interface for the API linter.
type apiLinterRunner struct{}

func (linter *apiLinterRunner) Run(req *style.LinterRequest) (*style.LinterResponse, error) {
	return linter.RunImpl(req, runApiLinter)
}

func (linter *apiLinterRunner) RunImpl(
	req *style.LinterRequest,
	runLinter runLinter,
) (*style.LinterResponse, error) {
	lintFiles := make([]*style.LintFile, 0)

	//log.Infof(context.TODO(), "REQUEST %+v", req)
	// Traverse the files in the directory
	err := filepath.Walk(req.GetSpecDirectory(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".proto") {
			// Currently, only proto files are supported by API Linter.
			return nil
		}

		protoPath := strings.TrimPrefix(path, req.GetSpecDirectory()+"/")
		if strings.HasPrefix(protoPath, "google/api/") ||
			strings.HasPrefix(protoPath, "google/longrunning/") ||
			strings.HasPrefix(protoPath, "google/rpc/") {
			// Skip common includes.
			return nil
		}

		// Execute the API linter.
		lintProblems, err := runLinter(protoPath, req.GetSpecDirectory())
		if err != nil {
			return err
		}

		// Filter the problems only those that were enabled by the user.
		filteredProblems := linter.filterProblems(lintProblems, req.GetRuleIds())

		// Formulate the response.
		lintFiles = append(lintFiles, &style.LintFile{
			FilePath: path,
			Problems: filteredProblems,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &style.LinterResponse{
		Lint: &style.Lint{
			Name:  "registry-lint-api-linter",
			Files: lintFiles,
		},
	}, nil
}

func runApiLinter(specPath, specDirectory string) ([]*style.LintProblem, error) {
	// TODO: Replace this new instance with a logger inherited from the context.
	logger := log.NewLogger()
	logger.Infof("Running api-linter on %s", specPath)

	data, err := createAndRunApiLinterCommand(specDirectory, specPath)
	if err != nil {
		return nil, err
	}
	return parseLinterOutput(data)
}

func main() {
	lint.Main(&apiLinterRunner{})
}

func (linter *apiLinterRunner) filterProblems(
	problems []*style.LintProblem,
	rules []string,
) []*style.LintProblem {
	// If no rules were specified, return the list without filtering.
	if len(rules) == 0 {
		return problems
	}

	// Construct a set of all the problems enabled for this mimetype
	// so we have efficient lookup.
	enabledProblems := make(map[string]bool)
	for _, rule := range rules {
		enabledProblems[rule] = true
	}

	// From a list of LintProblem objects, only return the rules that were
	// enabled by the caller via `addRule`.
	// We can do this in place.
	n := 0
	for i := 0; i < len(problems); i++ {
		if _, exists := enabledProblems[problems[i].GetRuleId()]; exists {
			problems[n] = problems[i]
			n++
		}
	}

	return problems[:n]
}

func createAndRunApiLinterCommand(specDirectory, specName string) ([]byte, error) {
	cmd := exec.Command("api-linter",
		specName,
		"-I", ".",
		"--output-format", "json",
	)
	cmd.Dir = specDirectory
	return cmd.CombinedOutput()
}

func parseLinterOutput(data []byte) ([]*style.LintProblem, error) {
	// Parse the API Linter output.
	if len(data) == 0 {
		return []*style.LintProblem{}, nil
	}
	var lintFiles []*style.LintFile
	err := json.Unmarshal(data, &lintFiles)
	if err != nil {
		return nil, err
	}

	// We only passed a single spec to the API linter. Thus
	// the LintFile array should only contain 1 element.
	if len(lintFiles) > 0 {
		lintFile := lintFiles[0]
		return lintFile.GetProblems(), nil
	}
	return []*style.LintProblem{}, nil
}
