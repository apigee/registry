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
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/pkg/remote"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/apigee/registry/server/registry/test/seeder"
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
	server := remote.NewProxy()
	server.Open(ctx)
	defer server.Close()
	var err error
	if _, err = server.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Setup: failed to delete test project: %s", err)
	}

	seed := []*rpc.Artifact{
		{Name: "projects/wipeout-test/locations/global/apis/a1/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v1/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v1/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v1/specs/s1/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v1/specs/s1/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v1/specs/s2/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v1/specs/s2/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v2/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v2/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v2/specs/s1/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v2/specs/s1/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v2/specs/s2/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/versions/v2/specs/s2/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/deployments/d1/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/deployments/d1/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/deployments/d2/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a1/deployments/d2/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v1/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v1/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v1/specs/s1/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v1/specs/s1/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v1/specs/s2/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v1/specs/s2/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v2/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v2/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v2/specs/s1/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v2/specs/s1/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v2/specs/s2/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/versions/v2/specs/s2/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/deployments/d1/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/deployments/d1/artifacts/a2"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/deployments/d2/artifacts/a1"},
		{Name: "projects/wipeout-test/locations/global/apis/a2/deployments/d2/artifacts/a2"},
	}
	if err := seeder.SeedArtifacts(ctx, server, seed...); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create registry client: %s", err)
	}
	defer registryClient.Close()

	t.Run("WipeoutProject", func(t *testing.T) {
		Wipeout(ctx, registryClient, projectID, 10)
		if _, ok := registryClient.ListApis(ctx, &rpc.ListApisRequest{Parent: parent}).Next(); ok != iterator.Done {
			t.Errorf("Error: APIs found after wipeout")
		}
		if _, ok := registryClient.ListArtifacts(ctx, &rpc.ListArtifactsRequest{Parent: parentName.Api("-").String()}).Next(); ok != iterator.Done {
			t.Errorf("Error: API artifacts found after wipeout")
		}
		if _, ok := registryClient.ListApiVersions(ctx, &rpc.ListApiVersionsRequest{Parent: parentName.Api("-").String()}).Next(); ok != iterator.Done {
			t.Errorf("Error: Versions found after wipeout")
		}
		if _, ok := registryClient.ListArtifacts(ctx, &rpc.ListArtifactsRequest{Parent: parentName.Api("-").Version("-").String()}).Next(); ok != iterator.Done {
			t.Errorf("Error: Version artifacts found after wipeout")
		}
		if _, ok := registryClient.ListApiSpecs(ctx, &rpc.ListApiSpecsRequest{Parent: parentName.Api("-").Version("-").String()}).Next(); ok != iterator.Done {
			t.Errorf("Error: Specs found after wipeout")
		}
		if _, ok := registryClient.ListArtifacts(ctx, &rpc.ListArtifactsRequest{Parent: parentName.Api("-").Version("-").Spec("-").String()}).Next(); ok != iterator.Done {
			t.Errorf("Error: Spec artifacts found after wipeout")
		}
		if _, ok := registryClient.ListApiDeployments(ctx, &rpc.ListApiDeploymentsRequest{Parent: parentName.Api("-").String()}).Next(); ok != iterator.Done {
			t.Errorf("Error: Deployments found after wipeout")
		}
		if _, ok := registryClient.ListArtifacts(ctx, &rpc.ListArtifactsRequest{Parent: parentName.Api("-").Deployment("-").String()}).Next(); ok != iterator.Done {
			t.Errorf("Error: Deployment artifacts found after wipeout")
		}
	})
}
