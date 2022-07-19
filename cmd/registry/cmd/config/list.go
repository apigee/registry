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
	"github.com/apigee/registry/pkg/connection"
	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List properties for the currently active configuration.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			logger := log.FromContext(ctx)

			settings, err := activeSettings()
			if err != nil {
				logger.Fatalf("reading settings: %v", err)
			}

			m, err := settings.AsMap()
			if err != nil {
				logger.Fatalf("decoding settings: %v", err)
			}
			for k, v := range m {
				if sv := fmt.Sprintf("%v", v); sv != "" {
					fmt.Println(k, "=", sv)
				}
			}
		},
	}
	return cmd
}

func activeSettings() (connection.Settings, error) {
	var err error
	name, _ := connection.Flags.GetString("config")
	if name == "" {
		name, err = connection.ActiveConfigName()
		if err != nil {
			return connection.Settings{}, err
		}
	}

	return connection.ReadSettings(name)
}