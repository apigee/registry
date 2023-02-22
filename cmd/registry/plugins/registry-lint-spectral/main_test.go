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
	"errors"
	"os"
	"testing"

	"github.com/apigee/registry/pkg/application/style"
	"github.com/stretchr/testify/assert"
)

func setupFakeSpec() (path string, err error) {
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	f, err := os.CreateTemp(tempDir, "*.yaml")
	if err != nil {
		return "", err
	}
	return f.Name(), err
}

func TestSpectralPluginLintSpec(t *testing.T) {
	specDirectory, err := setupFakeSpec()
	defer os.RemoveAll(specDirectory)
	assert.Equal(t, err, nil)
	lintSpecTests := []struct {
		linter           *spectralLinterRunner
		runLinter        runLinter
		request          *style.LinterRequest
		expectedResponse *style.LinterResponse
		expectedError    error
	}{
		{
			&spectralLinterRunner{},
			func(specPath, configPath string) ([]*spectralLintResult, error) {
				return []*spectralLintResult{
					{
						Code:    "test",
						Message: "test",
						Source:  "test",
						Range: spectralLintRange{
							Start: spectralLintLocation{
								Line:      0,
								Character: 0},
							End: spectralLintLocation{
								Line:      2,
								Character: 10,
							},
						},
					},
				}, nil
			},
			&style.LinterRequest{
				SpecDirectory: specDirectory,
			},
			&style.LinterResponse{
				Lint: &style.Lint{
					Name: "registry-lint-spectral",
					Files: []*style.LintFile{
						{
							FilePath: specDirectory,
							Problems: []*style.LintProblem{
								{
									Message:    "test",
									RuleId:     "test",
									RuleDocUri: "https://meta.stoplight.io/docs/spectral/docs/reference/openapi-rules.md#test",
									Location: &style.LintLocation{
										StartPosition: &style.LintPosition{
											LineNumber:   1,
											ColumnNumber: 1,
										},
										EndPosition: &style.LintPosition{
											LineNumber:   3,
											ColumnNumber: 10,
										},
									},
								},
							},
						},
					},
				},
			},
			nil,
		},
		{
			&spectralLinterRunner{},
			func(specPath, configPath string) ([]*spectralLintResult, error) {
				return nil, errors.New("test")
			},
			&style.LinterRequest{
				SpecDirectory: specDirectory,
			},
			nil,
			errors.New("test"),
		},
	}

	for _, tt := range lintSpecTests {
		response, err := tt.linter.RunImpl(tt.request, tt.runLinter)
		assert.Equal(t, tt.expectedError, err)
		assert.EqualValues(t, tt.expectedResponse, response)
	}
}
