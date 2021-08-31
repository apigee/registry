// Copyright 2021 Google LLC
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

package controller

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var sortActions = cmpopts.SortSlices(func(a, b *Action) bool { return a.Command < b.Command })

const gzipOpenAPIv3 = "application/x.openapi+gzip;version=3.0.0"

var manifests = []*rpc.Manifest{
	{
		Name: "controller-test",
		GeneratedResources: []*rpc.GeneratedResource{
			{
				Pattern: "apis/-/versions/-/specs/-/artifacts/lint-gnostic",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.spec",
						Filter:  "mime_type.contains('openapi')",
					},
				},
				Action: "compute lint $0 --linter gnostic",
			},
		},
	},
	{
		Name: "controller-test",
		GeneratedResources: []*rpc.GeneratedResource{
			{
				Pattern: "apis/-/artifacts/vocabulary",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.api/versions/-/specs/-",
					},
				},
				Action: "compute vocabulary $0.api",
			},
		},
	},
	{
		Name: "controller-test",
		GeneratedResources: []*rpc.GeneratedResource{
			{
				Pattern: "apis/-/versions/-/specs/-/artifacts/score",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.spec/artifacts/lint-gnostic",
					},
					{
						Pattern: "$resource.spec/artifacts/complexity",
					},
				},
				Action: "compute score $0 $1",
			},
		},
	},
	{
		Name: "controller-test",
		GeneratedResources: []*rpc.GeneratedResource{
			{
				Pattern:     "apis/-/versions/-/specs/-/artifacts/custom-artifact",
				Placeholder: true,
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.spec",
					},
				},
				Action: "exec command $0",
			},
		},
	},
	{
		Name: "controller-test",
		GeneratedResources: []*rpc.GeneratedResource{
			{
				Pattern:     "artifacts/search-index",
				Placeholder: true,
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "apis/-/versions/-/specs/-",
					},
				},
				// TODO: Add support for "$0.pattern" reference
				Action: "compute search-index projects/controller-test/locations/global/apis/-/versions/-/specs/-",
			},
		},
	},
}

func deleteProject(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	projectID string) {
	t.Helper()
	req := &rpc.DeleteProjectRequest{
		Name: "projects/" + projectID,
	}
	err := client.DeleteProject(ctx, req)
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Failed DeleteProject(%v): %s", req, err.Error())
	}
}

func createProject(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	projectID string) {
	t.Helper()
	req := &rpc.CreateProjectRequest{
		ProjectId: projectID,
		Project: &rpc.Project{
			DisplayName: "Demo",
			Description: "A demo catalog",
		},
	}
	project, err := client.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed CreateProject(%v): %s", req, err.Error())
	}
	if project.GetName() != "projects/"+"controller-test" {
		t.Fatalf("Invalid project name %s", project.GetName())
	}
}

func createApi(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	parent, apiID string) {
	t.Helper()
	req := &rpc.CreateApiRequest{
		Parent: parent,
		ApiId:  apiID,
		Api: &rpc.Api{
			DisplayName:  apiID,
			Description:  fmt.Sprintf("Sample Test API: %s", apiID),
			Availability: "GENERAL",
		},
	}
	_, err := client.CreateApi(ctx, req)
	if err != nil {
		t.Fatalf("Failed CreateApi(%v): %s", req, err.Error())
	}
}

func createVersion(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	parent, versionID string) {
	t.Helper()
	req := &rpc.CreateApiVersionRequest{
		Parent:       parent,
		ApiVersionId: versionID,
		ApiVersion:   &rpc.ApiVersion{},
	}
	_, err := client.CreateApiVersion(ctx, req)
	if err != nil {
		t.Fatalf("Failed CreateApiVersion(%v): %s", req, err.Error())
	}
}

func createSpec(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	parent string,
	specId string,
	mimeType string,
) {
	t.Helper()
	// Create a spec entry with empty data
	req := &rpc.CreateApiSpecRequest{
		Parent:    parent,
		ApiSpecId: specId,
		ApiSpec: &rpc.ApiSpec{
			MimeType: mimeType,
		},
	}
	_, err := client.CreateApiSpec(ctx, req)
	if err != nil {
		t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
	}
}

func updateSpec(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	specName string) {
	t.Helper()
	req := &rpc.UpdateApiSpecRequest{
		ApiSpec: &rpc.ApiSpec{
			Name:     specName,
			MimeType: gzipOpenAPIv3,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"mime_type"}},
	}
	_, err := client.UpdateApiSpec(ctx, req)
	if err != nil {
		t.Fatalf("Failed UpdateApiSpec(%v): %s", req, err.Error())
	}
}

func createUpdateArtifact(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	artifactName string) {
	t.Helper()
	// Creates an artifact entry with empty data
	artifact := &rpc.Artifact{
		Name: artifactName,
	}
	err := core.SetArtifact(ctx, client, artifact)
	if err != nil {
		t.Fatalf("Failed SetArtifact(%v): %s", artifact, err.Error())
	}
}

// Tests for artifacts and resources and specs as dependencies

