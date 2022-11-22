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

	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	metrics "github.com/google/gnostic/metrics"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

var (
	integerDefinition = &rpc.ScoreDefinition{
		Id:             "lint-error",
		Kind:           "ScoreDefinition",
		DisplayName:    "Lint Error",
		Description:    "Number of errors found by linter",
		Uri:            "http://some/test/uri",
		UriDisplayName: "Test URI",
		Type: &rpc.ScoreDefinition_Integer{
			Integer: &rpc.IntegerType{
				MinValue: 0,
				MaxValue: 10,
				Thresholds: []*rpc.NumberThreshold{
					{
						Severity: rpc.Severity_OK,
						Range: &rpc.NumberThreshold_NumberRange{
							Min: 0,
							Max: 3,
						},
					},
					{
						Severity: rpc.Severity_WARNING,
						Range: &rpc.NumberThreshold_NumberRange{
							Min: 4,
							Max: 6,
						},
					},
					{
						Severity: rpc.Severity_ALERT,
						Range: &rpc.NumberThreshold_NumberRange{
							Min: 6,
							Max: 10,
						},
					},
				},
			},
		},
	}

	percentDefinition = &rpc.ScoreDefinition{
		Id:             "lint-error-percent",
		Kind:           "ScoreDefinition",
		DisplayName:    "Lint Error Percentage",
		Description:    "Percentage errors found by linter",
		Uri:            "http://some/test/uri",
		UriDisplayName: "Test URI",
		Type: &rpc.ScoreDefinition_Percent{
			Percent: &rpc.PercentType{
				Thresholds: []*rpc.NumberThreshold{
					{
						Severity: rpc.Severity_OK,
						Range: &rpc.NumberThreshold_NumberRange{
							Min: 0,
							Max: 30,
						},
					},
					{
						Severity: rpc.Severity_WARNING,
						Range: &rpc.NumberThreshold_NumberRange{
							Min: 31,
							Max: 60,
						},
					},
					{
						Severity: rpc.Severity_ALERT,
						Range: &rpc.NumberThreshold_NumberRange{
							Min: 61,
							Max: 100,
						},
					},
				},
			},
		},
	}

	booleanDefinition = &rpc.ScoreDefinition{
		Id:             "lint-approval",
		Kind:           "ScoreDefinition",
		DisplayName:    "Lint Approval",
		Description:    "Approval by linter",
		Uri:            "http://some/test/uri",
		UriDisplayName: "Test URI",
		Type: &rpc.ScoreDefinition_Boolean{
			Boolean: &rpc.BooleanType{
				DisplayTrue:  "Approved",
				DisplayFalse: "Denied",
				Thresholds: []*rpc.BooleanThreshold{
					{
						Severity: rpc.Severity_WARNING,
						Value:    false,
					},
					{
						Severity: rpc.Severity_OK,
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
		definitionProto *rpc.ScoreDefinition
		wantScore       *rpc.Score
	}{
		{
			desc: "nonexistent score ScoreArtifact",
			seed: []seeder.RegistryResource{
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&rpc.Lint{
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
					}),
				},
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/artifacts/lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition",
					Contents: protoMarshal(&rpc.ScoreDefinition{
						Id: "lint-error",
						TargetResource: &rpc.ResourcePattern{
							Pattern: "apis/-/versions/-/specs/-",
						},
						Formula: &rpc.ScoreDefinition_ScoreFormula{
							ScoreFormula: &rpc.ScoreFormula{
								Artifact: &rpc.ResourcePattern{
									Pattern: "$resource.spec/artifacts/lint-spectral",
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
					}),
				},
			},
			wantScore: &rpc.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DefinitionName: "projects/score-formula-test/locations/global/artifacts/lint-error",
				Severity:       rpc.Severity_SEVERITY_UNSPECIFIED,
				Value: &rpc.Score_IntegerValue{
					IntegerValue: &rpc.IntegerValue{
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
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&rpc.Lint{
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
					}),
				},
				// score artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: []byte{},
				},
				// definition artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/artifacts/lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition",
					Contents: protoMarshal(&rpc.ScoreDefinition{
						Id: "lint-error",
						TargetResource: &rpc.ResourcePattern{
							Pattern: "apis/-/versions/-/specs/-",
						},
						Formula: &rpc.ScoreDefinition_ScoreFormula{
							ScoreFormula: &rpc.ScoreFormula{
								Artifact: &rpc.ResourcePattern{
									Pattern: "$resource.spec/artifacts/lint-spectral",
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
					}),
				},
			},
			wantScore: &rpc.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DefinitionName: "projects/score-formula-test/locations/global/artifacts/lint-error",
				Severity:       rpc.Severity_SEVERITY_UNSPECIFIED,
				Value: &rpc.Score_IntegerValue{
					IntegerValue: &rpc.IntegerValue{
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
					Contents: protoMarshal(&rpc.ScoreDefinition{
						Id: "lint-error",
						TargetResource: &rpc.ResourcePattern{
							Pattern: "apis/-/versions/-/specs/-",
						},
						Formula: &rpc.ScoreDefinition_ScoreFormula{
							ScoreFormula: &rpc.ScoreFormula{
								Artifact: &rpc.ResourcePattern{
									Pattern: "$resource.spec/artifacts/lint-spectral",
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
					}),
				},
				// score artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: []byte{},
				},
				// score formula artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&rpc.Lint{
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
					}),
				},
			},
			wantScore: &rpc.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DefinitionName: "projects/score-formula-test/locations/global/artifacts/lint-error",
				Severity:       rpc.Severity_SEVERITY_UNSPECIFIED,
				Value: &rpc.Score_IntegerValue{
					IntegerValue: &rpc.IntegerValue{
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
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: []byte{},
				},
				// definition artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/artifacts/lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition",
					Contents: protoMarshal(&rpc.ScoreDefinition{
						Id: "lint-error",
						TargetResource: &rpc.ResourcePattern{
							Pattern: "apis/-/versions/-/specs/-",
						},
						Formula: &rpc.ScoreDefinition_ScoreFormula{
							ScoreFormula: &rpc.ScoreFormula{
								Artifact: &rpc.ResourcePattern{
									Pattern: "$resource.spec/artifacts/lint-spectral",
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
					}),
				},
				// score formula artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&rpc.Lint{
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
					}),
				},
			},
			wantScore: &rpc.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DefinitionName: "projects/score-formula-test/locations/global/artifacts/lint-error",
				Severity:       rpc.Severity_SEVERITY_UNSPECIFIED,
				Value: &rpc.Score_IntegerValue{
					IntegerValue: &rpc.IntegerValue{
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
			registryClient, err := connection.NewRegistryClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			t.Cleanup(func() { registryClient.Close() })

			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			t.Cleanup(func() { adminClient.Close() })

			deleteProject(ctx, adminClient, t, "score-formula-test")
			t.Cleanup(func() { deleteProject(ctx, adminClient, t, "score-formula-test") })

			client := seeder.Client{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
			}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			resource := patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
				"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error", true)
			if err != nil {
				t.Errorf("failed to get the result scoreArtifact from registry")
			}

			gotScore := &rpc.Score{}
			err = proto.Unmarshal(scoreArtifact.GetContents(), gotScore)
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
	registryClient, err := connection.NewRegistryClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	t.Cleanup(func() { registryClient.Close() })

	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	t.Cleanup(func() { adminClient.Close() })

	deleteProject(ctx, adminClient, t, "score-formula-test")
	t.Cleanup(func() { deleteProject(ctx, adminClient, t, "score-formula-test") })

	client := seeder.Client{
		RegistryClient: registryClient,
		AdminClient:    adminClient,
	}

	seed := []seeder.RegistryResource{
		&rpc.Artifact{
			Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
			MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
			Contents: protoMarshal(&rpc.Lint{
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
			}),
		},
	}

	if err := seeder.SeedRegistry(ctx, client, seed...); err != nil {
		t.Fatalf("Setup: failed to seed registry: %s", err)
	}

	// arguments
	formula := &rpc.ScoreFormula{
		Artifact: &rpc.ResourcePattern{
			Pattern: "$resource.spec/artifacts/lint-spectral",
		},
		ScoreExpression: "size(files[0].problems)",
	}
	resource := patterns.SpecResource{
		Spec: &rpc.ApiSpec{
			Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
		formula  *rpc.ScoreFormula
		resource patterns.ResourceInstance
	}{
		{
			desc: "invalid reference",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.specs/artifacts/lint-spectral", //error
				},
				ScoreExpression: "size(files[0].problems)",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
		{
			desc: "invalid extended pattern",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifact/lint-spectral", // error
				},
				ScoreExpression: "size(files[0].problems)",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
		{
			desc: "missing artifact",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problems)",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
		{
			desc: "unsupported artifact type",
			seed: []seeder.RegistryResource{
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-definition",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.ScoreDefinition",
					Contents: protoMarshal(&rpc.ScoreDefinition{
						Id:             "dummy-score-definition",
						TargetResource: &rpc.ResourcePattern{},
						Formula:        nil,
						Type:           nil,
					}),
				},
			},
			formula: &rpc.ScoreFormula{},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
		{
			desc: "invalid expression",
			seed: []seeder.RegistryResource{
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&rpc.Lint{
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
					}),
				},
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problem)", // invalid expression
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
		{
			desc: "missing expression",
			seed: []seeder.RegistryResource{
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&rpc.Lint{
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
					}),
				},
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewRegistryClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			t.Cleanup(func() { registryClient.Close() })

			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			t.Cleanup(func() { adminClient.Close() })

			deleteProject(ctx, adminClient, t, "score-formula-test")
			t.Cleanup(func() { deleteProject(ctx, adminClient, t, "score-formula-test") })

			client := seeder.Client{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
			}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

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
	registryClient, err := connection.NewRegistryClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	t.Cleanup(func() { registryClient.Close() })

	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	t.Cleanup(func() { adminClient.Close() })

	deleteProject(ctx, adminClient, t, "rollup-formula-test")
	t.Cleanup(func() { deleteProject(ctx, adminClient, t, "rollup-formula-test") })

	client := seeder.Client{
		RegistryClient: registryClient,
		AdminClient:    adminClient,
	}

	seed := []seeder.RegistryResource{
		// lint artifact
		&rpc.Artifact{
			Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
			MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
			Contents: protoMarshal(&rpc.Lint{
				Name: "openapi.yaml",
				Files: []*rpc.LintFile{
					{
						FilePath: "openapi.yaml",
						Problems: []*rpc.LintProblem{
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
			Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
			MimeType: "application/octet-stream;type=gnostic.metrics.Complexity",
			Contents: protoMarshal(&metrics.Complexity{
				GetCount:    1,
				PostCount:   1,
				PutCount:    1,
				DeleteCount: 1,
			}),
		},
	}

	if err := seeder.SeedRegistry(ctx, client, seed...); err != nil {
		t.Fatalf("Setup: failed to seed registry: %s", err)
	}

	// arguments
	formula := &rpc.RollUpFormula{
		ScoreFormulas: []*rpc.ScoreFormula{
			{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problems)",
				ReferenceId:     "numErrors",
			},
			{
				Artifact: &rpc.ResourcePattern{
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
			Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
		formula  *rpc.RollUpFormula
		resource patterns.ResourceInstance
	}{
		{
			desc: "missing score_formulas",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			formula: &rpc.RollUpFormula{
				RollupExpression: "double(numErrors)/numOperations",
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
		{
			desc: "missing rollup_expression",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			formula: &rpc.RollUpFormula{
				ScoreFormulas: []*rpc.ScoreFormula{
					{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-spectral",
						},
						ScoreExpression: "size(files[0].problems)",
						ReferenceId:     "numErrors",
					},
					{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/complexity",
						},
						ScoreExpression: "getCount + postCount + putCount + deleteCount",
						ReferenceId:     "numOperations",
					},
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
		{
			desc: "invalid score_expression",
			seed: []seeder.RegistryResource{
				// lint artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&rpc.Lint{
						Name: "openapi.yaml",
						Files: []*rpc.LintFile{
							{
								FilePath: "openapi.yaml",
								Problems: []*rpc.LintProblem{
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
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					MimeType: "application/octet-stream;type=gnostic.metrics.Complexity",
					Contents: protoMarshal(&metrics.Complexity{
						GetCount:    1,
						PostCount:   1,
						PutCount:    1,
						DeleteCount: 1,
					}),
				},
			},
			formula: &rpc.RollUpFormula{
				ScoreFormulas: []*rpc.ScoreFormula{
					{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-spectral",
						},
						ScoreExpression: "size(files.problems)", // invalid field reference
						ReferenceId:     "numErrors",
					},
					{
						Artifact: &rpc.ResourcePattern{
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
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
		{
			desc: "invalid rollup_expression",
			seed: []seeder.RegistryResource{
				// lint artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.Lint",
					Contents: protoMarshal(&rpc.Lint{
						Name: "openapi.yaml",
						Files: []*rpc.LintFile{
							{
								FilePath: "openapi.yaml",
								Problems: []*rpc.LintProblem{
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
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					MimeType: "application/octet-stream;type=gnostic.metrics.Complexity",
					Contents: protoMarshal(&metrics.Complexity{
						GetCount:    1,
						PostCount:   1,
						PutCount:    1,
						DeleteCount: 1,
					}),
				},
			},
			formula: &rpc.RollUpFormula{
				ScoreFormulas: []*rpc.ScoreFormula{
					{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-spectral",
						},
						ScoreExpression: "size(files[0].problems)",
						ReferenceId:     "numErrors",
					},
					{
						Artifact: &rpc.ResourcePattern{
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
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
		{
			desc: "invalid reference_id",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			formula: &rpc.RollUpFormula{
				ScoreFormulas: []*rpc.ScoreFormula{
					{
						Artifact: &rpc.ResourcePattern{
							Pattern: "$resource.spec/artifacts/lint-spectral",
						},
						ScoreExpression: "size(files[0].problems)",
						ReferenceId:     "num-errors",
					},
					{
						Artifact: &rpc.ResourcePattern{
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
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewRegistryClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			t.Cleanup(func() { registryClient.Close() })

			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			t.Cleanup(func() { adminClient.Close() })

			deleteProject(ctx, adminClient, t, "rollup-formula-test")
			t.Cleanup(func() { deleteProject(ctx, adminClient, t, "rollup-formula-test") })

			client := seeder.Client{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
			}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

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
		definition *rpc.ScoreDefinition
		scoreValue interface{}
		wantScore  *rpc.Score
	}{
		{
			desc:       "happy path integer",
			definition: integerDefinition,
			scoreValue: int64(1),
			wantScore: &rpc.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DisplayName:    "Lint Error",
				Description:    "Number of errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error",
				Severity:       rpc.Severity_OK,
				Value: &rpc.Score_IntegerValue{
					IntegerValue: &rpc.IntegerValue{
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
			wantScore: &rpc.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DisplayName:    "Lint Error",
				Description:    "Number of errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error",
				Severity:       rpc.Severity_OK,
				Value: &rpc.Score_IntegerValue{
					IntegerValue: &rpc.IntegerValue{
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
			wantScore: &rpc.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DisplayName:    "Lint Error",
				Description:    "Number of errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error",
				Severity:       rpc.Severity_ALERT,
				Value: &rpc.Score_IntegerValue{
					IntegerValue: &rpc.IntegerValue{
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
			wantScore: &rpc.Score{
				Id:             "score-lint-error",
				Kind:           "Score",
				DisplayName:    "Lint Error",
				Description:    "Number of errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error",
				Severity:       rpc.Severity_ALERT,
				Value: &rpc.Score_IntegerValue{
					IntegerValue: &rpc.IntegerValue{
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
			wantScore: &rpc.Score{
				Id:             "score-lint-error-percent",
				Kind:           "Score",
				DisplayName:    "Lint Error Percentage",
				Description:    "Percentage errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error-percent",
				Severity:       rpc.Severity_WARNING,
				Value: &rpc.Score_PercentValue{
					PercentValue: &rpc.PercentValue{
						Value: 50,
					},
				},
			},
		},
		{
			desc:       "happy path percent with integer value",
			definition: percentDefinition,
			scoreValue: int64(50),
			wantScore: &rpc.Score{
				Id:             "score-lint-error-percent",
				Kind:           "Score",
				DisplayName:    "Lint Error Percentage",
				Description:    "Percentage errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error-percent",
				Severity:       rpc.Severity_WARNING,
				Value: &rpc.Score_PercentValue{
					PercentValue: &rpc.PercentValue{
						Value: 50,
					},
				},
			},
		},
		{
			desc:       "greater than max percent",
			definition: percentDefinition,
			scoreValue: int64(101),
			wantScore: &rpc.Score{
				Id:             "score-lint-error-percent",
				Kind:           "Score",
				DisplayName:    "Lint Error Percentage",
				Description:    "Percentage errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error-percent",
				Severity:       rpc.Severity_ALERT,
				Value: &rpc.Score_PercentValue{
					PercentValue: &rpc.PercentValue{
						Value: 101,
					},
				},
			},
		},
		{
			desc:       "less than min percent",
			definition: percentDefinition,
			scoreValue: int64(-1),
			wantScore: &rpc.Score{
				Id:             "score-lint-error-percent",
				Kind:           "Score",
				DisplayName:    "Lint Error Percentage",
				Description:    "Percentage errors found by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-error-percent",
				Severity:       rpc.Severity_ALERT,
				Value: &rpc.Score_PercentValue{
					PercentValue: &rpc.PercentValue{
						Value: -1,
					},
				},
			},
		},
		{
			desc:       "happy path boolean",
			definition: booleanDefinition,
			scoreValue: true,
			wantScore: &rpc.Score{
				Id:             "score-lint-approval",
				Kind:           "Score",
				DisplayName:    "Lint Approval",
				Description:    "Approval by linter",
				Uri:            "http://some/test/uri",
				UriDisplayName: "Test URI",
				DefinitionName: "projects/score-type-test/locations/global/artifacts/lint-approval",
				Severity:       rpc.Severity_OK,
				Value: &rpc.Score_BooleanValue{
					BooleanValue: &rpc.BooleanValue{
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
		definition *rpc.ScoreDefinition
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
