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

	"github.com/apigee/registry/pkg/config"
	"github.com/spf13/cobra"
)

func setCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set PROPERTY VALUE",
		Short: "Set a property value in the active configuration.",
		Long: "Set a property value in the active configuration. The following are valid properties:\n" +
			"  - registry.address\n" +
			"  - registry.insecure\n" +
			"  - registry.location\n" +
			"  - registry.project\n" +
			"  - token-source",
		Example: "registry config set registry.address localhost:8080\n" +
			"registry config set token-source 'gcloud auth print-access-token email@example.com'",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			target, c, err := config.ActiveRaw()
			if err != nil {
				if err == config.NoActiveConfigurationError {
					return fmt.Errorf(`No active configuration. Use 'registry config configurations' to manage.`)
				}
				return fmt.Errorf("Cannot read config: %v", err)
			}

			name, value := args[0], args[1]
			if !contains(c.Properties(), name) {
				return UnknownPropertyError{name}
			}

			if err = c.Set(name, value); err != nil {
				return fmt.Errorf("Cannot set value: %v", err)
			}

			if err = c.Write(target); err != nil {
				return fmt.Errorf("Cannot write config: %v", err)
			}

			cmd.Printf("Updated property %q.\n", name)
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
