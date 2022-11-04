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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const gzipOpenAPIv3 = "application/x.openapi+gzip;version=3.0.0"
const gzipProtobuf = "application/x.protobuf+gzip"
const scoreDefinitionType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.ScoreDefinition"
const conformanceReportType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.style.ConformanceReport"

var (
	conformanceReport = &rpc.ConformanceReport{
		Id:         "conformance-report",
		Kind:       "ConformanceReport",
		Styleguide: "projects/score-test/locations/global/artifacts/styleguide",
		GuidelineReportGroups: []*rpc.GuidelineReportGroup{
			{State: rpc.Guideline_STATE_UNSPECIFIED},
			{State: rpc.Guideline_PROPOSED},
			{
				State: rpc.Guideline_ACTIVE,
				GuidelineReports: []*rpc.GuidelineReport{
					{
						GuidelineId: "sample-guideline",
						RuleReportGroups: []*rpc.RuleReportGroup{
							{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
							{
								Severity: rpc.Rule_ERROR,
								RuleReports: []*rpc.RuleReport{
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

	scoreAll = &rpc.ScoreDefinition{
		Id:   "lint-error",
		Kind: "ScoreDefinition",
		TargetResource: &rpc.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
		},
		Formula: &rpc.ScoreDefinition_ScoreFormula{
			ScoreFormula: &rpc.ScoreFormula{
				Artifact:        &rpc.ResourcePattern{Pattern: "$resource.spec/artifacts/conformance-report"},
				ScoreExpression: "sum(guidelineReportGroups[2].guidelineReports.map(r, size(r.ruleReportGroups[1].ruleReports)))",
			},
		},
		Type: &rpc.ScoreDefinition_Integer{
			Integer: &rpc.IntegerType{
				MinValue: 0,
				MaxValue: 10,
			},
		},
	}

	scoreOpenAPI = &rpc.ScoreDefinition{
		Id:   "lint-error-openapi",
		Kind: "ScoreDefinition",
		TargetResource: &rpc.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
			Filter:  "mime_type.contains('openapi')",
		},
		Formula: &rpc.ScoreDefinition_ScoreFormula{
			ScoreFormula: &rpc.ScoreFormula{
				Artifact:        &rpc.ResourcePattern{Pattern: "$resource.spec/artifacts/conformance-report"},
				ScoreExpression: "sum(guidelineReportGroups[2].guidelineReports.map(r, size(r.ruleReportGroups[1].ruleReports)))",
			},
		},
		Type: &rpc.ScoreDefinition_Integer{
			Integer: &rpc.IntegerType{
				MinValue: 0,
				MaxValue: 10,
			},
		},
	}

	scoreProto = &rpc.ScoreDefinition{
		Id:   "lint-error-proto",
		Kind: "ScoreDefinition",
		TargetResource: &rpc.ResourcePattern{
			Pattern: "apis/-/versions/-/specs/-",
			Filter:  "mime_type.contains('protobuf')",
		},
		Formula: &rpc.ScoreDefinition_ScoreFormula{
			ScoreFormula: &rpc.ScoreFormula{
				Artifact:        &rpc.ResourcePattern{Pattern: "$resource.spec/artifacts/conformance-report"},
				ScoreExpression: "sum(guidelineReportGroups[2].guidelineReports.map(r, size(r.ruleReportGroups[1].ruleReports)))",
			},
		},
		Type: &rpc.ScoreDefinition_Integer{
			Integer: &rpc.IntegerType{
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
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-report",
					MimeType: conformanceReportType,
					Contents: protoMarshal(conformanceReport),
				},
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/conformance-report",
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
				"projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml@([a-z0-9-]+)/artifacts/score-lint-error",
				"projects/score-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml@([a-z0-9-]+)/artifacts/score-lint-error",
			},
		},
		{
			desc: "only openapi scores with single definition",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-report",
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
				"projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml@([a-z0-9-]+)/artifacts/score-lint-error-openapi",
			},
		},
		{
			desc: "only proto scores with single definition",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-report",
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
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name:     "projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-report",
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
				"projects/score-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml@([a-z0-9-]+)/artifacts/score-lint-error-openapi",
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
			args := []string{"score", "projects/score-test/locations/global/apis/-/versions/-/specs/-"}
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
