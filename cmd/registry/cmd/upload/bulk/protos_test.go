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

package bulk

import (
	"context"
	"log"
	"sort"
	"strings"
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
	tests := []struct {
		desc              string
		spec              string
		wantProtoCount    int
		wantMetadataCount int
	}{
		{
			desc:              "Apigee Registry",
			spec:              "apis/apigeeregistry/versions/v1/specs/google-cloud-apigeeregistry-v1.zip",
			wantProtoCount:    11,
			wantMetadataCount: 2,
		},
		{
			desc:              "Example Library",
			spec:              "apis/library-example/versions/v1/specs/google-example-library-v1.zip",
			wantProtoCount:    6,
			wantMetadataCount: 3,
		},
	}
	for _, test := range tests {
		// Get the uploaded spec
		result, err := registryClient.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
			Name: "projects/" + projectID + "/locations/global/" + test.spec,
		})
		if err != nil {
			t.Fatalf("unable to fetch spec %s", test.spec)
		}
		// Verify the content type.
		if result.ContentType != "application/x.protobuf+zip" {
			t.Errorf("Invalid mime type for %s: %s", test.spec, result.ContentType)
		}
		// Verify that the zip contains the expected number of files
		m, err := core.UnzipArchiveToMap(result.Data)
		if err != nil {
			t.Fatalf("unable to unzip spec %s", test.spec)
		}
		proto_filenames := make([]string, 0)
		metadata_filenames := make([]string, 0)
		for filename := range m {
			if strings.HasSuffix(filename, ".proto") {
				proto_filenames = append(proto_filenames, filename)
			} else {
				metadata_filenames = append(metadata_filenames, filename)
			}
		}
		if len(proto_filenames) != test.wantProtoCount {
			t.Errorf("Archive contains incorrect number of proto files (%d, expected %d)",
				len(proto_filenames),
				test.wantProtoCount)
			sort.Strings(proto_filenames)
			for i, s := range proto_filenames {
				log.Printf("%d: %s", i, s)
			}
		}
		if len(metadata_filenames) != test.wantMetadataCount {
			t.Errorf("Archive contains incorrect number of metadata files (%d, expected %d)",
				len(metadata_filenames),
				test.wantMetadataCount)
			sort.Strings(metadata_filenames)
			for i, s := range metadata_filenames {
				log.Printf("%d: %s", i, s)
			}
		}
	}
	// Delete the test project.
	req := &rpc.DeleteProjectRequest{
		Name:  projectName,
		Force: true,
	}
	err = adminClient.DeleteProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to delete test project: %s", err)
	}
}
