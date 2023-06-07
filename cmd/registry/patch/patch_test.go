// Copyright 2022 Google LLC.
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

package patch

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/apigee/registry/pkg/application/apihub"
	"github.com/apigee/registry/pkg/application/controller"
	"github.com/apigee/registry/pkg/application/scoring"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/encoding"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
	metrics "github.com/google/gnostic/metrics"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
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

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			registryClient, adminClient := grpctest.SetupRegistry(ctx, t, test.resourceID, nil)

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
			err = visitor.GetProject(ctx, adminClient, projectName, nil,
				func(ctx context.Context, project *rpc.Project) error {
					if !cmp.Equal(test.message, project, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, project, opts))
					}
					model, err := NewProject(ctx, registryClient, project)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := encoding.EncodeYAML(model)
					if err != nil {
						t.Errorf("encoding.EncodeYAML(%+v) returned an error: %s", model, err)
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
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "projects/patch-api-test", nil)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyApiPatchBytes(ctx, registryClient, b, root, "patch.yaml")
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
			err = visitor.GetAPI(ctx, registryClient, apiName,
				func(ctx context.Context, api *rpc.Api) error {
					if !cmp.Equal(test.message, api, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, api, opts))
					}
					model, err := NewApi(ctx, registryClient, api, test.nested)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := encoding.EncodeYAML(model)
					if err != nil {
						t.Errorf("encoding.EncodeYAML(%+v) returned an error: %s", model, err)
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
				PrimarySpec: "apis/registry/versions/v1/specs/openapi",
				Labels:      map[string]string{"apihub-team": "apigee"},
				Annotations: map[string]string{"release-date": "2021-12-15"},
			},
		},
	}
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "patch-version-test", []seeder.RegistryResource{
		&rpc.Api{
			Name: root + "/apis/registry",
		},
	})

	for _, test := range tests {
		t.Run(test.resourceID, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyApiVersionPatchBytes(ctx, registryClient, b, root, "patch.yaml")
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
			err = visitor.GetVersion(ctx, registryClient, versionName,
				func(ctx context.Context, version *rpc.ApiVersion) error {
					if !cmp.Equal(test.message, version, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, version, opts))
					}
					model, err := NewApiVersion(ctx, registryClient, version, test.nested)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := encoding.EncodeYAML(model)
					if err != nil {
						t.Errorf("encoding.EncodeYAML(%+v) returned an error: %s", model, err)
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
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "patch-spec-test", []seeder.RegistryResource{
		&rpc.ApiVersion{
			Name: root + "/apis/registry/versions/v1",
		},
	})

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
			err = visitor.GetSpec(ctx, registryClient, specName, false,
				func(ctx context.Context, spec *rpc.ApiSpec) error {
					if !cmp.Equal(test.message, spec, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, spec, opts))
					}
					model, err := NewApiSpec(ctx, registryClient, spec, test.nested)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := encoding.EncodeYAML(model)
					if err != nil {
						t.Errorf("encoding.EncodeYAML(%+v) returned an error: %s", model, err)
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
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "patch-deployment-test", []seeder.RegistryResource{
		&rpc.Api{
			Name: root + "/apis/registry",
		},
	})

	for _, test := range tests {
		t.Run(test.resourceID, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyApiDeploymentPatchBytes(ctx, registryClient, b, root, "patch.yaml")
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
			err = visitor.GetDeployment(ctx, registryClient, deploymentName,
				func(ctx context.Context, deployment *rpc.ApiDeployment) error {
					if !cmp.Equal(test.message, deployment, opts) {
						t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, deployment, opts))
					}
					model, err := NewApiDeployment(ctx, registryClient, deployment, test.nested)
					if err != nil {
						t.Fatalf("%s", err)
					}
					if model.Header.Metadata.Parent != test.parent {
						t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
					}
					if model.Header.Metadata.Name != test.resourceID {
						t.Errorf("Incorrect export name. Wanted %s, got %s", test.resourceID, model.Header.Metadata.Name)
					}
					out, err := encoding.EncodeYAML(model)
					if err != nil {
						t.Errorf("encoding.EncodeYAML(%+v) returned an error: %s", model, err)
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
			artifactID: "fieldset",
			parent:     "apis/a",
			yamlFile:   "testdata/artifacts/fieldset.yaml",
			message: &apihub.FieldSet{
				Id:             "fieldset",
				Kind:           "FieldSet",
				DefinitionName: "artifacts/fieldset",
				Values: map[string]string{
					"creator":       "Wile E. Coyote",
					"creator-email": "wiley@acme.com",
					"hometown":      "Las Vegas, NV",
					"website":       "[ACME](https://acme.com)",
				},
			},
		},
		{
			artifactID: "fieldset",
			yamlFile:   "testdata/artifacts/fieldset-definition.yaml",
			message: &apihub.FieldSetDefinition{
				Id:          "fieldset",
				Kind:        "FieldSetDefinition",
				DisplayName: "Interesting Information",
				Description: "Additional topical information about this API.",
				Fields: []*apihub.FieldDefinition{
					{
						Id:          "creator",
						DisplayName: "Creator",
					}, {
						Id:          "creator-email",
						DisplayName: "Creator Email",
					}, {
						Id:          "hometown",
						DisplayName: "Hometown",
					}, {
						Id:          "website",
						DisplayName: "Website",
					},
				},
			},
		},
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
			artifactID: "conformancereport",
			yamlFile:   "testdata/artifacts/conformancereport.yaml",
			message: &style.ConformanceReport{
				Id:         "conformancereport",
				Kind:       "ConformanceReport",
				Styleguide: "projects/demo/locations/global/artifacts/styleguide",
				GuidelineReportGroups: []*style.GuidelineReportGroup{
					{
						State: style.Guideline_ACTIVE,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "sample-guideline",
								RuleReportGroups: []*style.RuleReportGroup{
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:     "no-ref-siblings",
												Spec:       "projects/demo/locations/global/apis/petstore/versions/v1/specs/openapi",
												File:       "openapi.yaml",
												Suggestion: "",
												Location: &style.LintLocation{
													StartPosition: &style.LintPosition{
														LineNumber:   10,
														ColumnNumber: 5,
													},
													EndPosition: &style.LintPosition{
														LineNumber:   10,
														ColumnNumber: 25,
													},
												},
												DisplayName: "No ref siblings",
												Description: "Represents a sample rule",
												DocUri:      "https://meta.stoplight.io/docs/spectral/4dec24461f3af-open-api-rules#no-ref-siblings",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			artifactID: "display-settings",
			yamlFile:   "testdata/artifacts/displaysettings.yaml",
			message: &apihub.DisplaySettings{
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
			message: &apihub.ApiSpecExtensionList{
				Id:          "extensions",           // deprecated field
				Kind:        "ApiSpecExtensionList", // deprecated field
				DisplayName: "Sample Extensions",
				Description: "Extensions connect external tools to registry applications",
				Extensions: []*apihub.ApiSpecExtensionList_ApiSpecExtension{
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
			message: &apihub.Lifecycle{
				Id:          "lifecycle", // deprecated field
				Kind:        "Lifecycle", // deprecated field
				DisplayName: "Lifecycle",
				Description: "A series of stages that an API typically moves through in its lifetime",
				Stages: []*apihub.Lifecycle_Stage{
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
			message: &controller.Manifest{
				Id:          "manifest", // deprecated field
				Kind:        "Manifest", // deprecated field
				DisplayName: "Sample Manifest",
				Description: "A sample manifest",
				GeneratedResources: []*controller.GeneratedResource{
					{
						Pattern: "apis/-/versions/-/specs/-/artifacts/lint-spectral",
						Filter:  "invalid-filter",
						Receipt: false,
						Dependencies: []*controller.Dependency{
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
			artifactID: "receipt",
			parent:     "apis/a/versions/v/specs/s",
			yamlFile:   "testdata/artifacts/receipt.yaml",
			message: &controller.Receipt{
				Id:          "receipt", // deprecated field
				Kind:        "Receipt", // deprecated field
				DisplayName: "Sample Receipt",
				Action:      "registry compute scorecard RESOURCE",
				Description: "Description",
				ResultUri:   "https://example.com",
			},
		},
		{
			artifactID: "references",
			parent:     "apis/a",
			yamlFile:   "testdata/artifacts/references.yaml",
			message: &apihub.ReferenceList{
				Id:          "references",    // deprecated field
				Kind:        "ReferenceList", // deprecated field
				DisplayName: "Related References",
				Description: "References related to this API",
				References: []*apihub.ReferenceList_Reference{
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
			artifactID: "score",
			yamlFile:   "testdata/artifacts/score.yaml",
			message: &scoring.Score{
				Id:             "score",
				Kind:           "Score",
				DisplayName:    "Sample Score",
				Description:    "Represents sample Score artifact",
				Uri:            "https://docs.stoplight.io/docs/spectral/4dec24461f3af-open-api-rules",
				UriDisplayName: "Spectral rules",
				DefinitionName: "projects/demo/locations/global/artifacts/sample-score-definition",
				Severity:       scoring.Severity_ALERT,
				Value: &scoring.Score_IntegerValue{
					IntegerValue: &scoring.IntegerValue{
						Value:    10,
						MinValue: 0,
						MaxValue: 100,
					},
				},
			},
		},
		{
			artifactID: "scorecard",
			yamlFile:   "testdata/artifacts/scorecard.yaml",
			message: &scoring.ScoreCard{
				Id:             "scorecard",
				Kind:           "ScoreCard",
				DisplayName:    "Sample ScoreCard",
				Description:    "Represents sample ScoreCard artifact",
				DefinitionName: "projects/demo/locations/global/artifacts/sample-scorecard-definition",
				Scores: []*scoring.Score{
					{
						Id:             "score1",
						Kind:           "Score",
						DisplayName:    "Sample Score 1",
						Description:    "Represents sample Score artifact",
						Uri:            "https://docs.stoplight.io/docs/spectral/4dec24461f3af-open-api-rules",
						UriDisplayName: "Spectral rules",
						DefinitionName: "projects/demo/locations/global/artifacts/sample-score-definition",
						Severity:       scoring.Severity_ALERT,
						Value: &scoring.Score_IntegerValue{
							IntegerValue: &scoring.IntegerValue{
								Value:    10,
								MinValue: 0,
								MaxValue: 100,
							},
						},
					},
					{
						Id:             "score2",
						Kind:           "Score",
						DisplayName:    "Sample Score 2",
						Description:    "Represents sample Score artifact",
						Uri:            "https://docs.stoplight.io/docs/spectral/4dec24461f3af-open-api-rules",
						UriDisplayName: "Spectral rules",
						DefinitionName: "projects/demo/locations/global/artifacts/sample-score-definition",
						Severity:       scoring.Severity_WARNING,
						Value: &scoring.Score_IntegerValue{
							IntegerValue: &scoring.IntegerValue{
								Value:    20,
								MinValue: 0,
								MaxValue: 100,
							},
						},
					},
				},
			},
		},
		{
			artifactID: "scorecarddefinition",
			yamlFile:   "testdata/artifacts/scorecarddefinition.yaml",
			message: &scoring.ScoreCardDefinition{
				Id:          "scorecarddefinition",
				Kind:        "ScoreCardDefinition",
				DisplayName: "Sample ScoreCard definition",
				Description: "Represents sample ScoreCard definition artifact",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
					Filter:  "mime_type.contains('openapi')",
				},
				ScorePatterns: []string{
					"$resource.spec/artifacts/sample-score-1",
					"$resource.spec/artifacts/sample-score-2",
				},
			},
		},
		{
			artifactID: "scoredefinition",
			yamlFile:   "testdata/artifacts/scoredefinition.yaml",
			message: &scoring.ScoreDefinition{
				Id:             "scoredefinition",
				Kind:           "ScoreDefinition",
				DisplayName:    "Sample Score definition",
				Description:    "Represents sample Score definition artifact",
				Uri:            "https://docs.stoplight.io/docs/spectral/4dec24461f3af-open-api-rules",
				UriDisplayName: "Spectral rules",
				TargetResource: &scoring.ResourcePattern{
					Pattern: "apis/-/versions/-/specs/-",
					Filter:  "mime_type.contains('openapi')",
				},
				Formula: &scoring.ScoreDefinition_ScoreFormula{
					ScoreFormula: &scoring.ScoreFormula{
						Artifact: &scoring.ResourcePattern{
							Pattern: "$resource.spec/artifacts/conformance-styleguide",
						},
						ScoreExpression: "sample expression",
					},
				},
				Type: &scoring.ScoreDefinition_Integer{
					Integer: &scoring.IntegerType{
						MinValue: 0,
						MaxValue: 100,
						Thresholds: []*scoring.NumberThreshold{
							{
								Severity: scoring.Severity_ALERT,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 0,
									Max: 30,
								},
							},
							{
								Severity: scoring.Severity_WARNING,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 31,
									Max: 60,
								},
							},
							{
								Severity: scoring.Severity_OK,
								Range: &scoring.NumberThreshold_NumberRange{
									Min: 61,
									Max: 100,
								},
							},
						},
					},
				},
			},
		},
		{
			artifactID: "styleguide",
			yamlFile:   "testdata/artifacts/styleguide.yaml",
			message: &style.StyleGuide{
				Id:          "styleguide", // deprecated field
				Kind:        "StyleGuide", // deprecated field
				DisplayName: "Sample Style Guide",
				MimeTypes: []string{
					"application/x.openapi+gzip;version=2",
				},
				Guidelines: []*style.Guideline{
					{
						Id:          "refproperties",
						DisplayName: "Govern Ref Properties",
						Description: "This guideline governs properties for ref fields on specs.",
						Rules: []*style.Rule{
							{
								Id:             "norefsiblings",
								DisplayName:    "No Ref Siblings",
								Description:    "An object exposing a $ref property cannot be further extended with additional properties.",
								Linter:         "spectral",
								LinterRulename: "no-$ref-siblings",
								Severity:       style.Rule_ERROR,
								DocUri:         "https://meta.stoplight.io/docs/spectral/4dec24461f3af-open-api-rules#no-ref-siblings",
							},
						},
						State: style.Guideline_ACTIVE,
					},
				},
				Linters: []*style.Linter{
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
			message: &apihub.TaxonomyList{
				Id:          "taxonomies",   // deprecated field
				Kind:        "TaxonomyList", // deprecated field
				DisplayName: "TaxonomyList",
				Description: "A list of taxonomies that can be used to classify resources in the registry",
				Taxonomies: []*apihub.TaxonomyList_Taxonomy{
					{
						Id:              "target-users",
						DisplayName:     "Target users",
						Description:     "The intended users (consumers) of an API",
						AdminApplied:    false,
						SingleSelection: false,
						SearchExcluded:  false,
						SystemManaged:   true,
						DisplayOrder:    0,
						Elements: []*apihub.TaxonomyList_Taxonomy_Element{
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
						Elements: []*apihub.TaxonomyList_Taxonomy_Element{
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

	storageTypes := []struct {
		mimeBase string
		ctx      context.Context
	}{
		{"application/octet-stream", context.Background()},
		{"application/yaml", SetStoreArchivesAsYaml(context.Background())},
	}

	for _, storage := range storageTypes {
		t.Run(storage.mimeBase, func(t *testing.T) {
			ctx := storage.ctx

			registryClient, _ := grpctest.SetupRegistry(ctx, t, "patch-message-artifact-test", []seeder.RegistryResource{
				&rpc.ApiSpec{
					Name: root + "/apis/a/versions/v/specs/s",
				},
			})

			for _, test := range tests {
				t.Run(test.artifactID, func(t *testing.T) {
					b, err := os.ReadFile(test.yamlFile)
					if err != nil {
						t.Fatalf("%s", err)
					}
					err = applyArtifactPatchBytes(ctx, registryClient, b, root, "patch.yaml")
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
					err = visitor.GetArtifact(ctx, registryClient, artifactName, true,
						func(ctx context.Context, artifact *rpc.Artifact) error {
							// sanity check that we're really testing the right mime types
							if !strings.HasPrefix(artifact.GetMimeType(), storage.mimeBase) {
								t.Fatalf("unexpected storage type, want: %s, got: %s", storage.mimeBase, artifact.GetMimeType())
							}
							contents, err := getArtifactMessageContents(artifact)
							if err != nil {
								t.Fatalf("%s", err)
							}
							opts := cmp.Options{protocmp.Transform()}
							if !cmp.Equal(test.message, contents, opts) {
								t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.message, contents, opts))
							}
							model, err := NewArtifact(ctx, registryClient, artifact)
							if err != nil {
								t.Fatalf("%s", err)
							}
							if model.Header.Metadata.Parent != test.parent {
								t.Errorf("Incorrect export parent. Wanted %s, got %s", test.parent, model.Header.Metadata.Parent)
							}
							if model.Header.Metadata.Name != test.artifactID {
								t.Errorf("Incorrect export name. Wanted %s, got %s", test.artifactID, model.Header.Metadata.Name)
							}
							out, err := encoding.EncodeYAML(model)
							if err != nil {
								t.Errorf("encoding.EncodeYAML(%+v) returned an error: %s", model, err)
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
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "patch-message-artifact-test", []seeder.RegistryResource{
		&rpc.ApiSpec{
			Name: root + "/apis/a/versions/v/specs/s",
		},
	})

	for _, test := range tests {
		t.Run(test.artifactID, func(t *testing.T) {
			b, err := os.ReadFile(test.yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyArtifactPatchBytes(ctx, registryClient, b, root, "patch.yaml")
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
			err = visitor.GetArtifact(ctx, registryClient, artifactName, true,
				func(ctx context.Context, artifact *rpc.Artifact) error {
					model, err := NewArtifact(ctx, registryClient, artifact)
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
					out, err := encoding.EncodeYAML(model)
					if err != nil {
						t.Errorf("encoding.EncodeYAML(%+v) returned an error: %s", model, err)
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
		},
		{
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
	registryClient, _ := grpctest.SetupRegistry(ctx, t, "patch-message-artifact-test", []seeder.RegistryResource{
		&rpc.ApiSpec{
			Name: root + "/apis/a/versions/v/specs/s",
		},
	})

	for _, test := range tests {
		t.Run(test.artifactID, func(t *testing.T) {
			yamlFile := "testdata/invalid-artifacts/" + test.artifactID + ".yaml"
			b, err := os.ReadFile(yamlFile)
			if err != nil {
				t.Fatalf("%s", err)
			}
			err = applyArtifactPatchBytes(ctx, registryClient, b, root, "patch.yaml")
			if err == nil {
				t.Fatalf("expected error, received none")
			}
		})
	}
}

func getArtifactMessageContents(artifact *rpc.Artifact) (proto.Message, error) {
	message, err := mime.MessageForMimeType(artifact.GetMimeType())
	if err != nil {
		return nil, err
	}
	err = UnmarshalContents(artifact.GetContents(), artifact.GetMimeType(), message)

	// restore id and kind
	s := strings.Split(artifact.GetName(), "/")
	id := s[len(s)-1]
	kind := mime.KindForMimeType(artifact.GetMimeType())
	fields := message.ProtoReflect().Descriptor().Fields()
	if fd := fields.ByTextName("id"); fd != nil {
		message.ProtoReflect().Set(fd, protoreflect.ValueOfString(id))
	}
	if fd := fields.ByTextName("kind"); fd != nil {
		message.ProtoReflect().Set(fd, protoreflect.ValueOfString(kind))
	}

	return message, err
}

func TestEmptyArtifactPatches(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "empty directory",
			path: "testdata/empty",
		},
		{
			name: "unrecognized yaml",
			path: "testdata/sample-hierarchical/apis/registry/versions/v1/specs/openapi/openapi.yaml",
		},
	}
	ctx := context.Background()
	registryClient, adminClient := grpctest.SetupRegistry(ctx, t, "patch-empty-test", []seeder.RegistryResource{
		&rpc.Project{
			Name: "projects/patch-empty-test",
		},
	})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := Apply(ctx, registryClient, adminClient, nil, "projects/patch-empty-test/locations/global", true, 10, test.path); err == nil {
				t.Errorf("Apply() succeeded and should have failed")
			}
		})
	}
}

func TestDeploymentImports(t *testing.T) {
	tests := []struct {
		desc string
		root string
	}{
		{
			desc: "sample-nested",
			root: "testdata/deployments-nested",
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			project := names.Project{ProjectID: "patch-deployments-test"}

			registryClient, adminClient := grpctest.SetupRegistry(ctx, t, "patch-empty-test", []seeder.RegistryResource{
				&rpc.Project{
					Name: project.String(),
				},
			})

			// set the configured registry.project to the test project
			config, err := connection.ActiveConfig()
			if err != nil {
				t.Fatalf("Setup: Failed to get registry configuration: %s", err)
			}
			config.Project = project.ProjectID
			connection.SetConfig(config)

			// apply the api and deployments
			if err := Apply(ctx, registryClient, adminClient, nil, project.String()+"/locations/global", true, 10, test.root); err != nil {
				t.Fatalf("Apply() returned error: %s", err)
			}

			// verify that all the spec references are to specific revisions
			it := registryClient.ListApiDeployments(ctx, &rpc.ListApiDeploymentsRequest{
				Parent: project.Api("registry").String(),
			})
			for d, err := it.Next(); err != iterator.Done; d, err = it.Next() {
				specName, err := names.ParseSpecRevision(d.ApiSpecRevision)
				if err != nil {
					t.Errorf("failed to parse spec name %s", d.ApiSpecRevision)
				}
				if specName.RevisionID == "" {
					t.Errorf("spec revision ID should not be empty: %s", d.ApiSpecRevision)
				}
			}
		})
	}
}
