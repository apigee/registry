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

package visitor

import (
	"bytes"
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
)

func TestFetch(t *testing.T) {
	projectID := "fetch-test"
	project := names.Project{ProjectID: projectID}
	parent := project.String() + "/locations/global"

	specContents := []byte("hello")
	gzippedSpecContents, err := compress.GZippedBytes(specContents)
	if err != nil {
		t.Fatalf("Setup: Failed to compress test data: %s", err)
	}
	artifactContents := "hello"

	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, projectID, nil)

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
			Contents: gzippedSpecContents,
			MimeType: "text/plain+gzip",
		},
	})
	if err != nil {
		t.Fatalf("Setup: Failed to create test spec: %s", err)
	}
	specRevisionID := spec.RevisionId
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
	t.Run("fetch-spec-contents", func(t *testing.T) {
		spec := &rpc.ApiSpec{Name: "projects/fetch-test/locations/global/apis/a/versions/v/specs/s"}
		err := FetchSpecContents(ctx, registryClient, spec)
		if err != nil {
			t.Fatalf("Failed to fetch spec contents: %s", err)
		}
		if !bytes.Equal(spec.Contents, specContents) {
			t.Fatalf("Fetched unexpected spec contents: wanted %q got %q", specContents, spec.Contents)
		}
	})
	t.Run("fetch-artifact-contents", func(t *testing.T) {
		artifact := &rpc.Artifact{Name: "projects/fetch-test/locations/global/apis/a/versions/v/specs/s/artifacts/x"}
		err := FetchArtifactContents(ctx, registryClient, artifact)
		if err != nil {
			t.Fatalf("Failed to fetch artifact contents: %s", err)
		}
		if string(artifact.Contents) != artifactContents {
			t.Fatalf("Fetched unexpected spec contents: wanted %q got %q", artifactContents, artifact.Contents)
		}
	})
	t.Run("get-spec-with-contents", func(t *testing.T) {
		specName, err := names.ParseSpec("projects/fetch-test/locations/global/apis/a/versions/v/specs/s")
		if err != nil {
			t.Fatalf("Failed to parse spec name: %s", err)
		}
		err = GetSpec(ctx, registryClient, specName, true, func(ctx context.Context, spec *rpc.ApiSpec) error {
			if !bytes.Equal(spec.Contents, specContents) {
				t.Fatalf("Fetched unexpected spec contents: wanted %q got %q", specContents, spec.Contents)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to get spec with contents: %s", err)
		}
	})
	t.Run("list-specs-with-contents", func(t *testing.T) {
		specName, err := names.ParseSpec("projects/fetch-test/locations/global/apis/a/versions/v/specs/-")
		if err != nil {
			t.Fatalf("Failed to parse spec name: %s", err)
		}
		count := 0
		err = ListSpecs(ctx, registryClient, specName, 0, "", true, func(ctx context.Context, spec *rpc.ApiSpec) error {
			count++
			if !bytes.Equal(spec.Contents, specContents) {
				t.Fatalf("Fetched unexpected spec contents: wanted %q got %q", specContents, spec.Contents)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to list specs with contents: %s", err)
		}
		if count != 1 {
			t.Fatalf("Failed to list specs: got %d expected 1", count)
		}
	})
	t.Run("get-spec-revision-with-contents", func(t *testing.T) {
		specRevisionName, err := names.ParseSpecRevision("projects/fetch-test/locations/global/apis/a/versions/v/specs/s@" + specRevisionID)
		if err != nil {
			t.Fatalf("Failed to parse spec revision name: %s", err)
		}
		err = GetSpecRevision(ctx, registryClient, specRevisionName, true, func(ctx context.Context, spec *rpc.ApiSpec) error {
			if !bytes.Equal(spec.Contents, specContents) {
				t.Fatalf("Fetched unexpected spec contents: wanted %q got %q", specContents, spec.Contents)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to get spec revision with contents: %s", err)
		}
	})
	t.Run("list-spec-revisions-with-contents", func(t *testing.T) {
		specRevisionName, err := names.ParseSpecRevision("projects/fetch-test/locations/global/apis/a/versions/v/specs/s@-")
		if err != nil {
			t.Fatalf("Failed to parse spec revision name: %s", err)
		}
		count := 0
		err = ListSpecRevisions(ctx, registryClient, specRevisionName, 0, "", true, func(ctx context.Context, spec *rpc.ApiSpec) error {
			count++
			if !bytes.Equal(spec.Contents, specContents) {
				t.Fatalf("Fetched unexpected spec contents: wanted %q got %q", specContents, spec.Contents)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to list spec revisions with contents: %s", err)
		}
		if count != 1 {
			t.Fatalf("Failed to list spec revisions: got %d expected 1", count)
		}
	})
	t.Run("get-artifact-with-contents", func(t *testing.T) {
		artifactName, err := names.ParseArtifact("projects/fetch-test/locations/global/apis/a/versions/v/specs/s/artifacts/x")
		if err != nil {
			t.Fatalf("Failed to parse artifact name: %s", err)
		}
		err = GetArtifact(ctx, registryClient, artifactName, true, func(ctx context.Context, artifact *rpc.Artifact) error {
			if string(artifact.Contents) != artifactContents {
				t.Fatalf("Fetched unexpected artifact contents: wanted %q got %q", artifactContents, artifact.Contents)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to get artifact with contents: %s", err)
		}
	})
	t.Run("list-artifacts-with-contents", func(t *testing.T) {
		artifactName, err := names.ParseArtifact("projects/fetch-test/locations/global/apis/a/versions/v/specs/s/artifacts/-")
		if err != nil {
			t.Fatalf("Failed to parse artifact name: %s", err)
		}
		count := 0
		err = ListArtifacts(ctx, registryClient, artifactName, 0, "", true, func(ctx context.Context, artifact *rpc.Artifact) error {
			count++
			if string(artifact.Contents) != artifactContents {
				t.Fatalf("Fetched unexpected artifact contents: wanted %q got %q", artifactContents, artifact.Contents)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to list artifacts with contents: %s", err)
		}
		if count != 1 {
			t.Fatalf("Failed to list artifacts: got %d expected 1", count)
		}
	})
}
