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
	"fmt"

	"github.com/apigee/registry/cmd/registry/plugins/linter"
	"github.com/apigee/registry/pkg/application/style"
)

// SampleLinterRunner implements the LinterRunner interface for the sample linter.
type SampleLinterRunner struct{}

func (*SampleLinterRunner) Run(req *style.LinterRequest) (*style.LinterResponse, error) {
	// Formulate the response. In this sample plugin, we will simply return a fake rule violation /
	// lint problem for every rule that the user specifies, on the given file that is provided.
	lintFile := &style.LintFile{
		FilePath: req.SpecDirectory,
	}

	for _, rule := range req.RuleIds {
		lintFile.Problems = append(lintFile.Problems, &style.LintProblem{
			RuleId:  rule,
			Message: fmt.Sprintf("This is a sample violation of the rule %s", rule),
		})
	}

	return &style.LinterResponse{
		Lint: &style.Lint{
			Name: "registry-lint-sample",
			Files: []*style.LintFile{
				lintFile,
			},
		},
	}, nil
}

func main() {
	linter.Main(&SampleLinterRunner{})
}
