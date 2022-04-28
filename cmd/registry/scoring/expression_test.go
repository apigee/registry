package scoring

import (
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGetStructProto(t *testing.T) {
	tests := []struct {
		desc          string
		contentsProto proto.Message
		mimeType      string
		wantStruct    *structpb.Struct
	}{
		{
			desc: "happy path rpc.Lint",
			contentsProto: &rpc.Lint{
				Name: "openapi.yaml",
				Files: []*rpc.LintFile{
					{
						FilePath: "openapi.yaml",
						Problems: []*rpc.LintProblem{
							{
								Message: "lint-error",
							},
						},
					},
				},
			},
			mimeType: "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint",
			wantStruct: func() *structpb.Struct {
				problem, _ := structpb.NewStruct(map[string]interface{}{
					"message": "lint-error",
				})
				problemsList, _ := structpb.NewList([]interface{}{
					problem.AsMap(),
				})
				file, _ := structpb.NewStruct(
					map[string]interface{}{
						"filePath": "openapi.yaml",
						"problems": problemsList.AsSlice(),
					})
				fileList, _ := structpb.NewList(
					[]interface{}{
						file.AsMap(),
					})
				lint, _ := structpb.NewStruct(
					map[string]interface{}{
						"name":  "openapi.yaml",
						"files": fileList.AsSlice(),
					})
				return lint
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			contents, _ := proto.Marshal(test.contentsProto)

			gotStruct, gotErr := getStructProto(contents, test.mimeType)
			if gotErr != nil {
				t.Errorf("processScoreType() returned unexpected error: %s", gotErr)
			}
			opts := cmp.Options{protocmp.Transform()}
			if !cmp.Equal(test.wantStruct, gotStruct, opts) {
				t.Errorf("processScoreType returned unexpected response (-want +got):\n%s", cmp.Diff(test.wantStruct, gotStruct, opts))
			}
		})
	}
}

func TestGetStructProtoError(t *testing.T) {
	tests := []struct {
		desc          string
		contentsProto proto.Message
		mimeType      string
	}{
		{
			desc: "unsupported artifact type",
			contentsProto: &rpc.ScoreDefinition{
				TargetResource: &rpc.ResourcePattern{},
				Formula: &rpc.ScoreDefinition_ScoreFormula{
					ScoreFormula: &rpc.ScoreFormula{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-error",
						},
						ScoreExpression: "size(files[0].problems)",
					},
				},
				Type: &rpc.ScoreDefinition_Integer{
					Integer: &rpc.IntegerType{
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

			_, gotErr := getStructProto(contents, test.mimeType)
			if gotErr == nil {
				t.Errorf("expected getStructProto() to return an error")
			}
		})
	}
}

func TestEvaluateScoreExpression(t *testing.T) {
	tests := []struct {
		desc          string
		expression    string
		mimeType      string
		contentsProto proto.Message
		wantValue     interface{}
	}{
		{
			desc:       "int happy path",
			expression: "size(files[0].problems)",
			mimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint",
			contentsProto: &rpc.Lint{
				Name: "openapi.yaml",
				Files: []*rpc.LintFile{
					{
						FilePath: "openapi.yaml",
						Problems: []*rpc.LintProblem{
							{
								Message: "lint-error",
							},
						},
					},
				},
			},
			wantValue: 1,
		},
		{
			desc:       "double happy path",
			expression: "double(size(files[0].problems))",
			mimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint",
			contentsProto: &rpc.Lint{
				Name: "openapi.yaml",
				Files: []*rpc.LintFile{
					{
						FilePath: "openapi.yaml",
						Problems: []*rpc.LintProblem{
							{
								Message: "lint-error",
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
			mimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint",
			contentsProto: &rpc.Lint{
				Name: "openapi.yaml",
				Files: []*rpc.LintFile{
					{
						FilePath: "openapi.yaml",
						Problems: []*rpc.LintProblem{
							{
								Message: "lint-error",
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
			contents, _ := proto.Marshal(test.contentsProto)

			gotValue, gotErr := evaluateScoreExpression(test.expression, test.mimeType, contents)
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
		desc          string
		expression    string
		mimeType      string
		contentsProto proto.Message
	}{
		{
			desc:          "error in structProto",
			expression:    "",
			mimeType:      "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.ScoreDefinition",
			contentsProto: &rpc.ScoreDefinition{},
		},
		{
			desc:       "error in expression",
			expression: "size(files.problems)",
			mimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint",
			contentsProto: &rpc.Lint{
				Name: "openapi.yaml",
				Files: []*rpc.LintFile{
					{
						FilePath: "openapi.yaml",
						Problems: []*rpc.LintProblem{
							{
								Message: "lint-error",
							},
						},
					},
				},
			},
		},
		{
			desc:       "unsupported type",
			expression: "files[0].problems", //this will return a list value
			mimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint",
			contentsProto: &rpc.Lint{
				Name: "openapi.yaml",
				Files: []*rpc.LintFile{
					{
						FilePath: "openapi.yaml",
						Problems: []*rpc.LintProblem{
							{
								Message: "lint-error",
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

			_, gotErr := evaluateScoreExpression(test.expression, test.mimeType, contents)
			if gotErr == nil {
				t.Errorf("expected evaluateScoreExpression() to return an error")
			}
		})
	}
}
