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
package score

import (
	"context"
	"regexp"
	"testing"

	"github.com/apigee/registry/pkg/artifacts"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

const gzipOpenAPIv3 = "application/x.openapi+gzip;version=3.0.0"
const gzipProtobuf = "application/x.protobuf+gzip"
const scoreDefinitionType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.ScoreDefinition"
const conformanceReportType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.ConformanceReport"

var (
	conformanceReport = &artifacts.ConformanceReport{
		Id:         "conformance-report",
		Kind:       "ConformanceReport",
		Styleguide: "projects/score-test/locations/global/artifacts/styleguide",
		GuidelineReportGroups: []*artifacts.GuidelineReportGroup{
			{State: artifacts.Guideline_STATE_UNSPECIFIED},
			{State: artifacts.Guideline_PROPOSED},
			{
				State: artifacts.Guideline_ACTIVE,
				GuidelineReports: []*artifacts.GuidelineReport{
					{
						GuidelineId: "sample-guideline",
						RuleReportGroups: []*artifacts.RuleReportGroup{
							{Severity: artifacts.Rule_SEVERITY_UNSPECIFIED},
							{
								Severity: artifacts.Rule_ERROR,
								RuleReports: []*artifacts.RuleReport{
									{
										RuleId: "openapi-tags",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	scoreAll = &artifacts.ScoreDefinition{
		Id:   "lint-error",
		Kind: "ScoreDefinition",
		TargetResource: &artifacts.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
		},
		Formula: &artifacts.ScoreDefinition_ScoreFormula{
			ScoreFormula: &artifacts.ScoreFormula{
				Artifact:        &artifacts.ResourcePattern{Pattern: "$resource.spec/artifacts/conformance-report"},
				ScoreExpression: "sum(guidelineReportGroups[2].guidelineReports.map(r, size(r.ruleReportGroups[1].ruleReports)))",
			},
		},
		Type: &artifacts.ScoreDefinition_Integer{
			Integer: &artifacts.IntegerType{
				MinValue: 0,
				MaxValue: 10,
			},
		},
	}

	scoreOpenAPI = &artifacts.ScoreDefinition{
		Id:   "lint-error-openapi",
		Kind: "ScoreDefinition",
		TargetResource: &artifacts.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
			Filter:  "mime_type.contains('openapi')",
		},
		Formula: &artifacts.ScoreDefinition_ScoreFormula{
			ScoreFormula: &artifacts.ScoreFormula{
				Artifact:        &artifacts.ResourcePattern{Pattern: "$resource.spec/artifacts/conformance-report"},
				ScoreExpression: "sum(guidelineReportGroups[2].guidelineReports.map(r, size(r.ruleReportGroups[1].ruleReports)))",
			},
		},
		Type: &artifacts.ScoreDefinition_Integer{
			Integer: &artifacts.IntegerType{
				MinValue: 0,
				MaxValue: 10,
			},
		},
	}

	scoreProto = &artifacts.ScoreDefinition{
		Id:   "lint-error-proto",
		Kind: "ScoreDefinition",
		TargetResource: &artifacts.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
			Filter:  "mime_type.contains('protobuf')",
		},
		Formula: &artifacts.ScoreDefinition_ScoreFormula{
			ScoreFormula: &artifacts.ScoreFormula{
				Artifact:        &artifacts.ResourcePattern{Pattern: "$resource.spec/artifacts/conformance-report"},
				ScoreExpression: "sum(guidelineReportGroups[2].guidelineReports.map(r, size(r.ruleReportGroups[1].ruleReports)))",
			},
		},
		Type: &artifacts.ScoreDefinition_Integer{
			Integer: &artifacts.IntegerType{
				MinValue: 0,
				MaxValue: 10,
			},
		},
	}
)

func protoMarshal(m proto.Message) []byte {
	b, _ := proto.Marshal(m)
	return b
}

func deleteProject(
	ctx context.Context,
	client connection.AdminClient,
	t *testing.T,
	projectID string) {
	t.Helper()
	req := &rpc.DeleteProjectRequest{
		Name:  "projects/" + projectID,
		Force: true,
	}
	err := client.DeleteProject(ctx, req)
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Failed DeleteProject(%v): %s", req, err.Error())
	}
}

func TestScore(t *testing.T) {
	tests := []struct {
		desc string
		seed []seeder.RegistryResource
		want []string
	}{
		{
			desc: "all spec scores",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/conformance-report",
					MimeType: conformanceReportType,
					Contents: protoMarshal(conformanceReport),
				},
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/conformance-report",
					MimeType: conformanceReportType,
					Contents: protoMarshal(conformanceReport),
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/artifacts/lint-error",
					MimeType: scoreDefinitionType,
					Contents: protoMarshal(scoreAll),
				},
			},
			want: []string{
				"projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi@([a-z0-9-]+)/artifacts/score-lint-error",
				"projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi@([a-z0-9-]+)/artifacts/score-lint-error",
			},
		},
		{
			desc: "only openapi scores with single definition",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/conformance-report",
					MimeType: conformanceReportType,
					Contents: protoMarshal(conformanceReport),
				},
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml",
					MimeType: gzipProtobuf,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml/artifacts/conformance-report",
					MimeType: conformanceReportType,
					Contents: protoMarshal(conformanceReport),
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/artifacts/lint-error-openapi",
					MimeType: scoreDefinitionType,
					Contents: protoMarshal(scoreOpenAPI),
				},
			},
			want: []string{
				"projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi@([a-z0-9-]+)/artifacts/score-lint-error-openapi",
			},
		},
		{
			desc: "only proto scores with single definition",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/conformance-report",
					MimeType: conformanceReportType,
					Contents: protoMarshal(conformanceReport),
				},
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml",
					MimeType: gzipProtobuf,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml/artifacts/conformance-report",
					MimeType: conformanceReportType,
					Contents: protoMarshal(conformanceReport),
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/artifacts/lint-error-proto",
					MimeType: scoreDefinitionType,
					Contents: protoMarshal(scoreProto),
				},
			},
			want: []string{
				"projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml@([a-z0-9-]+)/artifacts/score-lint-error-proto",
			},
		},
		{
			desc: "proto and openapi scores with both definitions",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/conformance-report",
					MimeType: conformanceReportType,
					Contents: protoMarshal(conformanceReport),
				},
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml",
					MimeType: gzipProtobuf,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml/artifacts/conformance-report",
					MimeType: conformanceReportType,
					Contents: protoMarshal(conformanceReport),
				},

				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/artifacts/lint-error-openapi",
					MimeType: scoreDefinitionType,
					Contents: protoMarshal(scoreOpenAPI),
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/artifacts/lint-error-proto",
					MimeType: scoreDefinitionType,
					Contents: protoMarshal(scoreProto),
				},
			},
			want: []string{
				"projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi@([a-z0-9-]+)/artifacts/score-lint-error-openapi",
				"projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/proto.yaml@([a-z0-9-]+)/artifacts/score-lint-error-proto",
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

			deleteProject(ctx, adminClient, t, "score-test")
			t.Cleanup(func() { deleteProject(ctx, adminClient, t, "score-test") })

			client := seeder.Client{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
			}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			// setup the score command
			scoreCmd := Command()
			args := []string{"projects/score-test/locations/global/apis/-/versions/-/specs/-"}
			scoreCmd.SetArgs(args)

			if err = scoreCmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}

			// list score artifacts
			it := registryClient.ListArtifacts(ctx, &rpc.ListArtifactsRequest{
				Parent: "projects/score-test/locations/global/apis/-/versions/-/specs/-",
				Filter: "mime_type.contains('Score')",
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
				t.Errorf("compute score command returned unexpected diff (-want +got): %s", diff)
			}
		})
	}
}
