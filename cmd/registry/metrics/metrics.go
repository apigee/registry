package metrics

import (
	"github.com/apigee/registry/rpc"
)

// ComputeStats will compute the ChangeStats proto for a list of Classified Diffs.
func ComputeStats(diffs ...*rpc.ChangeDetails) *rpc.ChangeStats {
	var breaking int64 = 0
	var nonbreaking int64 = 0
	var unknown int64 = 0
	for _, diff := range diffs {

		breaking += int64(len(diff.BreakingChanges.Additions))
		breaking += int64(len(diff.BreakingChanges.Deletions))
		breaking += int64(len(diff.BreakingChanges.Modifications))

		nonbreaking += int64(len(diff.NonBreakingChanges.Additions))
		nonbreaking += int64(len(diff.NonBreakingChanges.Deletions))
		nonbreaking += int64(len(diff.NonBreakingChanges.Modifications))

		unknown += int64(len(diff.UnknownChanges.Additions))
		unknown += int64(len(diff.UnknownChanges.Deletions))
		unknown += int64(len(diff.UnknownChanges.Modifications))
	}

	return &rpc.ChangeStats{
		BreakingChangeCount: breaking,
		// Default Unknown Changes to Nonbreaking.
		NonbreakingChangeCount: nonbreaking + unknown,
		DiffCount:              int64(len(diffs)),
	}
}

// ComputeMetrics will compute the metrics proto for a list of Classified Diffs.
func ComputeMetrics(stats *rpc.ChangeStats) *rpc.ChangeMetrics {
	breakingChangePercentage := (float64(stats.BreakingChangeCount) /
		float64(stats.BreakingChangeCount+stats.NonbreakingChangeCount))
	breakingChangeRate := float64(stats.BreakingChangeCount) / float64(stats.DiffCount)
	return &rpc.ChangeMetrics{
		BreakingChangePercentage: breakingChangePercentage,
		BreakingChangeRate:       breakingChangeRate,
	}
}
