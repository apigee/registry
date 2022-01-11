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
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/proto"
)

func initReport(t *testing.T) *rpc.ConformanceReport {
	t.Helper()
	return &rpc.ConformanceReport {
		GuidelineReportGroups: []*rpc.GuidelineReportGroup{
			&rpc.GuidelineReportGroup{
				Status: rpc.Guideline_STATUS_UNSPECIFIED,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
			&rpc.GuidelineReportGroup{
				Status: rpc.Guideline_PROPOSED,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
			&rpc.GuidelineReportGroup{
				Status: rpc.Guideline_ACTIVE,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
			&rpc.GuidelineReportGroup{
				Status: rpc.Guideline_DEPRECATED,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
			&rpc.GuidelineReportGroup{
				Status: rpc.Guideline_DISABLED,
				GuidelineReports: []*rpc.GuidelineReport{},
			},
		},
	}
}

func initRuleReportGroups(t *testing.T) []*rpc.RuleReportGroup {
	t.Helper()
	return []*rpc.RuleReportGroup{
		&rpc.RuleReportGroup{
			Severity: rpc.Rule_SEVERITY_UNSPECIFIED,
			RuleReports: []*rpc.RuleReport{},
		},
		&rpc.RuleReportGroup{
			Severity: rpc.Rule_ERROR,
			RuleReports: []*rpc.RuleReport{},
		},
		&rpc.RuleReportGroup{
			Severity: rpc.Rule_WARNING,
			RuleReports: []*rpc.RuleReport{},
		},
		&rpc.RuleReportGroup{
			Severity: rpc.Rule_INFO,
			RuleReports: []*rpc.RuleReport{},
		},
		&rpc.RuleReportGroup{
			Severity: rpc.Rule_HINT,
			RuleReports: []*rpc.RuleReport{},
		},
	}
}

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
		desc         string
		conformancePath string
		getPattern  string
		wantProto         *rpc.ConformanceReport
	}{
		//Tests the normal use case with one guideline defined with status: ACTIVE and one Rule defined with severity:ERROR
		{
			desc:         "normal case",
			conformancePath: filepath.Join("testdata", "styleguide.yaml"),
			getPattern:  "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest",
			wantProto: func() *rpc.ConformanceReport {
				conformance  := initReport(t)
				conformance.Name = "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest"
				conformance.StyleguideName = "openapitest"


				ruleReportGroups := initRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "norefsiblings",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}

				//Populate the expected guideline reports
				guidelineReports := []*rpc.GuidelineReport{
					&rpc.GuidelineReport{
						GuidelineId:  "refproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_ACTIVE].GuidelineReports = guidelineReports

				return conformance
			}(),
		},
		//Tests if default status and severity values are assigned properly in the absence of defined values
		{
			desc:         "default case",
			conformancePath: filepath.Join("testdata", "styleguide-default.yaml"),
			getPattern:  "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-default",
			wantProto: func() *rpc.ConformanceReport {
				conformance  := initReport(t)
				conformance.Name = "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-default"
				conformance.StyleguideName = "openapitest-default"


				ruleReportGroups := initRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_SEVERITY_UNSPECIFIED].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "norefsiblings",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}

				//Populate the expected guideline reports
				guidelineReports := []*rpc.GuidelineReport{
					&rpc.GuidelineReport{
						GuidelineId:  "refproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_STATUS_UNSPECIFIED].GuidelineReports = guidelineReports

				return conformance
			}(),
		},
		//Tests if multiple severity levels are populated correctly in severity report
		{
			desc:         "multiple severity",
			conformancePath: filepath.Join("testdata", "styleguide-multiple-severity.yaml"),
			getPattern:  "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-severity",
			wantProto: func() *rpc.ConformanceReport {
				conformance  := initReport(t)
				conformance.Name = "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-severity"
				conformance.StyleguideName = "openapitest-multiple-severity"


				ruleReportGroups := initRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_INFO].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "openapitags",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
					&rpc.RuleReport{
						RuleId: "openapitagsalphabetical",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "operationtags",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
					&rpc.RuleReport{
						RuleId: "operationtagdefined",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}

				//Populate the expected guideline reports
				guidelineReports := []*rpc.GuidelineReport{
					&rpc.GuidelineReport{
						GuidelineId:  "tagproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_ACTIVE].GuidelineReports = guidelineReports

				return conformance
			}(),
		},
		//Tests if multiple status entries are populated correctly in severity report
		{
			desc:         "multiple status",
			conformancePath: filepath.Join("testdata", "styleguide-multiple-status.yaml"),
			getPattern:  "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-status",
			wantProto: func() *rpc.ConformanceReport {
				conformance  := initReport(t)
				conformance.Name = "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-status"
				conformance.StyleguideName = "openapitest-multiple-status"

				// EXPECTED RULE REPORTS FOR STATUS = "ACTIVE"
				ruleReportGroups := initRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "norefsiblings",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}

				//Populate the expected guideline reports
				guidelineReports := []*rpc.GuidelineReport{
					&rpc.GuidelineReport{
						GuidelineId:  "refproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_ACTIVE].GuidelineReports = guidelineReports

				//EXPECTED RULE REPORTS FOR STATUS = "PROPOSED"
				ruleReportGroups = initRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_INFO].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "openapitags",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "operationtags",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}

				//Populate the expected guideline reports
				guidelineReports = []*rpc.GuidelineReport{
					&rpc.GuidelineReport{
						GuidelineId:  "tagproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_PROPOSED].GuidelineReports = guidelineReports


				return conformance
			}(),
		},
		//Tests a guideline which defines rules from multiple linters
		{
			desc:         "multiple linter",
			conformancePath: filepath.Join("testdata", "styleguide-multiple-linter.yaml"),
			getPattern:  "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-linter",
			wantProto: func() *rpc.ConformanceReport {
				conformance  := initReport(t)
				conformance.Name = "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml/artifacts/conformance-openapitest-multiple-linter"
				conformance.StyleguideName = "openapitest-multiple-linter"

				// EXPECTED RULE REPORTS FOR STATUS = "ACTIVE"
				ruleReportGroups := initRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "norefsiblings",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}

				//Populate the expected guideline reports
				guidelineReports := []*rpc.GuidelineReport{
					&rpc.GuidelineReport{
						GuidelineId:  "refproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_ACTIVE].GuidelineReports = guidelineReports

				//EXPECTED RULE REPORTS FOR STATUS = "PROPOSED"
				ruleReportGroups = initRuleReportGroups(t)

				// Populate the expected severity entry
				ruleReportGroups[rpc.Rule_INFO].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "tagdescription",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}
				ruleReportGroups[rpc.Rule_ERROR].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "operationdescription",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
					&rpc.RuleReport{
						RuleId: "infodescription",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}

				ruleReportGroups[rpc.Rule_WARNING].RuleReports = []*rpc.RuleReport{
					&rpc.RuleReport{
						RuleId: "descriptiontags",
						SpecName: "projects/conformance-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi.yaml",
					},
				}

				//Populate the expected guideline reports
				guidelineReports = []*rpc.GuidelineReport{
					&rpc.GuidelineReport{
						GuidelineId:  "descriptionproperties",
						RuleReportGroups: ruleReportGroups,
					},
				}

				//Populate the expected status entry
				conformance.GuidelineReportGroups[rpc.Guideline_PROPOSED].GuidelineReports = guidelineReports


				return conformance
			}(),
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
				Name: "projects/" + testProject,
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

			// List all the artifacts
			// got := make([]string, 0)
			artifactName, err := names.ParseArtifact(test.getPattern)
			if err != nil {
				t.Fatalf("Invalid artifact pattern: %s", test.getPattern)
			}
			// _ = core.ListArtifacts(ctx, client, artifact, "", false,
			// 	func(artifact *rpc.Artifact) {
			// 		got = append(got, artifact.Name)
			// 	},
			// )

			artifact, err := core.GetArtifact(ctx, client, artifactName, true, nil)
			gotProto :=  &rpc.ConformanceReport{}
			err = proto.Unmarshal(artifact.GetContents(), gotProto)
			if err!= nil {
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
				Name: "projects/" + testProject,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}
		})
	}
}
