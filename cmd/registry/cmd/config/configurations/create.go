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

	"github.com/apigee/registry/pkg/config"
	"github.com/spf13/cobra"
)

func createCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create CONFIGURATION_NAME",
		Short: "Creates and activates a new named configuration.",
		Long: "Creates and activates a new named configuration. Values will be populated from active " +
			"configuration (if any) and any passed property flags.",
		Example: "registry config configurations create localhost --registry.address='locahost:8080'",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			name := args[0]
			if err := config.ValidateName(name); err != nil {
				return err
			}

			if _, err := config.Read(name); err == nil {
				return fmt.Errorf("Cannot create config %q, it already exists.", name)
			}

			// attempt to clone the active config or at least the current flags
			s, err := config.Active()
			if err != nil {
				if s, err = config.ReadValid(""); err != nil {
					s = config.Configuration{}
				}
			}

			err = s.Write(name)
			if err != nil {
				return fmt.Errorf("Cannot create config %q: %v", name, err)
			}
			cmd.Printf("Created %q.\n", name)

			err = config.Activate(name)
			if err != nil {
				return fmt.Errorf("Cannot activate config %q: %v", name, err)
			}
			cmd.Printf("Activated %q.\n", name)
			return nil
		},
	}
	return cmd
}
