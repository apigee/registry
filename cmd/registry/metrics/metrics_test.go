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
		desc        string
		diffProtos  []*rpc.ClassifiedChanges
		wantMetrics *rpc.ChangeMetrics
		wantStats   *rpc.ChangeStats
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
			wantMetrics: &rpc.ChangeMetrics{
				BreakingChangePercentage: .25,
				BreakingChangeRate:       1.5,
			},
			wantStats: &rpc.ChangeStats{
				BreakingChangeCount:    3,
				NonbreakingChangeCount: 9,
				DiffCount:              2,
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
			wantMetrics: &rpc.ChangeMetrics{
				BreakingChangePercentage: 0,
				BreakingChangeRate:       0,
			},
			wantStats: &rpc.ChangeStats{
				BreakingChangeCount:    0,
				NonbreakingChangeCount: 6,
				DiffCount:              2,
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
			wantMetrics: &rpc.ChangeMetrics{
				BreakingChangePercentage: 0,
				BreakingChangeRate:       0,
			},
			wantStats: &rpc.ChangeStats{
				BreakingChangeCount:    0,
				NonbreakingChangeCount: 6,
				DiffCount:              2,
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
			wantMetrics: &rpc.ChangeMetrics{
				BreakingChangePercentage: 1,
				BreakingChangeRate:       3,
			},
			wantStats: &rpc.ChangeStats{
				BreakingChangeCount:    6,
				NonbreakingChangeCount: 0,
				DiffCount:              2,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotStats := ComputeStats(test.diffProtos...)
			gotMetrics := ComputeMetrics(gotStats)
			opts := cmp.Options{
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.wantMetrics, gotMetrics, opts) {
				t.Errorf("ComputeMetrics returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantMetrics, gotMetrics, opts))
			}
			if !cmp.Equal(test.wantStats, gotStats, opts) {
				t.Errorf("ComputeStats returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantStats, gotStats, opts))
			}
		})
	}
}
