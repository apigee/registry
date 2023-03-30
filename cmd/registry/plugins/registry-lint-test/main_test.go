// Copyright 2023 Google LLC
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
	"os"
	"path/filepath"
	"testing"

	"github.com/apigee/registry/pkg/application/style"
	"github.com/stretchr/testify/assert"
)

func TestTestLinter(t *testing.T) {
	var err error
	specDirectory := t.TempDir()
	err = os.WriteFile(filepath.Join(specDirectory, "openapi.yaml"), []byte(petstore), 0666)
	if err != nil {
		t.Fatal("Failed to create test file")
	}
	err = os.WriteFile(filepath.Join(specDirectory, "petstore.yaml"), []byte(petstore), 0666)
	if err != nil {
		t.Fatal("Failed to create test file")
	}
	request := &style.LinterRequest{
		SpecDirectory: specDirectory,
	}
	expectedResponse := &style.LinterResponse{
		Lint: &style.Lint{
			Name: "registry-lint-test",
			Files: []*style.LintFile{
				{
					FilePath: "openapi.yaml",
					Problems: []*style.LintProblem{
						{
							Message:    "2618",
							RuleId:     "size",
							RuleDocUri: "https://github.com/apigee/registry",
							Suggestion: "This is the size of openapi.yaml.",
							Location: &style.LintLocation{
								StartPosition: &style.LintPosition{
									LineNumber:   1,
									ColumnNumber: 1,
								},
								EndPosition: &style.LintPosition{
									LineNumber:   113,
									ColumnNumber: 1,
								},
							},
						},
					},
				},
				{
					FilePath: "petstore.yaml",
					Problems: []*style.LintProblem{
						{
							Message:    "2618",
							RuleId:     "size",
							RuleDocUri: "https://github.com/apigee/registry",
							Suggestion: "This is the size of petstore.yaml.",
							Location: &style.LintLocation{
								StartPosition: &style.LintPosition{
									LineNumber:   1,
									ColumnNumber: 1,
								},
								EndPosition: &style.LintPosition{
									LineNumber:   113,
									ColumnNumber: 1,
								},
							},
						},
					},
				},
			},
		},
	}
	response, err := (&testLinterRunner{}).Run(request)
	if err != nil {
		t.Fatalf("Linter failed with error %s", err)
	}
	assert.EqualValues(t, expectedResponse, response)
}

const petstore = `openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: How many items to return at one time (max 100)
          required: false
          schema:
            type: integer
            maximum: 100
            format: int32
      responses:
        '200':
          description: A paged array of pets
          headers:
            x-next:
              description: A link to the next page of responses
              schema:
                type: string
          content:
            application/json:    
              schema:
                $ref: "#/components/schemas/Pets"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
    post:
      summary: Create a pet
      operationId: createPets
      tags:
        - pets
      responses:
        '201':
          description: Null response
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
  /pets/{petId}:
    get:
      summary: Info for a specific pet
      operationId: showPetById
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          required: true
          description: The id of the pet to retrieve
          schema:
            type: string
      responses:
        '200':
          description: Expected response to a valid request
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
components:
  schemas:
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
    Pets:
      type: array
      maxItems: 100
      items:
        $ref: "#/components/schemas/Pet"
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
`
