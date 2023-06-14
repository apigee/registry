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
	"fmt"
	"sort"

	"github.com/apigee/registry/pkg/config"
	"github.com/spf13/cobra"
)

func describeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "describe CONFIGURATION",
		Short:   "Describes a named configuration by listing its properties",
		Example: `registry config configurations describe localhost`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			s, err := config.Read(name)
			if err != nil {
				return fmt.Errorf("cannot read config %q: %v", name, err)
			}
			settingsMap, err := s.FlatMap()
			if err != nil {
				return fmt.Errorf("cannot decode config %q: %v", name, err)
			}

			activeName, err := config.ActiveName()
			if err != nil {
				return fmt.Errorf("cannot read active config %q: %v", name, err)
			}

			sortedNames := make([]string, 0, len(settingsMap))
			for n := range settingsMap {
				if n != "registry.token" {
					sortedNames = append(sortedNames, n)
				}
			}
			sort.Strings(sortedNames)

			cmd.Printf("is_active: %v\n", name == activeName)
			cmd.Printf("name: %v\n", name)
			cmd.Printf("properties:\n")
			for _, name := range sortedNames {
				if settingsMap[name] != "" {
					cmd.Printf("  %s: %v\n", name, settingsMap[name])
				}
			}
			return nil
		},
	}
	return cmd
}
