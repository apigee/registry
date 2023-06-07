// Copyright 2022 Google LLC.
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

package upload

import (
	"context"
	"log"
	"net/http"
	"testing"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"google.golang.org/api/iterator"
)

func startTestServer() *http.Server {
	http.Handle("/", http.FileServer(http.Dir("testdata/discovery")))
	server := &http.Server{Addr: ":8081", Handler: nil}
	go func() {
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("Test server failed to start %v", err)
		}
	}()
	return server
}

func TestDiscoveryUpload(t *testing.T) {
	projectID := "disco-test"
	projectName := "projects/" + projectID
	args := []string{
		"discovery",
		"--service",
		"http://localhost:8081/apis.json",
		"--parent",
		"projects/disco-test/locations/global",
	}
	ctx := context.Background()
	// Start a test server to mock the Discovery Service.
	testServer := startTestServer()
	defer func() { _ = testServer.Shutdown(ctx) }()
	registryClient, adminClient := grpctest.SetupRegistry(ctx, t, projectID, nil)

	// Run the upload command.
	cmd := Command()
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Error running upload %v", err)
	}
	targets := []struct {
		desc     string
		spec     string
		wantType string
	}{
		{
			desc:     "Apigee Registry",
			spec:     "apis/apigeeregistry/versions/v1/specs/discovery",
			wantType: "application/x.discovery",
		},
		{
			desc:     "Petstore OpenAPI",
			spec:     "apis/discovery/versions/v1/specs/discovery",
			wantType: "application/x.discovery",
		},
	}
	for _, target := range targets {
		// Get the uploaded spec
		result, err := registryClient.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
			Name: "projects/" + projectID + "/locations/global/" + target.spec,
		})
		if err != nil {
			t.Fatalf("unable to fetch spec %s", target.spec)
		}
		// Verify the content type.
		if result.ContentType != target.wantType {
			t.Errorf("Invalid mime type for %s: %s (wanted %s)", target.spec, result.ContentType, target.wantType)
		}
	}
	// Run the upload a second time to ensure there are no errors or duplicated specs.
	cmd = Command()
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		t.Errorf("Error running second upload %v", err)
	}
	{
		iter := registryClient.ListApiSpecRevisions(ctx, &rpc.ListApiSpecRevisionsRequest{
			Name: "projects/" + projectID + "/locations/global/apis/-/versions/-/specs/-@-",
		})
		count := countSpecRevisions(iter)
		if count != 2 {
			t.Errorf("expected 2 versions, got %d", count)
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
	testServer.Close()
}

func countSpecRevisions(iter *gapic.ApiSpecIterator) int {
	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			break
		}
		count++
	}
	return count
}

func TestDiscoveryMissingParent(t *testing.T) {
	const (
		projectID   = "missing"
		projectName = "projects/" + projectID
		parent      = projectName + "/locations/global"
	)
	tests := []struct {
		desc string
		args []string
	}{
		{
			desc: "parent",
			args: []string{"discovery", "--parent", parent},
		},
		{
			desc: "project-id",
			args: []string{"discovery", "--project-id", projectID},
		},
		{
			desc: "unspecified",
			args: []string{"discovery"},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			cmd.SetArgs(test.args)
			if cmd.Execute() == nil {
				t.Error("expected error, none reported")
			}
		})
	}
}
