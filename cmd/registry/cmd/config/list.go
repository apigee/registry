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

package config

import (
	"fmt"

	"github.com/apigee/registry/pkg/config"
	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List properties in the active configuration",
		Example: `registry config list`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, c, err := config.ActiveRaw()
			if err != nil {
				if err == config.ErrNoActiveConfiguration {
					return fmt.Errorf(`no active configuration, use 'registry config configurations' to manage`)
				}
				return fmt.Errorf("cannot read config: %v", err)
			}

			m, err := c.FlatMap()
			if err != nil {
				return fmt.Errorf("cannot decode config: %v", err)
			}
			for _, p := range c.Properties() {
				if m[p] != nil {
					if sv := fmt.Sprintf("%v", m[p]); sv != "" {
						cmd.Println(p, "=", sv)
					}
				}
			}

			cmd.Printf("\nYour active configuration is: %q.\n", target)
			return nil
		},
	}
	return cmd
}
