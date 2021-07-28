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
					&rpc.ClassifiedChanges{
						BreakingChanges: &rpc.Diff{
							Deletions: []string{"components.schemas.x.x"},
						},
						NonBreakingChanges: &rpc.Diff{},
						UnknownChanges:     &rpc.Diff{},
					},
				},
				{
					&rpc.ClassifiedChanges{
						BreakingChanges: &rpc.Diff{
							Deletions: []string{"components.schemas.x.x"},
						},
						NonBreakingChanges: &rpc.Diff{},
						UnknownChanges:     &rpc.Diff{},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			CaluclateMetrics(test.diffProtos)
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
