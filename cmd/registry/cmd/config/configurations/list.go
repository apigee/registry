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

package configurations

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists existing named configurations.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			ctx := cmd.Context()
			logger := log.FromContext(ctx)

			allConfigs, err := connection.AllConfigs()
			if err != nil {
				logger.Fatalf("Cannot read configurations: %v", err)
			}

			activeName, err := connection.ActiveConfigName()
			if err != nil {
				logger.Fatalf("Cannot read active config: %v", err)
			}

			sortedNames := make([]string, 0, len(allConfigs))
			for n := range allConfigs {
				sortedNames = append(sortedNames, n)
			}
			sort.Strings(sortedNames)

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tIS_ACTIVE\tADDRESS\tINSECURE")
			for _, name := range sortedNames {
				config := allConfigs[name]
				fmt.Fprintf(w, "%s\t%t\t%s\t%t\n", name, name == activeName, config.Address, config.Insecure)
			}
			w.Flush()
		},
	}
	return cmd
}
