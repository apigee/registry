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

package export

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestExportCSV(t *testing.T) {
	ctx := context.Background()
	client, err := connection.NewRegistryClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create client: %s", err)
	}
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create client: %s", err)
	}

	const (
		projectID = "export-csv-test-project"
		apiID     = "my-api"
		versionID = "v1"
		specID    = "my-spec"
	)

	// Setup
	err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  "projects/" + projectID,
		Force: true,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Setup: Failed to delete test project: %s", err)
	}

	project, err := adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: projectID,
		Project:   &rpc.Project{},
	})
	if err != nil {
		t.Fatalf("Setup: Failed to create project: %s", err)
	}

	api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: project.GetName() + "/locations/global",
		ApiId:  apiID,
		Api:    &rpc.Api{},
	})
	if err != nil {
		t.Fatalf("Setup: Failed to create api: %s", err)
	}

	version, err := client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
		Parent:       api.GetName(),
		ApiVersionId: versionID,
		ApiVersion:   &rpc.ApiVersion{},
	})
	if err != nil {
		t.Fatalf("Setup: Failed to create version: %s", err)
	}

	spec, err := client.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
		Parent:    version.GetName(),
		ApiSpecId: specID,
		ApiSpec:   &rpc.ApiSpec{},
	})
	if err != nil {
		t.Fatalf("Setup: Failed to create spec: %s", err)
	}

	// Execute
	cmd := Command()
	out := bytes.NewBuffer(make([]byte, 0))
	args := []string{"csv", version.GetName()}
	cmd.SetOut(out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", args, err)
	}

	// Verify
	var (
		expectedHeader = []string{"api_id", "version_id", "spec_id", "contents_path"}
		expectedRow    = []string{apiID, versionID, specID, fmt.Sprintf("$APG_REGISTRY_ADDRESS/%s", spec.GetName())}
	)

	r := csv.NewReader(out)
	header, err := r.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV header: %s", err)
	}

	sortStrings := cmpopts.SortSlices(func(a, b string) bool { return a < b })
	if diff := cmp.Diff(expectedHeader, header, sortStrings); diff != "" {
		t.Errorf("Command printed unexpected header (-want +got):\n%s", diff)
	}

	row, err := r.Read()
	if err != nil {
		t.Fatalf("Failed to read expected CSV row: %s", err)
	}

	if diff := cmp.Diff(expectedRow, row, sortStrings); diff != "" {
		t.Errorf("Command printed unexpected row (-want +got):\n%s", diff)
	}
}
