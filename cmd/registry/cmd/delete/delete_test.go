// Copyright 2023 Google LLC. All Rights Reserved.
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

	"github.com/apigee/registry/cmd/registry/types"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	"google.golang.org/protobuf/proto"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func setup(t *testing.T) (context.Context, connection.RegistryClient) {
	// Seed a registry with a list of leaf-level artifacts.
	displaySettingsBytes, err := proto.Marshal(&rpc.DisplaySettings{Organization: "Sample"})
	if err != nil {
		t.Fatalf("Failed to prepare test data: %+v", err)
	}
	seed := []seeder.RegistryResource{
		&rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s", MimeType: "text/plain", Contents: []byte("hello")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/artifacts/x", MimeType: types.MimeTypeForKind("DisplaySettings"), Contents: displaySettingsBytes},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/a/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/a/versions/v/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/a/deployments/d/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		&rpc.Artifact{Name: "projects/my-project/locations/global/apis/b/versions/v/specs/s/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
	}
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
	client := seeder.Client{
		RegistryClient: registryClient,
		AdminClient:    adminClient,
	}
	t.Cleanup(func() {
		_ = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{Name: "projects/my-project", Force: true})
	})
	_ = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{Name: "projects/my-project", Force: true})
	if err := seeder.SeedRegistry(ctx, client, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
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
	// try to delete each resource
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
