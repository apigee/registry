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

package conformance

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/apply"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func TestConformance(t *testing.T) {
	tests := []struct {
		desc            string
		conformancePath string
		getPattern      string
		wantProto       *style.ConformanceReport
	}{
		//Tests the normal use case with one guideline defined with state: ACTIVE and one Rule defined with severity:ERROR
		{
			desc:            "normal case",
			conformancePath: filepath.Join("..", "testdata", "styleguide.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/conformance-openapitest",
			wantProto: &style.ConformanceReport{
				Id:         "conformance-openapitest",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest",
				GuidelineReportGroups: []*style.GuidelineReportGroup{
					{State: style.Guideline_STATE_UNSPECIFIED},
					{State: style.Guideline_PROPOSED},
					{
						State: style.Guideline_ACTIVE,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "norefsiblings",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "No $ref siblings",
												Description: "An object exposing a $ref property cannot be further extended with additional properties.",
												DocUri:      "https://meta.stoplight.io/docs/spectral/4dec24461f3af-open-api-rules#no-ref-siblings",
											},
										},
									},
									{Severity: style.Rule_WARNING},
									{Severity: style.Rule_INFO},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{State: style.Guideline_DEPRECATED},
					{State: style.Guideline_DISABLED},
				},
			},
		},
		//Tests if default state and severity values are assigned properly in the absence of defined values
		{
			desc:            "default case",
			conformancePath: filepath.Join("..", "testdata", "styleguide-default.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/conformance-openapitest-default",
			wantProto: &style.ConformanceReport{
				Id:         "conformance-openapitest-default",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest-default",
				GuidelineReportGroups: []*style.GuidelineReportGroup{
					{
						State: style.Guideline_STATE_UNSPECIFIED,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{
										Severity: style.Rule_SEVERITY_UNSPECIFIED,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "norefsiblings",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "No $ref siblings",
												Description: "An object exposing a $ref property cannot be further extended with additional properties.",
											},
										},
									},
									{Severity: style.Rule_ERROR},
									{Severity: style.Rule_WARNING},
									{Severity: style.Rule_INFO},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{State: style.Guideline_PROPOSED},
					{State: style.Guideline_ACTIVE},
					{State: style.Guideline_DEPRECATED},
					{State: style.Guideline_DISABLED},
				},
			},
		},
		//Tests if multiple severity levels are populated correctly in severity report
		{
			desc:            "multiple severity",
			conformancePath: filepath.Join("..", "testdata", "styleguide-multiple-severity.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/conformance-openapitest-multiple-severity",
			wantProto: &style.ConformanceReport{
				Id:         "conformance-openapitest-multiple-severity",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest-multiple-severity",
				GuidelineReportGroups: []*style.GuidelineReportGroup{
					{State: style.Guideline_STATE_UNSPECIFIED},
					{State: style.Guideline_PROPOSED},
					{
						State: style.Guideline_ACTIVE,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "tagproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "operationtags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "Operation tags",
												Description: "Operation should have non-empty tags array.",
											},
											{
												RuleId:      "operationtagdefined",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "Operation tag defined",
												Description: "Operation tags should be defined in global tags.",
											},
										},
									},
									{Severity: style.Rule_WARNING},
									{
										Severity: style.Rule_INFO,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "openapitags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "OpenAPI tags",
												Description: "OpenAPI object should have non-empty tags array.",
											},
											{
												RuleId:      "openapitagsalphabetical",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "OpenAPI tags alphabetical",
												Description: "OpenAPI object should have alphabetical tags. This will be sorted by the name property.",
											},
										},
									},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{State: style.Guideline_DEPRECATED},
					{State: style.Guideline_DISABLED},
				},
			},
		},
		//Tests if multiple state entries are populated correctly in severity report
		{
			desc:            "multiple state",
			conformancePath: filepath.Join("..", "testdata", "styleguide-multiple-state.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/conformance-openapitest-multiple-state",
			wantProto: &style.ConformanceReport{
				Id:         "conformance-openapitest-multiple-state",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest-multiple-state",
				GuidelineReportGroups: []*style.GuidelineReportGroup{
					{State: style.Guideline_STATE_UNSPECIFIED},
					{
						State: style.Guideline_PROPOSED,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "tagproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "operationtags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "Operation tags",
												Description: "Operation should have non-empty tags array.",
											},
										},
									},
									{Severity: style.Rule_WARNING},
									{
										Severity: style.Rule_INFO,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "openapitags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "OpenAPI tags",
												Description: "OpenAPI object should have non-empty tags array.",
											},
										},
									},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{
						State: style.Guideline_ACTIVE,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "norefsiblings",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "No $ref siblings",
												Description: "An object exposing a $ref property cannot be further extended with additional properties.",
											},
										},
									},
									{Severity: style.Rule_WARNING},
									{Severity: style.Rule_INFO},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{State: style.Guideline_DEPRECATED},
					{State: style.Guideline_DISABLED},
				},
			},
		},
		//Tests a guideline which defines rules from multiple linters
		{
			desc:            "multiple linter",
			conformancePath: filepath.Join("..", "testdata", "styleguide-multiple-linter.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/conformance-openapitest-multiple-linter",
			wantProto: &style.ConformanceReport{
				Id:         "conformance-openapitest-multiple-linter",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest-multiple-linter",
				GuidelineReportGroups: []*style.GuidelineReportGroup{
					{State: style.Guideline_STATE_UNSPECIFIED},
					{
						State: style.Guideline_PROPOSED,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "descriptionproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "operationdescription",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "Operation description",
												Description: "Operation should have non-empty description.",
											},
											{
												RuleId:      "infodescription",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "Info description",
												Description: "OpenAPI object info description must be present and non-empty string.",
											},
										},
									},
									{
										Severity: style.Rule_WARNING,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "descriptiontags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "Description tags",
												Description: "Ensures that description fields in the OpenAPI spec contain no tags (such as HTML tags).",
											},
										},
									},
									{
										Severity: style.Rule_INFO,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "tagdescription",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "Tag description",
												Description: "Tags alone are not very descriptive. Give folks a bit more information to work with.",
											},
										},
									},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{
						State: style.Guideline_ACTIVE,
						GuidelineReports: []*style.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*style.RuleReportGroup{
									{Severity: style.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: style.Rule_ERROR,
										RuleReports: []*style.RuleReport{
											{
												RuleId:      "norefsiblings",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
												DisplayName: "No $ref siblings",
												Description: "An object exposing a $ref property cannot be further extended with additional properties.",
											},
										},
									},
									{Severity: style.Rule_WARNING},
									{Severity: style.Rule_INFO},
									{Severity: style.Rule_HINT},
								},
							},
						},
					},
					{State: style.Guideline_DEPRECATED},
					{State: style.Guideline_DISABLED},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()

			testProject := "conformance-test"
			client, adminClient := grpctest.SetupRegistry(ctx, t, testProject, nil)
			project := rpc.Project{
				Name: "projects/" + testProject,
			}

			// Setup some resources in the project

			// Create API
			api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
				Parent: project.Name + "/locations/global",
				ApiId:  "petstore",
				Api: &rpc.Api{
					DisplayName:  "petstore",
					Description:  "Sample Test API",
					Availability: "GENERAL",
				},
			})
			if err != nil {
				t.Fatalf("Failed CreateApi %s: %s", "petstore", err.Error())
			}

			// Create Versions 1.0.0
			v1, err := client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
				Parent:       api.Name,
				ApiVersionId: "1.0.0",
				ApiVersion:   &rpc.ApiVersion{},
			})
			if err != nil {
				t.Fatalf("Failed CreateVersion 1.0.0: %s", err.Error())
			}

			// Create Spec in each of the versions
			buf, err := readAndGZipFile(t, filepath.Join("..", "testdata", "openapi.yaml"))
			if err != nil {
				t.Fatalf("Failed reading API contents: %s", err.Error())
			}

			req := &rpc.CreateApiSpecRequest{
				Parent:    v1.Name,
				ApiSpecId: "openapi",
				ApiSpec: &rpc.ApiSpec{
					MimeType: "application/x.openapi+gzip;version=3.0.0",
					Contents: buf.Bytes(),
					Filename: "openapi.yaml",
				},
			}

			spec, err := client.CreateApiSpec(ctx, req)
			if err != nil {
				t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
			}

			// Apply the styleguide to the registry
			args := []string{"-f", test.conformancePath, "--parent", "projects/" + testProject + "/locations/global"}
			applyCmd := apply.Command()
			applyCmd.SetArgs(args)
			if err = applyCmd.Execute(); err != nil {
				t.Fatalf("Failed to apply the styleguide: %s", err)
			}

			// setup the command
			conformanceCmd := Command()
			args = []string{spec.Name}
			conformanceCmd.SetArgs(args)

			if err = conformanceCmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}

			contents, err := client.GetArtifactContents(ctx, &rpc.GetArtifactContentsRequest{
				Name: test.getPattern,
			})
			if err != nil {
				t.Fatalf("Failed getting artifact contents %s: %s", test.getPattern, err)
			}

			// Add revision ID in wantProto
			for _, g := range test.wantProto.GuidelineReportGroups {
				for _, gr := range g.GetGuidelineReports() {
					for _, r := range gr.RuleReportGroups {
						for _, report := range r.RuleReports {
							report.Spec = fmt.Sprintf("%s@%s", report.Spec, spec.RevisionId)
						}
					}
				}
			}

			gotProto := &style.ConformanceReport{}
			if err := patch.UnmarshalContents(contents.GetData(), contents.GetContentType(), gotProto); err != nil {
				t.Fatalf("Failed to unmarshal artifact: %s", err)
			}

			opts := cmp.Options{
				protocmp.IgnoreFields(&style.RuleReport{}, "file", "suggestion", "location"),
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.wantProto, gotProto, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantProto, gotProto, opts))
			}

			// Delete the demo project
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
