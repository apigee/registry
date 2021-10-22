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

// mockApiLinterExecuter implements the apiLinterCommandExecuter interface.
// It returns mock results according to data provided in tests.
type mockApiLinterExecuter struct {
	mock.Mock
	results []*rpc.LintProblem
	err error
}

func (runner *mockApiLinterExecuter) Execute(specPath string) ([]*rpc.LintProblem, error) {
	return runner.results, runner.err
}

func newMockApiLinterExecuter(
	results []*rpc.LintProblem, 
	err error,
) apiLinterCommandExecuter {
	return &mockApiLinterExecuter{
		results: results, 
		err: err,
	}
}

func TestApiLinterPluginLintSpec(t *testing.T) {
	lintSpecTests := []struct {
		linter *apiLinterRunner
		request *rpc.LinterRequest
		executer apiLinterCommandExecuter
		expectedResponse *rpc.LinterResponse
		expectedError error
    }{
        {
			&apiLinterRunner{},
			&rpc.LinterRequest{
				SpecPath: "test",
				RuleIds: []string{"test"},
			},
			newMockApiLinterExecuter(
				[]*rpc.LintProblem {
					{
						Message:    "test",
						RuleId:     "test",
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
			nil,
			),
			&rpc.LinterResponse{
				Lint: &rpc.Lint{
					Name: "registry-lint-api-linter",
					Files: []*rpc.LintFile {
						{
							FilePath: "test",
							Problems: []*rpc.LintProblem{
								{
									Message:    "test",
									RuleId:     "test",
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
			&apiLinterRunner{},
			&rpc.LinterRequest{
				SpecPath: "test",
			},
			newMockApiLinterExecuter(
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
