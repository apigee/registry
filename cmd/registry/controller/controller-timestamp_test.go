package controller

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/apigee/registry/pkg/application/controller"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeLister struct {
	projects    []*rpc.Project
	apis        []*rpc.Api
	versions    []*rpc.ApiVersion
	specs       []*rpc.ApiSpec
	artifacts   []*rpc.Artifact
	deployments []*rpc.ApiDeployment
}

// This implementation doesn't support filter functionality
func (f *fakeLister) ListAPIs(ctx context.Context, api names.Api, filter string, handler visitor.ApiHandler) error {
	for _, a := range f.apis {
		name, _ := names.ParseApi(a.GetName())
		if strings.Contains(filter, name.Parent()) || (api.ApiID != "-" && name.ApiID != api.ApiID) {
			continue
		}
		if err := handler(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeLister) ListVersions(ctx context.Context, version names.Version, filter string, handler visitor.VersionHandler) error {
	for _, v := range f.versions {
		name, _ := names.ParseVersion(v.GetName())
		if strings.Contains(filter, name.Parent()) || (version.VersionID != "-" && name.VersionID != version.VersionID) {
			continue
		}
		if err := handler(ctx, v); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeLister) ListSpecs(ctx context.Context, spec names.Spec, filter string, handler visitor.SpecHandler) error {
	for _, s := range f.specs {
		name, _ := names.ParseSpec(s.GetName())
		if strings.Contains(filter, name.Parent()) || (spec.SpecID != "-" && name.SpecID != spec.SpecID) {
			continue
		}

		if err := handler(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeLister) ListArtifacts(ctx context.Context, artifact names.Artifact, filter string, contents bool, handler visitor.ArtifactHandler) error {
	for _, a := range f.artifacts {
		name, _ := names.ParseArtifact(a.GetName())
		if strings.Contains(filter, name.Parent()) || (artifact.ArtifactID() != "-" && name.ArtifactID() != artifact.ArtifactID()) {
			continue
		}

		if err := handler(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

// These functions are needed to use the fakeLister with the seeder package.
func (f *fakeLister) CreateProject(ctx context.Context, req *rpc.CreateProjectRequest) (*rpc.Project, error) {
	project := &rpc.Project{
		Name: fmt.Sprintf("projects/%s", req.GetProjectId()),
	}
	f.projects = append(f.projects, project)
	return project, nil
}

func (f *fakeLister) UpdateApi(ctx context.Context, req *rpc.UpdateApiRequest) (*rpc.Api, error) {
	api := req.Api
	if api.UpdateTime == nil {
		api.UpdateTime = timestamppb.Now()
	}
	for i, a := range f.apis {
		if a.GetName() == api.GetName() {
			// remove the old copy
			f.apis = append(f.apis[:i], f.apis[i+1:]...)
			break
		}
	}
	f.apis = append(f.apis, api)
	return api, nil
}

func (f *fakeLister) UpdateApiVersion(ctx context.Context, req *rpc.UpdateApiVersionRequest) (*rpc.ApiVersion, error) {
	version := req.ApiVersion
	if version.UpdateTime == nil {
		version.UpdateTime = timestamppb.Now()
	}
	for i, v := range f.versions {
		if v.GetName() == version.GetName() {
			// remove the old copy
			f.versions = append(f.versions[:i], f.versions[i+1:]...)
			break
		}
	}
	f.versions = append(f.versions, version)
	return version, nil
}

func (f *fakeLister) UpdateApiSpec(ctx context.Context, req *rpc.UpdateApiSpecRequest) (*rpc.ApiSpec, error) {
	spec := req.ApiSpec
	if spec.RevisionUpdateTime == nil {
		spec.RevisionUpdateTime = timestamppb.Now()
	}
	for i, s := range f.specs {
		if s.GetName() == spec.GetName() {
			// remove the old copy
			f.specs = append(f.specs[:i], f.specs[i+1:]...)
			break
		}
	}
	f.specs = append(f.specs, spec)
	return spec, nil
}

func (f *fakeLister) UpdateApiDeployment(ctx context.Context, req *rpc.UpdateApiDeploymentRequest) (*rpc.ApiDeployment, error) {
	deployment := req.ApiDeployment
	if deployment.RevisionUpdateTime == nil {
		deployment.RevisionUpdateTime = timestamppb.Now()
	}
	for i, d := range f.deployments {
		if d.GetName() == deployment.GetName() {
			// remove the old copy
			f.deployments = append(f.deployments[:i], f.deployments[i+1:]...)
			break
		}
	}
	f.deployments = append(f.deployments, deployment)
	return deployment, nil
}

func (f *fakeLister) CreateArtifact(ctx context.Context, req *rpc.CreateArtifactRequest) (*rpc.Artifact, error) {
	artifact := req.Artifact
	if artifact.UpdateTime == nil {
		artifact.UpdateTime = timestamppb.Now()
	}
	f.artifacts = append(f.artifacts, artifact)
	return artifact, nil
}

func (f *fakeLister) ReplaceArtifact(ctx context.Context, req *rpc.ReplaceArtifactRequest) (*rpc.Artifact, error) {
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

// Tests for artifacts as resources and specs as dependencies
func TestTimestampArtifacts(t *testing.T) {
	tests := []struct {
		desc string
		seed []seeder.RegistryResource
		want []*Action
	}{
		{
			desc: "partially existing artifacts",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:     "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					MimeType:           gzipOpenAPIv3,
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 2)),
				},
				&rpc.ApiSpec{
					Name:     "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
			},
			want: []*Action{
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
				},
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/lint-gnostic",
				},
			},
		},
		{
			desc: "outdated artifacts",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType:           gzipOpenAPIv3,
					RevisionUpdateTime: timestamppb.New(time.Now()),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 3)),
				},
				&rpc.ApiSpec{
					Name:     "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.Artifact{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/lint-gnostic",
				},
				&rpc.ApiSpec{
					Name:     "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
				&rpc.ApiSpec{
					Name:     "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					MimeType: gzipOpenAPIv3,
				},
			},
			want: []*Action{
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/lint-gnostic",
				},
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/lint-gnostic",
				},
			},
		},
		{
			desc: "not recent enough artifacts",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					MimeType:           gzipOpenAPIv3,
					RevisionUpdateTime: timestamppb.New(time.Now()),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					MimeType:           gzipOpenAPIv3,
					RevisionUpdateTime: timestamppb.New(time.Now()),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
					MimeType:           gzipOpenAPIv3,
					RevisionUpdateTime: timestamppb.New(time.Now()),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
				},
			},
			want: []*Action{
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
				},
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/lint-gnostic",
				},
				{
					Command:           "registry compute lint projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi --linter gnostic",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/lint-gnostic",
				},
			},
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			client := new(fakeLister)

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			manifest := &controller.Manifest{
				Id: "controller-test",
				GeneratedResources: []*controller.GeneratedResource{
					{
						Pattern: "apis/-/versions/-/specs/-/artifacts/lint-gnostic",
						Dependencies: []*controller.Dependency{
							{
								Pattern: "$resource.spec",
								Filter:  "mime_type.contains('openapi')",
							},
						},
						Action: "registry compute lint $resource.spec --linter gnostic",
					},
				},
			}
			actions := ProcessManifest(ctx, client, projectID, manifest, 10)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}
		})
	}
}

