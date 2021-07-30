package metrics

import (
	"github.com/apigee/registry/rpc"
)
func ComputeStability(calculatedDiffs []*rpc.ClassifiedChanges)*rpc.Stability{
	stats := calculateStats(calculatedDiffs)
	metrics := caluclateMetrics(stats)
}

func calculateStats(calculatedDiffs []*rpc.ClassifiedChanges)*rpc.Stats{
	numberOfDiffs := len(calculatedDiffs)
	totalChanges := 0
	totalBreakingChanges := 0
	totalNonbreakingChanges := 0
	for _, caluclatedDiff := range calculatedDiffs {
		breakingAdditions := caluclatedDiff.BreakingChanges.Additions
		breakingDeletions := caluclatedDiff.BreakingChanges.Deletions
		breakingModificaions := caluclatedDiff.BreakingChanges.Modifications

		totalBreakingChanges = (totalBreakingChanges + len(breakingAdditions) +
			len(breakingDeletions) + len(breakingModificaions))

		nonBreakingAdditions := caluclatedDiff.BreakingChanges.Additions
		nonBreakingDeletions := caluclatedDiff.BreakingChanges.Deletions
		nonBreakingModificaions := caluclatedDiff.BreakingChanges.Modifications

		unknownAdditions := caluclatedDiff.UnknownChanges.Additions
		unknownDeletions := caluclatedDiff.UnknownChanges.Deletions
		unknownModificaions := caluclatedDiff.UnknownChanges.Modifications

		totalNonbreakingChanges = (totalNonbreakingChanges + len(nonBreakingAdditions) +
			len(nonBreakingDeletions) + len(nonBreakingModificaions))
		// Default Unknown Changes to Nonbreaking.
		totalNonbreakingChanges = (totalNonbreakingChanges + len(unknownAdditions) +
			len(unknownDeletions) + len(unknownModificaions))

		totalChanges = totalChanges + totalBreakingChanges + totalNonbreakingChanges
	}

	return &rpc.Stats{
		TotalChanges: int64(totalChanges),
		TotalBreakingChanges: int64(totalBreakingChanges),
		TotalNonBreakingChanges: int64(totalNonbreakingChanges),
		NumberOfDiffs: int64(numberOfDiffs),
}
}

func caluclateMetrics(stats *rpc.Stats) *rpc.Metrics{
	breakingChangePercentage := float64(stats.TotalBreakingChanges/stats.TotalChanges)

	breakingChangeRate := float64(stats.TotalBreakingChanges/stats.NumberOfDiffs)
	return &rpc.Metrics{
		Stats: stats,
		BreakingChangePercentage: breakingChangePercentage,
		BreakingChangeRate: breakingChangeRate,
	}
}

func calculateStability(metrics *rpc.Metrics) *rpc.Stability{
	return &rpc.Stability{
		Metrics:metrics,
	}
}

