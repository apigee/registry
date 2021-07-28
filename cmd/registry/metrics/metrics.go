package metrics

import (
	"fmt"

	"github.com/apigee/registry/rpc"
)

func caluclateMetrics(calculatedDiffs []*rpc.ClassifiedChanges) {
	numberOfDiffs := len(calculatedDiffs)
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
	}
	fmt.Printf("Breaking Changes Introduced PerDiff %v", (totalBreakingChanges / numberOfDiffs))
}
