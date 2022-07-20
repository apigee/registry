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

	"github.com/apigee/registry/log"
	"github.com/spf13/cobra"
)

func setCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set PROPERTY VALUE",
		Short: "Set the value of a property.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			logger := log.FromContext(ctx)

			target, config, err := targetConfig()
			if err != nil {
				logger.Fatalf("Cannot read config: %v", err)
			}

			m := map[string]interface{}{
				args[0]: args[1],
			}

			err = config.FromMap(m)
			if err != nil {
				logger.Fatalf("Cannot set value: %v", err)
			}

			err = config.Write(target)
			if err != nil {
				logger.Fatalf("Cannot write config: %v", err)
			}

			fmt.Printf("Updated property %q.\n", m[args[0]])
		},
	}
	return cmd
}
