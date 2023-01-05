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

func TestProjectPatches(t *testing.T) {
	tests := []struct {
		desc       string
		resourceID string
		yamlFile   string
		parent     string
		message    proto.Message
	}{
		{
			desc:       "sample",
			resourceID: "sample",
			yamlFile:   "testdata/resources/projects-sample.yaml",
			parent:     "",
			message: &rpc.Project{
				Name:        "projects/sample",
				DisplayName: "Sample",
				Description: "This sample project is described by a YAML file.",
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
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyProjectPatchBytes(ctx, adminClient, b)
			if err != nil {
				t.Fatalf("%s", err)
			}
			collection := "projects/"
			projectName, err := names.ParseProject(collection + test.resourceID)
			if err != nil {
				t.Fatalf("%s", err)
			}
			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Project), "create_time", "update_time"),
			}
			err = core.GetProject(ctx, adminClient, projectName,
				func(project *rpc.Project) error {
					if !cmp.Equal(test.message, project, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, project, opts))
					}
					model, err := PatchForProject(ctx, registryClient, project)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := Encode(model)
					if err != nil {
						t.Errorf("Encode(%+v) returned an error: %s", model, err)
					}
					if !cmp.Equal(b, out, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(b, out, opts))
					}
					return nil
				})
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + test.resourceID,
				Force: true,
			})
			if err != nil {
				t.Fatalf("%s", err)
			}
		})
	}
}

func TestApiPatches(t *testing.T) {
	root := "projects/patch-api-test/locations/global"
	tests := []struct {
		desc       string
		resourceID string
		parent     string
		yamlFile   string
		nested     bool
		message    proto.Message
	}{
		{
			desc:       "registry",
			resourceID: "registry",
			parent:     "",
			yamlFile:   "testdata/resources/apis-registry.yaml",
			nested:     false,
			message: &rpc.Api{
				Name:                  root + "/apis/registry",
				DisplayName:           "Apigee Registry API",
				Description:           "The Registry API allows teams to track and manage machine-readable descriptions of APIs.",
				Labels:                map[string]string{"apihub-owner": "google"},
				Annotations:           map[string]string{"apihub-score": "99"},
				Availability:          "Preview",
				RecommendedDeployment: root + "/apis/registry/deployments/prod",
				RecommendedVersion:    root + "/apis/registry/versions/v1",
			},
		},
		{
			desc:       "registry-nested",
			resourceID: "registry",
			parent:     "",
			yamlFile:   "testdata/resources/apis-registry-nested.yaml",
			nested:     true,
			message: &rpc.Api{
				Name:                  root + "/apis/registry",
				DisplayName:           "Apigee Registry API",
				Description:           "The Registry API allows teams to track and manage machine-readable descriptions of APIs.",
				Labels:                map[string]string{"apihub-owner": "google"},
				Annotations:           map[string]string{"apihub-score": "99"},
				Availability:          "Preview",
				RecommendedDeployment: root + "/apis/registry/deployments/prod",
				RecommendedVersion:    root + "/apis/registry/versions/v1",
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
	project := &rpc.Project{
		Name: "projects/patch-api-test",
	}
	if err := seeder.SeedProjects(ctx, client, project); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyApiPatchBytes(ctx, registryClient, b, root)
			if err != nil {
				t.Fatalf("%s", err)
			}
			collection := root + "/apis/"
			apiName, err := names.ParseApi(collection + test.resourceID)
			if err != nil {
				t.Fatalf("%s", err)
			}
			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.Api), "create_time", "update_time"),
			}
			err = core.GetAPI(ctx, registryClient, apiName,
				func(api *rpc.Api) error {
					if !cmp.Equal(test.message, api, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, api, opts))
					}
					model, err := PatchForApi(ctx, registryClient, api, test.nested)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := Encode(model)
					if err != nil {
						t.Errorf("Encode(%+v) returned an error: %s", model, err)
					}
					if !cmp.Equal(b, out, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(b, out, opts))
					}
					return nil
				})
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = registryClient.DeleteApi(ctx, &rpc.DeleteApiRequest{
				Name:  root + "/apis/" + test.resourceID,
				Force: true,
			})
			if err != nil {
				t.Fatalf("%s", err)
			}
		})
	}
}

