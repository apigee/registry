package metrics

import (
	"github.com/apigee/registry/rpc"
)

// ComputeMetrics will compute the metrics proto for a list of Classified Diffs.
func ComputeMetrics(calculatedDiffs []*rpc.ClassifiedChanges) *rpc.Metrics {
	stats := calculateStats(calculatedDiffs)
	return caluclateMetrics(stats)
}

func calculateStats(calculatedDiffs []*rpc.ClassifiedChanges) *rpc.Stats {
	totalBreakingChanges := 0
	totalNonbreakingChanges := 0
	for _, caluclatedDiff := range calculatedDiffs {

		totalBreakingChanges = (totalBreakingChanges +
			len(caluclatedDiff.BreakingChanges.Additions) +
			len(caluclatedDiff.BreakingChanges.Deletions) +
			len(caluclatedDiff.BreakingChanges.Modifications))

		totalNonbreakingChanges = (totalNonbreakingChanges +
			len(caluclatedDiff.NonBreakingChanges.Additions) +
			len(caluclatedDiff.NonBreakingChanges.Deletions) +
			len(caluclatedDiff.NonBreakingChanges.Modifications))

		// Default Unknown Changes to Nonbreaking.
		totalNonbreakingChanges = (totalNonbreakingChanges +
			len(caluclatedDiff.UnknownChanges.Additions) +
			len(caluclatedDiff.UnknownChanges.Deletions) +
			len(caluclatedDiff.UnknownChanges.Modifications))
	}

	return &rpc.Stats{
		TotalChanges:            int64(totalBreakingChanges + totalNonbreakingChanges),
		TotalBreakingChanges:    int64(totalBreakingChanges),
		TotalNonBreakingChanges: int64(totalNonbreakingChanges),
		NumberOfDiffs:           int64(len(calculatedDiffs)),
	}
}

func caluclateMetrics(stats *rpc.Stats) *rpc.Metrics {
	breakingChangePercentage := float64(stats.TotalBreakingChanges) / float64(stats.TotalChanges)
	breakingChangeRate := float64(stats.TotalBreakingChanges) / float64(stats.NumberOfDiffs)
	return &rpc.Metrics{
		Stats:                    stats,
		BreakingChangePercentage: breakingChangePercentage,
		BreakingChangeRate:       breakingChangeRate,
	}
}
