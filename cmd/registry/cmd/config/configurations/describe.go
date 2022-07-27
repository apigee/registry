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
	"sort"

	"github.com/apigee/registry/pkg/connection"
	"github.com/spf13/cobra"
)

func describeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe CONFIGURATION_NAME",
		Short: "Describes a named configuration by listing its properties.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			name := args[0]
			s, err := connection.ReadConfig(name)
			if err != nil {
				return fmt.Errorf("Cannot read config %q: %v", name, err)
			}
			settingsMap, err := s.AsMap()
			if err != nil {
				return fmt.Errorf("Cannot decode config %q: %v", name, err)
			}

			activeName, err := connection.ActiveConfigName()
			if err != nil {
				return fmt.Errorf("Cannot read active config %q: %v", name, err)
			}

			sortedNames := make([]string, 0, len(settingsMap))
			for n := range settingsMap {
				if n != "token" {
					sortedNames = append(sortedNames, n)
				}
			}
			sort.Strings(sortedNames)

			cmd.Printf("is_active: %v\n", name == activeName)
			cmd.Printf("name: %v\n", name)
			cmd.Printf("properties:\n")
			for _, name := range sortedNames {
				cmd.Printf("  %s: %v\n", name, settingsMap[name])
			}
			return nil
		},
	}
	return cmd
}
