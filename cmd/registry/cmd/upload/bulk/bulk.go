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

package bulk

import (
	"context"

	"github.com/spf13/cobra"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk-upload API specs of selected styles",
	}

	cmd.AddCommand(discoveryCommand(ctx))
	cmd.AddCommand(openAPICommand(ctx))
	cmd.AddCommand(protosCommand(ctx))

	cmd.PersistentFlags().String("project-id", "", "Project ID to use for each upload")
	_ = cmd.MarkFlagRequired("project-id")

	cmd.PersistentFlags().Int("jobs", 10, "Number of upload jobs to run simultaneously")
	return cmd
}
