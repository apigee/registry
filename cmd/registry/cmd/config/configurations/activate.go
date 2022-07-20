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

func activateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate CONFIGURATION_NAME",
		Short: "Activates an existing named configuration.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			logger := log.FromContext(ctx)

			name := args[0]
			ensureValidConfigurationName(name, logger)

			_, err := connection.ReadSettings(name)
			if err != nil {
				logger.Fatalf("Cannot activate configuration %q: %v", name, err)
			}

			err = connection.ActivateConfig(name)
			if err != nil {
				logger.Fatalf("Cannot activate configuration %q: %v", name, err)
			}

			fmt.Printf("Activated %q.\n", name)
		},
	}
	return cmd
}
