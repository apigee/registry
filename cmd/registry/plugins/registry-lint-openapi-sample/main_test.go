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
	"io/ioutil"
	"os"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/stretchr/testify/assert"
)

func setupFakeSpec(contents string) (dirPath, specFilePath string, err error) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", "", err
	}

	f, err := ioutil.TempFile(tempDir, "*.yaml")
	if err != nil {
		return "", "", err
	}

	err = ioutil.WriteFile(f.Name(), []byte(contents), 0644)
	if err != nil {
		return "", "", err
	}

	return tempDir, f.Name(), nil
}

func TestRunDescriptionContainsNoTagsRule(t *testing.T) {
	contents := `
        openapi: "3.0.2"
        info:
            title: "Swagger Petstore <script></script> with eval("
        servers:
            - url: http://petstore.swagger.io/v1
        paths:
            /pets:
                get:
                description: Gets a <list> of all pets
    `
	specDirectory, specFilePath, err := setupFakeSpec(contents)
	defer os.RemoveAll(specDirectory)
	assert.Equal(t, err, nil)
	linter := &sampleOpenApiLinterRunner{}
	request := &rpc.LinterRequest{
		SpecDirectory: specDirectory,
		RuleIds:       []string{"description-contains-no-tags"},
	}
	expectedResponse := &rpc.LinterResponse{
		Lint: &rpc.Lint{
			Name: "registry-lint-openapi-sample",
			Files: []*rpc.LintFile{
				{
					FilePath: specFilePath,
					Problems: []*rpc.LintProblem{
						{
							Message:    "Description field should not contain any tags.",
							RuleId:     "description-contains-no-tags",
							Suggestion: "Ensure that your description field does not contain any tags (regex <[^>]*>)",
							Location: &rpc.LintLocation{
								StartPosition: &rpc.LintPosition{
									LineNumber:   10,
									ColumnNumber: 30,
								},
								EndPosition: &rpc.LintPosition{
									LineNumber:   11,
									ColumnNumber: 0,
								},
							},
						},
					},
				},
			},
		},
	}

	response, err := linter.Run(request)
	assert.Equal(t, nil, err)
	assert.EqualValues(t, expectedResponse, response)
}

func TestRunDescriptionLessThan1000CharsRule(t *testing.T) {
	contents := `
        openapi: "3.0.2"
        info:
            title: "Swagger Petstore <script></script> with eval("
        servers:
            - url: http://petstore.swagger.io/v1
        paths:
            /pets:
                get:
                description: >
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
                    eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
    `
	specDirectory, specFilePath, err := setupFakeSpec(contents)
	defer os.RemoveAll(specDirectory)
	assert.Equal(t, err, nil)
	linter := &sampleOpenApiLinterRunner{}
	request := &rpc.LinterRequest{
		SpecDirectory: specDirectory,
		RuleIds:       []string{"description-less-than-1000-chars"},
	}
	expectedResponse := &rpc.LinterResponse{
		Lint: &rpc.Lint{
			Name: "registry-lint-openapi-sample",
			Files: []*rpc.LintFile{
				{
					FilePath: specFilePath,
					Problems: []*rpc.LintProblem{
						{
							Message:    "Description field should be less than 1000 chars.",
							RuleId:     "description-less-than-1000-chars",
							Suggestion: "Ensure that your description field is less than 1000 chars in length.",
							Location: &rpc.LintLocation{
								StartPosition: &rpc.LintPosition{
									LineNumber:   10,
									ColumnNumber: 30,
								},
								EndPosition: &rpc.LintPosition{
									LineNumber:   11,
									ColumnNumber: 0,
								},
							},
						},
					},
				},
			},
		},
	}

	response, err := linter.Run(request)
	assert.Equal(t, nil, err)
	assert.EqualValues(t, expectedResponse, response)
}
