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

package export

import (
	"context"
	"io"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestExportYAML(t *testing.T) {
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
	if err := seeder.SeedArtifacts(ctx, client, artifacts...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	// Verify that export runs for each supported resource.
	resources := []string{
		"projects/my-project",
		"projects/my-project/locations/global/apis",
		"projects/my-project/locations/global/apis/-",
		"projects/my-project/locations/global/apis/a",
		"projects/my-project/locations/global/apis/a/versions",
		"projects/my-project/locations/global/apis/a/versions/-",
		"projects/my-project/locations/global/apis/a/versions/v",
		"projects/my-project/locations/global/apis/a/versions/v/specs",
		"projects/my-project/locations/global/apis/a/versions/v/specs/-",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/apis/a/deployments",
		"projects/my-project/locations/global/apis/a/deployments/-",
		"projects/my-project/locations/global/apis/a/deployments/d",
		"projects/my-project/locations/global/artifacts",
		"projects/my-project/locations/global/artifacts/-",
		"projects/my-project/locations/global/artifacts/x",
		"projects/my-project/locations/global/apis/a/artifacts",
		"projects/my-project/locations/global/apis/a/artifacts/-",
		"projects/my-project/locations/global/apis/a/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/-",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/-",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x",
		"projects/my-project/locations/global/apis/-/versions/-/specs/-/artifacts/-",
	}
	for _, r := range resources {
		t.Run(r, func(t *testing.T) {
			root := t.TempDir() // Use a new output directory for each export.
			cmd := Command()
			args := []string{r, "--root", root}
			cmd.SetArgs(args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}
		})
	}

	// Subsequent exports should all fail, so they share a common output directory.
	root := t.TempDir()

	// Verify that unsupported exports fail.
	unsupported := []string{
		"projects/my-project/locations/global/apis/a/deployments/d@",
		"projects/my-project/locations/global/apis/a/deployments/d@-",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s@",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s@-",
	}
	for _, r := range unsupported {
		t.Run("unsupported/"+r, func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			args := []string{r, "--root", root}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Fatalf("Execute() with args %v succeeded but should have failed", args)
			}
		})
	}

	// Verify that invalid exports fail.
	invalid := []string{
		"projects/my-missing-project",
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
		t.Run("invalid/"+r, func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			args := []string{r, "--root", root}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Fatalf("Execute() with args %v succeeded but should have failed", args)
			}
		})
	}
}
