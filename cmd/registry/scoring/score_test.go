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
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/pkg/application/scoring"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	metrics "github.com/google/gnostic/metrics"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

var (
	integerDefinition = &scoring.ScoreDefinition{
		Id:             "lint-error",
		Kind:           "ScoreDefinition",
		DisplayName:    "Lint Error",
		Description:    "Number of errors found by linter",
		Uri:            "http://some/test/uri",
		UriDisplayName: "Test URI",
		Type: &scoring.ScoreDefinition_Integer{
			Integer: &scoring.IntegerType{
				MinValue: 0,
				MaxValue: 10,
				Thresholds: []*scoring.NumberThreshold{
					{
						Severity: scoring.Severity_OK,
						Range: &scoring.NumberThreshold_NumberRange{
							Min: 0,
							Max: 3,
						},
					},
					{
						Severity: scoring.Severity_WARNING,
						Range: &scoring.NumberThreshold_NumberRange{
							Min: 4,
							Max: 6,
						},
					},
					{
						Severity: scoring.Severity_ALERT,
						Range: &scoring.NumberThreshold_NumberRange{
							Min: 6,
							Max: 10,
						},
					},
				},
			},
		},
	}

	percentDefinition = &scoring.ScoreDefinition{
		Id:             "lint-error-percent",
		Kind:           "ScoreDefinition",
		DisplayName:    "Lint Error Percentage",
		Description:    "Percentage errors found by linter",
		Uri:            "http://some/test/uri",
		UriDisplayName: "Test URI",
		Type: &scoring.ScoreDefinition_Percent{
			Percent: &scoring.PercentType{
				Thresholds: []*scoring.NumberThreshold{
					{
						Severity: scoring.Severity_OK,
						Range: &scoring.NumberThreshold_NumberRange{
							Min: 0,
							Max: 30,
						},
					},
					{
						Severity: scoring.Severity_WARNING,
						Range: &scoring.NumberThreshold_NumberRange{
							Min: 31,
							Max: 60,
						},
					},
					{
						Severity: scoring.Severity_ALERT,
						Range: &scoring.NumberThreshold_NumberRange{
							Min: 61,
							Max: 100,
						},
					},
				},
			},
		},
	}

	booleanDefinition = &scoring.ScoreDefinition{
		Id:             "lint-approval",
		Kind:           "ScoreDefinition",
		DisplayName:    "Lint Approval",
		Description:    "Approval by linter",
		Uri:            "http://some/test/uri",
		UriDisplayName: "Test URI",
		Type: &scoring.ScoreDefinition_Boolean{
			Boolean: &scoring.BooleanType{
				DisplayTrue:  "Approved",
				DisplayFalse: "Denied",
				Thresholds: []*scoring.BooleanThreshold{
					{
						Severity: scoring.Severity_WARNING,
						Value:    false,
					},
					{
						Severity: scoring.Severity_OK,
						Value:    true,
					},
				},
			},
		},
	}
)

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		desc            string
		seed            []seeder.RegistryResource
		definitionProto *scoring.ScoreDefinition
		wantScore       *scoring.Score
	}{
		{
			desc: "nonexistent score ScoreArtifact",
			seed: []seeder.RegistryResource{
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&style.Lint{
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
					}),
				},
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/artifacts/lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition",
					Contents: protoMarshal(&scoring.ScoreDefinition{
						Id: "lint-error",
						TargetResource: &scoring.ResourcePattern{
							Pattern: "apis/-/versions/-/specs/-",
						},
						Formula: &scoring.ScoreDefinition_ScoreFormula{
							ScoreFormula: &scoring.ScoreFormula{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.spec/artifacts/lint-spectral",
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
					}),
				},
			},
			wantScore: &scoring.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DefinitionName: "projects/score-formula-test/locations/global/artifacts/lint-error",
				Severity:       scoring.Severity_SEVERITY_UNSPECIFIED,
				Value: &scoring.Score_IntegerValue{
					IntegerValue: &scoring.IntegerValue{
						Value:    1,
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
		},
		{
			desc: "existing score updated definition",
			seed: []seeder.RegistryResource{
				// score formula artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&style.Lint{
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
					}),
				},
				// score artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: []byte{},
				},
				// definition artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/artifacts/lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition",
					Contents: protoMarshal(&scoring.ScoreDefinition{
						Id: "lint-error",
						TargetResource: &scoring.ResourcePattern{
							Pattern: "apis/-/versions/-/specs/-",
						},
						Formula: &scoring.ScoreDefinition_ScoreFormula{
							ScoreFormula: &scoring.ScoreFormula{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.spec/artifacts/lint-spectral",
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
					}),
				},
			},
			wantScore: &scoring.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DefinitionName: "projects/score-formula-test/locations/global/artifacts/lint-error",
				Severity:       scoring.Severity_SEVERITY_UNSPECIFIED,
				Value: &scoring.Score_IntegerValue{
					IntegerValue: &scoring.IntegerValue{
						Value:    1,
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
		},
		{
			desc: "existing score updated formula artifact",
			seed: []seeder.RegistryResource{
				// definition artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/artifacts/lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition",
					Contents: protoMarshal(&scoring.ScoreDefinition{
						Id: "lint-error",
						TargetResource: &scoring.ResourcePattern{
							Pattern: "apis/-/versions/-/specs/-",
						},
						Formula: &scoring.ScoreDefinition_ScoreFormula{
							ScoreFormula: &scoring.ScoreFormula{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.spec/artifacts/lint-spectral",
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
					}),
				},
				// score artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: []byte{},
				},
				// score formula artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&style.Lint{
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
					}),
				},
			},
			wantScore: &scoring.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DefinitionName: "projects/score-formula-test/locations/global/artifacts/lint-error",
				Severity:       scoring.Severity_SEVERITY_UNSPECIFIED,
				Value: &scoring.Score_IntegerValue{
					IntegerValue: &scoring.IntegerValue{
						Value:    1,
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
		},
		{
			desc: "existing score updated formula artifact and definition",
			seed: []seeder.RegistryResource{
				// score artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: []byte{},
				},
				// definition artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/artifacts/lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition",
					Contents: protoMarshal(&scoring.ScoreDefinition{
						Id: "lint-error",
						TargetResource: &scoring.ResourcePattern{
							Pattern: "apis/-/versions/-/specs/-",
						},
						Formula: &scoring.ScoreDefinition_ScoreFormula{
							ScoreFormula: &scoring.ScoreFormula{
								Artifact: &scoring.ResourcePattern{
									Pattern: "$resource.spec/artifacts/lint-spectral",
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
					}),
				},
				// score formula artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&style.Lint{
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
					}),
				},
			},
			wantScore: &scoring.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DefinitionName: "projects/score-formula-test/locations/global/artifacts/lint-error",
				Severity:       scoring.Severity_SEVERITY_UNSPECIFIED,
				Value: &scoring.Score_IntegerValue{
					IntegerValue: &scoring.IntegerValue{
						Value:    1,
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, _ := grpctest.SetupRegistry(ctx, t, "score-formula-test", test.seed)

			resource := patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			}

			artifactClient := &RegistryArtifactClient{RegistryClient: registryClient}

			//fetch definition artifact
			defArtifact, err := getArtifact(ctx, artifactClient, "projects/score-formula-test/locations/global/artifacts/lint-error", true)
			if err != nil {
				t.Errorf("failed to fetch the definition Artifact from setup: %s", err)
			}

			gotErr := CalculateScore(ctx, artifactClient, defArtifact, resource, false)
			if gotErr != nil {
				t.Errorf("CalculateScore(ctx, client, %v, %v) returned unexpected error: %s", defArtifact, resource, gotErr)
			}

			//fetch score artifact and check the value
			scoreArtifact, err := getArtifact(
				ctx, artifactClient,
				"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-lint-error", true)
			if err != nil {
				t.Errorf("failed to get the result scoreArtifact from registry")
			}

			gotScore := &scoring.Score{}
			err = patch.UnmarshalContents(scoreArtifact.GetContents(), scoreArtifact.GetMimeType(), gotScore)
			if err != nil {
				t.Errorf("failed unmarshalling score artifact from registry: %s", err)
			}

			opts := cmp.Options{protocmp.Transform()}
			if !cmp.Equal(test.wantScore, gotScore, opts) {
				t.Errorf("CalculateScore() returned unexpected response (-want +got):\n%s", cmp.Diff(test.wantScore, gotScore, opts))
			}
		})
	}
}

