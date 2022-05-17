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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCalculateScoreCard(t *testing.T) {
	tests := []struct {
		desc          string
		setup         func(context.Context, connection.Client, connection.AdminClient)
		wantScoreCard *rpc.ScoreCard
	}{
		{
			desc: "non existent ScoreCard artifact",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-card-test")
				createProject(ctx, adminClient, t, "score-card-test")
				createApi(ctx, client, t, "projects/score-card-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-card-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// ScoreCard definition
				artifactBytes, _ := proto.Marshal(&rpc.ScoreCardDefinition{
					Id:          "quality",
					Kind:        "ScoreCardDefinition",
					DisplayName: "Quality",
					Description: "Quality ScoreCard",
					TargetResource: &rpc.ResourcePattern{
						Pattern: "apis/-/versions/-/specs/-",
					},
					ScorePatterns: []string{
						"$resource.spec/artifacts/score-lint-error",
						"$resource.spec/artifacts/score-lang-reuse",
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/artifacts/quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCardDefinition")
				// score lint-error
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// score lang-reuse
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lang-reuse",
					Kind:           "Score",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
					Severity:       rpc.Severity_OK,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 70,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			wantScoreCard: &rpc.ScoreCard{
				Id:             "scorecard-quality",
				Kind:           "ScoreCard",
				DisplayName:    "Quality",
				Description:    "Quality ScoreCard",
				DefinitionName: "projects/score-card-test/locations/global/artifacts/quality",
				Scores: []*rpc.Score{
					{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					},
					{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					},
				},
			},
		},
		{
			desc: "updated definition",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-card-test")
				createProject(ctx, adminClient, t, "score-card-test")
				createApi(ctx, client, t, "projects/score-card-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-card-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// score lint-error
				artifactBytes, _ := proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// score lang-reuse
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lang-reuse",
					Kind:           "Score",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
					Severity:       rpc.Severity_OK,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 70,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// scorecard quality
				artifactBytes, _ = proto.Marshal(&rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 50,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard")
				// ScoreCard definition
				artifactBytes, _ = proto.Marshal(&rpc.ScoreCardDefinition{
					Id:          "quality",
					Kind:        "ScoreCardDefinition",
					DisplayName: "Quality",
					Description: "Quality ScoreCard",
					TargetResource: &rpc.ResourcePattern{
						Pattern: "apis/-/versions/-/specs/-",
					},
					ScorePatterns: []string{
						"$resource.spec/artifacts/score-lint-error",
						"$resource.spec/artifacts/score-lang-reuse",
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/artifacts/quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCardDefinition")
			},
			wantScoreCard: &rpc.ScoreCard{
				Id:             "scorecard-quality",
				Kind:           "ScoreCard",
				DisplayName:    "Quality",
				Description:    "Quality ScoreCard",
				DefinitionName: "projects/score-card-test/locations/global/artifacts/quality",
				Scores: []*rpc.Score{
					{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					},
					{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					},
				},
			},
		},
		{
			desc: "updated score artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-card-test")
				createProject(ctx, adminClient, t, "score-card-test")
				createApi(ctx, client, t, "projects/score-card-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-card-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// ScoreCard definition
				artifactBytes, _ := proto.Marshal(&rpc.ScoreCardDefinition{
					Id:          "quality",
					Kind:        "ScoreCardDefinition",
					DisplayName: "Quality",
					Description: "Quality ScoreCard",
					TargetResource: &rpc.ResourcePattern{
						Pattern: "apis/-/versions/-/specs/-",
					},
					ScorePatterns: []string{
						"$resource.spec/artifacts/score-lint-error",
						"$resource.spec/artifacts/score-lang-reuse",
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/artifacts/quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCardDefinition")
				// scorecard quality
				artifactBytes, _ = proto.Marshal(&rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 50,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard")
				// score lint-error
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// score lang-reuse
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lang-reuse",
					Kind:           "Score",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
					Severity:       rpc.Severity_OK,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 70,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			wantScoreCard: &rpc.ScoreCard{
				Id:             "scorecard-quality",
				Kind:           "ScoreCard",
				DisplayName:    "Quality",
				Description:    "Quality ScoreCard",
				DefinitionName: "projects/score-card-test/locations/global/artifacts/quality",
				Scores: []*rpc.Score{
					{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					},
					{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					},
				},
			},
		},
		{
			desc: "updated definition and score artifacts",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-card-test")
				createProject(ctx, adminClient, t, "score-card-test")
				createApi(ctx, client, t, "projects/score-card-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-card-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// scorecard quality
				artifactBytes, _ := proto.Marshal(&rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 50,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard")
				// score lint-error
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// score lang-reuse
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lang-reuse",
					Kind:           "Score",
					DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
					Severity:       rpc.Severity_OK,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 70,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// ScoreCard definition
				artifactBytes, _ = proto.Marshal(&rpc.ScoreCardDefinition{
					Id:          "quality",
					Kind:        "ScoreCardDefinition",
					DisplayName: "Quality",
					Description: "Quality ScoreCard",
					TargetResource: &rpc.ResourcePattern{
						Pattern: "apis/-/versions/-/specs/-",
					},
					ScorePatterns: []string{
						"$resource.spec/artifacts/score-lint-error",
						"$resource.spec/artifacts/score-lang-reuse",
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-card-test/locations/global/artifacts/quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCardDefinition")
			},
			wantScoreCard: &rpc.ScoreCard{
				Id:             "scorecard-quality",
				Kind:           "ScoreCard",
				DisplayName:    "Quality",
				Description:    "Quality ScoreCard",
				DefinitionName: "projects/score-card-test/locations/global/artifacts/quality",
				Scores: []*rpc.Score{
					{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					},
					{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
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
						ProjectID: "score-card-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			}

			//fetch definition artifact
			defArtifact, err := getArtifact(ctx, registryClient, "projects/score-card-test/locations/global/artifacts/quality", true)
			if err != nil {
				t.Errorf("failed to fetch the definition Artifact from setup: %s", err)
			}

			gotErr := CalculateScoreCard(ctx, registryClient, defArtifact, resource)
			if gotErr != nil {
				t.Errorf("CalculateScore(ctx, client, %v, %v) returned unexpected error: %s", defArtifact, resource, gotErr)
			}

			//fetch score artifact and check the value
			scoreCardArtifact, err := getArtifact(
				ctx, registryClient,
				"projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality", true)
			if err != nil {
				t.Errorf("failed to get the result scoreCardArtifact from registry")
			}

			gotScoreCard := &rpc.ScoreCard{}
			err = proto.Unmarshal(scoreCardArtifact.GetContents(), gotScoreCard)
			if err != nil {
				t.Errorf("failed unmarshalling ScoreCard artifact from registry: %s", err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.wantScoreCard, gotScoreCard, opts) {
				t.Errorf("CalculateScoreCard() returned unexpected response (-want +got):\n%s", cmp.Diff(test.wantScoreCard, gotScoreCard, opts))
			}
		})
	}
}

func TestProcessScorePatterns(t *testing.T) {
	tests := []struct {
		desc       string
		setup      func(context.Context, connection.Client, connection.AdminClient)
		resource   patterns.ResourceInstance
		takeAction bool
		wantResult scoreCardResult
	}{
		{
			desc: "takeAction and scoreCard is up-to-date",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-patterns-test")
				createProject(ctx, adminClient, t, "score-patterns-test")
				createApi(ctx, client, t, "projects/score-patterns-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// score lint-error
				artifactBytes, _ := proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// score lang-reuse
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lang-reuse",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
					Severity:       rpc.Severity_OK,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 70,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// ScoreCard quality
				artifactBytes, _ = proto.Marshal(&rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 70,
								},
							},
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-patterns-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: true,
			wantResult: scoreCardResult{
				scoreCard: &rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 70,
								},
							},
						},
					},
				},
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "takeAction and scoreCard is outdated",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-patterns-test")
				createProject(ctx, adminClient, t, "score-patterns-test")
				createApi(ctx, client, t, "projects/score-patterns-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// ScoreCard quality
				artifactBytes, _ := proto.Marshal(&rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 70,
								},
							},
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard")
				// score lint-error
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// score lang-reuse
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lang-reuse",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
					Severity:       rpc.Severity_OK,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 70,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-patterns-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: true,
			wantResult: scoreCardResult{
				scoreCard: &rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 70,
								},
							},
						},
					},
				},
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and scoreCard is up-to-date",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-patterns-test")
				createProject(ctx, adminClient, t, "score-patterns-test")
				createApi(ctx, client, t, "projects/score-patterns-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// score lint-error
				artifactBytes, _ := proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// score lang-reuse
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lang-reuse",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
					Severity:       rpc.Severity_OK,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 70,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// ScoreCard quality
				artifactBytes, _ = proto.Marshal(&rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 70,
								},
							},
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-patterns-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: false,
			wantResult: scoreCardResult{
				scoreCard:   nil,
				needsUpdate: false,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and scoreCard is outdated",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-patterns-test")
				createProject(ctx, adminClient, t, "score-patterns-test")
				createApi(ctx, client, t, "projects/score-patterns-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// ScoreCard quality
				artifactBytes, _ := proto.Marshal(&rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 70,
								},
							},
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard")
				// score lint-error
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// score lang-reuse
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lang-reuse",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
					Severity:       rpc.Severity_OK,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 70,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-patterns-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: false,
			wantResult: scoreCardResult{
				scoreCard: &rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 70,
								},
							},
						},
					},
				},
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and scoreCard is partially outdated",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-patterns-test")
				createProject(ctx, adminClient, t, "score-patterns-test")
				createApi(ctx, client, t, "projects/score-patterns-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// score lint-error
				artifactBytes, _ := proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
				// ScoreCard quality
				artifactBytes, _ = proto.Marshal(&rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 70,
								},
							},
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard")
				// score lang-reuse
				artifactBytes, _ = proto.Marshal(&rpc.Score{
					Id:             "score-lang-reuse",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
					Severity:       rpc.Severity_OK,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 70,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			resource: patterns.SpecResource{
				SpecName: patterns.SpecName{
					Name: names.Spec{
						ProjectID: "score-patterns-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			},
			takeAction: false,
			wantResult: scoreCardResult{
				scoreCard: &rpc.ScoreCard{
					Id:             "scorecard-quality",
					Kind:           "ScoreCard",
					DisplayName:    "Quality",
					Description:    "Quality ScoreCard",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/quality",
					Scores: []*rpc.Score{
						{
							Id:             "score-lint-error",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
							Severity:       rpc.Severity_ALERT,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 60,
								},
							},
						},
						{
							Id:             "score-lang-reuse",
							Kind:           "Score",
							DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
							Severity:       rpc.Severity_OK,
							Value: &rpc.Score_PercentValue{
								PercentValue: &rpc.PercentValue{
									Value: 70,
								},
							},
						},
					},
				},
				needsUpdate: true,
				err:         nil,
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

			definition := &rpc.ScoreCardDefinition{
				Id:          "quality",
				Kind:        "ScoreCardDefinition",
				DisplayName: "Quality",
				Description: "Quality ScoreCard",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				ScorePatterns: []string{
					"$resource.spec/artifacts/score-lint-error",
					"$resource.spec/artifacts/score-lang-reuse",
				},
			}

			//fetch the ScoreCard artifact
			scoreCardArtifact, err := getArtifact(ctx, registryClient, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality", false)
			if err != nil {
				t.Errorf("failed to fetch the scoreCardArtifact from setup: %s", err)
			}

			gotResult := processScorePatterns(ctx, registryClient, definition, test.resource, scoreCardArtifact, test.takeAction, "projects/score-patterns-test/locations/global")

			opts := cmp.Options{
				cmp.AllowUnexported(scoreCardResult{}),
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}

			if !cmp.Equal(test.wantResult, gotResult, opts) {
				t.Errorf("processScorePatterns() returned unexpected response, (-want +got):\n%s", cmp.Diff(test.wantResult, gotResult, opts))
			}
		})
	}
}

func TestProcessScorePatternsError(t *testing.T) {
	tests := []struct {
		desc       string
		setup      func(context.Context, connection.Client, connection.AdminClient)
		takeAction bool
		definition *rpc.ScoreCardDefinition
	}{
		{
			desc:       "Invalid reference pattern in ScoreCardDefinition",
			setup:      func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {},
			takeAction: true,
			definition: &rpc.ScoreCardDefinition{
				Id:          "quality",
				Kind:        "ScoreCardDefinition",
				DisplayName: "Quality",
				Description: "Quality ScoreCard",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				ScorePatterns: []string{
					"$resource.spec/artifact/score-lint-error",
					"$resource.spec/artifact/score-lang-reuse",
				},
			},
		},
		{
			desc: "Missing score artifact",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-patterns-test")
				createProject(ctx, adminClient, t, "score-patterns-test")
				createApi(ctx, client, t, "projects/score-patterns-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// score lint-error
				artifactBytes, _ := proto.Marshal(&rpc.Score{
					Id:             "score-lint-error",
					Kind:           "Score",
					DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
					Severity:       rpc.Severity_ALERT,
					Value: &rpc.Score_PercentValue{
						PercentValue: &rpc.PercentValue{
							Value: 60,
						},
					},
				})
				createUpdateArtifact(
					ctx, client, t,
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			takeAction: true,
			definition: &rpc.ScoreCardDefinition{
				Id:          "quality",
				Kind:        "ScoreCardDefinition",
				DisplayName: "Quality",
				Description: "Quality ScoreCard",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				ScorePatterns: []string{
					"$resource.spec/artifact/score-lint-error",
					"$resource.spec/artifact/score-lang-reuse",
				},
			},
		},
		{
			desc: "Invalid score artifact",
			setup: func(ctx context.Context, client connection.Client, adminClient connection.AdminClient) {
				//setup
				deleteProject(ctx, adminClient, t, "score-patterns-test")
				createProject(ctx, adminClient, t, "score-patterns-test")
				createApi(ctx, client, t, "projects/score-patterns-test/locations/global", "petstore")
				createVersion(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore", "1.0.0")
				createSpec(ctx, client, t, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
				// score lint-error
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
					"projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					artifactBytes, "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score")
			},
			takeAction: true,
			definition: &rpc.ScoreCardDefinition{
				Id:          "quality",
				Kind:        "ScoreCardDefinition",
				DisplayName: "Quality",
				Description: "Quality ScoreCard",
				TargetResource: &rpc.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
				},
				ScorePatterns: []string{
					"$resource.spec/artifact/score-lint-error",
					"$resource.spec/artifact/score-lang-reuse",
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
						ProjectID: "score-patterns-test",
						ApiID:     "petstore",
						VersionID: "1.0.0",
						SpecID:    "openapi.yaml",
					},
				},
			}

			gotResult := processScorePatterns(ctx, registryClient, test.definition, resource, &rpc.Artifact{}, test.takeAction, "projects/score-patterns-test/locations/global")

			if gotResult.err == nil {
				t.Errorf("processScorePatterns(ctx, client, %v, %v) did not return an error", test.definition, resource)
			}
		})
	}
}
