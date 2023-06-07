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

func unsetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset PROPERTY",
		Short: "Clear the property value from the active configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, c, err := config.ActiveRaw()
			if err != nil {
				if err == config.ErrNoActiveConfiguration {
					return fmt.Errorf(`no active configuration, use 'registry config configurations' to manage`)
				}
				return fmt.Errorf("cannot read config: %v", err)
			}

			if err = c.Unset(args[0]); err != nil {
				return fmt.Errorf("cannot unset property: %v", err)
			}

			if err = c.Write(target); err != nil {
				return fmt.Errorf("cannot write config: %v", err)
			}

			cmd.Printf("Unset property %q.\n", args[0])
			return nil
		},
	}
	return cmd
}
