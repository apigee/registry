package patch

import (
	"context"
	"os"
	"testing"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/names"
	"github.com/apigee/registry/server/registry/test/seeder"
	metrics "github.com/google/gnostic/metrics"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestSpecArtifactPatches(t *testing.T) {
	tests := []struct {
		artifactName string
		yamlFileName string
		message      proto.Message
	}{
		{
			artifactName: "complexity",
			yamlFileName: "testdata/artifacts/complexity.yaml",
			message: &metrics.Complexity{
				PathCount:           76,
				GetCount:            25,
				PostCount:           27,
				PutCount:            11,
				DeleteCount:         13,
				SchemaCount:         1150,
				SchemaPropertyCount: 964,
			},
		},
		{
			artifactName: "vocabulary",
			yamlFileName: "testdata/artifacts/vocabulary.yaml",
			message: &metrics.Vocabulary{
				Name:       "sample-name",
				Schemas:    []*metrics.WordCount{{Word: "sample-schema", Count: 1}},
				Properties: []*metrics.WordCount{{Word: "sample-property", Count: 2}},
				Operations: []*metrics.WordCount{{Word: "sample-operation", Count: 3}},
				Parameters: []*metrics.WordCount{{Word: "sample-parameter", Count: 4}},
			},
		},
	}
	ctx := context.Background()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Setup: failed to create client: %+v", err)
	}
	defer adminClient.Close()
	registryClient, err := connection.NewRegistryClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create registry client: %s", err)
	}
	defer registryClient.Close()
	client := seeder.Client{
		RegistryClient: registryClient,
		AdminClient:    adminClient,
	}
	spec := &rpc.ApiSpec{
		Name: "projects/patch-project-test/locations/global/apis/a/versions/v/specs/s",
	}
	if err := seeder.SeedSpecs(ctx, client, spec); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	for _, test := range tests {
		t.Run(test.artifactName, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFileName)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyArtifactPatchBytes(ctx, registryClient, b, "projects/patch-project-test/locations/global")
			if err != nil {
				t.Fatalf("%s", err)
			}
			artifactName, err := names.ParseArtifact("projects/patch-project-test/locations/global/apis/a/versions/v/specs/s/artifacts/" + test.artifactName)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = core.GetArtifact(ctx, registryClient, artifactName, true,
				func(artifact *rpc.Artifact) error {
					contents, err := core.GetArtifactMessageContents(artifact)
					if err != nil {
						t.Fatalf("%s", err)
					}
					opts := cmp.Options{protocmp.Transform()}
					if !cmp.Equal(test.message, contents, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, contents, opts))
					}
					out, header, err := ExportArtifact(ctx, registryClient, artifact)
					if err != nil {
						t.Fatalf("%s", err)
					}
					wantedParent := "apis/a/versions/v/specs/s"
					if header.Metadata.Parent != wantedParent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", wantedParent, header.Metadata.Parent)
					}
					if header.Metadata.Name != test.artifactName {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.artifactName, header.Metadata.Name)
					}
					if !cmp.Equal(b, out, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(b, out, opts))
					}
					return nil
				})
			if err != nil {
				t.Fatalf("%s", err)
			}
		})
	}
}
