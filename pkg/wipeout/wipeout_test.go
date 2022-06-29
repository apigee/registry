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

package wipeout

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWipeout(t *testing.T) {
	projectID := "wipeout-test"
	project := names.Project{ProjectID: projectID}
	parent := project.String() + "/locations/global"
	parentName, nil := names.ParseProjectWithLocation(parent)

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
	for i := 0; i <= 2; i++ {
		api, err := registryClient.CreateApi(ctx, &rpc.CreateApiRequest{
			ApiId:  fmt.Sprintf("a%d", i),
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
		for k := 0; k < 2; k++ {
			_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
				ArtifactId: fmt.Sprintf("a%d", k),
				Parent:     apiName.String(),
				Artifact:   &rpc.Artifact{},
			})
			if err != nil {
				t.Fatalf("Setup: Failed to create test artifact: %s", err)
			}
		}
		for j := 0; j < 2; j++ {
			deployment, err := registryClient.CreateApiDeployment(ctx, &rpc.CreateApiDeploymentRequest{
				ApiDeploymentId: fmt.Sprintf("d%d", j),
				Parent:          apiName.String(),
				ApiDeployment:   &rpc.ApiDeployment{},
			})
			if err != nil {
				t.Fatalf("Setup: Failed to create test deployment: %s", err)
			}
			deploymentName, err := names.ParseDeployment(deployment.Name)
			if err != nil {
				t.Fatalf("Setup: Failed to create test deployment: %s", err)
			}
			for k := 0; k < 2; k++ {
				_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
					ArtifactId: fmt.Sprintf("a%d", k),
					Parent:     deploymentName.String(),
					Artifact:   &rpc.Artifact{},
				})
				if err != nil {
					t.Fatalf("Setup: Failed to create test artifact: %s", err)
				}
			}
			version, err := registryClient.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
				ApiVersionId: fmt.Sprintf("v%d", j),
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
			for k := 0; k < 2; k++ {
				_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
					ArtifactId: fmt.Sprintf("a%d", k),
					Parent:     versionName.String(),
					Artifact:   &rpc.Artifact{},
				})
				if err != nil {
					t.Fatalf("Setup: Failed to create test artifact: %s", err)
				}
			}
			for k := 0; k < 2; k++ {
				spec, err := registryClient.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
					ApiSpecId: fmt.Sprintf("s%d", k),
					Parent:    versionName.String(),
					ApiSpec:   &rpc.ApiSpec{},
				})
				if err != nil {
					t.Fatalf("Setup: Failed to create test spec: %s", err)
				}
				specName, err := names.ParseSpec(spec.Name)
				if err != nil {
					t.Fatalf("Setup: Failed to create test spec: %s", err)
				}
				for l := 0; l < 2; l++ {
					_, err := registryClient.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
						ArtifactId: fmt.Sprintf("a%d", l),
						Parent:     specName.String(),
						Artifact:   &rpc.Artifact{},
					})
					if err != nil {
						t.Fatalf("Setup: Failed to create test artifact: %s", err)
					}
				}
			}
		}
	}
	t.Run("WipeoutProject", func(t *testing.T) {
		err = Wipeout(ctx, registryClient, projectID, 10)
		if err != nil {
			t.Fatalf("Setup: Failed to wipeout project: %s", err)
		}
		it := registryClient.ListApis(ctx, &rpc.ListApisRequest{Parent: parent})
		if _, ok := it.Next(); ok != iterator.Done {
			t.Errorf("Error: APIs found after wipeout")
		}
		it2 := registryClient.ListApiVersions(ctx, &rpc.ListApiVersionsRequest{Parent: parentName.Api("-").String()})
		if _, ok := it2.Next(); ok != iterator.Done {
			t.Errorf("Error: Versions found after wipeout")
		}
		it3 := registryClient.ListArtifacts(ctx, &rpc.ListArtifactsRequest{Parent: parentName.Api("-").Version("-").String()})
		if _, ok := it3.Next(); ok != iterator.Done {
			t.Errorf("Error: Artifacts found after wipeout")
		}
	})

}
