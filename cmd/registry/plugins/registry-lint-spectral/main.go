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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	lint "github.com/apigee/registry/cmd/registry/plugins/linter"
	"github.com/apigee/registry/pkg/application/style"
)

// spectralConfiguration describes a spectral ruleset that is used to lint
// a given API Spec.
type spectralConfiguration struct {
	Extends [][]string      `json:"extends"`
	Rules   map[string]bool `json:"rules"`
}

// spectralLintResult contains metadata related to a rule violation.
type spectralLintResult struct {
	Code     string            `json:"code"`
	Path     []string          `json:"path"`
	Message  string            `json:"message"`
	Severity int32             `json:"severity"`
	Range    spectralLintRange `json:"range"`
	Source   string            `json:"source"`
}

// spectralLintRange is the start and end location for a rule violation.
type spectralLintRange struct {
	Start spectralLintLocation `json:"start"`
	End   spectralLintLocation `json:"end"`
}

// spectralLintLocation is the location in a file for a rule violation.
type spectralLintLocation struct {
	Line      int32 `json:"line"`
	Character int32 `json:"character"`
}

// Runs the spectral linter with a provided spec and configuration path
type runLinter func(specPath, configPath string) ([]*spectralLintResult, error)

// spectralLinterRunner implements the LinterRunner interface for the Spectral linter.
type spectralLinterRunner struct{}

func (linter *spectralLinterRunner) Run(req *style.LinterRequest) (*style.LinterResponse, error) {
	return linter.RunImpl(req, runSpectralLinter)
}

func (linter *spectralLinterRunner) RunImpl(
	req *style.LinterRequest,
	runLinter runLinter,
) (*style.LinterResponse, error) {
	lintFiles := make([]*style.LintFile, 0)

	// Create a temporary directory to store the configuration.
	root, err := os.MkdirTemp("", "spectral-config-")
	if err != nil {
		return nil, err
	}

	// Defer the deletion of the temporary directory.
	defer os.RemoveAll(root)

	// Create configuration file for Spectral to execute the correct rules
	configPath, err := linter.createConfigurationFile(root, req.GetRuleIds())
	if err != nil {
		return nil, err
	}

	// Traverse the files in the directory
	err = filepath.Walk(req.GetSpecDirectory(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Execute the spectral linter.
		lintResults, err := runLinter(path, configPath)
		if err != nil {
			return err
		}

		// Get the lint results as a LintFile object from the spectral output file
		lintProblems := getLintProblemsFromSpectralResults(lintResults)

		// Formulate the response.
		lintFile := &style.LintFile{
			FilePath: path,
			Problems: lintProblems,
		}

		lintFiles = append(lintFiles, lintFile)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &style.LinterResponse{
		Lint: &style.Lint{
			Name:  "registry-lint-spectral",
			Files: lintFiles,
		},
	}, nil
}

func (linter *spectralLinterRunner) createConfigurationFile(root string, ruleIds []string) (string, error) {
	// Create the spectral configuration.
	configuration := spectralConfiguration{}
	configuration.Rules = make(map[string]bool)
	if len(ruleIds) == 0 {
		// if no rules were specified, use the default rules.
		configuration.Extends = [][]string{{"spectral:oas", "all"}, {"spectral:asyncapi", "all"}}
	} else {
		configuration.Extends = [][]string{{"spectral:oas", "off"}, {"spectral:asyncapi", "off"}}
	}
	for _, ruleName := range ruleIds {
		configuration.Rules[ruleName] = true
	}
	// Marshal the configuration into a file.
	file, err := json.MarshalIndent(configuration, "", " ")
	if err != nil {
		return "", err
	}
	// Write the configuration to the temporary directory.
	configPath := filepath.Join(root, "spectral.json")
	err = os.WriteFile(configPath, file, 0644)
	if err != nil {
		return "", err
	}
	return configPath, nil
}

func getLintProblemsFromSpectralResults(
	lintResults []*spectralLintResult,
) []*style.LintProblem {
	problems := make([]*style.LintProblem, len(lintResults))
	for i, result := range lintResults {
		problem := &style.LintProblem{
			Message:    result.Message,
			RuleId:     result.Code,
			RuleDocUri: "https://meta.stoplight.io/docs/spectral/docs/reference/openapi-rules.md#" + result.Code,
			Location: &style.LintLocation{
				StartPosition: &style.LintPosition{
					LineNumber:   result.Range.Start.Line + 1,
					ColumnNumber: result.Range.Start.Character + 1,
				},
				EndPosition: &style.LintPosition{
					LineNumber:   result.Range.End.Line + 1,
					ColumnNumber: result.Range.End.Character,
				},
			},
		}
		problems[i] = problem
	}
	return problems
}

func runSpectralLinter(specPath, configPath string) ([]*spectralLintResult, error) {
	// Create a temporary destination directory to store the output.
	root, err := os.MkdirTemp("", "spectral-output-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(root)

	// Set the destination path of the spectral output.
	outputPath := filepath.Join(root, "spectral-lint.json")

	cmd := exec.Command("spectral",
		"lint", specPath,
		"--r", configPath,
		"--f", "json",
		"--output", outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		switch v := err.(type) {
		case *exec.ExitError:
			code := v.ExitCode()
			if code == 1 {
				// This just means the linter found errors
			} else {
				log.Printf("linter error %T (%s)", err, specPath)
				log.Printf("%s", string(output))
			}
		case *exec.Error:
			if strings.Contains(v.Err.Error(), "executable file not found") {
				return nil, v.Err
			}
			log.Printf("linter error %T (%s)", err, specPath)
			log.Printf("%s", string(output))
		default:
			log.Printf("linter error %T (%s)", err, specPath)
			log.Printf("%s", string(output))
		}
	}

	// Read and parse the spectral output.
	b, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}
	var lintResults []*spectralLintResult
	err = json.Unmarshal(b, &lintResults)
	if err != nil {
		return nil, err
	}

	return lintResults, nil
}

func main() {
	lint.Main(&spectralLinterRunner{})
}