func TestProcessScoreFormula(t *testing.T) {
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "score-formula-test", []seeder.RegistryResource{
		&rpc.Artifact{
			Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
			MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
			Contents: protoMarshal(&style.Lint{
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
			}),
		},
	})

	// arguments
	formula := &scoring.ScoreFormula{
		Artifact: &scoring.ResourcePattern{
			Pattern: "$resource.spec/artifacts/lint-spectral",
		},
		ScoreExpression: "size(files[0].problems)",
	}
	resource := patterns.SpecResource{
		Spec: &rpc.ApiSpec{
			Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
	}

	wantResult := scoreResult{
		value:       int64(1),
		needsUpdate: true,
		err:         nil,
	}

	artifactClient := &RegistryArtifactClient{RegistryClient: registryClient}

	gotResult := processScoreFormula(ctx, artifactClient, formula, resource, &rpc.Artifact{}, true)

	opts := cmp.AllowUnexported(scoreResult{})
	if !cmp.Equal(wantResult, gotResult, opts) {
		t.Errorf("processScoreFormula() returned unexpected response, (-want +got):\n%s", cmp.Diff(wantResult, gotResult, opts))
	}
}

func TestProcessScoreFormulaError(t *testing.T) {
	tests := []struct {
		desc     string
		seed     []seeder.RegistryResource
		setup    func(context.Context, connection.RegistryClient, connection.AdminClient)
		formula  *scoring.ScoreFormula
		resource patterns.ResourceInstance
	}{
		{
			desc: "invalid reference",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
			formula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.specs/artifacts/lint-spectral", //error
				},
				ScoreExpression: "size(files[0].problems)",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
		{
			desc: "invalid extended pattern",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
			formula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifact/lint-spectral", // error
				},
				ScoreExpression: "size(files[0].problems)",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
		{
			desc: "missing artifact",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
			formula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problems)",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
		{
			desc: "unsupported artifact type",
			seed: []seeder.RegistryResource{
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-definition",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.ScoreDefinition",
					Contents: protoMarshal(&scoring.ScoreDefinition{
						Id:             "dummy-score-definition",
						TargetResource: &scoring.ResourcePattern{},
						Formula:        nil,
						Type:           nil,
					}),
				},
			},
			formula: &scoring.ScoreFormula{},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
		{
			desc: "invalid expression",
			seed: []seeder.RegistryResource{
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&style.Lint{
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
					}),
				},
			},
			formula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problem)", // invalid expression
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
		{
			desc: "missing expression",
			seed: []seeder.RegistryResource{
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&style.Lint{
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
					}),
				},
			},
			formula: &scoring.ScoreFormula{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, _ := grpctest.SetupRegistry(ctx, t, "score-formula-test", test.seed)

			artifactClient := &RegistryArtifactClient{RegistryClient: registryClient}

			gotResult := processScoreFormula(ctx, artifactClient, test.formula, test.resource, &rpc.Artifact{}, true)
			if gotResult.err == nil {
				t.Errorf("processScoreFormula(ctx, client, %v, %v) did not return an error", test.formula, test.resource)
			}
		})
	}
}

