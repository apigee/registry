package breakingchangedetector

import (
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestChanges(t *testing.T) {
	tests := []struct {
		desc      string
		diffProto *rpc.Diff
		wantProto *rpc.ChangeDetails
	}{
		{
			desc: "Components.Required field Addition Breaking Test",
			diffProto: &rpc.Diff{
				Additions: []string{"components.schemas.x.required.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{
					Additions: []string{"components.schemas.x.required.x"},
				},
				NonBreakingChanges: &rpc.Diff{},
				UnknownChanges:     &rpc.Diff{},
			},
		},
		{
			desc: "Components.Schemas field Deletion Breaking Test",
			diffProto: &rpc.Diff{
				Deletions: []string{"components.schemas.x.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{
					Deletions: []string{"components.schemas.x.x"},
				},
				NonBreakingChanges: &rpc.Diff{},
				UnknownChanges:     &rpc.Diff{},
			},
		},
		{
			desc: "Components.Schema.Type field Modification Breaking Test",
			diffProto: &rpc.Diff{
				Modifications: map[string]*rpc.Diff_ValueChange{
					"components.schemas.x.properties.type": {
						To:   "float",
						From: "int64",
					},
				},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{
					Modifications: map[string]*rpc.Diff_ValueChange{
						"components.schemas.x.properties.type": {
							To:   "float",
							From: "int64",
						},
					},
				},
				NonBreakingChanges: &rpc.Diff{},
				UnknownChanges:     &rpc.Diff{},
			},
		},
		{
			desc: "Info field Addition NonBreaking Test",
			diffProto: &rpc.Diff{
				Additions: []string{"info.x.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{},
				NonBreakingChanges: &rpc.Diff{
					Additions: []string{"info.x.x"},
				},
				UnknownChanges: &rpc.Diff{},
			},
		},
		{
			desc: "Info field Deletion NonBreaking Test",
			diffProto: &rpc.Diff{
				Deletions: []string{"info.x.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{},
				NonBreakingChanges: &rpc.Diff{
					Deletions: []string{"info.x.x"},
				},
				UnknownChanges: &rpc.Diff{},
			},
		},
		{
			desc: "Info field Modification NonBreaking Test",
			diffProto: &rpc.Diff{
				Modifications: map[string]*rpc.Diff_ValueChange{
					"info.x.x.x": {
						To:   "to",
						From: "from",
					},
				},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{},
				NonBreakingChanges: &rpc.Diff{
					Modifications: map[string]*rpc.Diff_ValueChange{
						"info.x.x.x": {
							To:   "to",
							From: "from",
						},
					},
				},
				UnknownChanges: &rpc.Diff{},
			},
		},
		{
			desc: "Components.Schemas field Addition NonBreaking Test",
			diffProto: &rpc.Diff{
				Additions: []string{"components.schemas.x.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{},
				NonBreakingChanges: &rpc.Diff{
					Additions: []string{"components.schemas.x.x"},
				},
				UnknownChanges: &rpc.Diff{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotProto := GetChangeDetails(test.diffProto)
			opts := cmp.Options{
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.wantProto, gotProto, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantProto, gotProto, opts))
			}
		})
	}
}
