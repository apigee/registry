// Copyright 2023 Google LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package visitor

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

const (
	ProjectCount    = 1
	ApiCount        = 3
	VersionCount    = 2
	DeploymentCount = 2
	SpecCount       = 2
	ArtifactCount   = 2
)

func TestVisit(t *testing.T) {
	ctx, registryClient, adminClient, spec, deployment := setupVisitTests(t)

	tests := []struct {
		pattern string
		filter  string
		count   int
		fails   bool
	}{
		// Visit all resources of each type using collections.
		{
			pattern: "projects",
			count:   ProjectCount,
		},
		{
			pattern: "projects/-/locations/global/artifacts",
			count:   ProjectCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis",
			count:   ProjectCount * ApiCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/artifacts",
			count:   ProjectCount * ApiCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions",
			count:   ProjectCount * ApiCount * VersionCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/artifacts",
			count:   ProjectCount * ApiCount * VersionCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/specs",
			count:   ProjectCount * ApiCount * VersionCount * SpecCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/specs/-@",
			count:   ProjectCount * ApiCount * VersionCount * SpecCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/specs/-/artifacts",
			count:   ProjectCount * ApiCount * VersionCount * SpecCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/deployments",
			count:   ProjectCount * ApiCount * DeploymentCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/deployments/-@",
			count:   ProjectCount * ApiCount * DeploymentCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/deployments/-/artifacts",
			count:   ProjectCount * ApiCount * DeploymentCount * ArtifactCount,
		},
		// Visit all resources of each type using collections implied by a final dash.
		{
			pattern: "projects/-",
			count:   ProjectCount,
		},
		{
			pattern: "projects/-/locations/global/artifacts/-",
			count:   ProjectCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-",
			count:   ProjectCount * ApiCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/artifacts/-",
			count:   ProjectCount * ApiCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-",
			count:   ProjectCount * ApiCount * VersionCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/artifacts/-",
			count:   ProjectCount * ApiCount * VersionCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/specs/-",
			count:   ProjectCount * ApiCount * VersionCount * SpecCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/specs/-@-",
			count:   ProjectCount * ApiCount * VersionCount * SpecCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/specs/-/artifacts/-",
			count:   ProjectCount * ApiCount * VersionCount * SpecCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/deployments/-",
			count:   ProjectCount * ApiCount * DeploymentCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/deployments/-@-",
			count:   ProjectCount * ApiCount * DeploymentCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/deployments/-/artifacts/-",
			count:   ProjectCount * ApiCount * DeploymentCount * ArtifactCount,
		},
		// visit collections under specific resources
		{
			pattern: "projects/visit-test/locations/global/artifacts",
			count:   ArtifactCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis",
			count:   ApiCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/artifacts",
			count:   ArtifactCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions",
			count:   VersionCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0/artifacts",
			count:   ArtifactCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0/specs",
			count:   SpecCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0/specs/s0/artifacts",
			count:   ArtifactCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0/specs/s0@",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0/specs/s0@" + spec.RevisionId + "/artifacts",
			count:   ArtifactCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/deployments",
			count:   DeploymentCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/deployments/d0/artifacts",
			count:   ArtifactCount,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/deployments/d0@",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/deployments/d0@" + deployment.RevisionId + "/artifacts",
			count:   ArtifactCount,
		},
		// Visit specific resources
		{
			pattern: "projects/visit-test",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/artifacts/x0",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/artifacts/x0",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0/artifacts/x0",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0/specs/s0",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0/specs/s0@" + spec.RevisionId,
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/versions/v0/specs/s0/artifacts/x0",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/deployments/d0",
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/deployments/d0@" + deployment.RevisionId,
			count:   1,
		},
		{
			pattern: "projects/visit-test/locations/global/apis/a0/deployments/d0/artifacts/x0",
			count:   1,
		},
		// test filters
		{
			pattern: "projects/-",
			filter:  "project_id.contains('visit')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis",
			filter:  "api_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/-",
			filter:  "api_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0",
			filter:  "api_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/artifacts",
			filter:  "artifact_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/artifacts/-",
			filter:  "artifact_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/artifacts/x0",
			filter:  "artifact_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/versions",
			filter:  "version_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/versions/-",
			filter:  "version_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/versions/v0",
			filter:  "version_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/versions/v0/specs",
			filter:  "spec_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/versions/v0/specs/-",
			filter:  "spec_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/versions/v0/specs/s0",
			filter:  "spec_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/deployments",
			filter:  "deployment_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/deployments/-",
			filter:  "deployment_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/deployments/d0",
			filter:  "deployment_id.contains('0')",
			count:   1,
		},
		// filters used with specific resources should fail
		{
			pattern: "projects/my-project/locations/global/apis/a0",
			filter:  "api_id.contains('0')",
			fails:   true,
		},
		{
			pattern: "projects/my-project/locations/global/apis/a0/versions/v0",
			filter:  "version_id.contains('0')",
			fails:   true,
		},
		{
			pattern: "projects/my-project/locations/global/apis/a0/deployments/d0",
			filter:  "deployment_id.contains('0')",
			fails:   true,
		},
		// invalid patterns should fail
		{
			pattern: "projects/myproject/invalid",
			fails:   true,
		},
		{
			pattern: "projects/myproject/locations/global/invalid",
			fails:   true,
		},
		{
			pattern: "projects/myproject/locations/global/apis/-/invalid",
			fails:   true,
		},
	}
	for _, test := range tests {
		testname := test.pattern
		if test.filter != "" {
			testname = fmt.Sprintf("%s(--filter=%s)", test.pattern, test.filter)
		}
		t.Run(testname, func(t *testing.T) {
			v := &testVisitor{}
			err := Visit(ctx, v, VisitorOptions{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
				Pattern:        test.pattern,
				Filter:         test.filter,
			})
			if err != nil && !test.fails {
				t.Errorf("Visit() failed with error %s", err)
			}
			if err == nil && test.fails {
				t.Errorf("Visit() succeeded when it should have failed")
			}
			if err == nil && v.count != test.count {
				t.Errorf("Visit() visited %d resources, expected %d", v.count, test.count)
			}
		})
	}
	for _, test := range tests {
		testname := "unsupported:" + test.pattern
		if test.filter != "" {
			testname = fmt.Sprintf("unsupported:%s(--filter=%s)", test.pattern, test.filter)
		}
		t.Run(testname, func(t *testing.T) {
			v := &Unsupported{}
			err := Visit(ctx, v, VisitorOptions{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
				Pattern:        test.pattern,
				Filter:         test.filter,
			})
			if err == nil {
				t.Errorf("Visit() of Unsupported succeeded when it should have failed")
			}
		})
	}
}

