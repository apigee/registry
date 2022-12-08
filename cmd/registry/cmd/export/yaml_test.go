package export

import (
	"context"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
)

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

	// Verify that yaml export runs for each resource.
	resources := []string{
		"projects/my-project/locations/global/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/apis/a/deployments/d",
		"projects/my-project/locations/global/apis/a",
		"projects/my-project/locations/global/apis/a/versions/v",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/apis/a/deployments/d",
		"projects/my-project/locations/global/apis/a/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/x",
	}
	for _, r := range resources {
		cmd := Command()
		args := []string{"yaml", r}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %v returned error: %s", args, err)
		}
	}

	// Repeat with --nested export enabled.
	for _, r := range resources {
		cmd := Command()
		args := []string{"yaml", r, "--nested"}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %v returned error: %s", args, err)
		}
	}

	// Verify that invalid exports fail.
	invalid := []string{
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
		args := []string{"yaml", r}
		cmd.SetArgs(args)
		if err := cmd.Execute(); err == nil {
			t.Fatalf("Execute() with args %v succeeded but should have failed", args)
		}
	}
}
