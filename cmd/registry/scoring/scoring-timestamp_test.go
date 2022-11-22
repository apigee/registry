package scoring

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patterns"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/apigee/registry/server/registry/test/seeder"
	metrics "github.com/google/gnostic/metrics"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeArtifactClient struct {
	artifacts []*rpc.Artifact
}

func (f *fakeArtifactClient) GetArtifact(ctx context.Context, artifact names.Artifact, getContents bool, handler core.ArtifactHandler) error {
	for _, a := range f.artifacts {
		if a.GetName() == artifact.String() {
			err := handler(a)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

func (f *fakeArtifactClient) SetArtifact(ctx context.Context, artifact *rpc.Artifact) error {
	if artifact.UpdateTime == nil {
		artifact.UpdateTime = timestamppb.Now()
	}
	for i, a := range f.artifacts {
		if a.GetName() == artifact.GetName() {
			// remove the old copy
			f.artifacts = append(f.artifacts[:i], f.artifacts[i+1:]...)
			break
		}
	}
	f.artifacts = append(f.artifacts, artifact)
	return nil
}

// This implementation doesn't support filter functionality
func (f *fakeArtifactClient) ListArtifacts(ctx context.Context, artifact names.Artifact, filter string, contents bool, handler core.ArtifactHandler) error {
	for _, a := range f.artifacts {
		name, _ := names.ParseArtifact(a.GetName())
		if strings.Contains(filter, name.Parent()) || (artifact.ArtifactID() != "-" && name.ArtifactID() != artifact.ArtifactID()) {
			continue
		}

		if err := handler(a); err != nil {
			return err
		}
	}
	return nil
}

// These functions are needed to use the fakeLister with the seeder package.
func (f *fakeArtifactClient) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	project := &rpc.Project{
		Name: fmt.Sprintf("projects/%s", req.GetProjectId()),
	}
	return project, nil
}

func (f *fakeArtifactClient) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	api := req.Api
	if api.UpdateTime == nil {
		api.UpdateTime = timestamppb.Now()
	}
	return api, nil
}

func (f *fakeArtifactClient) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	version := req.ApiVersion
	if version.UpdateTime == nil {
		version.UpdateTime = timestamppb.Now()
	}
	return version, nil
}

func (f *fakeArtifactClient) UpdateApiSpec(ctx context.Context, req *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	spec := req.ApiSpec
	if spec.RevisionUpdateTime == nil {
		spec.RevisionUpdateTime = timestamppb.Now()
	}
	return spec, nil
}

func (f *fakeArtifactClient) UpdateApiDeployment(ctx context.Context, req *rpc.UpdateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	deployment := req.ApiDeployment
	if deployment.RevisionUpdateTime == nil {
		deployment.RevisionUpdateTime = timestamppb.Now()
	}
	return deployment, nil
}

func (f *fakeArtifactClient) CreateArtifact(ctx context.Context, req *rpc.CreateArtifactRequest) (*rpc.Artifact, error) {
	artifact := req.Artifact
	if artifact.UpdateTime == nil {
		artifact.UpdateTime = timestamppb.Now()
	}
	f.artifacts = append(f.artifacts, artifact)
	return artifact, nil
}

func (f *fakeArtifactClient) ReplaceArtifact(ctx context.Context, req *rpc.ReplaceArtifactRequest) (*rpc.Artifact, error) {
	artifact := req.Artifact
	if artifact.UpdateTime == nil {
		artifact.UpdateTime = timestamppb.Now()
	}
	for i, a := range f.artifacts {
		if a.GetName() == artifact.GetName() {
			// remove the old copy
			f.artifacts = append(f.artifacts[:i], f.artifacts[i+1:]...)
			break
		}
	}
	f.artifacts = append(f.artifacts, artifact)
	return artifact, nil
}

