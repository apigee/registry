// Copyright 2022 Google LLC. All Rights Reserved.
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

package apply

import (
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/connection"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var fileName string
	var project string
	var recursive bool
	var jobs int
	cmd := &cobra.Command{
		Use:   "apply (-f FILENAME | -f -)",
		Short: "Apply an object to the API Registry",
		Long:  "Apply an object to the API Registry by file name or stdin. Resources will be created if they don't exist yet.\n\nMore info and example usage at https://github.com/apigee/registry/wiki/registry-apply",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if project == "" {
				c, err := connection.ActiveConfig()
				if err != nil {
					return err
				}
				project, err = c.ProjectWithLocation()
				if err != nil {
					return fmt.Errorf("%s: please use --project or set registry.project in configuration", err)
				}
			} else if !strings.Contains(project, "/locations/") {
				project += "/locations/global"
			}
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				return err
			}
			if err := core.VerifyLocation(ctx, client, project); err != nil {
				return fmt.Errorf("parent project %q does not exist: %s", project, err)
			}
			return patch.Apply(ctx, client, fileName, project, recursive, jobs)
		},
	}
	cmd.Flags().StringVarP(&fileName, "file", "f", "", "file or directory containing the patch(es) to apply. Use '-' to read from standard input")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringVar(&project, "project", "", "GCP project containing the API registry")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false, "process the directory used in -f, --file recursively")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
}
