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

package export

import (
	"fmt"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/server/registry/names"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var jobs int
	var root string
	cmd := &cobra.Command{
		Use:   "export PATTERN",
		Short: "Export resources from the API Registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}
			name := c.FQName(args[0])
			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			taskQueue, wait := core.WorkerPool(ctx, jobs)
			defer wait()
			if project, err := names.ParseProject(name); err == nil {
				return patch.ExportProject(ctx, client, project, root, taskQueue)
			} else if api, err := names.ParseApi(name); err == nil {
				return patch.ExportAPI(ctx, client, api, root, taskQueue)
			} else {
				return fmt.Errorf("unsupported pattern %+s", args[0])
			}
		},
	}
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "Number of file exports to perform simultaneously")
	cmd.Flags().StringVar(&root, "root", "", "Root directory for export")
	return cmd
}
