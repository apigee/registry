package compute

import (
	"context"
	"regexp"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/iterator"
)

const scoreCardDefinitionType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.ScoreCardDefinition"
const scoreType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.Score"

var (
	scoreCardAll = &rpc.ScoreCardDefinition{
		Id:   "lint-summary",
		Kind: "ScoreCardDefinition",
		TargetResource: &rpc.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
		},
		ScorePatterns: []string{
			"$resource.spec/artifacts/score-lint-error",
		},
	}

	scoreCardOpenAPI = &rpc.ScoreCardDefinition{
		Id:   "lint-summary-openapi",
		Kind: "ScoreCardDefinition",
		TargetResource: &rpc.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
			Filter:  "mime_type.contains('openapi')",
		},
		ScorePatterns: []string{
			"$resource.spec/artifacts/score-lint-error",
		},
	}

	scoreCardProto = &rpc.ScoreCardDefinition{
		Id:   "lint-summary-proto",
		Kind: "ScoreCardDefinition",
		TargetResource: &rpc.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
			Filter:  "mime_type.contains('protobuf')",
		},
		ScorePatterns: []string{
			"$resource.spec/artifacts/score-lint-error",
		},
	}

	scoreLintError = &rpc.Score{
		Id:   "score-lint-error",
		Kind: "Score",
		Value: &rpc.Score_IntegerValue{
			IntegerValue: &rpc.IntegerValue{
				Value:    1,
				MinValue: 0,
				MaxValue: 10,
			},
		},
	}
)

func TestScoreCard(t *testing.T) {
	tests := []struct {
		desc string
		seed []seeder.RegistryResource
		want []string
	}{
		{
			desc: "all spec scores",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: scoreType,
					Contents: protoMarshal(scoreLintError),
				},
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: scoreType,
					Contents: protoMarshal(scoreLintError),
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/artifacts/lint-summary",
					MimeType: scoreCardDefinitionType,
					Contents: protoMarshal(scoreCardAll),
				},
			},
			want: []string{
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml@([a-z0-9-]+)/artifacts/scorecard-lint-summary",
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml@([a-z0-9-]+)/artifacts/scorecard-lint-summary",
			},
		},
		{
			desc: "only openapi scores with single definition",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: scoreType,
					Contents: protoMarshal(scoreLintError),
				},
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml",
					MimeType: gzipProtobuf,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml/artifacts/score-lint-error",
					MimeType: scoreType,
					Contents: protoMarshal(scoreLintError),
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/artifacts/lint-summary-openapi",
					MimeType: scoreCardDefinitionType,
					Contents: protoMarshal(scoreCardOpenAPI),
				},
			},
			want: []string{
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml@([a-z0-9-]+)/artifacts/scorecard-lint-summary-openapi",
			},
		},
		{
			desc: "only proto scores with single definition",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: scoreType,
					Contents: protoMarshal(scoreLintError),
				},
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml",
					MimeType: gzipProtobuf,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml/artifacts/score-lint-error",
					MimeType: scoreType,
					Contents: protoMarshal(scoreLintError),
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/artifacts/lint-summary-proto",
					MimeType: scoreCardDefinitionType,
					Contents: protoMarshal(scoreCardProto),
				},
			},
			want: []string{
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml@([a-z0-9-]+)/artifacts/scorecard-lint-summary-proto",
			},
		},
		{
			desc: "proto and openapi scores with both definitions",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: scoreType,
					Contents: protoMarshal(scoreLintError),
				},
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml",
					MimeType: gzipProtobuf,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml/artifacts/score-lint-error",
					MimeType: scoreType,
					Contents: protoMarshal(scoreLintError),
				},

				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/artifacts/lint-summary-openapi",
					MimeType: scoreCardDefinitionType,
					Contents: protoMarshal(scoreCardOpenAPI),
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/artifacts/lint-summary-proto",
					MimeType: scoreCardDefinitionType,
					Contents: protoMarshal(scoreCardProto),
				},
			},
			want: []string{
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml@([a-z0-9-]+)/artifacts/scorecard-lint-summary-openapi",
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml@([a-z0-9-]+)/artifacts/scorecard-lint-summary-proto",
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

			deleteProject(ctx, adminClient, t, "scorecard-test")
			t.Cleanup(func() { deleteProject(ctx, adminClient, t, "scorecard-test") })

			client := seeder.Client{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
			}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			// setup the score command
			scoreCardCmd := Command()
			args := []string{"scorecard", "projects/scorecard-test/locations/global/apis/-/versions/-/specs/-"}
			scoreCardCmd.SetArgs(args)

			if err = scoreCardCmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}

			// list scorecard artifacts
			it := registryClient.ListArtifacts(ctx, &rpc.ListArtifactsRequest{
				Parent: "projects/scorecard-test/locations/global/apis/-/versions/-/specs/-",
				Filter: "mime_type.contains('ScoreCard')",
			})

			got := []string{}
			for a, err := it.Next(); err != iterator.Done; a, err = it.Next() {
				if err != nil {
					break
				}
				got = append(got, a.GetName())
			}

			regexComparer := cmp.Comparer(func(a, b string) bool {
				return regexp.MustCompile(a).MatchString(b) || regexp.MustCompile(b).MatchString(a)
			})

			if diff := cmp.Diff(test.want, got, regexComparer); diff != "" {
				t.Errorf("compute scorecard command returned unexpected diff (-want +got): %s", diff)
			}
		})
	}
}
