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
	"testing"

	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
)

func TestOpenAPI(t *testing.T) {
	const (
		projectID   = "openapi-test"
		projectName = "projects/" + projectID
		parent      = projectName + "/locations/global"
	)
	// Create a registry client.
	ctx := context.Background()
	registryClient, adminClient := grpctest.SetupRegistry(ctx, t, projectID, nil)

	cmd := Command()
	args := []string{"openapi", "testdata/openapi", "--parent", parent}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %+v returned error: %s", args, err)
	}
	tests := []struct {
		desc     string
		spec     string
		wantType string
	}{
		{
			desc:     "Apigee Registry",
			spec:     "apis/apigee-registry/versions/v1/specs/openapi",
			wantType: "application/x.openapi;version=3",
		},
		{
			desc:     "Petstore OpenAPI",
			spec:     "apis/petstore/versions/3.0/specs/openapi",
			wantType: "application/x.openapi;version=3",
		},
		{
			desc:     "Petstore Swagger",
			spec:     "apis/petstore/versions/2.0/specs/openapi",
			wantType: "application/x.openapi;version=2",
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
		if result.ContentType != test.wantType {
			t.Errorf("Invalid mime type for %s: %s (wanted %s)", test.spec, result.ContentType, test.wantType)
		}
	}
	// Delete the test project.
	req := &rpc.DeleteProjectRequest{
		Name:  projectName,
		Force: true,
	}
	err := adminClient.DeleteProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed to delete test project: %s", err)
	}
}

func TestOpenAPIMissingParent(t *testing.T) {
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
			args: []string{"openapi", "nonexistent-specs-dir", "--parent", parent},
		},
		{
			desc: "project-id",
			args: []string{"openapi", "nonexistent-specs-dir", "--project-id", projectID},
		},
		{
			desc: "unspecified",
			args: []string{"openapi", "nonexistent-specs-dir"},
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
