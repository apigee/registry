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

package auth

import (
	"fmt"

	"github.com/apigee/registry/pkg/config"
	"github.com/spf13/cobra"
)

func tokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Generate and print an auth token.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			_, c, err := config.ActiveRaw()
			if err != nil {
				return err
			}

			if c.TokenCmd == "" {
				return fmt.Errorf("No token generator found. Use `registry auth generator` to define.")
			}

			token, err := genToken(c.TokenCmd)
			if err != nil {
				return err
			}
			cmd.Printf(token)
			return nil
		},
	}
	return cmd
}