func TestTimestampCalculateScore(t *testing.T) {
	tests := []struct {
		desc            string
		seed            []seeder.RegistryResource
		definitionProto *rpc.ScoreDefinition
		wantScore       *rpc.Score
	}{
		{
			desc: "existing up-to-date score",
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
					UpdateTime: timestamppb.Now(),
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
					UpdateTime: timestamppb.Now(),
				},
				// score artifact
				&rpc.Artifact{
					Name:       "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents:   []byte{},
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 3)),
				},
			},
			wantScore: &rpc.Score{},
		},
		{
			desc: "not recent enough score",
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
					UpdateTime: timestamppb.Now(),
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
					UpdateTime: timestamppb.Now(),
				},
				// score artifact
				&rpc.Artifact{
					Name:       "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents:   []byte{},
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
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
			client := &fakeArtifactClient{}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			resource := patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			}

			//fetch definition artifact
			defArtifact, err := getArtifact(ctx, client, "projects/score-formula-test/locations/global/artifacts/lint-error", true)
			if err != nil {
				t.Errorf("failed to fetch the definition Artifact from setup: %s", err)
			}

			gotErr := CalculateScore(ctx, client, defArtifact, resource, false)
			if gotErr != nil {
				t.Errorf("CalculateScore(ctx, client, %v, %v) returned unexpected error: %s", defArtifact, resource, gotErr)
			}

			//fetch score artifact and check the value
			scoreArtifact, err := getArtifact(
				ctx, client,
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

func TestProcessScoreFormulaTimestamp(t *testing.T) {
	tests := []struct {
		desc       string
		seed       []seeder.RegistryResource
		resource   patterns.ResourceInstance
		takeAction bool
		wantResult scoreResult
	}{
		// When takeAction is true, the score value should be always updated
		{
			desc: "takeAction is true and score is outdated",
			seed: []seeder.RegistryResource{
				// score artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents: []byte{},
				},
				// score  formula artifact
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
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: true,
			wantResult: scoreResult{
				value:       int64(1),
				needsUpdate: true,
				err:         nil,
			},
		},
		// When takeAction is true, the score value should be always updated
		{
			desc: "takeAction and score is up-to-date",
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
					UpdateTime: timestamppb.Now(),
				},
				// score artifact
				&rpc.Artifact{
					Name:       "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					Contents:   []byte{},
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 3)),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: true,
			wantResult: scoreResult{
				value:       int64(1),
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and score is outdated",
			seed: []seeder.RegistryResource{
				// score artifact
				&rpc.Artifact{
					Name:     "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
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
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: false,
			wantResult: scoreResult{
				value:       int64(1),
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and score is up-to-date",
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
					UpdateTime: timestamppb.Now(),
				},
				// score artifact
				&rpc.Artifact{
					Name:       "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 3)),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: false,
			wantResult: scoreResult{
				value:       int64(1),
				needsUpdate: false,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and score not recent enough",
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
					UpdateTime: timestamppb.Now(),
				},
				// score artifact
				&rpc.Artifact{
					Name:       "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: false,
			wantResult: scoreResult{
				value:       int64(1),
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "takeAction and score not recent enough",
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
					UpdateTime: timestamppb.Now(),
				},
				// score artifact
				&rpc.Artifact{
					Name:       "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: true,
			wantResult: scoreResult{
				value:       int64(1),
				needsUpdate: true,
				err:         nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			client := &fakeArtifactClient{}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			//fetch score artifact
			scoreArtifact, err := getArtifact(ctx, client, "projects/score-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error", false)
			if err != nil {
				t.Errorf("failed to fetch the scoreArtifact from setup: %s", err)
			}

			formula := &rpc.ScoreFormula{
				Artifact: &rpc.ResourcePattern{
					Pattern: "$resource.spec/artifacts/lint-spectral",
				},
				ScoreExpression: "size(files[0].problems)",
			}

			gotResult := processScoreFormula(ctx, client, formula, test.resource, scoreArtifact, test.takeAction)

			opts := cmp.AllowUnexported(scoreResult{})
			if !cmp.Equal(test.wantResult, gotResult, opts) {
				t.Errorf("processScoreFormula() returned unexpected response, (-want, + got):\n%s", cmp.Diff(test.wantResult, gotResult, opts))
			}
		})
	}
}

func TestProcessRollUpFormulaTimestamp(t *testing.T) {
	tests := []struct {
		desc       string
		seed       []seeder.RegistryResource
		resource   patterns.ResourceInstance
		takeAction bool
		wantResult scoreResult
	}{
		{
			desc: "takeAction and score completely outdated",
			seed: []seeder.RegistryResource{
				// score artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
				},
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
				//  complexity artifact
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
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: true,
			wantResult: scoreResult{
				value:       float64(0.5),
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "takeAction and score partially outdated",
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
				// score artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
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
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: true,
			wantResult: scoreResult{
				value:       float64(0.5),
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "takeAction and score up-to-date",
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
					UpdateTime: timestamppb.Now(),
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
					UpdateTime: timestamppb.Now(),
				},
				// score artifact
				&rpc.Artifact{
					Name:       "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 3)),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: true,
			wantResult: scoreResult{
				value:       float64(0.5),
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and score completely outdated",
			seed: []seeder.RegistryResource{
				// score artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
				},
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
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: false,
			wantResult: scoreResult{
				value:       float64(0.5),
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and score partially outdated",
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
				// score artifact
				&rpc.Artifact{
					Name:     "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType: "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
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
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: false,
			wantResult: scoreResult{
				value:       float64(0.5),
				needsUpdate: true,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and score up-to-date",
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
					UpdateTime: timestamppb.Now(),
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
					UpdateTime: timestamppb.Now(),
				},
				// score artifact
				&rpc.Artifact{
					Name:       "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 3)),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: false,
			wantResult: scoreResult{
				value:       nil,
				needsUpdate: false,
				err:         nil,
			},
		},
		{
			desc: "!takeAction and score not recent enough",
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
					UpdateTime: timestamppb.Now(),
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
					UpdateTime: timestamppb.Now(),
				},
				// score artifact
				&rpc.Artifact{
					Name:       "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error",
					MimeType:   "application/octet-stream;type=google.cloud.apigeeregistry.v1.Score",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
				},
			},
			takeAction: false,
			wantResult: scoreResult{
				value:       float64(0.5),
				needsUpdate: true,
				err:         nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			client := &fakeArtifactClient{}

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			//fetch score artifact
			scoreArtifact, err := getArtifact(ctx, client, "projects/rollup-formula-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score-lint-error", false)
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

			gotResult := processRollUpFormula(ctx, client, formula, test.resource, scoreArtifact, test.takeAction)

			opts := cmp.AllowUnexported(scoreResult{})
			if !cmp.Equal(test.wantResult, gotResult, opts) {
				t.Errorf("processScoreFormula() returned unexpected response, (-want, +got):\n%s", cmp.Diff(test.wantResult, gotResult, opts))
			}
		})
	}
}

func TestProcessScorePatternsTimestamp(t *testing.T) {
	tests := []struct {
		desc       string
		seed       []seeder.RegistryResource
		resource   patterns.ResourceInstance
		takeAction bool
		wantResult scoreCardResult
	}{
		{
			desc: "!takeAction and scoreCard is up-to-date",
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
					UpdateTime: timestamppb.Now(),
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
					UpdateTime: timestamppb.Now(),
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
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 3)),
				},
			},
			resource: patterns.SpecResource{
				Spec: &rpc.ApiSpec{
					Name: "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
			desc: "!takeAction and scoreCard is not recent enough",
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
					UpdateTime: timestamppb.Now(),
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
					UpdateTime: timestamppb.Now(),
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
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
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
			client := &fakeArtifactClient{}

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

			//fetch the ScoreCard artifact
			scoreCardArtifact, err := getArtifact(ctx, client, "projects/score-patterns-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/scorecard-quality", false)
			if err != nil {
				t.Errorf("failed to fetch the scoreCardArtifact from setup: %s", err)
			}

			gotResult := processScorePatterns(ctx, client, definition, test.resource, scoreCardArtifact, test.takeAction, "projects/score-patterns-test/locations/global")

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
