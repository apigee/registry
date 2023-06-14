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

func getCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get PROPERTY",
		Short:   "Print a property value in the active configuration",
		Example: `registry config get address`,
		Long: `Print a property value in the active configuration.

The following are valid configuration properties:
	- address
	- insecure
	- location
	- project
	- token-source`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, c, err := config.ActiveRaw()
			if err != nil {
				if err == config.ErrNoActiveConfiguration {
					return fmt.Errorf(`no active configuration, use 'registry config configurations' to manage`)
				}
				return fmt.Errorf("cannot read config: %v", err)
			}

			v, err := c.Get(args[0])
			if err != nil {
				return err
			}

			cmd.Println(v)
			return nil
		},
	}
	return cmd
}