type testVisitor struct {
	count int
}

func (v *testVisitor) ProjectHandler() ProjectHandler {
	return func(ctx context.Context, message *rpc.Project) error {
		fmt.Printf("+ %s\n", message.GetName())
		v.count++
		return nil
	}
}

func (v *testVisitor) ApiHandler() ApiHandler {
	return func(ctx context.Context, message *rpc.Api) error {
		fmt.Printf("+ %s\n", message.GetName())
		v.count++
		return nil
	}
}

func (v *testVisitor) VersionHandler() VersionHandler {
	return func(ctx context.Context, message *rpc.ApiVersion) error {
		fmt.Printf("+ %s\n", message.GetName())
		v.count++
		return nil
	}
}

func (v *testVisitor) DeploymentHandler() DeploymentHandler {
	return func(ctx context.Context, message *rpc.ApiDeployment) error {
		fmt.Printf("+ %s\n", message.GetName())
		v.count++
		return nil
	}
}

func (v *testVisitor) DeploymentRevisionHandler() DeploymentHandler {
	return func(ctx context.Context, message *rpc.ApiDeployment) error {
		fmt.Printf("+ %s\n", message.GetName())
		v.count++
		return nil
	}
}

func (v *testVisitor) SpecHandler() SpecHandler {
	return func(ctx context.Context, message *rpc.ApiSpec) error {
		fmt.Printf("+ %s\n", message.GetName())
		v.count++
		return nil
	}
}

func (v *testVisitor) SpecRevisionHandler() SpecHandler {
	return func(ctx context.Context, message *rpc.ApiSpec) error {
		fmt.Printf("+ %s\n", message.GetName())
		v.count++
		return nil
	}
}

func (v *testVisitor) ArtifactHandler() ArtifactHandler {
	return func(ctx context.Context, message *rpc.Artifact) error {
		fmt.Printf("+ %s\n", message.GetName())
		v.count++
		return nil
	}
}

