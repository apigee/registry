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
				t.Errorf("expected matchResourceWithTarget() to return errors")
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
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
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
	gotValue, gotErr := processScoreFormula(ctx, registryClient, formula, resource)
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
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer adminClient.Close()

			test.setup(ctx, registryClient, adminClient)

			_, gotErr := processScoreFormula(ctx, registryClient, test.formula, test.resource)
			if gotErr == nil {
				t.Errorf("expected processScoreFormula() to return an error")
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
				t.Errorf("expected processScoreType() to return an error")
			}
		})
	}
}
