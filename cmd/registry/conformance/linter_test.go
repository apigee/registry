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
	"testing"

	"github.com/apigee/registry/pkg/artifacts"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

var noRefSiblingsRule = &artifacts.Rule{
	Id:             "norefsiblings",
	Linter:         "sample",
	LinterRulename: "no-$ref-siblings",
	Severity:       artifacts.Rule_ERROR,
}

var noRefCyclesRule = &artifacts.Rule{
	Id:             "norefcycles",
	Linter:         "spectral",
	LinterRulename: "no-ref-cycles",
	Severity:       artifacts.Rule_ERROR,
}

var operationDescriptionRule = &artifacts.Rule{
	Id:             "operationdescription",
	Linter:         "spectral",
	LinterRulename: "operation-description",
	Severity:       artifacts.Rule_ERROR,
}

func TestGenerateLinterMetadata(t *testing.T) {
	tests := []struct {
		desc       string
		styleguide *artifacts.StyleGuide
		want       map[string]*linterMetadata
		wantErr    bool
	}{
		{
			desc: "Normal case",
			styleguide: &artifacts.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*artifacts.Guideline{
					{
						Id:    "refproperties",
						Rules: []*artifacts.Rule{noRefSiblingsRule},
						State: artifacts.Guideline_ACTIVE,
					},
				},
			},
			want: map[string]*linterMetadata{
				"sample": {
					name:  "sample",
					rules: []string{noRefSiblingsRule.GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						noRefSiblingsRule.GetLinterRulename(): {
							guidelineRule: noRefSiblingsRule,
							guideline: &artifacts.Guideline{
								Id:    "refproperties",
								Rules: []*artifacts.Rule{noRefSiblingsRule},
								State: artifacts.Guideline_ACTIVE,
							},
						},
					},
				},
			},
		},
		{
			desc: "Multiple linters",
			styleguide: &artifacts.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*artifacts.Guideline{
					{
						Id:    "descriptionproperties",
						Rules: []*artifacts.Rule{noRefSiblingsRule, noRefCyclesRule},
						State: artifacts.Guideline_ACTIVE,
					},
				},
			},
			want: map[string]*linterMetadata{
				"sample": {
					name:  "sample",
					rules: []string{noRefSiblingsRule.GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						noRefSiblingsRule.GetLinterRulename(): {
							guidelineRule: noRefSiblingsRule,
							guideline: &artifacts.Guideline{
								Id:    "descriptionproperties",
								Rules: []*artifacts.Rule{noRefSiblingsRule, noRefCyclesRule},
								State: artifacts.Guideline_ACTIVE,
							},
						},
					},
				},
				"spectral": {
					name:  "spectral",
					rules: []string{noRefCyclesRule.GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						noRefCyclesRule.GetLinterRulename(): {
							guidelineRule: noRefCyclesRule,
							guideline: &artifacts.Guideline{
								Id:    "descriptionproperties",
								Rules: []*artifacts.Rule{noRefSiblingsRule, noRefCyclesRule},
								State: artifacts.Guideline_ACTIVE,
							},
						},
					},
				},
			},
		},
		{
			desc: "Multiple linters guidelines",
			styleguide: &artifacts.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*artifacts.Guideline{
					{
						Id:    "refproperties",
						Rules: []*artifacts.Rule{noRefSiblingsRule, noRefCyclesRule},
						State: artifacts.Guideline_ACTIVE,
					},
					{
						Id:    "descriptionproperties",
						Rules: []*artifacts.Rule{operationDescriptionRule},
						State: artifacts.Guideline_PROPOSED,
					},
				},
			},
			want: map[string]*linterMetadata{
				"sample": {
					name:  "sample",
					rules: []string{noRefSiblingsRule.GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						noRefSiblingsRule.GetLinterRulename(): {
							guidelineRule: noRefSiblingsRule,
							guideline: &artifacts.Guideline{
								Id:    "refproperties",
								Rules: []*artifacts.Rule{noRefSiblingsRule, noRefCyclesRule},
								State: artifacts.Guideline_ACTIVE,
							},
						},
					},
				},
				"spectral": {
					name:  "spectral",
					rules: []string{noRefCyclesRule.GetLinterRulename(), operationDescriptionRule.GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						noRefCyclesRule.GetLinterRulename(): {
							guidelineRule: noRefCyclesRule,
							guideline: &artifacts.Guideline{
								Id:    "refproperties",
								Rules: []*artifacts.Rule{noRefSiblingsRule, noRefCyclesRule},
								State: artifacts.Guideline_ACTIVE,
							},
						},
						operationDescriptionRule.GetLinterRulename(): {
							guidelineRule: operationDescriptionRule,
							guideline: &artifacts.Guideline{
								Id:    "descriptionproperties",
								Rules: []*artifacts.Rule{operationDescriptionRule},
								State: artifacts.Guideline_PROPOSED,
							},
						},
					},
				},
			},
		},
		{
			desc: "Empty metadata",
			styleguide: &artifacts.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*artifacts.Guideline{
					{
						Id:    "refproperties",
						Rules: []*artifacts.Rule{},
						State: artifacts.Guideline_ACTIVE,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			desc: "Missing linterName",
			styleguide: &artifacts.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*artifacts.Guideline{
					{
						Id: "refproperties",
						Rules: []*artifacts.Rule{
							{
								Id:             "norefsiblings",
								LinterRulename: "no-$ref-siblings",
								Severity:       artifacts.Rule_ERROR,
							},
						},
						State: artifacts.Guideline_ACTIVE,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			desc: "Missing linterRuleName",
			styleguide: &artifacts.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*artifacts.Guideline{
					{
						Id: "refproperties",
						Rules: []*artifacts.Rule{
							{
								Id:       "norefsiblings",
								Linter:   "spectral",
								Severity: artifacts.Rule_ERROR,
							},
							noRefCyclesRule,
						},
						State: artifacts.Guideline_ACTIVE,
					},
				},
			},
			want: map[string]*linterMetadata{
				"spectral": {
					name:  "spectral",
					rules: []string{noRefCyclesRule.GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						noRefCyclesRule.GetLinterRulename(): {
							guidelineRule: noRefCyclesRule,
							guideline: &artifacts.Guideline{
								Id: "refproperties",
								Rules: []*artifacts.Rule{
									{
										Id:       "norefsiblings",
										Linter:   "spectral",
										Severity: artifacts.Rule_ERROR,
									},
									noRefCyclesRule,
								},
								State: artifacts.Guideline_ACTIVE,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := GenerateLinterMetadata(test.styleguide)

			if test.wantErr && err == nil {
				t.Fatalf("Expected GenerateLinterMetadata() to return an error")
			} else if !test.wantErr && err != nil {
				t.Fatalf("Unexpected error from GenerateLinterMetadata(): %s", err)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
				cmp.AllowUnexported(linterMetadata{}),
				cmp.AllowUnexported(ruleMetadata{}),
				cmpopts.SortMaps(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.want, got, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.want, got, opts))
			}
		})
	}
}
