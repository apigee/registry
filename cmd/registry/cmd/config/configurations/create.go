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

	"github.com/apigee/registry/log"
	"github.com/apigee/registry/pkg/connection"
	"github.com/spf13/cobra"
)

func createCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create CONFIGURATION_NAME",
		Short: "Creates a new named configuration.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			logger := log.FromContext(ctx)

			name := args[0] // TODO: ensure simple name?
			ensureValidConfigurationName(name, logger)

			if _, err := connection.ReadConfig(name); err == nil {
				logger.Fatalf("Cannot create configuration %q, it already exists.", name)
			}

			s := connection.Config{}
			err := s.Write(name)
			if err != nil {
				logger.Fatalf("Cannot create configuration %q: %v", name, err)
			}

			err = connection.ActivateConfig(name)
			if err != nil {
				logger.Fatalf("Cannot set active configuration %q: %v", name, err)
			}

			fmt.Printf("Created %q.\n", name)
			fmt.Printf("Activated %q.\n", name)
		},
	}
	return cmd
}
