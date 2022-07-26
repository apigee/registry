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

package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

func setCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set PROPERTY VALUE",
		Short: "Set the value of a property.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			target, config, err := targetConfig()
			if err != nil {
				return fmt.Errorf("Cannot read config: %v", err)
			}

			name, value := args[0], args[1]
			if !contains(config.Properties(), name) {
				return fmt.Errorf("Config has no property %q.", name)
			}

			err = config.FromMap(map[string]interface{}{
				name: value,
			})
			if err != nil {
				return fmt.Errorf("Cannot set value: %v", err)
			}

			if err = config.Write(target); err != nil {
				return fmt.Errorf("Cannot write config: %v", err)
			}

			fmt.Printf("Updated property %q.\n", name)
			return nil
		},
	}
	return cmd
}

func contains(arr []string, x string) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}