func TestProcessRollUpFormula(t *testing.T) {
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "score-formula-test", []seeder.RegistryResource{
		// lint artifact
		&rpc.Artifact{
			Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
			MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
			Contents: protoMarshal(&style.Lint{
				Name: "openapi.yaml",
				Files: []*style.LintFile{
					{
						FilePath: "openapi.yaml",
						Problems: []*style.LintProblem{
							{
								Message: "lint-error",
							},
							{
								Message: "lint-error",
							},
						},
					},
				},
			}),
		},
		// complexity artifact
		&rpc.Artifact{
			Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity",
			MimeType: "application/octet-stream;type=gnostic.metrics.Complexity",
			Contents: protoMarshal(&metrics.Complexity{
				GetCount:    1,
				PostCount:   1,
				PutCount:    1,
				DeleteCount: 1,
			}),
		},
	})

	// arguments
	formula := &scoring.RollUpFormula{
		ScoreFormulas: []*scoring.ScoreFormula{
			{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problems)",
				ReferenceId:     "numErrors",
			},
			{
				Artifact: &scoring.ResourcePattern{
					Pattern: "$resource.spec/artifacts/complexity",
				},
				ScoreExpression: "getCount + postCount + putCount + deleteCount",
				ReferenceId:     "numOperations",
			},
		},
		RollupExpression: "double(numErrors)/numOperations",
	}
	resource := patterns.SpecResource{
		Spec: &rpc.ApiSpec{
			Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
		},
	}

	wantResult := scoreResult{
		value:       float64(0.5),
		needsUpdate: true,
		err:         nil,
	}

	artifactClient := &RegistryArtifactClient{RegistryClient: registryClient}

	gotResult := processRollUpFormula(ctx, artifactClient, formula, resource, &rpc.Artifact{}, true)

	opts := cmp.AllowUnexported(scoreResult{})
	if !cmp.Equal(wantResult, gotResult, opts) {
		t.Errorf("processRollUpFormula() returned unexpected value, (-want, +got):\n%s", cmp.Diff(wantResult, gotResult, opts))
	}
}

