// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package scorecard

import (
	"context"
	"regexp"
	"testing"

	"github.com/apigee/registry/pkg/application/scoring"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

const gzipOpenAPIv3 = "application/x.openapi+gzip;version=3.0.0"
const gzipProtobuf = "application/x.protobuf+gzip"
const scoreCardDefinitionType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.ScoreCardDefinition"
const scoreType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.Score"

func protoMarshal(m proto.Message) []byte {
	b, _ := proto.Marshal(m)
	return b
}

var (
	scoreCardAll = &scoring.ScoreCardDefinition{
		Id:   "lint-summary",
		Kind: "ScoreCardDefinition",
		TargetResource: &scoring.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
		},
		ScorePatterns: []string{
			"$resource.spec/artifacts/score-lint-error",
		},
	}

	scoreCardOpenAPI = &scoring.ScoreCardDefinition{
		Id:   "lint-summary-openapi",
		Kind: "ScoreCardDefinition",
		TargetResource: &scoring.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
			Filter:  "mime_type.contains('openapi')",
		},
		ScorePatterns: []string{
			"$resource.spec/artifacts/score-lint-error",
		},
	}

	scoreCardProto = &scoring.ScoreCardDefinition{
		Id:   "lint-summary-proto",
		Kind: "ScoreCardDefinition",
		TargetResource: &scoring.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
			Filter:  "mime_type.contains('protobuf')",
		},
		ScorePatterns: []string{
			"$resource.spec/artifacts/score-lint-error",
		},
	}

	scoreLintError = &scoring.Score{
		Id:   "score-lint-error",
		Kind: "Score",
		Value: &scoring.Score_IntegerValue{
			IntegerValue: &scoring.IntegerValue{
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
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-lint-error",
					MimeType: scoreType,
					Contents: protoMarshal(scoreLintError),
				},
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/score-lint-error",
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
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi@([a-z0-9-]+)/artifacts/scorecard-lint-summary",
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi@([a-z0-9-]+)/artifacts/scorecard-lint-summary",
			},
		},
		{
			desc: "only openapi scores with single definition",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-lint-error",
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
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi@([a-z0-9-]+)/artifacts/scorecard-lint-summary-openapi",
			},
		},
		{
			desc: "only proto scores with single definition",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-lint-error",
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
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-lint-error",
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
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi@([a-z0-9-]+)/artifacts/scorecard-lint-summary-openapi",
				"projects/scorecard-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml@([a-z0-9-]+)/artifacts/scorecard-lint-summary-proto",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, _ := grpctest.SetupRegistry(ctx, t, "scorecard-test", test.seed)

			// setup the score command
			scoreCardCmd := Command()
			args := []string{"projects/scorecard-test/locations/global/apis/-/versions/-/specs/-"}
			scoreCardCmd.SetArgs(args)

			if err := scoreCardCmd.Execute(); err != nil {
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
