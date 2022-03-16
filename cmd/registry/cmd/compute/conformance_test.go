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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/upload"
	"github.com/apigee/registry/connection"
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
	fileBytes, _ := ioutil.ReadFile(filename)
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
		//Tests the normal use case with one guideline defined with status: ACTIVE and one Rule defined with severity:ERROR
		{
			desc:            "normal case",
			conformancePath: filepath.Join("testdata", "styleguide.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest",
			wantProto: &rpc.ConformanceReport{
				Name:           "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest",
				StyleguideName: "openapitest",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{Status: rpc.Guideline_STATUS_UNSPECIFIED},
					{Status: rpc.Guideline_PROPOSED},
					{
						Status: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "norefsiblings",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
					{Status: rpc.Guideline_DEPRECATED},
					{Status: rpc.Guideline_DISABLED},
				},
			},
		},
		//Tests if default status and severity values are assigned properly in the absence of defined values
		{
			desc:            "default case",
			conformancePath: filepath.Join("testdata", "styleguide-default.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-default",
			wantProto: &rpc.ConformanceReport{
				Name:           "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-default",
				StyleguideName: "openapitest-default",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{
						Status: rpc.Guideline_STATUS_UNSPECIFIED,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{
										Severity: rpc.Rule_SEVERITY_UNSPECIFIED,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "norefsiblings",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
					{Status: rpc.Guideline_PROPOSED},
					{Status: rpc.Guideline_ACTIVE},
					{Status: rpc.Guideline_DEPRECATED},
					{Status: rpc.Guideline_DISABLED},
				},
			},
		},
		//Tests if multiple severity levels are populated correctly in severity report
		{
			desc:            "multiple severity",
			conformancePath: filepath.Join("testdata", "styleguide-multiple-severity.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-severity",
			wantProto: &rpc.ConformanceReport{
				Name:           "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-severity",
				StyleguideName: "openapitest-multiple-severity",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{Status: rpc.Guideline_STATUS_UNSPECIFIED},
					{Status: rpc.Guideline_PROPOSED},
					{
						Status: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "tagproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "operationtags",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
											{
												RuleId:   "operationtagdefined",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{
										Severity: rpc.Rule_INFO,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "openapitags",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
											{
												RuleId:   "openapitagsalphabetical",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
										},
									},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{Status: rpc.Guideline_DEPRECATED},
					{Status: rpc.Guideline_DISABLED},
				},
			},
		},
		//Tests if multiple status entries are populated correctly in severity report
		{
			desc:            "multiple status",
			conformancePath: filepath.Join("testdata", "styleguide-multiple-status.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-status",
			wantProto: &rpc.ConformanceReport{
				Name:           "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-status",
				StyleguideName: "openapitest-multiple-status",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{Status: rpc.Guideline_STATUS_UNSPECIFIED},
					{
						Status: rpc.Guideline_PROPOSED,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "tagproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "operationtags",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
										},
									},
									{Severity: rpc.Rule_WARNING},
									{
										Severity: rpc.Rule_INFO,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "openapitags",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
										},
									},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{
						Status: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "norefsiblings",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
					{Status: rpc.Guideline_DEPRECATED},
					{Status: rpc.Guideline_DISABLED},
				},
			},
		},
		//Tests a guideline which defines rules from multiple linters
		{
			desc:            "multiple linter",
			conformancePath: filepath.Join("testdata", "styleguide-multiple-linter.yaml"),
			getPattern:      "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-linter",
			wantProto: &rpc.ConformanceReport{
				Name:           "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-linter",
				StyleguideName: "openapitest-multiple-linter",
				GuidelineReportGroups: []*rpc.GuidelineReportGroup{
					{Status: rpc.Guideline_STATUS_UNSPECIFIED},
					{
						Status: rpc.Guideline_PROPOSED,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "descriptionproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "operationdescription",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
											{
												RuleId:   "infodescription",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
										},
									},
									{
										Severity: rpc.Rule_WARNING,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "descriptiontags",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
										},
									},
									{
										Severity: rpc.Rule_INFO,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "tagdescription",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
											},
										},
									},
									{Severity: rpc.Rule_HINT},
								},
							},
						},
					},
					{
						Status: rpc.Guideline_ACTIVE,
						GuidelineReports: []*rpc.GuidelineReport{
							{
								GuidelineId: "refproperties",
								RuleReportGroups: []*rpc.RuleReportGroup{
									{Severity: rpc.Rule_SEVERITY_UNSPECIFIED},
									{
										Severity: rpc.Rule_ERROR,
										RuleReports: []*rpc.RuleReport{
											{
												RuleId:   "norefsiblings",
												SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
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
					{Status: rpc.Guideline_DEPRECATED},
					{Status: rpc.Guideline_DISABLED},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			client, err := connection.NewClient(ctx)
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
			uploadCmd := upload.Command(ctx)
			uploadCmd.SetArgs(args)
			if err = uploadCmd.Execute(); err != nil {
				t.Fatalf("Failed to upload the styleguide: %s", err)
			}

			conformanceCmd := conformanceCommand(ctx)

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

			gotProto := &rpc.ConformanceReport{}
			err = proto.Unmarshal(contents.GetData(), gotProto)
			if err != nil {
				t.Fatalf("Failed to unmarshal artifact: %s", err)
			}
			opts := cmp.Options{
				protocmp.IgnoreFields(&rpc.RuleReport{}, "file_name", "suggestion", "location"),
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