func TestVisitSubtree(t *testing.T) {
	ctx, registryClient, adminClient, _, _ := setupVisitTests(t)

	tests := []struct {
		pattern string
		filter  string
		count   int
		fails   bool
	}{
		{
			pattern: "projects",
			count: ProjectCount * (1 + ArtifactCount +
				ApiCount*(1+ArtifactCount+
					VersionCount*(1+ArtifactCount+
						SpecCount+VersionCount*ArtifactCount)+
					DeploymentCount+DeploymentCount*ArtifactCount)),
		},
		{
			pattern: "projects/-/locations/global/artifacts",
			count:   ProjectCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis",
			count: ProjectCount * ApiCount * (1 + ArtifactCount +
				VersionCount*(1+ArtifactCount+
					SpecCount+VersionCount*ArtifactCount) +
				DeploymentCount + DeploymentCount*ArtifactCount),
		},
		{
			pattern: "projects/-/locations/global/apis/-/artifacts",
			count:   ProjectCount * ApiCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions",
			count: ProjectCount * ApiCount * VersionCount * (1 + ArtifactCount +
				SpecCount + SpecCount*ArtifactCount),
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/artifacts",
			count:   ProjectCount * ApiCount * VersionCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/specs",
			count:   ProjectCount * ApiCount * VersionCount * SpecCount * (1 + ArtifactCount),
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/specs/-@",
			count:   ProjectCount * ApiCount * VersionCount * SpecCount * (1 + ArtifactCount),
		},
		{
			pattern: "projects/-/locations/global/apis/-/versions/-/specs/-/artifacts",
			count:   ProjectCount * ApiCount * VersionCount * SpecCount * ArtifactCount,
		},
		{
			pattern: "projects/-/locations/global/apis/-/deployments",
			count:   ProjectCount * ApiCount * DeploymentCount * (1 + ArtifactCount),
		},
		{
			pattern: "projects/-/locations/global/apis/-/deployments/-@",
			count:   ProjectCount * ApiCount * DeploymentCount * (1 + ArtifactCount),
		},
		{
			pattern: "projects/-/locations/global/apis/-/deployments/-/artifacts",
			count:   ProjectCount * ApiCount * DeploymentCount * ArtifactCount,
		},
		// test filters
		{
			pattern: "projects/-",
			filter:  "project_id.contains('visit')",
			count: ProjectCount * (1 + ArtifactCount +
				ApiCount*(1+ArtifactCount+
					VersionCount*(1+ArtifactCount+
						SpecCount+VersionCount*ArtifactCount)+
					DeploymentCount+DeploymentCount*ArtifactCount)),
		},
		{
			pattern: "projects/-/locations/global/apis",
			filter:  "api_id.contains('0')",
			count: ProjectCount * (1 + ArtifactCount +
				VersionCount*(1+ArtifactCount+
					SpecCount+VersionCount*ArtifactCount) +
				DeploymentCount + DeploymentCount*ArtifactCount),
		},
		{
			pattern: "projects/-/locations/global/artifacts",
			filter:  "artifact_id.contains('0')",
			count:   1,
		},
		{
			pattern: "projects/-/locations/global/apis/a0/versions",
			filter:  "version_id.contains('0')",
			count: ProjectCount * (1 + ArtifactCount +
				SpecCount + SpecCount*ArtifactCount),
		},
		{
			pattern: "projects/-/locations/global/apis/a0/versions/v0/specs",
			filter:  "spec_id.contains('0')",
			count:   ProjectCount * (1 + ArtifactCount),
		},
		{
			pattern: "projects/-/locations/global/apis/a0/deployments",
			filter:  "deployment_id.contains('0')",
			count:   ProjectCount * (1 + ArtifactCount),
		},
	}
	for _, test := range tests {
		testname := test.pattern
		if test.filter != "" {
			testname = fmt.Sprintf("%s(--filter=%s)", test.pattern, test.filter)
		}
		t.Run(testname, func(t *testing.T) {
			opts := VisitorOptions{
				RegistryClient: registryClient,
				AdminClient:    adminClient,
				Pattern:        test.pattern,
				Filter:         test.filter,
			}
			v := &testVisitor{}
			subtreeVisitor := &SubtreeVisitor{
				Visitor: v,
				Options: opts,
			}
			err := Visit(ctx, subtreeVisitor, opts)
			if err != nil && !test.fails {
				t.Errorf("Visit() failed with error %s", err)
			}
			if err == nil && test.fails {
				t.Errorf("Visit() succeeded when it should have failed")
			}
			if err == nil && v.count != test.count {
				t.Errorf("Visit() visited %d resources, expected %d", v.count, test.count)
			}
		})
	}
}

