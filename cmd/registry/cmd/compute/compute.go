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
	"context"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Compute properties of resources in the API Registry",
	}

	cmd.AddCommand(complexityCommand(ctx))
	cmd.AddCommand(descriptorCommand(ctx))
	cmd.AddCommand(detailsCommand(ctx))
	cmd.AddCommand(indexCommand(ctx))
	cmd.AddCommand(lintCommand(ctx))
	cmd.AddCommand(lintStatsCommand(ctx))
	cmd.AddCommand(referencesCommand(ctx))
	cmd.AddCommand(searchIndexCommand(ctx))
	cmd.AddCommand(vocabularyCommand(ctx))

	cmd.PersistentFlags().String("filter", "", "Filter selected resources")
	cmd.PersistentFlags().String("something", "", "desc")
	return cmd
}
