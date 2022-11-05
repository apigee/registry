// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compute

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Compute properties of resources in the API Registry",
	}

	cmd.AddCommand(conformanceCommand())
	cmd.AddCommand(complexityCommand())
	cmd.AddCommand(lintCommand())
	cmd.AddCommand(lintStatsCommand())
	cmd.AddCommand(referencesCommand())
	cmd.AddCommand(scoreCommand())
	cmd.AddCommand(scoreCardCommand())
	cmd.AddCommand(vocabularyCommand())

	cmd.PersistentFlags().String("filter", "", "Filter selected resources")
	cmd.PersistentFlags().Bool("dry-run", false, "if set, computation results will only be printed and will not stored in the registry")
	cmd.PersistentFlags().Int("jobs", 10, "Number of actions to perform concurrently")
	return cmd
}