// Tests for aggregated artifacts at api level and specs as resources
func TestTimestampAggregateArtifacts(t *testing.T) {
	tests := []struct {
		desc string
		seed []seeder.RegistryResource
		want []*Action
	}{
		{
			desc: "outdated artifacts",
			seed: []seeder.RegistryResource{
				// test api 1
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.0/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-1/versions/1.1.0/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.1/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/test-api-1/artifacts/vocabulary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 3)),
				},
				// test api 2
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.0/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-2/versions/1.1.0/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.1/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 3)),
				},
				// Update underlying spec to make artifact outdated
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.1/specs/openapi",
					RevisionUpdateTime: timestamppb.New(time.Now().Add(time.Second * 4)),
				},
			},
			want: []*Action{
				{
					Command:           "registry compute vocabulary projects/controller-test/locations/global/apis/test-api-2",
					GeneratedResource: "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary",
				},
			},
		},
		{
			desc: "not recent enough artifacts",
			seed: []seeder.RegistryResource{
				// test api 1
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.0/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-1/versions/1.1.0/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.1/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/test-api-1/artifacts/vocabulary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)), // should be at least 2 seconds newer than dependencies.
				},
				// test api 2
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.0/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-2/versions/1.1.0/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.1/specs/openapi",
					RevisionUpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)), // should be at least 2 seconds newer than dependencies.
				},
			},
			want: []*Action{
				{
					Command:           "registry compute vocabulary projects/controller-test/locations/global/apis/test-api-1",
					GeneratedResource: "projects/controller-test/locations/global/apis/test-api-1/artifacts/vocabulary",
				},
				{
					Command:           "registry compute vocabulary projects/controller-test/locations/global/apis/test-api-2",
					GeneratedResource: "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary",
				},
			},
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			client := new(fakeLister)

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			manifest := &controller.Manifest{

				Id: "controller-test",
				GeneratedResources: []*controller.GeneratedResource{
					{
						Pattern: "apis/-/artifacts/vocabulary",
						Dependencies: []*controller.Dependency{
							{
								Pattern: "$resource.api/versions/-/specs/-",
							},
						},
						Action: "registry compute vocabulary $resource.api",
					},
				},
			}
			actions := ProcessManifest(ctx, client, projectID, manifest, 10)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}
		})
	}
}

