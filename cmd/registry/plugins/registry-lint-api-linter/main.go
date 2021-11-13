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
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
)

// Runs the API linter with a provided spec path
type runLinter func(specPath string) ([]*rpc.LintProblem, error)

// apiLinterRunner implements the LinterRunner interface for the API linter.
type apiLinterRunner struct{}

func (linter *apiLinterRunner) Run(req *rpc.LinterRequest) (*rpc.LinterResponse, error) {
	return linter.RunImpl(req, runApiLinter)
}

func (linter *apiLinterRunner) RunImpl(
	req *rpc.LinterRequest,
	runLinter runLinter,
) (*rpc.LinterResponse, error) {

	lintFiles := make([]*rpc.LintFile, 0)

	// Traverse the files in the directory
	err := filepath.Walk(req.GetSpecDirectory(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".proto") {
			// Currently, only proto files are supported by API Linter.
			return nil
		}

		// Execute the API linter.
		lintProblems, err := runLinter(path)
		if err != nil {
			return err
		}

		// Filter the problems only those that were enabled by the user.
		filteredProblems := linter.filterProblems(lintProblems, req.GetRuleIds())

		// Formulate the response.
		lintFiles = append(lintFiles, &rpc.LintFile{
			FilePath: path,
			Problems: filteredProblems,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &rpc.LinterResponse{
		Lint: &rpc.Lint{
			Name:  "registry-lint-api-linter",
			Files: lintFiles,
		},
	}, nil
}

func runApiLinter(specPath string) ([]*rpc.LintProblem, error) {
	// API-linter necessitates being ran on specs in the CWD to avoid many import errors,
	// so we change the directory of the command to the directory of the spec.
	specDirectory := filepath.Dir(specPath)
	specName := filepath.Base(specPath)

	data, err := createAndRunApiLinterCommand(specDirectory, specName)
	if err == nil {
		return parseLinterOutput(data)
	}

	// TODO: Replace this new instance with a logger inherited from the context.
	logger := log.NewLogger()

	// Unpack api-common-protos and try again if failure occurred
	logger.Info("API-linter failed due to an import error, unpacking API common protos and retrying.")
	if err = unpackApiCommonProtos(specDirectory); err == nil {
		data, err = createAndRunApiLinterCommand(specDirectory, specName)
		if err == nil {
			return parseLinterOutput(data)
		}
	}

	logger.Info("API-linter failed due to an import error, unpacking GoogleAPIs and retrying.")
	if err = unpackGoogleApisProtos(specDirectory); err == nil {
		data, err = createAndRunApiLinterCommand(specDirectory, specName)
		if err == nil {
			return parseLinterOutput(data)
		}
	}

	return []*rpc.LintProblem{}, nil
}

func main() {
	lint.Main(&apiLinterRunner{})
}

func unpackGoogleApisProtos(rootDir string) error {
	// Curl the entire folder as a zipped archive from Github (much faster than git checkout).
	curlCmd := exec.Command("curl", "-L", "https://github.com/googleapis/googleapis/archive/refs/heads/master.zip", "-O")
	curlCmd.Dir = rootDir
	err := curlCmd.Run()
	if err != nil {
		return err
	}

	// Unzip the contents of the zipped archive.
	unzipCmd := exec.Command("unzip", "-q", "master.zip")
	unzipCmd.Dir = rootDir
	err = unzipCmd.Run()
	if err != nil {
		return err
	}

	// Move up the google/ directory (the one we're interested in) into the cwd.
	mvCmd := exec.Command("mv", "googleapis-master/google", "google")
	mvCmd.Dir = rootDir
	return mvCmd.Run()
}

func unpackApiCommonProtos(rootDir string) error {
	cmd := exec.Command("git", "clone", "https://github.com/googleapis/api-common-protos")
	cmd.Dir = rootDir
	return cmd.Run()
}

func (linter *apiLinterRunner) filterProblems(
	problems []*rpc.LintProblem,
	rules []string,
) []*rpc.LintProblem {
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
		"-I", "google",
		"-I", "api-common-protos",
		"--output-format", "json",
	)
	cmd.Dir = specDirectory
	return cmd.CombinedOutput()
}

func parseLinterOutput(data []byte) ([]*rpc.LintProblem, error) {
	// Parse the API Linter output.
	if len(data) == 0 {
		return []*rpc.LintProblem{}, nil
	}
	var lintFiles []*rpc.LintFile
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
	return []*rpc.LintProblem{}, nil
}
