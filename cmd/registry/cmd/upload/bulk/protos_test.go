// Copyright 2021 Google LLC. All Rights Reserved.
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

package bulk

import (
	"context"
	"log"
	"sort"
	"testing"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestProtos(t *testing.T) {
	const (
		projectID   = "protos-test"
		projectName = "projects/" + projectID
	)

	// Create a registry client.
	ctx := context.Background()
	registryClient, err := connection.NewRegistryClient(ctx)
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
	cmd := Command()
	args := []string{"protos", "testdata/protos", "--project-id", projectID}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %+v returned error: %s", args, err)
	}
	// Get the uploaded spec
	result, err := registryClient.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
		Name: "projects/" + projectID + "/locations/global/apis/apigeeregistry/versions/v1/specs/google-cloud-apigeeregistry-v1.zip",
	})
	if err != nil {
		t.Fatal("unable to fetch spec")
	}
	// Verify the content type.
	if result.ContentType != "application/x.protobuf+zip" {
		t.Errorf("Invalid mime type: %s", result.ContentType)
	}
	// Verify that the zip contains the expected number of files
	expectedLength := 13
	m, err := core.UnzipArchiveToMap(result.Data)
	if err != nil {
		t.Fatal("unable to unzip spec")
	}
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	if len(keys) != expectedLength {
		t.Errorf("Archive contains incorrect number of files (%d, expected %d)", len(keys), expectedLength)
		sort.Strings(keys)
		for i, s := range keys {
			log.Printf("%d: %s", i, s)
		}
	}
	// Delete the test project.
	if false {
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
