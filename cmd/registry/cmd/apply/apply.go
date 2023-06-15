// Copyright 2022 Google LLC.
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

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var files []string
	var project string
	var recursive bool
	var jobs int
	var yamlArchives bool
	cmd := &cobra.Command{
		Use:   "apply (-f FILE | -f -)",
		Short: "Apply YAML to the API Registry",
		Long: `Apply YAML to the API Registry by files / folder names or stdin.

Resources will be created if they don't exist yet. 
Multiple files may be specified by repeating the -f flag.

More info and example usage at https://github.com/apigee/registry/wiki/registry-apply.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(files) > 1 {
				for _, a := range files {
					if a == "-" {
						return fmt.Errorf("-f may include '-' or files, not both")
					}
				}
			}

			ctx := cmd.Context()
			c, err := connection.ActiveConfig()
			if err != nil {
				return err
			}
			if project == "" {
				project, err = c.ProjectWithLocation()
				if err != nil {
					return fmt.Errorf("%s: please use --parent or set registry.project in configuration", err)
				}
			} else if !strings.Contains(project, "/locations/") {
				project += "/locations/global"
			}

			client, err := connection.NewRegistryClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			adminClient, err := connection.NewAdminClientWithSettings(ctx, c)
			if err != nil {
				return err
			}
			if err := visitor.VerifyLocation(ctx, client, project); err != nil {
				return fmt.Errorf("parent project %q does not exist: %s", project, err)
			}
			if yamlArchives { // TODO: remove when default
				ctx = patch.SetStoreArchivesAsYaml(ctx)
			}
			return patch.Apply(ctx, client, adminClient, cmd.InOrStdin(), project, recursive, jobs, files...)
		},
	}
	cmd.Flags().StringSliceVarP(&files, "file", "f", nil, "file or directory containing the patch(es) to apply. Use '-' to read from standard input")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringVar(&project, "parent", "", "GCP project containing the API registry")
	cmd.Flags().BoolVarP(&recursive, "recursive", "R", false, "process the directory used in -f, --file recursively")
	cmd.Flags().BoolVarP(&yamlArchives, "yaml", "y", false, "store the archive data as yaml text instead of binary")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 10, "number of actions to perform concurrently")
	return cmd
}
