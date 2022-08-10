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
	"github.com/apigee/registry/pkg/config"
	"github.com/spf13/cobra"
)

func execCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec COMMAND",
		Short: "Set the shell command that generates an auth token.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			name, c, err := config.ActiveRaw()
			if err != nil {
				return err
			}

			token, err := genToken(args[0])
			if err != nil {
				return err
			}

			c.TokenCmd = args[0]
			if err := c.Write(name); err != nil {
				return err
			}

			cmd.Println("Updated auth exec command.")
			cmd.Printf("Command output: %q\n", token)
			return nil
		},
	}
	return cmd
}
