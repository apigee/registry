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

	"github.com/apigee/registry/cmd/registry/compress"
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
	// get names for each resource
	for _, r := range resources {
		t.Run(r, func(t *testing.T) {
			cmd := Command()
			args := []string{r, "-o", "name"}
			cmd.SetArgs(args)
			out := bytes.NewBuffer(make([]byte, 0))
			cmd.SetOut(out)
			if err := cmd.Execute(); err != nil {
				t.Errorf("Execute() with args %v returned error: %s", args, err)
			}
			if len(out.Bytes()) == 0 {
				t.Errorf("Execute() with args %v failed to return expected value(s)", args)
			}
		})
	}
	// get yaml for each resource
	for _, r := range resources {
		t.Run(r, func(t *testing.T) {
			cmd := Command()
			args := []string{r, "-o", "yaml"}
			cmd.SetArgs(args)
			out := bytes.NewBuffer(make([]byte, 0))
			cmd.SetOut(out)
			if err := cmd.Execute(); err != nil {
				t.Errorf("Execute() with args %v returned error: %s", args, err)
			}
			if len(out.Bytes()) == 0 {
				t.Errorf("Execute() with args %v failed to return expected value(s)", args)
			}
		})
	}
	resourcesWithContents := []string{
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/artifacts/x",
		"projects/my-project/locations/global/apis/a/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/x",
	}
	// Get the contents of these resources.
	for _, r := range resourcesWithContents {
		t.Run(r, func(t *testing.T) {
			cmd := Command()
			args := []string{r, "-o", "contents"}
			cmd.SetArgs(args)
			out := bytes.NewBuffer(make([]byte, 0))
			cmd.SetOut(out)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}
			if len(out.Bytes()) == 0 {
				t.Errorf("Execute() with args %v failed to return expected value(s)", args)
			}
		})
	}
	resourcesWithChildren := []string{
		"projects/my-project/locations/global/apis/a",
		"projects/my-project/locations/global/apis/a/versions/v",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/apis/a/deployments/d",
	}
	// Get the contents of these resources.
	for _, r := range resourcesWithChildren {
		t.Run(r, func(t *testing.T) {
			cmd := Command()
			args := []string{r, "-o", "yaml", "--nested"}
			cmd.SetArgs(args)
			out := bytes.NewBuffer(make([]byte, 0))
			cmd.SetOut(out)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}
			if len(out.Bytes()) == 0 {
				t.Errorf("Execute() with args %v failed to return expected value(s)", args)
			}
		})
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
		"projects/my-project/locations/global/apis/-/artifacts/xx",
		"projects/my-project/locations/global/apis/-/versions/-/artifacts/xx",
		"projects/my-project/locations/global/apis/-/versions/-/specs/-/artifacts/xx",
		"projects/my-project/locations/global/apis/-/deployments/-/artifacts/xx",
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
	for i, r := range invalid {
		t.Run(r, func(t *testing.T) {
			// cycle through output types
			format := []string{"name", "yaml", "contents"}[i%3]
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			args := []string{r, "-o", format}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Errorf("Execute() with args %v succeeded but should have failed", args)
			}
		})
	}
	// attempts to get contents of resources that don't support it should fail
	resourcesWithoutContents := []string{
		"projects/my-project",
		"projects/my-project/locations/global/apis/a",
		"projects/my-project/locations/global/apis/a/versions/v",
		"projects/my-project/locations/global/apis/a/deployments/d",
	}
	for _, r := range resourcesWithoutContents {
		t.Run(r+"--output-contents", func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			args := []string{r, "-o", "contents"}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Errorf("Execute() with args %v succeeded but should have failed", args)
			}
		})
	}
	// attempts to get an unsupported output type should fail
	resources := []string{
		"projects/my-project",
		"projects/my-project/locations/global/apis/a",
		"projects/my-project/locations/global/apis/a/versions/v",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/apis/a/deployments/d",
		"projects/my-project",
		"projects/my-project/locations/global/apis/a/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/x",
	}
	for _, r := range resources {
		t.Run(r+"--output-invalid", func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			args := []string{r, "-o", "invalid"}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Errorf("Execute() with args %v succeeded but should have failed", args)
			}
		})
	}
	// attempts to use `--nested` with unsupported output types should fail
	outputs := []string{"name", "contents"}
	for _, o := range outputs {
		t.Run(o+"--nested", func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			args := []string{resources[0], "-o", o, "--nested"}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Errorf("Execute() with args %v succeeded but should have failed", args)
			}
		})
	}
}

