// Copyright 2022 Google LLC. All Rights Reserved.
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

package get

import (
	"bytes"
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

func TestGetValidResources(t *testing.T) {
	// Seed a registry with a list of leaf-level artifacts.
	displaySettingsBytes, err := proto.Marshal(&rpc.DisplaySettings{Organization: "Sample"})
	if err != nil {
		t.Fatalf("Failed to prepare test data: %+v", err)
	}
	artifacts := []*rpc.Artifact{
		{Name: "projects/my-project/locations/global/artifacts/x", MimeType: types.MimeTypeForKind("DisplaySettings"), Contents: displaySettingsBytes},
		{Name: "projects/my-project/locations/global/apis/a/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		{Name: "projects/my-project/locations/global/apis/a/versions/v/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		{Name: "projects/my-project/locations/global/apis/a/deployments/d/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
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
	if err := seeder.SeedArtifacts(ctx, client, artifacts...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	spec, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s"})
	if err != nil {
		t.Fatalf("Failed to prepare test data: %+v", err)
	}
	deployment, err := registryClient.GetApiDeployment(ctx, &rpc.GetApiDeploymentRequest{Name: "projects/my-project/locations/global/apis/a/deployments/d"})
	if err != nil {
		t.Fatalf("Failed to prepare test data: %+v", err)
	}
	// Verify that get runs for each resource.
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
		"projects/my-project/locations/global/apis/a/versions/v/specs/-@-",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s@",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s@-",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s@" + spec.RevisionId,
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/-",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x",
		"projects/my-project/locations/global/apis/a/deployments",
		"projects/my-project/locations/global/apis/a/deployments/-",
		"projects/my-project/locations/global/apis/a/deployments/-@-",
		"projects/my-project/locations/global/apis/a/deployments/d",
		"projects/my-project/locations/global/apis/a/deployments/d@",
		"projects/my-project/locations/global/apis/a/deployments/d@-",
		"projects/my-project/locations/global/apis/a/deployments/d@" + deployment.RevisionId,
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/-",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/x",
	}
	for _, r := range resources {
		cmd := Command()
		args := []string{r}
		cmd.SetArgs(args)
		out := bytes.NewBuffer(make([]byte, 0))
		cmd.SetOutput(out)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %v returned error: %s", args, err)
		}
	}
	resourcesWithContents := []string{
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/artifacts/x",
		"projects/my-project/locations/global/apis/a/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/x",
	}
	// Get the raw contents of these resources.
	for _, r := range resourcesWithContents {
		cmd := Command()
		args := []string{r, "--raw"}
		cmd.SetArgs(args)
		out := bytes.NewBuffer(make([]byte, 0))
		cmd.SetOutput(out)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %v returned error: %s", args, err)
		}
	}
	// Print the contents of these resources.
	for _, r := range resources {
		cmd := Command()
		args := []string{r, "--print"}
		cmd.SetArgs(args)
		out := bytes.NewBuffer(make([]byte, 0))
		cmd.SetOutput(out)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %v returned error: %s", args, err)
		}
	}
}

func TestGetInvalidResources(t *testing.T) {
	// Seed a registry with a list of leaf-level artifacts.
	const scoreType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.Score"
	artifacts := []*rpc.Artifact{
		{Name: "projects/my-project/locations/global/artifacts/x", MimeType: scoreType},
		{Name: "projects/my-project/locations/global/apis/a/artifacts/x", MimeType: scoreType},
		{Name: "projects/my-project/locations/global/apis/a/versions/v/artifacts/x", MimeType: scoreType},
		{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x", MimeType: scoreType},
		{Name: "projects/my-project/locations/global/apis/a/deployments/d/artifacts/x", MimeType: scoreType},
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
	if err := seeder.SeedArtifacts(ctx, client, artifacts...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	// Verify that invalid gets fail.
	invalid := []string{
		"projects/my-project/locations/global/invalid",
		"projects/my-project/locations/global/apis/-/invalid",
		"projects/my-project/locations/global/apis/a/invalid",
		"projects/my-project/locations/global/apis/a/versions/v/invalid",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/invalid",
		"projects/my-project/locations/global/apis/a/deployments/d/invalid",
		"projects/my-project/locations/global/artifacts/xx",
		"projects/my-project/locations/global/apis/a/versions/vv",
		"projects/my-project/locations/global/apis/a/versions/v/specs/ss",
		"projects/my-project/locations/global/apis/a/deployments/dd",
		"projects/my-project/locations/global/apis/aa",
		"projects/my-project/locations/global/apis/a/versions/vv",
		"projects/my-project/locations/global/apis/a/versions/v/specs/ss",
		"projects/my-project/locations/global/apis/a/deployments/dd",
		"projects/my-project/locations/global/apis/a/artifacts/xx",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/xx",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/xx",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/xx",
	}
	for _, r := range invalid {
		cmd := Command()
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		args := []string{r}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err == nil {
			t.Fatalf("Execute() with args %v succeeded but should have failed", args)
		}
	}
}
