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

package complexity

import (
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/apply"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	metrics "github.com/google/gnostic/metrics"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestComputeComplexityWithNoArgs(t *testing.T) {
	command := Command()
	command.SilenceErrors = true
	command.SilenceUsage = true
	if err := command.Execute(); err == nil {
		t.Fatalf("Execute() with no args succeeded and should have failed")
	}
}

func TestComputeComplexity(t *testing.T) {
	project := names.Project{ProjectID: "complexity-test"}
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, project.ProjectID, nil)

	config, err := connection.ActiveConfig()
	if err != nil {
		t.Fatalf("Setup: Failed to get registry configuration: %s", err)
	}
	config.Project = project.ProjectID
	connection.SetConfig(config)

	applyCmd := apply.Command()
	applyCmd.SetArgs([]string{"-f", "testdata/apigeeregistry", "-R"})
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("Failed to apply test API")
	}

	t.Run("protos", func(t *testing.T) {
		complexityCmd := Command()
		complexityCmd.SetArgs([]string{project.Api("apigeeregistry").Version("v1").Spec("protos").String()})
		if err := complexityCmd.Execute(); err != nil {
			t.Fatalf("Compute complexity failed: %s", err)
		}

		artifactName := project.Api("apigeeregistry").Version("v1").Spec("protos").Artifact("complexity")
		err = visitor.GetArtifact(ctx, registryClient, artifactName, true, func(ctx context.Context, message *rpc.Artifact) error {
			complexity := &metrics.Complexity{}
			err = patch.UnmarshalContents(message.Contents, message.MimeType, complexity)
			if err != nil {
				return err
			}
			if complexity.PathCount == 0 ||
				complexity.SchemaCount == 0 {
				t.Errorf("Failed to compute %s", artifactName.String())
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Error getting artifact: %s", err)
		}
	})

	t.Run("openapi", func(t *testing.T) {
		complexityCmd := Command()
		complexityCmd.SetArgs([]string{project.Api("apigeeregistry").Version("v1").Spec("openapi").String()})
		if err := complexityCmd.Execute(); err != nil {
			t.Fatalf("Compute complexity failed: %s", err)
		}

		artifactName := project.Api("apigeeregistry").Version("v1").Spec("openapi").Artifact("complexity")
		err = visitor.GetArtifact(ctx, registryClient, artifactName, true, func(ctx context.Context, message *rpc.Artifact) error {
			complexity := &metrics.Complexity{}
			err = patch.UnmarshalContents(message.Contents, message.MimeType, complexity)
			if err != nil {
				return err
			}
			if complexity.PathCount == 0 ||
				complexity.SchemaCount == 0 {
				t.Errorf("Failed to compute %s", artifactName.String())
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Error getting artifact: %s", err)
		}
	})

	t.Run("discovery", func(t *testing.T) {
		complexityCmd := Command()
		complexityCmd.SetArgs([]string{project.Api("apigeeregistry").Version("v1").Spec("discovery").String()})
		if err := complexityCmd.Execute(); err != nil {
			t.Fatalf("Compute complexity failed: %s", err)
		}

		artifactName := project.Api("apigeeregistry").Version("v1").Spec("discovery").Artifact("complexity")
		err = visitor.GetArtifact(ctx, registryClient, artifactName, true, func(ctx context.Context, message *rpc.Artifact) error {
			complexity := &metrics.Complexity{}
			err = patch.UnmarshalContents(message.Contents, message.MimeType, complexity)
			if err != nil {
				return err
			}
			if complexity.PathCount == 0 ||
				complexity.SchemaCount == 0 {
				t.Errorf("Failed to compute %s", artifactName.String())
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Error getting artifact: %s", err)
		}
	})
}

func readAndGZipFile(t *testing.T, filename string) (*bytes.Buffer, error) {
	t.Helper()
	fileBytes, _ := os.ReadFile(filename)
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, err := zw.Write(fileBytes)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}
func TestComputeComplexityValues(t *testing.T) {
	tests := []struct {
		desc       string
		apiId      string
		versionId  string
		specId     string
		specFile   string
		mimeType   string
		getPattern string
		wantProto  *metrics.Complexity
	}{
		{
			desc:      "petstore-openapi",
			apiId:     "petstore",
			versionId: "1.0.0",
			specId:    "openapi",
			specFile:  "openapi.yaml",
			mimeType:  "application/x.openapi+gzip;version=3.0.0",
			wantProto: &metrics.Complexity{
				PathCount:           2,
				GetCount:            2,
				PostCount:           1,
				PutCount:            0,
				DeleteCount:         0,
				SchemaCount:         8,
				SchemaPropertyCount: 5,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			testProject := "complexity-test"
			client, adminClient := grpctest.SetupRegistry(ctx, t, testProject, nil)
			project := rpc.Project{
				Name: "projects/" + testProject,
			}

			api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
				Parent: project.Name + "/locations/global",
				ApiId:  test.apiId,
				Api: &rpc.Api{
					DisplayName: test.apiId,
				},
			})
			if err != nil {
				t.Fatalf("Failed to create API %s: %s", test.apiId, err.Error())
			}
			v1, err := client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
				Parent:       api.Name,
				ApiVersionId: test.versionId,
				ApiVersion:   &rpc.ApiVersion{},
			})
			if err != nil {
				t.Fatalf("Failed to create version %s: %s", test.versionId, err.Error())
			}
			buf, err := readAndGZipFile(t, filepath.Join("..", "testdata", test.specFile))
			if err != nil {
				t.Fatalf("Failed reading spec contents: %s", err.Error())
			}
			req := &rpc.CreateApiSpecRequest{
				Parent:    v1.Name,
				ApiSpecId: test.specId,
				ApiSpec: &rpc.ApiSpec{
					MimeType: test.mimeType,
					Contents: buf.Bytes(),
				},
			}
			spec, err := client.CreateApiSpec(ctx, req)
			if err != nil {
				t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
			}
			complexityCommand := Command()
			args := []string{spec.Name}
			complexityCommand.SetArgs(args)
			if err = complexityCommand.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}
			specName, err := names.ParseSpec(spec.Name)
			if err != nil {
				t.Fatalf("Failed parsing spec name %s: %s", spec.Name, err)
			}
			contents, err := client.GetArtifactContents(ctx, &rpc.GetArtifactContentsRequest{
				Name: specName.Artifact("complexity").String(),
			})
			if err != nil {
				t.Fatalf("Failed getting artifact contents %s: %s", test.getPattern, err)
			}
			gotProto := &metrics.Complexity{}
			if err := patch.UnmarshalContents(contents.GetData(), contents.GetContentType(), gotProto); err != nil {
				t.Fatalf("Failed to unmarshal artifact: %s", err)
			}
			opts := cmp.Options{protocmp.Transform()}
			if !cmp.Equal(test.wantProto, gotProto, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantProto, gotProto, opts))
			}
			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + testProject,
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}
		})
	}
}