func TestProcessRollUpFormulaError(t *testing.T) {
	tests := []struct {
		desc     string
		seed     []seeder.RegistryResource
		formula  *scoring.RollUpFormula
		resource patterns.ResourceInstance
	}{
		{
			desc: "missing score_formulas",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
			formula: &scoring.RollUpFormula{
				RollupExpression: "double(numErrors)/numOperations",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
		{
			desc: "missing rollup_expression",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
			formula: &scoring.RollUpFormula{
				ScoreFormulas: []*scoring.ScoreFormula{
					{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-spectral",
						},
						ScoreExpression: "size(files[0].problems)",
						ReferenceId:     "numErrors",
					},
					{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/complexity",
						},
						ScoreExpression: "getCount + postCount + putCount + deleteCount",
						ReferenceId:     "numOperations",
					},
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
		{
			desc: "invalid score_expression",
			seed: []seeder.RegistryResource{
				// lint artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&style.Lint{
						Name: "openapi.yaml",
						Files: []*style.LintFile{
							{
								FilePath: "openapi.yaml",
								Problems: []*style.LintProblem{
									{
										Message: "lint-error",
									},
									{
										Message: "lint-error",
									},
								},
							},
						},
					}),
				},
				// complexity artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity",
					MimeType: "application/octet-stream;type=gnostic.metrics.Complexity",
					Contents: protoMarshal(&metrics.Complexity{
						GetCount:    1,
						PostCount:   1,
						PutCount:    1,
						DeleteCount: 1,
					}),
				},
			},
			formula: &scoring.RollUpFormula{
				ScoreFormulas: []*scoring.ScoreFormula{
					{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-spectral",
						},
						ScoreExpression: "size(files.problems)", // invalid field reference
						ReferenceId:     "numErrors",
					},
					{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/complexity",
						},
						ScoreExpression: "getCount + postCount + putCount + deleteCount",
						ReferenceId:     "numOperations",
					},
				},
				RollupExpression: "double(numErrors)/numOperations",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
		{
			desc: "invalid rollup_expression",
			seed: []seeder.RegistryResource{
				// lint artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&style.Lint{
						Name: "openapi.yaml",
						Files: []*style.LintFile{
							{
								FilePath: "openapi.yaml",
								Problems: []*style.LintProblem{
									{
										Message: "lint-error",
									},
									{
										Message: "lint-error",
									},
								},
							},
						},
					}),
				},
				// complexity artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity",
					MimeType: "application/octet-stream;type=gnostic.metrics.Complexity",
					Contents: protoMarshal(&metrics.Complexity{
						GetCount:    1,
						PostCount:   1,
						PutCount:    1,
						DeleteCount: 1,
					}),
				},
			},
			formula: &scoring.RollUpFormula{
				ScoreFormulas: []*scoring.ScoreFormula{
					{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-spectral",
						},
						ScoreExpression: "size(files[0].problems)",
						ReferenceId:     "numErrors",
					},
					{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/complexity",
						},
						ScoreExpression: "getCount + postCount + putCount + deleteCount",
						ReferenceId:     "numOperations",
					},
				},
				RollupExpression: "numError/numOperation", // should be numErrors/numOperations
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
		{
			desc: "invalid reference_id",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
			formula: &scoring.RollUpFormula{
				ScoreFormulas: []*scoring.ScoreFormula{
					{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-spectral",
						},
						ScoreExpression: "size(files[0].problems)",
						ReferenceId:     "num-errors",
					},
					{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/complexity",
						},
						ScoreExpression: "getCount + postCount + putCount + deleteCount",
						ReferenceId:     "num-operations",
					},
				},
				RollupExpression: "num-errors/num-operations",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, _ := grpctest.SetupRegistry(ctx, t, "rollup-formula-test", test.seed)

			artifactClient := &RegistryArtifactClient{RegistryClient: registryClient}

			gotResult := processRollUpFormula(ctx, artifactClient, test.formula, test.resource, &rpc.Artifact{}, true)
			if gotResult.err == nil {
				t.Errorf("processRollUpFormula(ctx, client, %v, %v) did not return an error", test.formula, test.resource)
			}
		})
	}
}

