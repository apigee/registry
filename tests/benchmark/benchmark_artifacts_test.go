// Copyright 2023 Google LLC.
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

package benchmark

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/cmd/registry/tasks"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"google.golang.org/api/iterator"
)

func listArtifacts(b *testing.B, ctx context.Context, client connection.RegistryClient, parent, filter string) error {
	b.Helper()
	count := 0
	it := client.ListArtifacts(ctx, &rpc.ListArtifactsRequest{Parent: parent, Filter: filter})
	for _, err := it.Next(); err != iterator.Done; _, err = it.Next() {
		if err != nil {
			return err
		}
		count++
	}
	return nil
}

type seedArtifactsTask struct {
	client      connection.RegistryClient
	b           *testing.B
	start       int
	count       int
	projectName names.Project
}

func (task *seedArtifactsTask) String() string {
	return fmt.Sprintf("seed %d-%d", task.start, task.start+task.count-1)
}

func (task *seedArtifactsTask) Run(ctx context.Context) error {
	b := task.b
	client := task.client
	for i := 0; i < task.count; i++ {
		apiName := task.projectName.Api(fmt.Sprintf("a%d", task.start+i))
		if err := createApi(b, ctx, client, apiName); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, apiName.Artifact("a")); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, apiName.Artifact("b")); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, apiName.Artifact("c")); err != nil {
			return err
		}
		versionName := apiName.Version("v")
		if err := createApiVersion(b, ctx, client, versionName); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, versionName.Artifact("a")); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, versionName.Artifact("b")); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, versionName.Artifact("c")); err != nil {
			return err
		}
		specName := versionName.Spec("s")
		if err := createApiSpecRevisions(b, ctx, client, specName, 3); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, specName.Artifact("a")); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, specName.Artifact("b")); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, specName.Artifact("c")); err != nil {
			return err
		}
		deploymentName := apiName.Deployment("d")
		if err := createApiDeploymentRevisions(b, ctx, client, deploymentName, 3); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, deploymentName.Artifact("a")); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, deploymentName.Artifact("b")); err != nil {
			return err
		}
		if err := createArtifact(b, ctx, client, deploymentName.Artifact("c")); err != nil {
			return err
		}
	}
	return nil
}

func BenchmarkListArtifacts(b *testing.B) {
	ctx, client := setup(b)
	// Try to speed up creation with a worker pool (this might not help much but seemed worth a try).
	taskQueue, wait := tasks.WorkerPoolIgnoreError(ctx, 5)
	projectName := names.Project{ProjectID: projectID}
	total := 1000
	step := 5
	for i := 0; i < total; i += step {
		taskQueue <- &seedArtifactsTask{
			client:      client,
			b:           b,
			start:       i,
			count:       step,
			projectName: projectName,
		}
	}
	wait()
	// Tests will list one of three artifacts created under each resource for all instances of the resource.
	filter := "artifact_id == 'b'"
	b.Run("ListApiArtifacts", func(b *testing.B) {
		parent := root().Api("-").String()
		for i := 1; i <= b.N; i++ {
			if err := listArtifacts(b, ctx, client, parent, filter); err != nil {
				b.Fatalf("%s", err)
			}
		}
	})
	b.Run("ListVersionArtifacts", func(b *testing.B) {
		parent := root().Api("-").Version("-").String()
		for i := 1; i <= b.N; i++ {
			if err := listArtifacts(b, ctx, client, parent, filter); err != nil {
				b.Fatalf("%s", err)
			}
		}
	})
	b.Run("ListSpecArtifacts", func(b *testing.B) {
		parent := root().Api("-").Version("-").Spec("-").String()
		for i := 1; i <= b.N; i++ {
			if err := listArtifacts(b, ctx, client, parent, filter); err != nil {
				b.Fatalf("%s", err)
			}
		}
	})
	b.Run("ListDeploymentArtifacts", func(b *testing.B) {
		parent := root().Api("-").Deployment("-").String()
		for i := 1; i <= b.N; i++ {
			if err := listArtifacts(b, ctx, client, parent, filter); err != nil {
				b.Fatalf("%s", err)
			}
		}
	})
	teardown(ctx, b, client)
}

func createApi(b *testing.B, ctx context.Context, client connection.RegistryClient, api names.Api) error {
	b.Helper()
	_, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: api.Parent(),
		ApiId:  api.ApiID,
		Api: &rpc.Api{
			DisplayName: api.ApiID,
			Description: fmt.Sprintf("Description for %s", api.ApiID),
		},
	})
	return err
}

func createApiVersion(b *testing.B, ctx context.Context, client connection.RegistryClient, version names.Version) error {
	b.Helper()
	_, err := client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
		Parent:       version.Parent(),
		ApiVersionId: version.VersionID,
		ApiVersion: &rpc.ApiVersion{
			DisplayName: version.VersionID,
			Description: fmt.Sprintf("Description for %s", version.VersionID),
		},
	})
	return err
}

func createApiSpecRevisions(b *testing.B, ctx context.Context, client connection.RegistryClient, spec names.Spec, revisions int) error {
	b.Helper()
	for r := 0; r < revisions; r++ {
		_, err := client.UpdateApiSpec(ctx, &rpc.UpdateApiSpecRequest{
			ApiSpec: &rpc.ApiSpec{
				Name:        spec.String(),
				Filename:    spec.SpecID,
				Description: fmt.Sprintf("Description for %s", spec.SpecID),
				MimeType:    "text/plain",
				Contents:    []byte(fmt.Sprintf("Revision %d", r)),
			},
			AllowMissing: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func createApiDeploymentRevisions(b *testing.B, ctx context.Context, client connection.RegistryClient, deployment names.Deployment, revisions int) error {
	b.Helper()
	for r := 0; r < revisions; r++ {
		_, err := client.UpdateApiDeployment(ctx, &rpc.UpdateApiDeploymentRequest{
			ApiDeployment: &rpc.ApiDeployment{
				Name:            deployment.String(),
				DisplayName:     deployment.DeploymentID,
				Description:     fmt.Sprintf("Description for %s", deployment.DeploymentID),
				EndpointUri:     fmt.Sprintf("https://r%d.example.com", r),
				ApiSpecRevision: fmt.Sprintf(deployment.Project().String()+"/apis/a/versions/v/specs/s@%d", r),
			},
			AllowMissing: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func createArtifact(b *testing.B, ctx context.Context, client connection.RegistryClient, artifact names.Artifact) error {
	b.Helper()
	_, err := client.CreateArtifact(ctx, &rpc.CreateArtifactRequest{
		Parent:     artifact.Parent(),
		ArtifactId: artifact.ArtifactID(),
		Artifact:   &rpc.Artifact{},
	})
	return err
}