func setupVisitTests(t *testing.T) (context.Context, *gapic.RegistryClient, *gapic.AdminClient, *rpc.ApiSpec, *rpc.ApiDeployment) {
	projectID := "visit-test"
	project := names.Project{ProjectID: projectID}
	ctx := context.Background()
	registryClient, adminClient := grpctest.SetupRegistry(ctx, t, projectID, nil)
	parent := project.String() + "/locations/global"

	for k := 0; k < ArtifactCount; k++ {
		_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
			ArtifactId: fmt.Sprintf("x%d", k),
			Parent:     parent,
			Artifact:   &rpc.Artifact{},
		})
		if err != nil {
			t.Fatalf("Setup: Failed to create test artifact: %s", err)
		}
	}
	for i := 0; i < ApiCount; i++ {
		api, err := registryClient.CreateApi(ctx, &rpc.CreateApiRequest{
			ApiId:  fmt.Sprintf("a%d", i),
			Parent: parent,
			Api:    &rpc.Api{},
		})
		if err != nil {
			t.Fatalf("Setup: Failed to create test api: %s", err)
		}
		apiName, err := names.ParseApi(api.Name)
		if err != nil {
			t.Fatalf("Setup: Failed to create test api: %s", err)
		}
		for k := 0; k < ArtifactCount; k++ {
			_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
				ArtifactId: fmt.Sprintf("x%d", k),
				Parent:     apiName.String(),
				Artifact:   &rpc.Artifact{},
			})
			if err != nil {
				t.Fatalf("Setup: Failed to create test artifact: %s", err)
			}
		}
		for j := 0; j < DeploymentCount; j++ {
			deployment, err := registryClient.CreateApiDeployment(ctx, &rpc.CreateApiDeploymentRequest{
				ApiDeploymentId: fmt.Sprintf("d%d", j),
				Parent:          apiName.String(),
				ApiDeployment:   &rpc.ApiDeployment{},
			})
			if err != nil {
				t.Fatalf("Setup: Failed to create test deployment: %s", err)
			}
			deploymentName, err := names.ParseDeployment(deployment.Name)
			if err != nil {
				t.Fatalf("Setup: Failed to create test deployment: %s", err)
			}
			for k := 0; k < ArtifactCount; k++ {
				_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
					ArtifactId: fmt.Sprintf("x%d", k),
					Parent:     deploymentName.String(),
					Artifact:   &rpc.Artifact{},
				})
				if err != nil {
					t.Fatalf("Setup: Failed to create test artifact: %s", err)
				}
			}
		}
		for j := 0; j < VersionCount; j++ {
			version, err := registryClient.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
				ApiVersionId: fmt.Sprintf("v%d", j),
				Parent:       apiName.String(),
				ApiVersion:   &rpc.ApiVersion{},
			})
			if err != nil {
				t.Fatalf("Setup: Failed to create test version: %s", err)
			}
			versionName, err := names.ParseVersion(version.Name)
			if err != nil {
				t.Fatalf("Setup: Failed to create test version: %s", err)
			}
			for k := 0; k < ArtifactCount; k++ {
				_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
					ArtifactId: fmt.Sprintf("x%d", k),
					Parent:     versionName.String(),
					Artifact:   &rpc.Artifact{},
				})
				if err != nil {
					t.Fatalf("Setup: Failed to create test artifact: %s", err)
				}
			}
			for k := 0; k < SpecCount; k++ {
				spec, err := registryClient.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
					ApiSpecId: fmt.Sprintf("s%d", k),
					Parent:    versionName.String(),
					ApiSpec:   &rpc.ApiSpec{},
				})
				if err != nil {
					t.Fatalf("Setup: Failed to create test spec: %s", err)
				}
				specName, err := names.ParseSpec(spec.Name)
				if err != nil {
					t.Fatalf("Setup: Failed to create test spec: %s", err)
				}
				for l := 0; l < ArtifactCount; l++ {
					_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
						ArtifactId: fmt.Sprintf("x%d", l),
						Parent:     specName.String(),
						Artifact:   &rpc.Artifact{},
					})
					if err != nil {
						t.Fatalf("Setup: Failed to create test artifact: %s", err)
					}
				}
			}
		}
	}
	// get the latest spec and revision ID
	spec, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{Name: "projects/visit-test/locations/global/apis/a0/versions/v0/specs/s0"})
	if err != nil {
		t.Fatalf("Failed to prepare test data: %+v", err)
	}
	deployment, err := registryClient.GetApiDeployment(ctx, &rpc.GetApiDeploymentRequest{Name: "projects/visit-test/locations/global/apis/a0/deployments/d0"})
	if err != nil {
		t.Fatalf("Failed to prepare test data: %+v", err)
	}

	return ctx, registryClient, adminClient, spec, deployment
}
