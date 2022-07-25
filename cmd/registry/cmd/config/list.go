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

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List properties for the currently active configuration.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			target, config, err := targetConfig()
			if err != nil {
				return fmt.Errorf("Cannot read config: %v", err)
			}

			m, err := config.AsMap()
			if err != nil {
				return fmt.Errorf("Cannot decode config: %v", err)
			}
			for _, p := range config.Properties() {
				if sv := fmt.Sprintf("%v", m[p]); sv != "" {
					fmt.Println(p, "=", sv)
				}
			}

			fmt.Printf("\nYour active configuration is: %q.\n", target)
			return nil
		},
	}
	return cmd
}
