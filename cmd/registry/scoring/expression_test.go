// Copyright 2022 Google LLC
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

package scoring

import (
	"testing"

	"github.com/apigee/registry/pkg/application/scoring"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetMap(t *testing.T) {
	tests := []struct {
		desc          string
		contentsProto proto.Message
		mimeType      string
		wantMap       map[string]interface{}
	}{
		{
			desc: "happy path style.Lint",
			contentsProto: &style.Lint{
				Name: "openapi.yaml",
				Files: []*style.LintFile{
					{
						FilePath: "openapi.yaml",
						Problems: []*style.LintProblem{
							{
								Message: "lint-error",
							},
						},
					},
				},
			},
			mimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
			wantMap: map[string]interface{}{
				"name": "openapi.yaml",
				"files": []interface{}{
					map[string]interface{}{
						"filePath": "openapi.yaml",
						"problems": []interface{}{
							map[string]interface{}{
								"message": "lint-error",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			contents, _ := proto.Marshal(test.contentsProto)

			gotMap, gotErr := getMap(contents, test.mimeType)
			if gotErr != nil {
				t.Errorf("getMap() returned unexpected error: %s", gotErr)
			}
			opts := protocmp.Transform()
			if !cmp.Equal(test.wantMap, gotMap, opts) {
				t.Errorf("getMap returned unexpected response (-want +got):\n%s", cmp.Diff(test.wantMap, gotMap, opts))
			}
		})
	}
}

func TestGetMapError(t *testing.T) {
	tests := []struct {
		desc          string
		contentsProto proto.Message
		mimeType      string
	}{
		{
			desc: "unsupported artifact type",
			contentsProto: &scoring.ScoreDefinition{
				TargetResource: &scoring.ResourcePattern{},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-error",
						},
						ScoreExpression: "size(files[0].problems)",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
			mimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.ScoreDefinition",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			contents, _ := proto.Marshal(test.contentsProto)

			_, gotErr := getMap(contents, test.mimeType)
			if gotErr == nil {
				t.Errorf("getMap(%v, %s) did not return an error", test.contentsProto, test.mimeType)
			}
		})
	}
}

func TestEvaluateScoreExpression(t *testing.T) {
	tests := []struct {
		desc        string
		expression  string
		artifactMap map[string]interface{}
		wantValue   interface{}
	}{
		{
			desc:       "int happy path",
			expression: "size(files[0].problems)",
			artifactMap: map[string]interface{}{
				"name": "openapi.yaml",
				"files": []map[string]interface{}{
					{
						"filePath": "openapi.yaml",
						"problems": []map[string]interface{}{
							{
								"message": "lint-error",
							},
						},
					},
				},
			},
			wantValue: int64(1),
		},
		{
			desc:       "double happy path",
			expression: "double(size(files[0].problems))",
			artifactMap: map[string]interface{}{
				"name": "openapi.yaml",
				"files": []map[string]interface{}{
					{
						"filePath": "openapi.yaml",
						"problems": []map[string]interface{}{
							{
								"message": "lint-error",
							},
						},
					},
				},
			},
			wantValue: float64(1),
		},
		{
			desc:       "bool happy path",
			expression: "size(files[0].problems)>0",
			artifactMap: map[string]interface{}{
				"name": "openapi.yaml",
				"files": []map[string]interface{}{
					{
						"filePath": "openapi.yaml",
						"problems": []map[string]interface{}{
							{
								"message": "lint-error",
							},
						},
					},
				},
			},
			wantValue: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotValue, gotErr := evaluateScoreExpression(test.expression, test.artifactMap)
			if gotErr != nil {
				t.Errorf("evaluateScoreExpression() returned unexpected error: %s", gotErr)
			}

			if test.wantValue != gotValue {
				t.Errorf("evaluateScoreExpression() returned unexpected value, want: %v, got: %v", test.wantValue, gotValue)
			}
		})
	}
}

func TestEvaluateScoreExpressionError(t *testing.T) {
	tests := []struct {
		desc        string
		expression  string
		artifactMap map[string]interface{}
	}{
		{
			desc:       "invalid field reference",
			expression: "size(files.problems)", // correct expression should be "size(files[0].problems)"
			artifactMap: map[string]interface{}{
				"name": "openapi.yaml",
				"files": []map[string]interface{}{
					{
						"filePath": "openapi.yaml",
						"problems": []map[string]interface{}{
							{
								"message": "lint-error",
							},
						},
					},
				},
			},
		},
		{
			desc:       "unsupported type (list)",
			expression: "files[0].problems", //this will return a list value
			artifactMap: map[string]interface{}{
				"name": "openapi.yaml",
				"files": []map[string]interface{}{
					{
						"filePath": "openapi.yaml",
						"problems": []map[string]interface{}{
							{
								"message": "lint-error",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, gotErr := evaluateScoreExpression(test.expression, test.artifactMap)
			if gotErr == nil {
				t.Errorf("evaluateScoreExpression(%s, %v) did not return an error", test.expression, test.artifactMap)
			}
		})
	}
}