func TestVersionPatches(t *testing.T) {
	root := "projects/patch-version-test/locations/global"
	tests := []struct {
		resourceID string
		parent     string
		yamlFile   string
		nested     bool
		message    proto.Message
	}{
		{
			resourceID: "v1",
			parent:     "apis/registry",
			yamlFile:   "testdata/resources/apis-registry-versions-v1.yaml",
			nested:     false,
			message: &rpc.ApiVersion{
				Name:        root + "/apis/registry/versions/v1",
				DisplayName: "v1",
				Description: "New in 2022",
				State:       "Staging",
				Labels:      map[string]string{"apihub-team": "apigee"},
				Annotations: map[string]string{"release-date": "2021-12-15"},
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
	api := &rpc.Api{
		Name: root + "/apis/registry",
	}
	if err := seeder.SeedApis(ctx, client, api); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	for _, test := range tests {
		t.Run(test.resourceID, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyApiVersionPatchBytes(ctx, registryClient, b, root)
			if err != nil {
				t.Fatalf("%s", err)
			}
			collection := root + "/apis/registry/versions/"
			versionName, err := names.ParseVersion(collection + test.resourceID)
			if err != nil {
				t.Fatalf("%s", err)
			}
			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiVersion), "create_time", "update_time"),
			}
			err = core.GetVersion(ctx, registryClient, versionName,
				func(version *rpc.ApiVersion) error {
					if !cmp.Equal(test.message, version, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, version, opts))
					}
					model, err := PatchForApiVersion(ctx, registryClient, version, test.nested)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := Encode(model)
					if err != nil {
						t.Errorf("Encode(%+v) returned an error: %s", model, err)
					}
					if !cmp.Equal(b, out, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(b, out, opts))
					}
					return nil
				})
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = registryClient.DeleteApiVersion(ctx, &rpc.DeleteApiVersionRequest{
				Name:  root + "/apis/registry/versions/" + test.resourceID,
				Force: true,
			})
			if err != nil {
				t.Fatalf("%s", err)
			}
		})
	}
}

func TestSpecPatches(t *testing.T) {
	root := "projects/patch-spec-test/locations/global"
	tests := []struct {
		resourceID string
		parent     string
		yamlFile   string
		nested     bool
		message    proto.Message
	}{
		{
			resourceID: "openapi",
			parent:     "apis/registry/versions/v1",
			yamlFile:   "testdata/resources/apis-registry-versions-v1-specs-openapi.yaml",
			nested:     false,
			message: &rpc.ApiSpec{
				Name:        root + "/apis/registry/versions/v1/specs/openapi",
				Description: "OpenAPI description of the Registry API",
				Filename:    "openapi.yaml",
				MimeType:    "application/x.openapi+gzip;version=3",
				SourceUri:   "https://raw.githubusercontent.com/apigee/registry/main/openapi.yaml",
				Labels:      map[string]string{"openapi-verified": "true"},
				Annotations: map[string]string{"linters": "spectral,gnostic"},
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
	version := &rpc.ApiVersion{
		Name: root + "/apis/registry/versions/v1",
	}
	if err := seeder.SeedVersions(ctx, client, version); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	for _, test := range tests {
		t.Run(test.resourceID, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyApiSpecPatchBytes(ctx, registryClient, b, root, "")
			if err != nil {
				t.Fatalf("%s", err)
			}
			collection := root + "/apis/registry/versions/v1/specs/"
			specName, err := names.ParseSpec(collection + test.resourceID)
			if err != nil {
				t.Fatalf("%s", err)
			}
			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiSpec), "hash", "size_bytes", "revision_id", "create_time", "revision_create_time", "revision_update_time"),
			}
			err = core.GetSpec(ctx, registryClient, specName, false,
				func(spec *rpc.ApiSpec) error {
					if !cmp.Equal(test.message, spec, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, spec, opts))
					}
					model, err := PatchForApiSpec(ctx, registryClient, spec, test.nested)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := Encode(model)
					if err != nil {
						t.Errorf("Encode(%+v) returned an error: %s", model, err)
					}
					if !cmp.Equal(b, out, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(b, out, opts))
					}
					return nil
				})
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = registryClient.DeleteApiSpec(ctx, &rpc.DeleteApiSpecRequest{
				Name:  root + "/apis/registry/versions/v1/specs/" + test.resourceID,
				Force: true,
			})
			if err != nil {
				t.Fatalf("%s", err)
			}
		})
	}
}

