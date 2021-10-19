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
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockSpectralExecuter implements the spectral runner interface.
// It returns mock results according to data provided in tests.
type mockSpectralExecuter struct {
	mock.Mock
	results []*spectralLintResult
	err error
}

func (runner *mockSpectralExecuter) Execute(
	spec, 
	config string,
) ([]*spectralLintResult, error) {
	return runner.results, runner.err
}

func NewMockSpectralRunner(
	results []*spectralLintResult, 
	err error,
) spectralLintCommandExecuter {
	test := &mockSpectralExecuter{
		results: results, 
		err: err,
	}
	return test
}

func TestSpectralPluginLintSpec(t *testing.T) {
	lintSpecTests := []struct {
        linter *spectralLinterRunner
		request *rpc.LinterRequest
		executer spectralLintCommandExecuter
		expectedResponse *rpc.LinterResponse
		expectedError error
    }{
        {
			&spectralLinterRunner{},
			&rpc.LinterRequest{
				SpecPath: "test",
			},
			NewMockSpectralRunner(
				[]*spectralLintResult {
					{
						Code: "test",
						Message: "test",
						Source: "test",
						Range: spectralLintRange {
							Start: spectralLintLocation {
								Line: 0, Character: 0,
							},
							End: spectralLintLocation {
								Line: 2, Character: 10,
							},
						},
					},
				},
			nil,
			),
			&rpc.LinterResponse{
				Lint: &rpc.Lint{
					Name: "registry-lint-sample",
					Files: []*rpc.LintFile {
						{
							FilePath: "test",
							Problems: []*rpc.LintProblem{
								{
									Message:    "test",
									RuleId:     "test",
									RuleDocUri: "https://meta.stoplight.io/docs/spectral/docs/reference/openapi-rules.md#test",
									Location: &rpc.LintLocation{
										StartPosition: &rpc.LintPosition{
											LineNumber:   1,
											ColumnNumber: 1,
										},
										EndPosition: &rpc.LintPosition{
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
			&rpc.LinterRequest{
				SpecPath: "test",
			},
			NewMockSpectralRunner(
				[]*spectralLintResult{},
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
