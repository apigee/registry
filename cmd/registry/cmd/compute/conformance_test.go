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

package compute

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/upload"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

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
		wantProto       *rpc.ConformanceReport
	}{
		//Tests the normal use case with one guideline defined with state: ACTIVE and one Rule defined with severity:ERROR
		{
			desc:            "normal case",
			conformancePath: filepath.Join("testdata", "styleguide.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest",
			wantProto: &rpc.ConformanceReport{
				Id:         "conformance-openapitest",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{State: rpc.Guideline_STATE_UNSPECIFIED},
					{State: rpc.Guideline_PROPOSED},
					{
						State: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "norefsiblings",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "No $ref siblings",
												Description: "An object exposing a $ref property cannot be further extended with additional properties.",
												DocUri:      "https://meta.stoplight.io/docs/spectral/4dec24461f3af-open-api-rules#no-ref-siblings",
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{Severity: rpc.Rule_INFO},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{State: rpc.Guideline_DEPRECATED},
					{State: rpc.Guideline_DISABLED},
				},
			},
		},
		//Tests if default state and severity values are assigned properly in the absence of defined values
		{
			desc:            "default case",
			conformancePath: filepath.Join("testdata", "styleguide-default.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-default",
			wantProto: &rpc.ConformanceReport{
				Id:         "conformance-openapitest-default",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest-default",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{
						State: rpc.Guideline_STATE_UNSPECIFIED,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{
										Severity: rpc.Rule_SEVERITY_UNSPECIFIED,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "norefsiblings",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "No $ref siblings",
												Description: "An object exposing a $ref property cannot be further extended with additional properties.",
											},
										},
									},
									{Severity: rpc.Rule_ERROR},
									{Severity: rpc.Rule_WARNING},
									{Severity: rpc.Rule_INFO},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{State: rpc.Guideline_PROPOSED},
					{State: rpc.Guideline_ACTIVE},
					{State: rpc.Guideline_DEPRECATED},
					{State: rpc.Guideline_DISABLED},
				},
			},
		},
		//Tests if multiple severity levels are populated correctly in severity report
		{
			desc:            "multiple severity",
			conformancePath: filepath.Join("testdata", "styleguide-multiple-severity.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-severity",
			wantProto: &rpc.ConformanceReport{
				Id:         "conformance-openapitest-multiple-severity",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest-multiple-severity",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{State: rpc.Guideline_STATE_UNSPECIFIED},
					{State: rpc.Guideline_PROPOSED},
					{
						State: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "tagproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "operationtags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "Operation tags",
												Description: "Operation should have non-empty tags array.",
											},
											{
												RuleId:      "operationtagdefined",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "Operation tag defined",
												Description: "Operation tags should be defined in global tags.",
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{
										Severity: rpc.Rule_INFO,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "openapitags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "OpenAPI tags",
												Description: "OpenAPI object should have non-empty tags array.",
											},
											{
												RuleId:      "openapitagsalphabetical",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "OpenAPI tags alphabetical",
												Description: "OpenAPI object should have alphabetical tags. This will be sorted by the name property.",
											},
										},
									},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{State: rpc.Guideline_DEPRECATED},
					{State: rpc.Guideline_DISABLED},
				},
			},
		},
		//Tests if multiple state entries are populated correctly in severity report
		{
			desc:            "multiple state",
			conformancePath: filepath.Join("testdata", "styleguide-multiple-state.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-state",
			wantProto: &rpc.ConformanceReport{
				Id:         "conformance-openapitest-multiple-state",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest-multiple-state",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{State: rpc.Guideline_STATE_UNSPECIFIED},
					{
						State: rpc.Guideline_PROPOSED,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "tagproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "operationtags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "Operation tags",
												Description: "Operation should have non-empty tags array.",
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{
										Severity: rpc.Rule_INFO,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "openapitags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "OpenAPI tags",
												Description: "OpenAPI object should have non-empty tags array.",
											},
										},
									},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{
						State: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "norefsiblings",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "No $ref siblings",
												Description: "An object exposing a $ref property cannot be further extended with additional properties.",
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{Severity: rpc.Rule_INFO},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{State: rpc.Guideline_DEPRECATED},
					{State: rpc.Guideline_DISABLED},
				},
			},
		},
		//Tests a guideline which defines rules from multiple linters
		{
			desc:            "multiple linter",
			conformancePath: filepath.Join("testdata", "styleguide-multiple-linter.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-linter",
			wantProto: &rpc.ConformanceReport{
				Id:         "conformance-openapitest-multiple-linter",
				Kind:       "ConformanceReport",
				Styleguide: "projects/conformance-test/locations/global/artifacts/openapitest-multiple-linter",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{State: rpc.Guideline_STATE_UNSPECIFIED},
					{
						State: rpc.Guideline_PROPOSED,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "descriptionproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "operationdescription",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "Operation description",
												Description: "Operation should have non-empty description.",
											},
											{
												RuleId:      "infodescription",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "Info description",
												Description: "OpenAPI object info description must be present and non-empty string.",
											},
										},
									},
									{
										Severity: rpc.Rule_WARNING,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "descriptiontags",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "Description tags",
												Description: "Ensures that description fields in the OpenAPI spec contain no tags (such as HTML tags).",
											},
										},
									},
									{
										Severity: rpc.Rule_INFO,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "tagdescription",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "Tag description",
												Description: "Tags alone are not very descriptive. Give folks a bit more information to work with.",
											},
										},
									},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{
						State: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:      "norefsiblings",
												Spec:        "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
												DisplayName: "No $ref siblings",
												Description: "An object exposing a $ref property cannot be further extended with additional properties.",
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{Severity: rpc.Rule_INFO},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{State: rpc.Guideline_DEPRECATED},
					{State: rpc.Guideline_DISABLED},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}

			testProject := "conformance-test"

			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + testProject,
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}

			project, err := adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
				ProjectId: testProject,
				Project: &rpc.Project{
					DisplayName: "Demo",
					Description: "A demo catalog",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create project %s: %s", testProject, err.Error())
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
			buf, err := readAndGZipFile(t, filepath.Join("testdata", "openapi.yaml"))
			if err != nil {
				t.Fatalf("Failed reading API contents: %s", err.Error())
			}

			req := &rpc.CreateApiSpecRequest{
				Parent:    v1.Name,
				ApiSpecId: "openapi.yaml",
				ApiSpec: &rpc.ApiSpec{
					MimeType: "application/x.openapi+gzip;version=3.0.0",
					Contents: buf.Bytes(),
				},
			}

			spec, err := client.CreateApiSpec(ctx, req)
			if err != nil {
				t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
			}

			// Upload the styleguide to registry
			args := []string{"styleguide", test.conformancePath, "--project-id=" + testProject}
			uploadCmd := upload.Command()
			uploadCmd.SetArgs(args)
			if err = uploadCmd.Execute(); err != nil {
				t.Fatalf("Failed to upload the styleguide: %s", err)
			}

			// setup the command
			conformanceCmd := Command()
			args = []string{"conformance", spec.Name}
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

			gotProto := &rpc.ConformanceReport{}
			if err := proto.Unmarshal(contents.GetData(), gotProto); err != nil {
				t.Fatalf("Failed to unmarshal artifact: %s", err)
			}

			opts := cmp.Options{
				protocmp.IgnoreFields(&rpc.RuleReport{}, "file", "suggestion", "location"),
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
