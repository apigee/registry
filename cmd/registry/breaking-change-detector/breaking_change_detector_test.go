package breakingChangeDetector

import (
	"io/ioutil"
	"testing"

	"github.com/apigee/registry/cmd/registry/diff"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

// Give a Diff Directly
func TestBreakingChange(t *testing.T) {
	tests := []struct {
		desc         string
		baseSpec     string
		revisionSpec string
		wantProto    *rpc.Changes
	}{
		{
		desc: "basic Test",
		baseSpec: "./test-specs/base-test.yaml",
		revisionSpec: "./test-specs/struct-test-add.yaml",
		wantProto: &rpc.Changes{},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			baseYaml, err := ioutil.ReadFile(test.baseSpec)
			if err != nil {
				t.Fatalf("Failed to get base spec yaml: %s", err)
			}
			revisionYaml, err := ioutil.ReadFile(test.revisionSpec)
			if err != nil {
				t.Fatalf("Failed to get revision spec yaml: %s", err)
			}
			diffProto, err := diff.GetDiff(baseYaml, revisionYaml)
			t.Logf("DiffProto: +%v \n", diffProto)
			if err != nil {
				t.Fatalf("Failed to get diff.Diff: %s", err)
			}
			gotProto := GetBreakingChanges(diffProto)
			opts := cmp.Options{
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}
			//t.Fail()
			if !cmp.Equal(test.wantProto, gotProto, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantProto, gotProto, opts))
			}
		})
	}
}
