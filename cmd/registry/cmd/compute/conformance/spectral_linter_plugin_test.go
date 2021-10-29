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

package conformance

import (
	"errors"
	"fmt"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSpectralPluginGetName(t *testing.T) {
	linter := SpectralLinter{}
	name := linter.GetName()
	expectedName := "spectral"
	if name != expectedName {
		t.Errorf(
			"Name is incorrect, got: %s, want: %s.",
			name,
			expectedName,
		)
	}
}

func TestSpectralPluginAddRule(t *testing.T) {
	addRuleTests := []struct {
		linter        SpectralLinter
		mimeType      string
		rule          string
		expectedError error
	}{
		{
			NewSpectralLinter(),
			"application/x.openapi+gzip;version=2",
			"testRule1",
			nil,
		},
		{
			NewSpectralLinter(),
			"application/x.protobuf+gzip",
			"testRule2",
			fmt.Errorf(
				"mime type %s is not supported by the spectral linter",
				"application/x.protobuf+gzip",
			),
		},
	}

	for _, tt := range addRuleTests {
		// Add the rule to the linter
		err := tt.linter.AddRule(tt.mimeType, tt.rule)

		// Ensure that the error output of AddRule is what we expect
		if err == nil || tt.expectedError == nil {
			if err != tt.expectedError {
				t.Errorf("got %s want %s", err, tt.expectedError)
			}
		} else if err.Error() != tt.expectedError.Error() {
			t.Errorf("got %s want %s", err, tt.expectedError)
		}

		if err != nil {
			continue
		}

		// Ensure that the rule was added to the linter
		ruleAdded := false
		for _, rule := range tt.linter.Rules[tt.mimeType] {
			if rule == tt.rule {
				ruleAdded = true
				break
			}
		}
		if !ruleAdded {
			t.Errorf(
				"AddRule was unable to add the rule \"%s\"",
				tt.rule,
			)
		}
	}
}

func TestSpectralPluginSupportsMimeType(t *testing.T) {
	supportsMimeTypeTests := []struct {
		linter   SpectralLinter
		mimeType string
		want     bool
	}{
		{
			NewSpectralLinter(),
			"application/x.openapi+gzip;version=2",
			true,
		},
		{
			NewSpectralLinter(),
			"application/x.openapi+gzip;version=3",
			true,
		},
		{
			NewSpectralLinter(),
			"application/x.asyncapi+gzip;version=2",
			true,
		},
		{
			NewSpectralLinter(),
			"application/x.protobuf+gzip",
			false,
		},
		{
			NewSpectralLinter(),
			"application/x.asyncapi+gzip;version=3",
			false,
		},
	}

	for _, tt := range supportsMimeTypeTests {
		if supports :=
			tt.linter.SupportsMimeType(tt.mimeType); supports != tt.want {
			t.Errorf(
				"SupportsMimeType returned %t for mime type %s, expected %t",
				supports,
				tt.mimeType,
				tt.want,
			)
		}
	}
}

// mockSpectralRunner implements the spectral runner interface.
// It returns mock results according to data provided in tests.
type mockSpectralRunner struct {
	mock.Mock
	results []*spectralLintResult
	err     error
}

func (runner *mockSpectralRunner) Run(
	spec,
	config string,
) ([]*spectralLintResult, error) {
	return runner.results, runner.err
}

func NewMockSpectralRunner(
	results []*spectralLintResult,
	err error,
) spectralRunner {
	test := &mockSpectralRunner{
		results: results,
		err:     err,
	}
	return test
}

func TestSpectralPluginLintSpec(t *testing.T) {
	lintSpecTests := []struct {
		linter               SpectralLinter
		mimeType             string
		runner               spectralRunner
		expectedLintProblems []*rpc.LintProblem
		expectedError        error
	}{
		{
			NewSpectralLinter(),
			"application/x.asyncapi+gzip;version=2",
			NewMockSpectralRunner(
				[]*spectralLintResult{
					{
						Code:    "test",
						Message: "test",
						Source:  "test",
						Range: spectralLintRange{
							Start: spectralLintLocation{
								Line: 0, Character: 0,
							},
							End: spectralLintLocation{
								Line: 2, Character: 10,
							},
						},
					},
				},
				nil,
			),
			[]*rpc.LintProblem{
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
			nil,
		},
		{
			NewSpectralLinter(),
			"application/x.asyncapi+gzip;version=2",
			NewMockSpectralRunner(
				[]*spectralLintResult{},
				errors.New("test"),
			),
			nil,
			errors.New("test"),
		},
	}

	for _, tt := range lintSpecTests {
		lintProblems, err := tt.linter.LintSpecImpl(tt.mimeType, "", tt.runner)
		assert.Equal(t, tt.expectedError, err)
		assert.EqualValues(t, tt.expectedLintProblems, lintProblems)
	}
}
