package conformance

import (
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

var rules = []*rpc.Rule{
	{
		Id:             "norefsiblings",
		Linter:         "sample",
		LinterRulename: "no-$ref-siblings",
		Severity:       rpc.Rule_ERROR,
	},
	{
		Id:             "norefcycles",
		Linter:         "spectral",
		LinterRulename: "no-ref-cycles",
		Severity:       rpc.Rule_ERROR,
	},
	{
		Id:             "operationdescription",
		Linter:         "spectral",
		LinterRulename: "operation-description",
		Severity:       rpc.Rule_ERROR,
	},
}

func TestGenerateLinterMetadata(t *testing.T) {
	tests := []struct {
		desc         string
		styleguide   *rpc.StyleGuide
		wantMetadata map[string]*linterMetadata
		wantErr      bool
	}{
		{
			desc: "Normal case",
			styleguide: &rpc.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*rpc.Guideline{
					{
						Id:     "refproperties",
						Rules:  []*rpc.Rule{rules[0]},
						Status: rpc.Guideline_ACTIVE,
					},
				},
			},
			wantMetadata: map[string]*linterMetadata{
				"sample": {
					name:  "sample",
					rules: []string{rules[0].GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						rules[0].GetLinterRulename(): {
							guidelineRule: rules[0],
							guideline: &rpc.Guideline{
								Id:     "refproperties",
								Rules:  []*rpc.Rule{rules[0]},
								Status: rpc.Guideline_ACTIVE,
							},
						},
					},
				},
			},
		},
		{
			desc: "Multiple linters",
			styleguide: &rpc.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*rpc.Guideline{
					{
						Id:     "descriptionproperties",
						Rules:  []*rpc.Rule{rules[0], rules[1]},
						Status: rpc.Guideline_ACTIVE,
					},
				},
			},
			wantMetadata: map[string]*linterMetadata{
				"sample": {
					name:  "sample",
					rules: []string{rules[0].GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						rules[0].GetLinterRulename(): {
							guidelineRule: rules[0],
							guideline: &rpc.Guideline{
								Id:     "descriptionproperties",
								Rules:  []*rpc.Rule{rules[0], rules[1]},
								Status: rpc.Guideline_ACTIVE,
							},
						},
					},
				},
				"spectral": {
					name:  "spectral",
					rules: []string{rules[1].GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						rules[1].GetLinterRulename(): {
							guidelineRule: rules[1],
							guideline: &rpc.Guideline{
								Id:     "descriptionproperties",
								Rules:  []*rpc.Rule{rules[0], rules[1]},
								Status: rpc.Guideline_ACTIVE,
							},
						},
					},
				},
			},
		},
		{
			desc: "Multiple linters guidelines",
			styleguide: &rpc.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*rpc.Guideline{
					{
						Id:     "refproperties",
						Rules:  []*rpc.Rule{rules[0], rules[1]},
						Status: rpc.Guideline_ACTIVE,
					},
					{
						Id:     "descriptionproperties",
						Rules:  []*rpc.Rule{rules[2]},
						Status: rpc.Guideline_PROPOSED,
					},
				},
			},
			wantMetadata: map[string]*linterMetadata{
				"sample": {
					name:  "sample",
					rules: []string{rules[0].GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						rules[0].GetLinterRulename(): {
							guidelineRule: rules[0],
							guideline: &rpc.Guideline{
								Id:     "refproperties",
								Rules:  []*rpc.Rule{rules[0], rules[1]},
								Status: rpc.Guideline_ACTIVE,
							},
						},
					},
				},
				"spectral": {
					name:  "spectral",
					rules: []string{rules[1].GetLinterRulename(), rules[2].GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						rules[1].GetLinterRulename(): {
							guidelineRule: rules[1],
							guideline: &rpc.Guideline{
								Id:     "refproperties",
								Rules:  []*rpc.Rule{rules[0], rules[1]},
								Status: rpc.Guideline_ACTIVE,
							},
						},
						rules[2].GetLinterRulename(): {
							guidelineRule: rules[2],
							guideline: &rpc.Guideline{
								Id:     "descriptionproperties",
								Rules:  []*rpc.Rule{rules[2]},
								Status: rpc.Guideline_PROPOSED,
							},
						},
					},
				},
			},
		},
		{
			desc: "Empty metadata",
			styleguide: &rpc.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*rpc.Guideline{
					{
						Id:     "refproperties",
						Rules:  []*rpc.Rule{},
						Status: rpc.Guideline_ACTIVE,
					},
				},
			},
			wantMetadata: nil,
			wantErr:      true,
		},
		{
			desc: "Missing linterName",
			styleguide: &rpc.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*rpc.Guideline{
					{
						Id: "refproperties",
						Rules: []*rpc.Rule{
							{
								Id:             "norefsiblings",
								LinterRulename: "no-$ref-siblings",
								Severity:       rpc.Rule_ERROR,
							},
						},
						Status: rpc.Guideline_ACTIVE,
					},
				},
			},
			wantMetadata: nil,
			wantErr:      true,
		},
		{
			desc: "Missing linterRuleName",
			styleguide: &rpc.StyleGuide{
				Id:        "openapitest",
				MimeTypes: []string{"application/x.openapi+gzip;version=3.0.0"},
				Guidelines: []*rpc.Guideline{
					{
						Id: "refproperties",
						Rules: []*rpc.Rule{
							{
								Id:       "norefsiblings",
								Linter:   "spectral",
								Severity: rpc.Rule_ERROR,
							},
							rules[1],
						},
						Status: rpc.Guideline_ACTIVE,
					},
				},
			},
			wantMetadata: map[string]*linterMetadata{
				"spectral": {
					name:  "spectral",
					rules: []string{rules[1].GetLinterRulename()},
					rulesMetadata: map[string]*ruleMetadata{
						rules[1].GetLinterRulename(): {
							guidelineRule: rules[1],
							guideline: &rpc.Guideline{
								Id: "refproperties",
								Rules: []*rpc.Rule{
									{
										Id:       "norefsiblings",
										Linter:   "spectral",
										Severity: rpc.Rule_ERROR,
									},
									rules[1],
								},
								Status: rpc.Guideline_ACTIVE,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {

			gotMetadata, err := GenerateLinterMetadata(test.styleguide)

			if test.wantErr {
				if err == nil {
					t.Errorf("Expected GenerateLinterMetadata() to return an error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error from GenerateLinterMetadata(): %s", err)
				}
			}

			opts := cmp.Options{
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
				cmp.AllowUnexported(linterMetadata{}),
				cmp.AllowUnexported(ruleMetadata{}),
				cmpopts.SortMaps(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.wantMetadata, gotMetadata, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantMetadata, gotMetadata, opts))
			}
		})
	}
}
