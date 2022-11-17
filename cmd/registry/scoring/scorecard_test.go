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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCalculateScoreCard(t *testing.T) {
	tests := []struct {
		desc          string
		seed          []seeder.RegistryResource
		wantScoreCard *rpc.ScoreCard
	}{
		{
			desc: "nonexistent ScoreCard artifact",
			seed: []seeder.RegistryResource{
				// ScoreCard definition
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/artifacts/quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCardDefinition",
					Contents: protoMarshal(&rpc.ScoreCardDefinition{
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
					}),
				},
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
				// Score lang-reuse
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					}),
				},
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
			seed: []seeder.RegistryResource{
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
				// Score lang-reuse
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					}),
				},
				// ScoreCard artifact
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard",
					Contents: protoMarshal(&rpc.ScoreCard{
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
					}),
				},
				// ScoreCard definition
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/artifacts/quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCardDefinition",
					Contents: protoMarshal(&rpc.ScoreCardDefinition{
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
					}),
				},
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
			seed: []seeder.RegistryResource{
				// ScoreCard definition
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/artifacts/quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCardDefinition",
					Contents: protoMarshal(&rpc.ScoreCardDefinition{
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
					}),
				},
				// ScoreCard artifact
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard",
					Contents: protoMarshal(&rpc.ScoreCard{
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
					}),
				},
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
				// Score lang-reuse
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					}),
				},
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
			seed: []seeder.RegistryResource{
				// ScoreCard artifact
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard",
					Contents: protoMarshal(&rpc.ScoreCard{
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
					}),
				},
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
				// Score lang-reuse
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-card-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					}),
				},
				// ScoreCard definition
				&rpc.Artifact{
					Name:     "projects/score-card-test/locations/global/artifacts/quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCardDefinition",
					Contents: protoMarshal(&rpc.ScoreCardDefinition{
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
					}),
				},
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

			deleteProject(ctx, adminClient, t, "score-card-test")
			t.Cleanup(func() { deleteProject(ctx, adminClient, t, "score-card-test") })

			client := seeder.Client{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
			}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			resource := patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-card-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			}

			artifactClient := &RegistryArtifactClient{RegistryClient: registryClient}

			//fetch definition artifact
			defArtifact, err := getArtifact(ctx, artifactClient, "projects/score-card-test/locations/global/artifacts/quality", true)
			if err != nil {
				t.Errorf("failed to fetch the definition Artifact from setup: %s", err)
			}

			gotErr := CalculateScoreCard(ctx, artifactClient, defArtifact, resource, false)
			if gotErr != nil {
				t.Errorf("CalculateScore(ctx, client, %v, %v) returned unexpected error: %s", defArtifact, resource, gotErr)
			}

			//fetch score artifact and check the value
			scoreCardArtifact, err := getArtifact(
				ctx, artifactClient,
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
		seed       []seeder.RegistryResource
		resource   patterns.ResourceInstance
		takeAction bool
		wantResult scoreCardResult
	}{
		{
			desc: "takeAction and scoreCard is up-to-date",
			seed: []seeder.RegistryResource{
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
				// Score lang-reuse
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					}),
				},
				// ScoreCard artifact
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard",
					Contents: protoMarshal(&rpc.ScoreCard{
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
										Value: 50,
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
										Value: 60,
									},
								},
							},
						},
					}),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
			seed: []seeder.RegistryResource{
				// ScoreCard artifact
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard",
					Contents: protoMarshal(&rpc.ScoreCard{
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
										Value: 50,
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
										Value: 60,
									},
								},
							},
						},
					}),
				},
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
				// Score lang-reuse
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					}),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
			desc: "!takeAction and scoreCard is outdated",
			seed: []seeder.RegistryResource{
				// ScoreCard artifact
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard",
					Contents: protoMarshal(&rpc.ScoreCard{
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
										Value: 50,
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
										Value: 60,
									},
								},
							},
						},
					}),
				},
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
				// Score lang-reuse
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					}),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
			seed: []seeder.RegistryResource{
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
				// ScoreCard artifact
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.ScoreCard",
					Contents: protoMarshal(&rpc.ScoreCard{
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
										Value: 50,
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
										Value: 60,
									},
								},
							},
						},
					}),
				},
				// Score lang-reuse
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lang-reuse",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lang-reuse",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lang-reuse",
						Severity:       rpc.Severity_OK,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 70,
							},
						},
					}),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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

			deleteProject(ctx, adminClient, t, "score-patterns-test")
			t.Cleanup(func() { deleteProject(ctx, adminClient, t, "score-patterns-test") })

			client := seeder.Client{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
			}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

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

			artifactClient := &RegistryArtifactClient{RegistryClient: registryClient}

			//fetch the ScoreCard artifact
			scoreCardArtifact, err := getArtifact(ctx, artifactClient, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality", false)
			if err != nil {
				t.Errorf("failed to fetch the scoreCardArtifact from setup: %s", err)
			}

			gotResult := processScorePatterns(ctx, artifactClient, definition, test.resource, scoreCardArtifact, test.takeAction, "projects/score-patterns-test/locations/global")

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
		seed       []seeder.RegistryResource
		takeAction bool
		definition *rpc.ScoreCardDefinition
	}{
		{
			desc:       "Invalid reference pattern in ScoreCardDefinition",
			seed:       []seeder.RegistryResource{},
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
			seed: []seeder.RegistryResource{
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
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
			seed: []seeder.RegistryResource{
				// Score lint-error
				&rpc.Artifact{
					Name:     "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: protoMarshal(&rpc.Score{
						Id:             "score-lint-error",
						Kind:           "Score",
						DefinitionName: "projects/score-patterns-test/locations/global/artifacts/lint-error",
						Severity:       rpc.Severity_ALERT,
						Value: &rpc.Score_PercentValue{
							PercentValue: &rpc.PercentValue{
								Value: 60,
							},
						},
					}),
				},
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

			deleteProject(ctx, adminClient, t, "score-patterns-test")
			t.Cleanup(func() { deleteProject(ctx, adminClient, t, "score-patterns-test") })

			client := seeder.Client{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
			}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			resource := patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			}

			artifactClient := &RegistryArtifactClient{RegistryClient: registryClient}

			gotResult := processScorePatterns(ctx, artifactClient, test.definition, resource, &rpc.Artifact{}, test.takeAction, "projects/score-patterns-test/locations/global")

			if gotResult.err == nil {
				t.Errorf("processScorePatterns(ctx, client, %v, %v) did not return an error", test.definition, resource)
			}
		})
	}
}
