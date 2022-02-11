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
	"github.com/apigee/registry/log"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestApply(t *testing.T) {
	const (
		projectID   = "apply-test"
		projectName = "projects/" + projectID
	)

	// Create a registry client.
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Fatalf("Error creating client: %+v", err)
	}
	defer registryClient.Close()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Error creating client: %+v", err)
	}
	defer adminClient.Close()
	// Clear the test project.
	err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  projectName,
		Force: true,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Error deleting test project: %+v", err)
	}
	// Create the test project.
	_, err = adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: projectID,
		Project: &rpc.Project{
			DisplayName: "Test",
			Description: "A test catalog",
		},
	})
	if err != nil {
		t.Fatalf("Error creating project %s", err)
	}
	cmd := Command(ctx)
	cmd.SetArgs([]string{"-f", "testdata/registry.yaml", "--parent", "projects/apply-test/locations/global"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
	}

	expected, err := ioutil.ReadFile("testdata/registry.yaml")
	if err != nil {
		t.Fatalf("failed to read testdata/registry.yaml")
	}

	client, err := connection.NewClient(ctx)

	if api, err := names.ParseApi("projects/apply-test/locations/global/apis/registry"); err == nil {
		_, err = core.GetAPI(ctx, client, api, func(message *rpc.Api) {
			actual, _, err := patch.ExportAPI(ctx, client, message)
			if err != nil {
				log.FromContext(ctx).WithError(err).Fatal("Failed to export artifact")
			} else {
				if diff := cmp.Diff(actual, expected); diff != "" {
					t.Errorf("API mismatch %+v", api)
					fmt.Printf("expected %d %s", len(expected), string(expected))
					fmt.Printf("actual %d %s", len(actual), string(actual))
				}
			}
		})
	}

	// Delete the test project.
	{
		req := &rpc.DeleteProjectRequest{
			Name:  projectName,
			Force: true,
		}
		err = adminClient.DeleteProject(ctx, req)
		if err != nil {
			t.Fatalf("Failed to delete test project: %s", err)
		}
	}
}