func TestSingleSpec(t *testing.T) {
	// Setup: Single spec in the project
	// Expect: One single command to compute artifact
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)

	// Test the manifest
	manifest := manifests[0]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []*Action{
		{
			Command:           "compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml --linter gnostic",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

func TestMultipleSpecs(t *testing.T) {
	// Setup: 3 specs in project
	// Expect: Create artifact command from scratch for each spec

	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")
	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

	// Test the manifest
	manifest := manifests[0]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}

	expectedActions := []*Action{
		{
			Command:           "compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml --linter gnostic",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
		},
		{
			Command:           "compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml --linter gnostic",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic",
		},
		{
			Command:           "compute lint projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml --linter gnostic",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
		},
	}

	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

func TestPartiallyExistingArtifacts(t *testing.T) {
	// Setup: 3 specs in project
	// Artifact already exists for one of the specs
	// Expect: Create artifact command for the remaining two specs

	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")
	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

	// Test the manifest
	manifest := manifests[0]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}

	expectedActions := []*Action{
		{
			Command:           "compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml --linter gnostic",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
		},
		{
			Command:           "compute lint projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml --linter gnostic",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

func TestOutdatedArtifacts(t *testing.T) {
	// Setup: 3 specs in project, 2 artifacts already exist, one of them is outdated
	// Expect: Create artifact command for the non-existing and the outdated artifacts.

	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")
	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
	// Update spec 1.0.1 to make the artifact outdated
	updateSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml")

	// Test the manifest
	manifest := manifests[0]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}

	expectedActions := []*Action{
		{
			Command:           "compute lint projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml --linter gnostic",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic",
		},
		{
			Command:           "compute lint projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml --linter gnostic",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

// Tests for aggregated artifacts at api level and specs as resources
func TestApiLevelArtifactsCreate(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "test-api-1")
	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

	// Test API 2
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "test-api-2")
	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

	// Test the manifest
	manifest := manifests[1]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []*Action{
		{
			Command:           "compute vocabulary projects/controller-test/locations/global/apis/test-api-1",
			GeneratedResource: "projects/controller-test/locations/global/apis/test-api-1/artifacts/vocabulary",
		},
		{
			Command:           "compute vocabulary projects/controller-test/locations/global/apis/test-api-2",
			GeneratedResource: "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary",
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

func TestApiLevelArtifactsOutdated(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "test-api-1")
	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-1/artifacts/vocabulary")

	// Test API 2
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "test-api-2")
	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary")
	// Update underlying spec to make artifact outdated
	updateSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/test-api-2/versions/1.0.1/specs/openapi.yaml")

	// Test the manifest
	manifest := manifests[1]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []*Action{
		{
			Command:           "compute vocabulary projects/controller-test/locations/global/apis/test-api-2",
			GeneratedResource: "projects/controller-test/locations/global/apis/test-api-2/artifacts/vocabulary",
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

// Tests for derived artifacts with artifacts as dependencies
func TestDerivedArtifactsCreate(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")
	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity")
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity")
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity")

	// Test the manifest
	manifest := manifests[2]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []*Action{
		{
			Command: fmt.Sprintf(
				"compute score %s %s",
				"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
				"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity"),
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score",
		},
		{
			Command: fmt.Sprintf(
				"compute score %s %s",
				"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic",
				"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity"),
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/score",
		},
		{
			Command: fmt.Sprintf(
				"compute score %s %s",
				"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
				"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity"),
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/score",
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

func TestDerivedArtifactsMissing(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")
	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity")
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity")

	// Test the manifest
	manifest := manifests[2]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []*Action{
		{
			Command: fmt.Sprintf(
				"compute score %s %s",
				"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic",
				"projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity"),
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/score",
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

func TestDerivedArtifactsOutdated(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")

	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score")
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/lint-gnostic")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/complexity")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/score")
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/score")

	// Make some artifacts outdated from the above setup
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic")
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity")

	// Test the manifest
	manifest := manifests[2]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}

	expectedActions := []*Action{
		{
			Command: fmt.Sprintf(
				"compute score %s %s",
				"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/lint-gnostic",
				"projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/complexity"),
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/score",
		},
		{
			Command: fmt.Sprintf(
				"compute score %s %s",
				"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/lint-gnostic",
				"projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/complexity"),
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/score",
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

func TestPlaceholderArtifacts(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")

	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

	// Test the manifest
	manifest := manifests[3]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []*Action{
		{
			Command:           "exec command projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/custom-artifact",
			Placeholder:       true,
		},
		{
			Command:           "exec command projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi.yaml/artifacts/custom-artifact",
			Placeholder:       true,
		},
		{
			Command:           "exec command projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml",
			GeneratedResource: "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi.yaml/artifacts/custom-artifact",
			Placeholder:       true,
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

func TestPlaceholderAggArtifacts(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")

	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.1.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

	// Test the manifest
	manifest := manifests[4]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []*Action{
		{
			Command:           "compute search-index projects/controller-test/locations/global/apis/-/versions/-/specs/-",
			GeneratedResource: "projects/controller-test/locations/global/artifacts/search-index",
			Placeholder:       true,
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}

func TestPlaceholderAggUpdatedArtifacts(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()

	// Setup
	deleteProject(ctx, registryClient, t, "controller-test")
	createProject(ctx, registryClient, t, "controller-test")
	createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")

	// Version 1.0.0
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
	// Version 1.0.1
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
	// Create target artifact
	createUpdateArtifact(ctx, registryClient, t, "projects/controller-test/locations/global/artifacts/search-index")

	// Add a new spec to make the artifact outdated
	createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
	createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

	// Test the manifest
	manifest := manifests[4]
	actions, err := ProcessManifest(ctx, registryClient, "controller-test", manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []*Action{
		{
			Command:           "compute search-index projects/controller-test/locations/global/apis/-/versions/-/specs/-",
			GeneratedResource: "projects/controller-test/locations/global/artifacts/search-index",
			Placeholder:       true,
		},
	}
	if diff := cmp.Diff(expectedActions, actions, sortActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

	deleteProject(ctx, registryClient, t, "controller-test")
}
