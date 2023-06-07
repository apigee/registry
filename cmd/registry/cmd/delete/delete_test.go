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

package delete

import (
	"context"
	"testing"

	"github.com/apigee/registry/pkg/application/apihub"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	"google.golang.org/protobuf/proto"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func setup(t *testing.T) (context.Context, connection.RegistryClient) {
	// Seed a registry with a list of leaf-level artifacts.
	displaySettingsBytes, err := proto.Marshal(&apihub.DisplaySettings{Organization: "Sample"})
	if err != nil {
		t.Fatalf("Failed to prepare test data: %+v", err)
	}
	seed := []seeder.RegistryResource{
		&rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s", MimeType: "text/plain", Contents: []byte("hello")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/x", MimeType: mime.MimeTypeForKind("DisplaySettings"), Contents: displaySettingsBytes},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/a/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/a/versions/v/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/a/deployments/d/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/b/versions/v/specs/s/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
	}
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "my-project", seed)
	return ctx, registryClient
}

func TestDeleteValidResources(t *testing.T) {
	// Verify that delete with --force succeeds for each resource.
	resources := []string{
		"projects",
		"projects/-",
		"projects/my-project",
		"projects/my-project/locations/global/artifacts",
		"projects/my-project/locations/global/artifacts/-",
		"projects/my-project/locations/global/artifacts/x",
		"projects/my-project/locations/global/apis",
		"projects/my-project/locations/global/apis/-",
		"projects/my-project/locations/global/apis/a",
		"projects/my-project/locations/global/apis/a/artifacts",
		"projects/my-project/locations/global/apis/a/artifacts/-",
		"projects/my-project/locations/global/apis/a/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions",
		"projects/my-project/locations/global/apis/a/versions/-",
		"projects/my-project/locations/global/apis/a/versions/v",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/-",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/specs",
		"projects/my-project/locations/global/apis/a/versions/v/specs/-",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/-",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x",
		"projects/my-project/locations/global/apis/a/deployments",
		"projects/my-project/locations/global/apis/a/deployments/-",
		"projects/my-project/locations/global/apis/a/deployments/d",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/-",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/x",
	}
	// Try to delete each resource.
	for _, r := range resources {
		t.Run(r, func(t *testing.T) {
			setup(t)
			cmd := Command()
			args := []string{r, "--force"}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				t.Errorf("Execute() with args %v should have succeeded but failed", args)
			}
		})
	}
}

func setupWithoutArtifacts(t *testing.T) (context.Context, connection.RegistryClient) {
	// Seed a registry with a list of leaf-level artifacts.
	seed := []seeder.RegistryResource{
		&rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s", MimeType: "text/plain", Contents: []byte("hello")},
		&rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
	}
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "my-project", seed)
	return ctx, registryClient
}

func TestDeleteValidRevisions(t *testing.T) {
	// Specifically delete a spec revision.
	t.Run("spec revision", func(t *testing.T) {
		ctx, registryClient := setupWithoutArtifacts(t)
		spec, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s"})
		if err != nil {
			t.Fatalf("Failed to prepare test data: %+v", err)
		}
		// Update spec to create a second revision.
		spec.Contents = []byte("goodbye")
		_, err = registryClient.UpdateApiSpec(ctx, &rpc.UpdateApiSpecRequest{ApiSpec: spec})
		if err != nil {
			t.Fatalf("Failed to prepare test data: %+v", err)
		}
		// Delete the original revision.
		r := "projects/my-project/locations/global/apis/a/versions/v/specs/s@" + spec.RevisionId
		cmd := Command()
		args := []string{r, "--force"}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Errorf("Execute() with args %v should have succeeded but failed", args)
		}
	})
	// Specifically delete a deployment revision.
	t.Run("deployment revision", func(t *testing.T) {
		ctx, registryClient := setupWithoutArtifacts(t)
		deployment, err := registryClient.GetApiDeployment(ctx, &rpc.GetApiDeploymentRequest{Name: "projects/my-project/locations/global/apis/a/deployments/d"})
		if err != nil {
			t.Fatalf("Failed to prepare test data: %+v", err)
		}
		// Update deployment to create a second revision
		deployment.EndpointUri = "https://another"
		_, err = registryClient.UpdateApiDeployment(ctx, &rpc.UpdateApiDeploymentRequest{ApiDeployment: deployment})
		if err != nil {
			t.Fatalf("Failed to prepare test data: %+v", err)
		}
		// Delete the original revision.
		r := "projects/my-project/locations/global/apis/a/deployments/d@" + deployment.RevisionId
		cmd := Command()
		args := []string{r, "--force"}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Errorf("Execute() with args %v should have succeeded but failed", args)
		}
	})
}

func TestDeleteInvalidResources(t *testing.T) {
	setup(t)
	resources := []string{
		"projects/my-project/locations/global/apis/-/versions/-/specs/missing-spec",
		"projects/my-project/locations/global/apis/-/invalid-collection",
		"projects/my-project/locations/global/apis/a/invalid-collection/x",
	}
	// Try to delete each resource.
	for _, r := range resources {
		t.Run(r, func(t *testing.T) {
			cmd := Command()
			args := []string{r, "--force"}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Errorf("Execute() with args %v succeeded but should have failed", args)
			}
		})
	}
	// Verify that we get an error if --filter is used with a specific resource.
	t.Run("filter-with-specific-resource", func(t *testing.T) {
		cmd := Command()
		args := []string{"projects/my-project/locations/global/apis/a", "--filter", "true"}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err == nil {
			t.Errorf("Execute() with args %v succeeded but should have failed", args)
		}
	})
}

func TestTaskString(t *testing.T) {
	task := &deleteApiTask{
		deleteTask: deleteTask{resourceName: "sample"},
	}
	if task.String() != "delete sample" {
		t.Errorf("deleteTask.String() returned incorrect value %s, expected %s", task.String(), "delete sample")
	}
}
