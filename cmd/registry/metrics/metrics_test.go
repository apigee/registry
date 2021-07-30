package metrics

import (
	"testing"

	"github.com/apigee/registry/rpc"
)

func TestMetrics(t *testing.T) {
	tests := []struct {
		desc       string
		diffProtos []*rpc.ClassifiedChanges
	}{
		{
			desc: "Test 1",
			diffProtos: []*rpc.ClassifiedChanges{
				{
					BreakingChanges:    &rpc.Diff{
						Deletions: []string{"breakingChange"},
						Additions: []string{"breakingChange"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"breakingChange":{To:"test", From:"test"},
						},
					},
					NonBreakingChanges: &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change":{To:"test", From:"test"},
						},
					},
					UnknownChanges:     &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change":{To:"test", From:"test"},
						},
					},
				},
				{
					BreakingChanges:    &rpc.Diff{},
					NonBreakingChanges: &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change":{To:"test", From:"test"},
						},
					},
					UnknownChanges:     &rpc.Diff{},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			metrics := CaluclateMetrics(test.diffProtos)
			t.Logf("Metrics: %+v \n",metrics)
			t.Fail()
			/*opts := cmp.Options{
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.wantProto, gotProto, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantProto, gotProto, opts))
			}*/
		})
	}
}
