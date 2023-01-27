// Copyright 2023 Google LLC. All Rights Reserved.
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

package visitor

import (
	"context"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFetch(t *testing.T) {
	projectID := "fetch-test"
	project := names.Project{ProjectID: projectID}
	parent := project.String() + "/locations/global"

	specContents := "hello"
	artifactContents := "hello"

	ctx := context.Background()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Setup: failed to create client: %+v", err)
	}
	if err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Setup: failed to delete test project: %s", err)
	}
	if _, err := adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: project.ProjectID,
		Project:   &rpc.Project{},
	}); err != nil {
		t.Fatalf("Setup: Failed to create test project: %s", err)
	}
	t.Cleanup(func() {
		if err := adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
			Name:  project.String(),
			Force: true,
		}); err != nil && status.Code(err) != codes.NotFound {
			t.Fatalf("Setup: failed to delete test project: %s", err)
		}
		adminClient.Close()
	})
	registryClient, err := connection.NewRegistryClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create registry client: %s", err)
	}
	defer registryClient.Close()
	api, err := registryClient.CreateApi(ctx, &rpc.CreateApiRequest{
		ApiId:  "a",
		Parent: parent,
		Api:    &rpc.Api{},
	})
	if err != nil {
		t.Fatalf("Setup: Failed to create test api: %s", err)
	}
	apiName, err := names.ParseApi(api.Name)
	if err != nil {
		t.Fatalf("Setup: Failed to create test api: %s", err)
	}
	version, err := registryClient.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
		ApiVersionId: "v",
		Parent:       apiName.String(),
		ApiVersion:   &rpc.ApiVersion{},
	})
	if err != nil {
		t.Fatalf("Setup: Failed to create test version: %s", err)
	}
	versionName, err := names.ParseVersion(version.Name)
	if err != nil {
		t.Fatalf("Setup: Failed to create test version: %s", err)
	}
	spec, err := registryClient.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
		ApiSpecId: "s",
		Parent:    versionName.String(),
		ApiSpec: &rpc.ApiSpec{
			Contents: []byte(specContents),
			MimeType: "text/plain",
		},
	})
	if err != nil {
		t.Fatalf("Setup: Failed to create test spec: %s", err)
	}
	specName, err := names.ParseSpec(spec.Name)
	if err != nil {
		t.Fatalf("Setup: Failed to create test spec: %s", err)
	}
	_, err = registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
		ArtifactId: "x",
		Parent:     specName.String(),
		Artifact: &rpc.Artifact{
			Contents: []byte(artifactContents),
			MimeType: "text/plain",
		},
	})
	if err != nil {
		t.Fatalf("Setup: Failed to create test artifact: %s", err)
	}
	t.Run("spec", func(t *testing.T) {
		spec := &rpc.ApiSpec{Name: "projects/fetch-test/locations/global/apis/a/versions/v/specs/s"}
		err := FetchSpecContents(ctx, registryClient, spec)
		if err != nil {
			t.Fatalf("Failed to fetch spec contents: %s", err)
		}
		if string(spec.Contents) != specContents {
			t.Fatalf("Fetched unexpected spec contents: wanted %q got %q", specContents, spec.Contents)
		}
	})
	t.Run("artifact", func(t *testing.T) {
		artifact := &rpc.Artifact{Name: "projects/fetch-test/locations/global/apis/a/versions/v/specs/s/artifacts/x"}
		err := FetchArtifactContents(ctx, registryClient, artifact)
		if err != nil {
			t.Fatalf("Failed to fetch artifact contents: %s", err)
		}
		if string(artifact.Contents) != artifactContents {
			t.Fatalf("Fetched unexpected spec contents: wanted %q got %q", artifactContents, artifact.Contents)
		}
	})
}
