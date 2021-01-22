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

package core

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/apigee/registry/rpc"
)

type SpectralLintResult struct {
	Code     string            `json:"code"`
	Path     []string          `json:"path"`
	Message  string            `json:"message"`
	Severity int32             `json:"severity"`
	Range    SpectralLintRange `json:"range"`
	Source   string            `json:"source"`
}

type SpectralLintRange struct {
	Start SpectralLintLocation `json:"start"`
	End   SpectralLintLocation `json:"end"`
}

type SpectralLintLocation struct {
	Line      int32 `json:"line"`
	Character int32 `json:"character"`
}

func lintFileForOpenAPIWithSpectral(path string, root string) (*rpc.LintFile, error) {
	cmd := exec.Command("spectral", "lint", path, "--f", "json", "--output", "spectral-lint.json")
	cmd.Dir = root
	// ignore errors from Spectral because Spectral returns an error result when APIs have errors.
	cmd.Run()
	b, err := ioutil.ReadFile(filepath.Join(root, "/spectral-lint.json"))
	if err != nil {
		return nil, err
	}
	var lintResults []*SpectralLintResult
	err = json.Unmarshal(b, &lintResults)
	if err != nil {
		return nil, err
	}
	problems := make([]*rpc.LintProblem, 0)
	for _, result := range lintResults {
		problem := &rpc.LintProblem{
			Message: result.Message,
			RuleId:  result.Code,
			Location: &rpc.LintLocation{
				StartPosition: &rpc.LintPosition{
					LineNumber:   result.Range.Start.Line + 1,
					ColumnNumber: result.Range.Start.Character + 1,
				},
				EndPosition: &rpc.LintPosition{
					LineNumber:   result.Range.End.Line + 1,
					ColumnNumber: result.Range.End.Character,
				},
			},
		}
		problems = append(problems, problem)
	}
	result := &rpc.LintFile{Problems: problems}
	return result, nil
}
