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
	"io/ioutil"
	"os"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockSampleOpenApiExecuter implements the Sample OpenAPI runner interface.
// It returns mock results according to data provided in tests.
type mockSampleOpenApiExecuter struct {
	mock.Mock
	results []*rpc.LintProblem
	err     error
}

func (runner *mockSampleOpenApiExecuter) Execute(
	spec string,
	ruleIds []string,
) ([]*rpc.LintProblem, error) {
	return runner.results, runner.err
}

func NewMockSampleOpenAPIExecuter(
	results []*rpc.LintProblem,
	err error,
) sampleOpenApiLintCommandExecuter {
	return &mockSampleOpenApiExecuter{
		results: results,
		err:     err,
	}
}

func setupFakeSpec() (path string, err error) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}

	f, err := ioutil.TempFile(tempDir, "*.yaml")
	if err != nil {
		return "", err
	}
	return f.Name(), err
}

func TestRunImpl(t *testing.T) {
	specDirectory, err := setupFakeSpec()
	defer os.RemoveAll(specDirectory)
	assert.Equal(t, err, nil)
	lintSpecTests := []struct {
		linter           *sampleOpenApiLinterRunner
		request          *rpc.LinterRequest
		executer         sampleOpenApiLintCommandExecuter
		expectedResponse *rpc.LinterResponse
		expectedError    error
	}{
		{
			&sampleOpenApiLinterRunner{},
			&rpc.LinterRequest{
				SpecDirectory: specDirectory,
			},
			NewMockSampleOpenAPIExecuter(
				[]*rpc.LintProblem{
					{
						Message: "test",
						RuleId:  "test",
						Location: &rpc.LintLocation{
							StartPosition: &rpc.LintPosition{
								LineNumber:   1,
								ColumnNumber: 1,
							},
						},
					},
				},
				nil,
			),
			&rpc.LinterResponse{
				Lint: &rpc.Lint{
					Name: "registry-lint-openapi-sample",
					Files: []*rpc.LintFile{
						{
							FilePath: specDirectory,
							Problems: []*rpc.LintProblem{
								{
									Message: "test",
									RuleId:  "test",
									Location: &rpc.LintLocation{
										StartPosition: &rpc.LintPosition{
											LineNumber:   1,
											ColumnNumber: 1,
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
			&sampleOpenApiLinterRunner{},
			&rpc.LinterRequest{
				SpecDirectory: specDirectory,
			},
			NewMockSampleOpenAPIExecuter(
				[]*rpc.LintProblem{},
				errors.New("test"),
			),
			nil,
			errors.New("test"),
		},
	}

	for _, tt := range lintSpecTests {
		response, err := tt.linter.RunImpl(tt.request, tt.executer)
		assert.Equal(t, tt.expectedError, err)
		assert.EqualValues(t, tt.expectedResponse, response)
	}
}