// Tests for derived artifacts with artifacts as dependencies
func TestTimestampDerivedArtifacts(t *testing.T) {
	tests := []struct {
		desc string
		seed []seeder.RegistryResource
		want []*Action
	}{
		{
			desc: "outdated artifacts",
			seed: []seeder.RegistryResource{
				// version 1.0.0
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/summary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 2)),
				},
				// version 1.0.1
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/complexity",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/summary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 2)),
				},
				// version 1.1.0
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/complexity",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/summary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 2)),
				},
				// Make some artifacts outdated from the above setup
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 4)),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/complexity",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 4)),
				},
			},
			want: []*Action{
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/summary",
				},
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/summary",
				},
			},
		},
		{
			desc: "not recent enough artifacts",
			seed: []seeder.RegistryResource{
				// version 1.0.0
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/summary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
				},
				// version 1.0.1
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/complexity",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/summary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
				},
				// version 1.1.0
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/lint-gnostic",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/complexity",
					UpdateTime: timestamppb.Now(),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/summary",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * 1)),
				},
			},
			want: []*Action{
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/summary",
				},
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/summary",
				},
				{
					Command: fmt.Sprintf(
						"registry compute summary %s %s",
						"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/lint-gnostic",
						"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/complexity"),
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/summary",
				},
			},
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			client := new(fakeLister)

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			manifest := &controller.Manifest{
				Id: "controller-test",
				GeneratedResources: []*controller.GeneratedResource{
					{
						Pattern: "apis/-/versions/-/specs/-/artifacts/summary",
						Dependencies: []*controller.Dependency{
							{
								Pattern: "$resource.spec/artifacts/lint-gnostic",
							},
							{
								Pattern: "$resource.spec/artifacts/complexity",
							},
						},
						Action: "registry compute summary $resource.spec/artifacts/lint-gnostic $resource.spec/artifacts/complexity",
					},
				},
			}
			actions := ProcessManifest(ctx, client, projectID, manifest, 10)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}
		})
	}
}

func TestRefreshArtifacts(t *testing.T) {
	tests := []struct {
		desc string
		seed []seeder.RegistryResource
		want []*Action
	}{
		{
			desc: "non-existing artifacts",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
				&rpc.ApiSpec{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
				},
				&rpc.ApiSpec{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
				},
			},
			want: []*Action{
				{
					Command:           "registry compute score projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-receipt",
					RequiresReceipt:   true,
				},
				{
					Command:           "registry compute score projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/score-receipt",
					RequiresReceipt:   true,
				},
				{
					Command:           "registry compute score projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/score-receipt",
					RequiresReceipt:   true,
				},
			},
		},
		{
			desc: "existing valid artifacts",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
				},
				&rpc.ApiSpec{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
				},
				&rpc.ApiSpec{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
				},
				&rpc.Artifact{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-receipt",
				},
				&rpc.Artifact{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/score-receipt",
				},
				&rpc.Artifact{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/score-receipt",
				},
			},
			want: nil,
		},
		{
			desc: "existing invalid artifacts",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					RevisionUpdateTime: timestamppb.New(time.Now().Add(time.Second * -5)),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					RevisionUpdateTime: timestamppb.New(time.Now().Add(time.Second * -5)),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
					RevisionUpdateTime: timestamppb.New(time.Now().Add(time.Second * -5)),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-receipt",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * -3)),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/score-receipt",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * -3)),
				},
				&rpc.Artifact{
					Name:       "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/score-receipt",
					UpdateTime: timestamppb.New(time.Now().Add(time.Second * -3)),
				},
			},
			want: []*Action{
				{
					Command:           "registry compute score projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-receipt",
					RequiresReceipt:   true,
				},
				{
					Command:           "registry compute score projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/score-receipt",
					RequiresReceipt:   true,
				},
				{
					Command:           "registry compute score projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
					GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/score-receipt",
					RequiresReceipt:   true,
				},
			},
		},
		{
			desc: "existing valid artifacts",
			seed: []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
					RevisionUpdateTime: timestamppb.New(time.Now().Add(time.Second * -2)),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
					RevisionUpdateTime: timestamppb.New(time.Now().Add(time.Second * -2)),
				},
				&rpc.ApiSpec{
					Name:               "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
					RevisionUpdateTime: timestamppb.New(time.Now().Add(time.Second * -2)),
				},
				&rpc.Artifact{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/score-receipt",
				},
				&rpc.Artifact{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/score-receipt",
				},
				&rpc.Artifact{
					Name: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/score-receipt",
				},
			},
			want: nil,
		},
	}

	const projectID = "controller-test"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			client := new(fakeLister)

			if err := seeder.SeedRegistry(ctx, client, test.seed...); err != nil {
				t.Fatalf("Setup: failed to seed registry: %s", err)
			}

			// time.Sleep(test.wait * time.Second)

			manifest := &controller.Manifest{
				Id: "controller-test",
				GeneratedResources: []*controller.GeneratedResource{
					{
						Pattern: "apis/-/versions/-/specs/-/artifacts/score-receipt",
						Receipt: true,
						Refresh: &durationpb.Duration{
							Seconds: 2,
						},
						Action: "registry compute score $resource.spec",
					},
				},
			}
			actions := ProcessManifest(ctx, client, projectID, manifest, 10)

			if diff := cmp.Diff(test.want, actions, sortActions); diff != "" {
				t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
			}
		})
	}
}
