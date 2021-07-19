package breakingChangeDetector

import (
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestChanges(t *testing.T) {
	tests := []struct {
		desc         string
		diffProto 	 *rpc.Diff
		wantProto    *rpc.Changes
	}{
		{
		desc: "basic Addition Breaking Test",
		diffProto: &rpc.Diff{
			Additions: []string{"components.schemas.x.required.x"},
		},
		wantProto: &rpc.Changes{
			BreakingChanges: &rpc.Diff{
			Additions: []string{"components.schemas.x.required.x"},
			},
			NonbreakingChanges: &rpc.Diff{
				},
			UnknownChanges: &rpc.Diff{
				},
			},
		},
		{
		desc: "basic Deletion Breaking Test",
		diffProto: &rpc.Diff{
				Deletions: []string{"components.schemas.x.x"},
			},
			wantProto: &rpc.Changes{
				BreakingChanges: &rpc.Diff{
				Deletions: []string{"components.schemas.x.x"},
				},
				NonbreakingChanges: &rpc.Diff{
					},
				UnknownChanges: &rpc.Diff{
					},
				},
			},
			{
				desc: "basic Modification Breaking Test",
				diffProto: &rpc.Diff{
						Modifications: map[string]*rpc.Diff_ValueChange{
							"components.schemas.x.properties.type":{
								To:"float",
								From:"int64",
							},
						},
					},
					wantProto: &rpc.Changes{
						BreakingChanges: &rpc.Diff{
							Modifications: map[string]*rpc.Diff_ValueChange{
								"components.schemas.x.properties.type":{
									To:"float",
									From:"int64",
								},
							},
						},
						NonbreakingChanges: &rpc.Diff{
							},
						UnknownChanges: &rpc.Diff{
							},
						},
					},

	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotProto := GetBreakingChanges(test.diffProto)
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
