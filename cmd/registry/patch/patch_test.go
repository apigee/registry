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

func TestArtifactPatches(t *testing.T) {
	tests := []struct {
		artifactID string
		parent     string
		yamlFile   string
		message    proto.Message
	}{
		{
			artifactID: "complexity",
			parent:     "apis/a/versions/v/specs/s",
			yamlFile:   "testdata/artifacts/complexity.yaml",
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
			artifactID: "display-settings",
			parent:     "",
			yamlFile:   "testdata/artifacts/displaysettings.yaml",
			message: &rpc.DisplaySettings{
				Id:              "display-settings", // deprecated field
				Kind:            "DisplaySettings",  // deprecated field
				Description:     "Defines display settings",
				Organization:    "Sample",
				ApiGuideEnabled: true,
				ApiScoreEnabled: true,
			},
		},
		{
			artifactID: "extensions",
			parent:     "",
			yamlFile:   "testdata/artifacts/extensions.yaml",
			message: &rpc.ApiSpecExtensionList{
				Id:          "extensions",           // deprecated field
				Kind:        "ApiSpecExtensionList", // deprecated field
				DisplayName: "Sample Extensions",
				Description: "Extensions connect external tools to registry applications",
				Extensions: []*rpc.ApiSpecExtensionList_ApiSpecExtension{
					{
						Id:          "sample",
						DisplayName: "Sample",
						Description: "A sample extension",
						Filter:      "mime_type.contains('openapi')",
						UriPattern:  "https://example.com",
					},
				},
			},
		},
		{
			artifactID: "lifecycle",
			yamlFile:   "testdata/artifacts/lifecycle.yaml",
			message: &rpc.Lifecycle{
				Id:          "lifecycle", // deprecated field
				Kind:        "Lifecycle", // deprecated field
				DisplayName: "Lifecycle",
				Description: "A series of stages that an API typically moves through in its lifetime",
				Stages: []*rpc.Lifecycle_Stage{
					{
						Id:           "concept",
						DisplayName:  "Concept",
						Description:  "Description of the business case and user needs for why an API should exist",
						Url:          "https://google.com",
						DisplayOrder: 0,
					},
					{
						Id:           "design",
						DisplayName:  "Design",
						Description:  "Definition of the interface details and proposal of the API contract",
						Url:          "https://google.com",
						DisplayOrder: 1,
					},
					{
						Id:           "develop",
						DisplayName:  "Develop",
						Description:  "Implementation of the service and its API",
						Url:          "https://google.com",
						DisplayOrder: 2,
					},
				},
			},
		},
		{
			artifactID: "manifest",
			yamlFile:   "testdata/artifacts/manifest.yaml",
			message: &rpc.Manifest{
				Id:          "manifest", // deprecated field
				Kind:        "Manifest", // deprecated field
				DisplayName: "Sample Manifest",
				Description: "A sample manifest",
				GeneratedResources: []*rpc.GeneratedResource{
					{
						Pattern: "apis/-/versions/-/specs/-/artifacts/lint-spectral",
						Filter:  "",
						Receipt: false,
						Dependencies: []*rpc.Dependency{
							{
								Pattern: "$resource.spec",
								Filter:  "mime_type.contains('openapi')",
							},
						},
						Action:  "registry compute lint $resource.spec --linter spectral",
						Refresh: nil,
					},
				},
			},
		},
		{
			artifactID: "references",
			parent:     "apis/a",
			yamlFile:   "testdata/artifacts/references.yaml",
			message: &rpc.ReferenceList{
				Id:          "references",    // deprecated field
				Kind:        "ReferenceList", // deprecated field
				DisplayName: "Related References",
				Description: "References related to this API",
				References: []*rpc.ReferenceList_Reference{
					{
						Id:          "github",
						DisplayName: "GitHub Repo",
						Category:    "apihub-source-code",
						Resource:    "",
						Uri:         "https://github.com/apigee/registry",
					},
					{
						Id:          "docs",
						DisplayName: "GitHub Documentation",
						Category:    "apihub-other",
						Resource:    "",
						Uri:         "https://apigee.github.io/registry/",
					},
				},
			},
		},
		{
			artifactID: "styleguide",
			yamlFile:   "testdata/artifacts/styleguide.yaml",
			message: &rpc.StyleGuide{
				Id:          "styleguide", // deprecated field
				Kind:        "StyleGuide", // deprecated field
				DisplayName: "Sample Style Guide",
				MimeTypes: []string{
					"application/x.openapi+gzip;version=2",
				},
				Guidelines: []*rpc.Guideline{
					{
						Id:          "refproperties",
						DisplayName: "Govern Ref Properties",
						Description: "This guideline governs properties for ref fields on specs.",
						Rules: []*rpc.Rule{
							{
								Id:             "norefsiblings",
								DisplayName:    "No Ref Siblings",
								Description:    "An object exposing a $ref property cannot be further extended with additional properties.",
								Linter:         "spectral",
								LinterRulename: "no-$ref-siblings",
								Severity:       rpc.Rule_ERROR,
								DocUri:         "https://meta.stoplight.io/docs/spectral/4dec24461f3af-open-api-rules#no-ref-siblings",
							},
						},
						State: rpc.Guideline_ACTIVE,
					},
				},
				Linters: []*rpc.Linter{
					{
						Name: "spectral",
						Uri:  "https://github.com/stoplightio/spectral",
					},
				},
			},
		},
		{
			artifactID: "taxonomies",
			yamlFile:   "testdata/artifacts/taxonomies.yaml",
			message: &rpc.TaxonomyList{
				Id:          "taxonomies",   // deprecated field
				Kind:        "TaxonomyList", // deprecated field
				DisplayName: "TaxonomyList",
				Description: "A list of taxonomies that can be used to classify resources in the registry",
				Taxonomies: []*rpc.TaxonomyList_Taxonomy{
					{
						Id:              "target-users",
						DisplayName:     "Target users",
						Description:     "The intended users (consumers) of an API",
						AdminApplied:    false,
						SingleSelection: false,
						SearchExcluded:  false,
						SystemManaged:   true,
						DisplayOrder:    0,
						Elements: []*rpc.TaxonomyList_Taxonomy_Element{
							{
								Id:          "team",
								DisplayName: "Team",
								Description: "Intended for exclusive use by the producing team",
							},
							{
								Id:          "internal",
								DisplayName: "Internal",
								Description: "Available to internal teams",
							},
							{
								Id:          "partner",
								DisplayName: "Partner",
								Description: "Available to select partners",
							},
							{
								Id:          "public",
								DisplayName: "Public",
								Description: "Published for discovery by the general public",
							},
						},
					},
					{
						Id:              "style",
						DisplayName:     "Style (primary)",
						Description:     "The primary architectural style of the API",
						AdminApplied:    false,
						SingleSelection: true,
						SearchExcluded:  false,
						SystemManaged:   true,
						DisplayOrder:    1,
						Elements: []*rpc.TaxonomyList_Taxonomy_Element{
							{
								Id:          "openapi",
								DisplayName: "OpenAPI",
								Description: "https://spec.openapis.org/oas/latest.html",
							},
							{
								Id:          "grpc",
								DisplayName: "gRPC",
								Description: "https://grpc.io",
							},
							{
								Id:          "graphql",
								DisplayName: "GraphQL",
								Description: "https://graphql.org",
							},
							{
								Id:          "asyncapi",
								DisplayName: "AsyncAPI",
								Description: "https://www.asyncapi.com",
							},
							{
								Id:          "soap",
								DisplayName: "SOAP",
								Description: "https://en.wikipedia.org/wiki/Web_Services_Description_Language",
							},
						},
					},
				},
			},
		},
		{
			artifactID: "vocabulary",
			parent:     "apis/a/versions/v/specs/s",
			yamlFile:   "testdata/artifacts/vocabulary.yaml",
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
		t.Run(test.artifactID, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyArtifactPatchBytes(ctx, registryClient, b, "projects/patch-project-test/locations/global")
			if err != nil {
				t.Fatalf("%s", err)
			}
			var collection string
			if test.parent != "" {
				collection = "projects/patch-project-test/locations/global/" + test.parent + "/artifacts/"
			} else {
				collection = "projects/patch-project-test/locations/global/artifacts/"
			}
			artifactName, err := names.ParseArtifact(collection + test.artifactID)
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
					if header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, header.Metadata.Parent)
					}
					if header.Metadata.Name != test.artifactID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.artifactID, header.Metadata.Name)
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
