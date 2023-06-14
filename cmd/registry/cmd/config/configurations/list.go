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

package configurations

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"text/tabwriter"

	"github.com/apigee/registry/pkg/config"
	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Lists existing named configurations",
		Example: `registry config configurations list`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			configs, err := config.Configurations()
			if errors.Is(err, fs.ErrNotExist) || len(configs) == 0 {
				cmd.Println("You don't have any configurations. Run 'registry config configurations create' to create a configuration.")
				return nil
			} else if err != nil {
				return fmt.Errorf("cannot read configs: %v", err)
			}

			activeName, err := config.ActiveName()
			if err != nil {
				return fmt.Errorf("cannot read active config: %v", err)
			}

			names := make([]string, 0, len(configs))
			for n := range configs {
				names = append(names, n)
			}
			sort.Strings(names)

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			defer w.Flush()
			fmt.Fprintln(w, "NAME\tIS_ACTIVE\tADDRESS\tINSECURE")
			for _, name := range names {
				c := configs[name].Registry
				fmt.Fprintf(w, "%s\t%t\t%s\t%t\n", name, name == activeName, c.Address, c.Insecure)
			}
			return nil
		},
	}
	return cmd
}
