package metrics

import (
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestMetrics(t *testing.T) {
	tests := []struct {
		desc       string
		diffProtos []*rpc.ClassifiedChanges
		wantProto  *rpc.Metrics
	}{
		{
			desc: "Breaking Change Percentage And Rate Test",
			diffProtos: []*rpc.ClassifiedChanges{
				{
					BreakingChanges: &rpc.Diff{
						Deletions: []string{"breakingChange"},
						Additions: []string{"breakingChange"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"breakingChange": {To: "test", From: "test"},
						},
					},
					NonBreakingChanges: &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change": {To: "test", From: "test"},
						},
					},
					UnknownChanges: &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change": {To: "test", From: "test"},
						},
					},
				},
				{
					BreakingChanges: &rpc.Diff{},
					NonBreakingChanges: &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change": {To: "test", From: "test"},
						},
					},
					UnknownChanges: &rpc.Diff{},
				},
			},
			wantProto: &rpc.Metrics{
				BreakingChangePercentage: .25,
				BreakingChangeRate:       1.5,
				Stats: &rpc.Stats{
					TotalBreakingChanges:    3,
					TotalNonBreakingChanges: 9,
					TotalChanges:            12,
					NumberOfDiffs:           2,
				},
			},
		},
		{
			desc: "NonBreaking Changes Test",
			diffProtos: []*rpc.ClassifiedChanges{
				{
					BreakingChanges: &rpc.Diff{},
					NonBreakingChanges: &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change": {To: "test", From: "test"},
						},
					},
					UnknownChanges: &rpc.Diff{},
				},
				{
					BreakingChanges: &rpc.Diff{},
					NonBreakingChanges: &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change": {To: "test", From: "test"},
						},
					},
					UnknownChanges: &rpc.Diff{},
				},
			},
			wantProto: &rpc.Metrics{
				BreakingChangePercentage: 0,
				BreakingChangeRate:       0,
				Stats: &rpc.Stats{
					TotalBreakingChanges:    0,
					TotalNonBreakingChanges: 6,
					TotalChanges:            6,
					NumberOfDiffs:           2,
				},
			},
		},
		{
			desc: "Unknown Default to NonBreaking Changes Test",
			diffProtos: []*rpc.ClassifiedChanges{
				{
					BreakingChanges:    &rpc.Diff{},
					NonBreakingChanges: &rpc.Diff{},
					UnknownChanges: &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change": {To: "test", From: "test"},
						},
					},
				},
				{
					BreakingChanges:    &rpc.Diff{},
					NonBreakingChanges: &rpc.Diff{},
					UnknownChanges: &rpc.Diff{
						Deletions: []string{"Change"},
						Additions: []string{"Change"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"Change": {To: "test", From: "test"},
						},
					},
				},
			},
			wantProto: &rpc.Metrics{
				BreakingChangePercentage: 0,
				BreakingChangeRate:       0,
				Stats: &rpc.Stats{
					TotalBreakingChanges:    0,
					TotalNonBreakingChanges: 6,
					TotalChanges:            6,
					NumberOfDiffs:           2,
				},
			},
		},
		{
			desc: "Breaking Changes Test",
			diffProtos: []*rpc.ClassifiedChanges{
				{
					BreakingChanges: &rpc.Diff{
						Deletions: []string{"breakingChange"},
						Additions: []string{"breakingChange"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"breakingChange": {To: "test", From: "test"},
						},
					},
					NonBreakingChanges: &rpc.Diff{},
					UnknownChanges:     &rpc.Diff{},
				},
				{
					BreakingChanges: &rpc.Diff{
						Deletions: []string{"breakingChange"},
						Additions: []string{"breakingChange"},
						Modifications: map[string]*rpc.Diff_ValueChange{
							"breakingChange": {To: "test", From: "test"},
						},
					},
					NonBreakingChanges: &rpc.Diff{},
					UnknownChanges:     &rpc.Diff{},
				},
			},
			wantProto: &rpc.Metrics{
				BreakingChangePercentage: 1,
				BreakingChangeRate:       3,
				Stats: &rpc.Stats{
					TotalBreakingChanges:    6,
					TotalNonBreakingChanges: 0,
					TotalChanges:            6,
					NumberOfDiffs:           2,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotProto := ComputeMetrics(test.diffProtos)
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
