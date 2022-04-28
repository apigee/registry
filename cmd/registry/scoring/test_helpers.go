// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scoring

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const gzipOpenAPIv3 = "application/x.openapi+gzip;version=3.0.0"

func deleteProject(
	ctx context.Context,
	client connection.AdminClient,
	t *testing.T,
	projectID string) {
	t.Helper()
	req := &rpc.DeleteProjectRequest{
		Name:  "projects/" + projectID,
		Force: true,
	}
	err := client.DeleteProject(ctx, req)
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Failed DeleteProject(%v): %s", req, err.Error())
	}
}

func createProject(
	ctx context.Context,
	client connection.AdminClient,
	t *testing.T,
	projectID string) {
	t.Helper()
	req := &rpc.CreateProjectRequest{
		ProjectId: projectID,
		Project: &rpc.Project{
			DisplayName: "Demo",
			Description: "A demo catalog",
		},
	}
	project, err := client.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("Failed CreateProject(%v): %s", req, err.Error())
	}
	if project.GetName() != fmt.Sprintf("projects/%s", projectID) {
		t.Fatalf("Invalid project name %s", project.GetName())
	}
}

func createApi(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	parent, apiID string) {
	t.Helper()
	req := &rpc.CreateApiRequest{
		Parent: parent,
		ApiId:  apiID,
		Api: &rpc.Api{
			DisplayName:  apiID,
			Description:  fmt.Sprintf("Sample Test API: %s", apiID),
			Availability: "GENERAL",
		},
	}
	_, err := client.CreateApi(ctx, req)
	if err != nil {
		t.Fatalf("Failed CreateApi(%v): %s", req, err.Error())
	}
}

func createVersion(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	parent, versionID string) {
	t.Helper()
	req := &rpc.CreateApiVersionRequest{
		Parent:       parent,
		ApiVersionId: versionID,
		ApiVersion:   &rpc.ApiVersion{},
	}
	_, err := client.CreateApiVersion(ctx, req)
	if err != nil {
		t.Fatalf("Failed CreateApiVersion(%v): %s", req, err.Error())
	}
}

func createSpec(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	parent string,
	specId string,
	mimeType string,
) {
	t.Helper()
	// Create a spec entry with empty data
	req := &rpc.CreateApiSpecRequest{
		Parent:    parent,
		ApiSpecId: specId,
		ApiSpec: &rpc.ApiSpec{
			MimeType: mimeType,
		},
	}
	_, err := client.CreateApiSpec(ctx, req)
	if err != nil {
		t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
	}
}

func createUpdateArtifact(
	ctx context.Context,
	client connection.Client,
	t *testing.T,
	artifactName string,
	data []byte,
	mimeType string) {
	t.Helper()
	// Creates an artifact entry with empty data
	artifact := &rpc.Artifact{
		Name:     artifactName,
		Contents: data,
		MimeType: mimeType,
	}
	err := core.SetArtifact(ctx, client, artifact)
	if err != nil {
		t.Fatalf("Failed SetArtifact(%v): %s", artifact, err.Error())
	}
}