func TestGetValidResourcesWithFilter(t *testing.T) {
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

	// Verify that a filter specified on a get of a collection is ok.
	valid_collections := []string{
		"projects/my-project/locations/global/apis",
		"projects/my-project/locations/global/apis/a/versions",
		"projects/my-project/locations/global/apis/a/versions/v/specs",
		"projects/my-project/locations/global/apis/a/deployments",
		"projects/my-project/locations/global/apis/a/artifacts",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts",
	}
	for _, c := range valid_collections {
		t.Run(c, func(t *testing.T) {
			cmd := Command()
			out := bytes.NewBuffer(make([]byte, 0))
			cmd.SetOut(out)
			args := []string{c, "--filter", "name.contains('a')"}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				t.Errorf("Execute() with args %v failed but should have succeeded", args)
			}
			if len(out.Bytes()) == 0 {
				t.Errorf("Execute() with args %v failed to return expected value(s)", args)
			}
		})
	}

	// Verify that a filter specified on a get of an individual resource is an error.
	valid_resources := []string{
		"projects/my-project/locations/global/apis/a",
		"projects/my-project/locations/global/apis/a/versions/v",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s",
		"projects/my-project/locations/global/apis/a/deployments/d",
		"projects/my-project/locations/global/apis/a/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/artifacts/x",
		"projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x",
		"projects/my-project/locations/global/apis/a/deployments/d/artifacts/x",
	}
	for _, r := range valid_resources {
		t.Run(r, func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			args := []string{r, "--filter", "name.contains('a')"}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Errorf("Execute() with args %v succeeded but should have failed", args)
			}
		})
	}
}

func TestGetGZippedSpec(t *testing.T) {
	payload := "hello"
	contents, err := compress.GZippedBytes([]byte(payload))
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	seed := []*rpc.ApiSpec{
		{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s", MimeType: "text/plain+gzip", Contents: contents},
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
	if err := seeder.SeedSpecs(ctx, client, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	cmd := Command()
	out := bytes.NewBuffer(make([]byte, 0))
	cmd.SetOut(out)
	args := []string{"projects/my-project/locations/global/apis/a/versions/v/specs/s", "-o", "contents"}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Errorf("Execute() with args %v failed but should have succeeded", args)
	}
	if len(out.Bytes()) == 0 {
		t.Errorf("Execute() with args %v failed to return expected value(s)", args)
	}
	if out.String() != payload {
		t.Errorf("Execute() with args %v returned spec %q, expected %q", out.String(), args, payload)
	}
}

func TestGetMultipleContentRequestsShouldFail(t *testing.T) {
	seed := []*rpc.Artifact{
		{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
		{Name: "projects/my-project/locations/global/apis/b/versions/v/specs/s/artifacts/x", MimeType: "application/yaml", Contents: []byte("hello: 123")},
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
	if err := seeder.SeedArtifacts(ctx, client, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	// Verify that a filter specified on a get of an individual resource is an error.
	multiple_resources := []string{
		"projects/my-project/locations/global/apis/-/versions/v/specs/s",
		"projects/my-project/locations/global/apis/-/versions/v/specs/s/artifacts/x",
	}
	for _, r := range multiple_resources {
		t.Run(r, func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			args := []string{r, "--output", "contents"}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Errorf("Execute() with args %v succeeded but should have failed", args)
			}
		})
	}
}
