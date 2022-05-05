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
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/golang/protobuf/proto"
	metrics "github.com/google/gnostic/metrics"
	"github.com/google/go-cmp/cmp"
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
		setup           func(context.Context, connection.Client, connection.AdminClient)
		definitionProto *rpc.ScoreDefinition
		wantScore       *rpc.Score
	}{
		{
			desc: "non existent score ScoreArtifact",
			setup: func(ctx context.Context, registryClient connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, registryClient, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
				defBytes, _ := proto.Marshal(&rpc.ScoreDefinition{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/artifacts/lint-error",
					defBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition")
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
			desc: "existing up-to-date score",
			setup: func(ctx context.Context, registryClient connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, registryClient, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create formula artifact
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
				// create definition artifact
				defBytes, _ := proto.Marshal(&rpc.ScoreDefinition{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/artifacts/lint-error",
					defBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition")
				// create score artifact
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			wantScore: &rpc.Score{},
		},
		{
			desc: "existing score updated definition",
			setup: func(ctx context.Context, registryClient connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, registryClient, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create formula artifact
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
				// create score artifact
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// create definition artifact
				defBytes, _ := proto.Marshal(&rpc.ScoreDefinition{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/artifacts/lint-error",
					defBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition")
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
			setup: func(ctx context.Context, registryClient connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, registryClient, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create definition artifact
				defBytes, _ := proto.Marshal(&rpc.ScoreDefinition{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/artifacts/lint-error",
					defBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition")
				// create score artifact
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// create formula artifact
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
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
			setup: func(ctx context.Context, registryClient connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, registryClient, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create score artifact
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// create definition artifact
				defBytes, _ := proto.Marshal(&rpc.ScoreDefinition{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/artifacts/lint-error",
					defBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreDefinition")
				// create formula artifact
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, registryClient, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
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
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			resource := patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			}

			//fetch definition artifact
			defArtifact, err := getArtifact(ctx, registryClient, "projects/score-formula-test/locations/global/artifacts/lint-error", true)
			if err != nil {
				t.Errorf("failed to fetch the scoreArtifact from setup: %s", err)
			}

			gotErr := CalculateScore(ctx, registryClient, defArtifact, resource)
			if gotErr != nil {
				t.Errorf("CalculateScore(ctx, client, %v, %v) returned unexpected error: %s", defArtifact, resource, gotErr)
			}

			//fetch score artifact and check the value
			scoreArtifact, err := getArtifact(
				ctx, registryClient,
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

func TestMatchResourceWithTarget(t *testing.T) {
	tests := []struct {
		desc          string
		targetPattern *rpc.ResourcePattern
		resourceName  string
		wantErr       bool
	}{
		{
			desc: "spec pattern",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
			},
			resourceName: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		},
		{
			desc: "specific api match",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "apis/petstore/versions/-/specs/-",
			},
			resourceName: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		},
		{
			desc: "specific api no match",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "apis/test/versions/-/specs/-",
			},
			resourceName: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			wantErr:      true,
		},
		{
			desc: "specific version match",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "apis/-/versions/1.0.0/specs/-",
			},
			resourceName: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		},
		{
			desc: "specific version no match",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "apis/-/versions/2.0.0/specs/-",
			},
			resourceName: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			wantErr:      true,
		},
		{
			desc: "specific spec match",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/openapi.yaml",
			},
			resourceName: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
		},
		{
			desc: "specific spec no match",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/swagger.yaml",
			},
			resourceName: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			wantErr:      true,
		},
		{
			desc: "artifact pattern error",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-/artifacts/lint-spectral",
			},
			resourceName: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
			wantErr:      true,
		},
		{
			desc: "target and resource mismatch",
			targetPattern: &rpc.ResourcePattern{
				Pattern: "apis/-/versions/-/specs/-",
			},
			resourceName: "projects/pattern-test/locations/global/apis/petstore/versions/1.0.0",
			wantErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			resourceName, _ := patterns.ParseResourcePattern(test.resourceName)
			gotErr := matchResourceWithTarget(test.targetPattern, resourceName, "projects/pattern-test/locations/global")
			if test.wantErr && gotErr == nil {
				t.Errorf("matchResourceWithTarget(%s, %v, %s) did not return an error", test.targetPattern, resourceName, "projects/pattern-test/locations/global")
			}

			if !test.wantErr && gotErr != nil {
				t.Errorf("matchResourceWithTarget() returned unexpected error: %s", gotErr)
			}
		})
	}
}

func TestProcessScoreFormula(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	defer registryClient.Close()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	defer adminClient.Close()

	//setup
	deleteProject(ctx, adminClient, t, "score-formula-test")
	createProject(ctx, adminClient, t, "score-formula-test")
	createApi(ctx, registryClient, t, "projects/score-formula-test/locations/global", "petstore")
	createVersion(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
	})
	createUpdateArtifact(
		ctx, registryClient, t,
		"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
		artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")

	// arguments
	formula := &rpc.ScoreFormula{
		Artifact: &rpc.ResourcePattern{
			Pattern: "$resource.spec/artifacts/lint-spectral",
		},
		ScoreExpression: "size(files[0].problems)",
	}
	resource := patterns.SpecResource{
		SpecName: patterns.SpecName{
			Name: names.Spec{
				ProjectID: "score-formula-test",
				ApiID:     "petstore",
				VersionID: "1.0.0",
				SpecID:    "openapi.yaml",
			},
		},
	}
	gotValue, _, gotErr := processScoreFormula(ctx, registryClient, formula, resource, &rpc.Artifact{}, true)
	if gotErr != nil {
		t.Errorf("processScoreFormula() returned unexpected error: %s", gotErr)
	}

	wantValue := int64(1)

	if gotValue != nil && wantValue != gotValue {
		t.Errorf("processScoreFormula() returned unexpected value, want: %v, got: %v", wantValue, gotValue)
	}
}

func TestProcessScoreFormulaError(t *testing.T) {
	tests := []struct {
		desc     string
		setup    func(context.Context, connection.Client, connection.AdminClient)
		formula  *rpc.ScoreFormula
		resource patterns.ResourceInstance
	}{
		{
			desc: "invalid reference",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.specs/artifacts/lint-spectral", //error
				},
				ScoreExpression: "size(files[0].problems)",
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
		{
			desc: "invalid extended pattern",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifact/lint-spectral", // error
				},
				ScoreExpression: "size(files[0].problems)",
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
		{
			desc: "missing artifact",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problems)",
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
		{
			desc: "unsupported artifact type",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				artifactBytes, _ := proto.Marshal(&rpc.ScoreDefinition{
					Id:             "dummy-score-definition",
					TargetResource: &rpc.ResourcePattern{},
					Formula:        nil,
					Type:           nil,
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/score-definition",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.ScoreDefinition")
			},
			formula: &rpc.ScoreFormula{},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
		{
			desc: "invalid expression",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problem)", // invalid expression
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
		{
			desc: "missing expression",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
			},
			formula: &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			_, _, gotErr := processScoreFormula(ctx, registryClient, test.formula, test.resource, &rpc.Artifact{}, true)
			if gotErr == nil {
				t.Errorf("processScoreFormula(ctx, client, %v, %v) did not return an error", test.formula, test.resource)
			}
		})
	}
}

func TestProcessScoreFormulaTimestamp(t *testing.T) {
	tests := []struct {
		desc       string
		setup      func(context.Context, connection.Client, connection.AdminClient)
		resource   patterns.ResourceInstance
		takeAction bool
		wantValue  interface{}
		wantUpdate bool
		wantErr    bool
	}{
		// When takeAction is true, the score value should be always updated
		{
			desc: "takeAction is true and score is outdated",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// create formula artifact
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: true,
			wantValue:  int64(1),
			wantUpdate: true,
			wantErr:    false,
		},
		// When takeAction is true, the score value should be always updated
		{
			desc: "takeAction and score is up-to-date",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create formula artifact
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: true,
			wantValue:  int64(1),
			wantUpdate: true,
			wantErr:    false,
		},
		{
			desc: "!takeAction and score is outdated",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// create formula artifact
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: false,
			wantValue:  int64(1),
			wantUpdate: true,
			wantErr:    false,
		},
		{
			desc: "!takeAction and score is up-to-date",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "score-formula-test")
				createProject(ctx, adminClient, t, "score-formula-test")
				createApi(ctx, client, t, "projects/score-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create formula artifact
				artifactBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs//openapi.yaml/artifacts/lint-spectral",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: false,
			wantValue:  int64(1),
			wantUpdate: false,
			wantErr:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			//fetch score artifact
			scoreArtifact, err := getArtifact(ctx, registryClient, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error", false)
			if err != nil {
				t.Errorf("failed to fetch the scoreArtifact from setup: %s", err)
			}

			formula := &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problems)",
			}

			gotValue, gotUpdate, gotErr := processScoreFormula(ctx, registryClient, formula, test.resource, scoreArtifact, test.takeAction)
			if test.wantErr {
				if gotErr == nil {
					t.Errorf("processScoreFormula(ctx, client, %v, %v, %v, %t) did not return an error", formula, test.resource, scoreArtifact, test.takeAction)
				}
			} else if gotValue != test.wantValue || gotUpdate != test.wantUpdate {
				t.Errorf("processScoreFormula() returned unexpected response, want: (%s, %t), got: (%s, %t)", test.wantValue, test.wantUpdate, gotValue, gotUpdate)
			}
		})
	}
}

func TestProcessRollUpFormula(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	defer registryClient.Close()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	defer adminClient.Close()

	//setup
	deleteProject(ctx, adminClient, t, "rollup-formula-test")
	createProject(ctx, adminClient, t, "rollup-formula-test")
	createApi(ctx, registryClient, t, "projects/rollup-formula-test/locations/global", "petstore")
	createVersion(ctx, registryClient, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// create lint artifact
	lintBytes, _ := proto.Marshal(&rpc.Lint{
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
	})
	createUpdateArtifact(
		ctx, registryClient, t,
		"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
		lintBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
	// create complexity artifact
	complexityBytes, _ := proto.Marshal(&metrics.Complexity{
		GetCount:    1,
		PostCount:   1,
		PutCount:    1,
		DeleteCount: 1,
	})
	createUpdateArtifact(
		ctx, registryClient, t,
		"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
		complexityBytes, "application/octet-stream;type=gnostic.metrics.Complexity")

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
		SpecName: patterns.SpecName{
			Name: names.Spec{
				ProjectID: "rollup-formula-test",
				ApiID:     "petstore",
				VersionID: "1.0.0",
				SpecID:    "openapi.yaml",
			},
		},
	}
	gotValue, _, gotErr := processRollUpFormula(ctx, registryClient, formula, resource, &rpc.Artifact{}, true)
	if gotErr != nil {
		t.Errorf("processRollUpFormula() returned unexpected error: %s", gotErr)
	}

	wantValue := float64(0.5)

	if gotValue != nil && wantValue != gotValue {
		t.Errorf("processRollUpFormula() returned unexpected value, want: %v, got: %v", wantValue, gotValue)
	}
}

func TestProcessRollUpFormulaError(t *testing.T) {
	tests := []struct {
		desc     string
		setup    func(context.Context, connection.Client, connection.AdminClient)
		formula  *rpc.RollUpFormula
		resource patterns.ResourceInstance
	}{
		{
			desc: "missing score_formulas",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
			},
			formula: &rpc.RollUpFormula{
				RollupExpression: "double(numErrors)/numOperations",
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
		{
			desc: "missing rollup_expression",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
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
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
		{
			desc: "invalid score_expression",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create lint artifact
				lintBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					lintBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
				// create complexity artifact
				complexityBytes, _ := proto.Marshal(&metrics.Complexity{
					GetCount:    1,
					PostCount:   1,
					PutCount:    1,
					DeleteCount: 1,
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					complexityBytes, "application/octet-stream;type=gnostic.metrics.Complexity")
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
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
		{
			desc: "invalid rollup_expression",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// create lint artifact
				lintBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					lintBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
				// create complexity artifact
				complexityBytes, _ := proto.Marshal(&metrics.Complexity{
					GetCount:    1,
					PostCount:   1,
					PutCount:    1,
					DeleteCount: 1,
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					complexityBytes, "application/octet-stream;type=gnostic.metrics.Complexity")
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
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
		{
			desc: "invalid reference_id",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
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
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			_, _, gotErr := processRollUpFormula(ctx, registryClient, test.formula, test.resource, &rpc.Artifact{}, true)
			if gotErr == nil {
				t.Errorf("processRollUpFormula(ctx, client, %v, %v) did not return an error", test.formula, test.resource)
			}
		})
	}
}

func TestProcessRollUpFormulaTimestamp(t *testing.T) {
	tests := []struct {
		desc       string
		setup      func(context.Context, connection.Client, connection.AdminClient)
		resource   patterns.ResourceInstance
		takeAction bool
		wantValue  interface{}
		wantUpdate bool
		wantErr    bool
	}{
		{
			desc: "takeAction and score completely outdated",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)

				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")

				// create lint artifact
				lintBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					lintBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
				// create complexity artifact
				complexityBytes, _ := proto.Marshal(&metrics.Complexity{
					GetCount:    1,
					PostCount:   1,
					PutCount:    1,
					DeleteCount: 1,
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					complexityBytes, "application/octet-stream;type=gnostic.metrics.Complexity")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: true,
			wantValue:  float64(0.5),
			wantUpdate: true,
			wantErr:    false,
		},
		{
			desc: "takeAction and score partially outdated",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)

				// create lint artifact
				lintBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					lintBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")

				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")

				// create complexity artifact
				complexityBytes, _ := proto.Marshal(&metrics.Complexity{
					GetCount:    1,
					PostCount:   1,
					PutCount:    1,
					DeleteCount: 1,
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					complexityBytes, "application/octet-stream;type=gnostic.metrics.Complexity")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: true,
			wantValue:  float64(0.5),
			wantUpdate: true,
			wantErr:    false,
		},
		{
			desc: "takeAction and score up-to-date",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)

				// create lint artifact
				lintBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					lintBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")

				// create complexity artifact
				complexityBytes, _ := proto.Marshal(&metrics.Complexity{
					GetCount:    1,
					PostCount:   1,
					PutCount:    1,
					DeleteCount: 1,
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					complexityBytes, "application/octet-stream;type=gnostic.metrics.Complexity")

				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: true,
			wantValue:  float64(0.5),
			wantUpdate: true,
			wantErr:    false,
		},
		{
			desc: "!takeAction and score completely outdated",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)

				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")

				// create lint artifact
				lintBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					lintBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")
				// create complexity artifact
				complexityBytes, _ := proto.Marshal(&metrics.Complexity{
					GetCount:    1,
					PostCount:   1,
					PutCount:    1,
					DeleteCount: 1,
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					complexityBytes, "application/octet-stream;type=gnostic.metrics.Complexity")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: false,
			wantValue:  float64(0.5),
			wantUpdate: true,
			wantErr:    false,
		},
		{
			desc: "!takeAction and score partially outdated",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)

				// create lint artifact
				lintBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					lintBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")

				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")

				// create complexity artifact
				complexityBytes, _ := proto.Marshal(&metrics.Complexity{
					GetCount:    1,
					PostCount:   1,
					PutCount:    1,
					DeleteCount: 1,
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					complexityBytes, "application/octet-stream;type=gnostic.metrics.Complexity")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: false,
			wantValue:  float64(0.5),
			wantUpdate: true,
			wantErr:    false,
		},
		{
			desc: "!takeAction and score up-to-date",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				deleteProject(ctx, adminClient, t, "rollup-formula-test")
				createProject(ctx, adminClient, t, "rollup-formula-test")
				createApi(ctx, client, t, "projects/rollup-formula-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)

				// create lint artifact
				lintBytes, _ := proto.Marshal(&rpc.Lint{
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
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-spectral",
					lintBytes, "application/octet-stream;type=google.cloud.apigeeregistry.applications.v1alpha1.Lint")

				// create complexity artifact
				complexityBytes, _ := proto.Marshal(&metrics.Complexity{
					GetCount:    1,
					PostCount:   1,
					PutCount:    1,
					DeleteCount: 1,
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity",
					complexityBytes, "application/octet-stream;type=gnostic.metrics.Complexity")

				// create score artifact
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					[]byte{}, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "rollup-formula-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: false,
			wantValue:  nil,
			wantUpdate: false,
			wantErr:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Failed to create client: %+v", err)
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			//fetch score artifact
			scoreArtifact, err := getArtifact(ctx, registryClient, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error", false)
			if err != nil {
				t.Errorf("failed to fetch the scoreArtifact from setup: %s", err)
			}

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

			gotValue, gotUpdate, gotErr := processRollUpFormula(ctx, registryClient, formula, test.resource, scoreArtifact, test.takeAction)
			if test.wantErr {
				if gotErr == nil {
					t.Errorf("processScoreFormula(ctx, client, %v, %v, %v, %t) did not return an error", formula, test.resource, scoreArtifact, test.takeAction)
				}
			} else if gotValue != test.wantValue || gotUpdate != test.wantUpdate {
				t.Errorf("processScoreFormula() returned unexpected response, want: (%s, %t), got: (%s, %t)", test.wantValue, test.wantUpdate, gotValue, gotUpdate)
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
			desc:       "greater than max integer",
			definition: integerDefinition,
			scoreValue: 11,
		},
		{
			desc:       "less than min integer",
			definition: integerDefinition,
			scoreValue: -1,
		},
		{
			desc:       "type mismatch percent",
			definition: percentDefinition,
			scoreValue: false,
		},
		{
			desc:       "greater than max percent",
			definition: integerDefinition,
			scoreValue: 101,
		},
		{
			desc:       "less than min percent",
			definition: integerDefinition,
			scoreValue: -1,
		},
		{
			desc:       "type mismatch boolean",
			definition: booleanDefinition,
			scoreValue: 1,
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
