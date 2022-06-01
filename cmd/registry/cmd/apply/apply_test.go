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

package apply

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const sampleDir = "testdata/sample"

func TestApply(t *testing.T) {
	project := names.Project{ProjectID: "apply-test"}
	parent := project.String() + "/locations/global"

	ctx := context.Background()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Setup: failed to create client: %+v", err)
	}
	defer adminClient.Close()

	if err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil && status.Code(err) != codes.NotFound {
		t.Errorf("Setup: failed to delete test project: %s", err)
	}

	if _, err := adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: project.ProjectID,
		Project:   &rpc.Project{},
	}); err != nil {
		t.Fatalf("Setup: Failed to create test project: %s", err)
	}

	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create registry client: %s", err)
	}
	defer registryClient.Close()

	// Test API creation and export.
	// TODO: This should be split into two parts: 1) testing API creation, and 2) testing API export.
	// When API creation breaks we want to see something like FAIL: TestApply/Create_API or
	// FAIL: TestApplyAPIs/Create, not FAIL: TestApply/Create_and_Export_API, or worse FAIL: TestApply.
	{
		const filename = sampleDir + "/apis/registry.yaml"
		cmd := Command()
		cmd.SetArgs([]string{"-f", filename, "--parent", parent})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
		}
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read API YAML: %s", err)
		}

		got, err := registryClient.GetApi(ctx, &rpc.GetApiRequest{
			Name: project.Api("registry").String(),
		})
		if status.Code(err) == codes.NotFound {
			t.Fatalf("Expected API doesn't exist: %s", err)
		} else if err != nil {
			t.Fatalf("Failed to verify API existence: %s", err)
		}

		actual, _, err := patch.ExportAPI(ctx, registryClient, got)
		if err != nil {
			t.Fatalf("ExportApi(%+v) returned an error: %s", got, err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("GetApi(%q) returned unexpected diff: (-want +got):\n%s", got, diff)
		}
	}

	// Test artifact creation and export.
	// TODO: These should run as separate subtests to make it clear exactly which artifact types are failing.
	// Creation and export should also be separated ideally. The error message should at least make it
	// clear whether create or export is failing.
	artifacts := []string{"lifecycle", "manifest", "taxonomies"}
	for _, a := range artifacts {
		filename := fmt.Sprintf("%s/artifacts/%s.yaml", sampleDir, a)
		cmd := Command()
		cmd.SetArgs([]string{"-f", filename, "--parent", parent})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
		}
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read artifact YAML %s", err)
		}

		message, err := registryClient.GetArtifact(ctx, &rpc.GetArtifactRequest{
			Name: project.Artifact(a).String(),
		})
		if status.Code(err) == codes.NotFound {
			t.Fatalf("Expected artifact doesn't exist: %s", err)
		} else if err != nil {
			t.Fatalf("Failed to verify artifact existence: %s", err)
		}

		actual, _, err := patch.ExportArtifact(ctx, registryClient, message)
		if err != nil {
			t.Fatalf("ExportArtifact(%+v) returned an error: %s", message, err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("GetArtifact(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
		}
	}

	if err := adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil {
		t.Logf("Cleanup: Failed to delete test project: %s", err)
	}
}

func TestApplyProject(t *testing.T) {
	ctx := context.Background()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Setup: failed to create client: %+v", err)
	}
	defer adminClient.Close()

	project := names.Project{ProjectID: "apply-project-test"}
	if err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil && status.Code(err) != codes.NotFound {
		t.Errorf("Setup: failed to delete test project: %s", err)
	}

	if _, err := adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: project.ProjectID,
		Project:   &rpc.Project{},
	}); err != nil {
		t.Fatalf("Setup: Failed to create test project: %s", err)
	}

	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create registry client: %s", err)
	}
	defer registryClient.Close()

	cmd := Command()
	cmd.SetArgs([]string{"-f", sampleDir, "-R", "--parent", project.String() + "/locations/global"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
	}

	filename := sampleDir + "/apis/registry.yaml"
	expected, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read API YAML: %s", err)
	}

	got, err := registryClient.GetApi(ctx, &rpc.GetApiRequest{
		Name: project.Api("registry").String(),
	})
	if err != nil {
		t.Fatalf("failed to get api: %s", err)
	}

	actual, _, err := patch.ExportAPI(ctx, registryClient, got)
	if err != nil {
		t.Fatalf("ExportApi(%+v) returned an error: %s", got, err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("GetApi(%q) returned unexpected diff: (-want +got):\n%s", got, diff)
	}

	artifacts := []string{"lifecycle", "manifest", "taxonomies"}
	for _, a := range artifacts {
		filename := fmt.Sprintf("%s/artifacts/%s.yaml", sampleDir, a)
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read artifact YAML %s", err)
		}

		message, err := registryClient.GetArtifact(ctx, &rpc.GetArtifactRequest{
			Name: project.Artifact(a).String(),
		})
		if status.Code(err) == codes.NotFound {
			t.Fatalf("Expected artifact doesn't exist: %s", err)
		} else if err != nil {
			t.Fatalf("Failed to verify artifact existence: %s", err)
		}

		actual, _, err := patch.ExportArtifact(ctx, registryClient, message)
		if err != nil {
			t.Fatalf("ExportArtifact(%+v) returned an error: %s", message, err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("GetArtifact(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
		}
	}

	if err := adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil {
		t.Logf("Cleanup: Failed to delete test project: %s", err)
	}
}