func TestProcessScoreType(t *testing.T) {
	tests := []struct {
		desc       string
		definition *scoring.ScoreDefinition
		scoreValue interface{}
		wantScore  *scoring.Score
	}{
		{
			desc:       "happy path integer",
			definition: integerDefinition,
			scoreValue: int64(1),
			wantScore: &scoring.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DisplayName:    "Lint Error",
				Description:    "Number of errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error",
				Severity:       scoring.Severity_OK,
				Value: &scoring.Score_IntegerValue{
					IntegerValue: &scoring.IntegerValue{
						Value:    1,
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
		},
		{
			desc:       "happy path integer with float value",
			definition: integerDefinition,
			scoreValue: float64(1),
			wantScore: &scoring.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DisplayName:    "Lint Error",
				Description:    "Number of errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error",
				Severity:       scoring.Severity_OK,
				Value: &scoring.Score_IntegerValue{
					IntegerValue: &scoring.IntegerValue{
						Value:    1,
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
		},
		{
			desc:       "greater than max integer",
			definition: integerDefinition,
			scoreValue: int64(11),
			wantScore: &scoring.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DisplayName:    "Lint Error",
				Description:    "Number of errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error",
				Severity:       scoring.Severity_ALERT,
				Value: &scoring.Score_IntegerValue{
					IntegerValue: &scoring.IntegerValue{
						Value:    11,
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
		},
		{
			desc:       "less than min integer",
			definition: integerDefinition,
			scoreValue: int64(-1),
			wantScore: &scoring.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DisplayName:    "Lint Error",
				Description:    "Number of errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error",
				Severity:       scoring.Severity_ALERT,
				Value: &scoring.Score_IntegerValue{
					IntegerValue: &scoring.IntegerValue{
						Value:    -1,
						MinValue: 0,
						MaxValue: 10,
					},
				},
			},
		},
		{
			desc:       "happy path percent",
			definition: percentDefinition,
			scoreValue: float64(50),
			wantScore: &scoring.Score{
				Id:             "score-lint-error-percent",
				Kind:           "Score",
				DisplayName:    "Lint Error Percentage",
				Description:    "Percentage errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error-percent",
				Severity:       scoring.Severity_WARNING,
				Value: &scoring.Score_PercentValue{
					PercentValue: &scoring.PercentValue{
						Value: 50,
					},
				},
			},
		},
		{
			desc:       "happy path percent with integer value",
			definition: percentDefinition,
			scoreValue: int64(50),
			wantScore: &scoring.Score{
				Id:             "score-lint-error-percent",
				Kind:           "Score",
				DisplayName:    "Lint Error Percentage",
				Description:    "Percentage errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error-percent",
				Severity:       scoring.Severity_WARNING,
				Value: &scoring.Score_PercentValue{
					PercentValue: &scoring.PercentValue{
						Value: 50,
					},
				},
			},
		},
		{
			desc:       "greater than max percent",
			definition: percentDefinition,
			scoreValue: int64(101),
			wantScore: &scoring.Score{
				Id:             "score-lint-error-percent",
				Kind:           "Score",
				DisplayName:    "Lint Error Percentage",
				Description:    "Percentage errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error-percent",
				Severity:       scoring.Severity_ALERT,
				Value: &scoring.Score_PercentValue{
					PercentValue: &scoring.PercentValue{
						Value: 101,
					},
				},
			},
		},
		{
			desc:       "less than min percent",
			definition: percentDefinition,
			scoreValue: int64(-1),
			wantScore: &scoring.Score{
				Id:             "score-lint-error-percent",
				Kind:           "Score",
				DisplayName:    "Lint Error Percentage",
				Description:    "Percentage errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error-percent",
				Severity:       scoring.Severity_ALERT,
				Value: &scoring.Score_PercentValue{
					PercentValue: &scoring.PercentValue{
						Value: -1,
					},
				},
			},
		},
		{
			desc:       "happy path boolean",
			definition: booleanDefinition,
			scoreValue: true,
			wantScore: &scoring.Score{
				Id:             "score-lint-approval",
				Kind:           "Score",
				DisplayName:    "Lint Approval",
				Description:    "Approval by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-approval",
				Severity:       scoring.Severity_OK,
				Value: &scoring.Score_BooleanValue{
					BooleanValue: &scoring.BooleanValue{
						Value:        true,
						DisplayValue: "Approved",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotScore, gotErr := processScoreType(test.definition, test.scoreValue, "projects/score-type-test/locations/global")
			if gotErr != nil {
				t.Errorf("processScoreType() returned unexpected error: %s", gotErr)
			}

			opts := cmp.Options{protocmp.Transform()}
			if !cmp.Equal(test.wantScore, gotScore, opts) {
				t.Errorf("processScoreType() returned unexpected response (-want +got):\n%s", cmp.Diff(test.wantScore, gotScore, opts))
			}
		})
	}
}

func TestProcessScoreTypeError(t *testing.T) {
	tests := []struct {
		desc       string
		definition *scoring.ScoreDefinition
		scoreValue interface{}
	}{
		{
			desc:       "type mismatch integer",
			definition: integerDefinition,
			scoreValue: true,
		},
		{
			desc:       "type mismatch percent",
			definition: percentDefinition,
			scoreValue: false,
		},
		{
			desc:       "type mismatch boolean",
			definition: booleanDefinition,
			scoreValue: int64(1),
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, gotErr := processScoreType(test.definition, test.scoreValue, "projects/score-type-test/locations/global")
			if gotErr == nil {
				t.Errorf("processScoreType(%v, %v, %s) did not return an error", test.definition, test.scoreValue, "projects/score-type-test/locations/global")
			}
		})
	}
}