func TestDeploymentPatches(t *testing.T) {
	root := "projects/patch-deployment-test/locations/global"
	tests := []struct {
		resourceID string
		parent     string
		yamlFile   string
		nested     bool
		message    proto.Message
	}{
		{
			resourceID: "prod",
			parent:     "apis/registry",
			yamlFile:   "testdata/resources/apis-registry-deployments-prod.yaml",
			nested:     false,
			message: &rpc.ApiDeployment{
				Name:               root + "/apis/registry/deployments/prod",
				AccessGuidance:     "See https://github.com/apigee/registry for tools and usage information.",
				ApiSpecRevision:    "projects/patch-deployment-test/locations/global/apis/registry/versions/v1/specs/openapi@latest",
				DisplayName:        "Production",
				Description:        "The hosted deployment of the Registry API",
				EndpointUri:        "https://apigeeregistry.googleapis.com",
				ExternalChannelUri: "https://apigee.github.io/registry/",
				IntendedAudience:   "Public",
				Labels:             map[string]string{"platform": "google"},
				Annotations:        map[string]string{"region": "us-central1"},
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
	api := &rpc.Api{
		Name: root + "/apis/registry",
	}
	if err := seeder.SeedApis(ctx, client, api); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	for _, test := range tests {
		t.Run(test.resourceID, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyApiDeploymentPatchBytes(ctx, registryClient, b, root)
			if err != nil {
				t.Fatalf("%s", err)
			}
			collection := root + "/apis/registry/deployments/"
			deploymentName, err := names.ParseDeployment(collection + test.resourceID)
			if err != nil {
				t.Fatalf("%s", err)
			}
			opts := cmp.Options{
				protocmp.Transform(),
				protocmp.IgnoreFields(new(rpc.ApiDeployment), "revision_id", "create_time", "revision_create_time", "revision_update_time"),
			}
			err = core.GetDeployment(ctx, registryClient, deploymentName,
				func(deployment *rpc.ApiDeployment) error {
					if !cmp.Equal(test.message, deployment, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, deployment, opts))
					}
					model, err := PatchForApiDeployment(ctx, registryClient, deployment, test.nested)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := Encode(model)
					if err != nil {
						t.Errorf("Encode(%+v) returned an error: %s", model, err)
					}
					if !cmp.Equal(b, out, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(b, out, opts))
					}
					return nil
				})
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = registryClient.DeleteApiDeployment(ctx, &rpc.DeleteApiDeploymentRequest{
				Name:  root + "/apis/registry/deployments/" + test.resourceID,
				Force: true,
			})
			if err != nil {
				t.Fatalf("%s", err)
			}
		})
	}
}

func TestMessageArtifactPatches(t *testing.T) {
	root := "projects/patch-message-artifact-test/locations/global"
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
						Filter:  "invalid-filter",
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
						Resource:    "invalid-resource",
						Uri:         "https://github.com/apigee/registry",
					},
					{
						Id:          "docs",
						DisplayName: "GitHub Documentation",
						Category:    "apihub-other",
						Resource:    "invalid-resource",
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
		Name: root + "/apis/a/versions/v/specs/s",
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
			err = applyArtifactPatchBytes(ctx, registryClient, b, root)
			if err != nil {
				t.Fatalf("%s", err)
			}
			var collection string
			if test.parent != "" {
				collection = root + "/" + test.parent + "/artifacts/"
			} else {
				collection = root + "/artifacts/"
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
					model, err := PatchForArtifact(ctx, registryClient, artifact)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.artifactID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.artifactID, model.Header.Metadata.Name)
					}
					out, err := Encode(model)
					if err != nil {
						t.Errorf("Encode(%+v) returned an error: %s", model, err)
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

func TestYamlArtifactPatches(t *testing.T) {
	root := "projects/patch-yaml-artifact-test/locations/global"
	tests := []struct {
		artifactID string
		kind       string
		parent     string
		yamlFile   string
	}{
		{
			artifactID: "struct",
			kind:       "Struct",
			parent:     "",
			yamlFile:   "testdata/artifacts/struct.yaml",
		},
		{
			artifactID: "untyped",
			kind:       "",
			parent:     "apis/a/versions/v/specs/s",
			yamlFile:   "testdata/artifacts/untyped.yaml",
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
		Name: root + "/apis/a/versions/v/specs/s",
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
			err = applyArtifactPatchBytes(ctx, registryClient, b, root)
			if err != nil {
				t.Fatalf("%s", err)
			}
			var collection string
			if test.parent != "" {
				collection = root + "/" + test.parent + "/artifacts/"
			} else {
				collection = root + "/artifacts/"
			}
			artifactName, err := names.ParseArtifact(collection + test.artifactID)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = core.GetArtifact(ctx, registryClient, artifactName, true,
				func(artifact *rpc.Artifact) error {
					model, err := PatchForArtifact(ctx, registryClient, artifact)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Kind != test.kind {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.kind, model.Header.Kind)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.artifactID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.artifactID, model.Header.Metadata.Name)
					}
					out, err := Encode(model)
					if err != nil {
						t.Errorf("Encode(%+v) returned an error: %s", model, err)
					}
					if !cmp.Equal(b, out) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(b, out))
					}
					return nil
				})
			if err != nil {
				t.Fatalf("%s", err)
			}
		})
	}
}

func TestInvalidArtifactPatches(t *testing.T) {
	root := "projects/patch-invalid-artifact-test/locations/global"
	tests := []struct {
		artifactID string
	}{
		{
			artifactID: "complexity-invalid-field",
		},
		{
			artifactID: "lifecycle-invalid-parent",
		}, {
			artifactID: "references-no-data",
		},
		{
			artifactID: "struct-invalid-structure",
		},
		{
			artifactID: "struct-no-metadata",
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
		Name: root + "/apis/a/versions/v/specs/s",
	}
	if err := seeder.SeedSpecs(ctx, client, spec); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}
	for _, test := range tests {
		t.Run(test.artifactID, func(t *testing.T) {
			yamlFile := "testdata/invalid-artifacts/" + test.artifactID + ".yaml"
			b, err := os.ReadFile(yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyArtifactPatchBytes(ctx, registryClient, b, root)
			if err == nil {
				t.Fatalf("expected error, received none")
			}
		})
	}
}
