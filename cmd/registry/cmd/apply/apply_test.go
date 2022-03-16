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
	"io/ioutil"
	"testing"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		const filename = "testdata/registry.yaml"
		cmd := Command(ctx)
		cmd.SetArgs([]string{"-f", filename, "--parent", parent})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
		}
		expected, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("failed to read %s: %s", filename, err)
		}
		api := project.Api("registry")
		// TODO: Prefer using the API client directly instead of relying on application code that isn't directly related to this test.
		err = core.GetAPI(ctx, registryClient, api, func(message *rpc.Api) {
			actual, _, err := patch.ExportAPI(ctx, registryClient, message)
			if err != nil {
				t.Fatalf("ExportApi(%+v) returned an error: %s", message, err)
			}
			if diff := cmp.Diff(actual, expected); diff != "" {
				t.Errorf("GetApi(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
			}
		})
		if err != nil {
			t.Fatalf("failed to get api: %s", err)
		}

	}

	// Test artifact creation and export.
	// TODO: These should run as separate subtests to make it clear exactly which artifact types are failing.
	// Creation and export should also be separated ideally. The error message should at least make it
	// clear whether create or export is failing.
	artifacts := []string{"lifecycle", "manifest", "taxonomies"}
	for _, a := range artifacts {
		filename := fmt.Sprintf("testdata/%s.yaml", a)
		cmd := Command(ctx)
		cmd.SetArgs([]string{"-f", filename, "--parent", parent})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
		}
		expected, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("failed to read %s", filename)
		}
		artifact := project.Artifact(a)
		// TODO: This subtly determines that we're testing ExportArtifact with messages that have
		// preloaded contents. Loading the contents directly using the client will make it clear
		// that's part of the test, and will provide a good place to write a comment about that choice.
		err = core.GetArtifact(ctx, registryClient, artifact, true, func(message *rpc.Artifact) {
			actual, _, err := patch.ExportArtifact(ctx, registryClient, message)
			if err != nil {
				t.Fatalf("ExportArtifact(%+v) returned an error: %s", message, err)
			}
			if diff := cmp.Diff(actual, expected); diff != "" {
				t.Errorf("GetArtifact(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
			}
		})
		if err != nil {
			t.Fatalf("failed to get artifact: %s", err)
		}
	}

	if err := adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil {
		t.Logf("Cleanup: Failed to delete test project: %s", err)
	}
}
