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

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/names"
	"github.com/apigee/registry/server/registry/test/remote"
	"github.com/apigee/registry/server/registry/test/seeder"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestWipeout(t *testing.T) {
	projectID := "wipeout-test"
	project := names.Project{ProjectID: projectID}
	parent := project.String() + "/locations/global"
	parentName, nil := names.ParseProjectWithLocation(parent)

	ctx := context.Background()
	server := remote.NewProxy()
	if err := server.Open(ctx); err != nil {
		t.Fatalf("Setup: failed to connect to remote server: %s", err)
	}
	defer server.Close()
	var err error
	if _, err = server.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Setup: failed to delete test project: %s", err)
	}

	seed := []*rpc.Artifact{}
	for a := 1; a < 2; a++ {
		api := project.Api(fmt.Sprintf("a%d", a))
		for x := 1; x < 2; x++ {
			seed = append(seed, &rpc.Artifact{Name: api.Artifact(fmt.Sprintf("a%d", x)).String()})
		}
		for v := 1; v < 2; v++ {
			version := api.Version(fmt.Sprintf("v%d", v))
			for x := 1; x < 2; x++ {
				seed = append(seed, &rpc.Artifact{Name: version.Artifact(fmt.Sprintf("a%d", x)).String()})
			}
			for s := 1; s < 2; s++ {
				spec := version.Spec(fmt.Sprintf("s%d", s))
				for x := 1; x < 2; x++ {
					seed = append(seed, &rpc.Artifact{Name: spec.Artifact(fmt.Sprintf("a%d", x)).String()})
				}
			}
		}
		for d := 1; d < 2; d++ {
			deployment := api.Deployment(fmt.Sprintf("d%d", d))
			for x := 1; x < 2; x++ {
				seed = append(seed, &rpc.Artifact{Name: deployment.Artifact(fmt.Sprintf("a%d", x)).String()})
			}
		}
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
