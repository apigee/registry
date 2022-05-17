// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
