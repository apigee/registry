// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exec

import (
	"context"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	osexec "os/exec"
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec COMMAND",
		Short: "execute any command",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			commandStr := args[0]
			if commandStr == "" {
				log.Fatal("Please provide a command to execute")
			}

			commandFields := strings.Fields(commandStr)
			command := commandFields[0]
			commandArgs := []string{}
			if len(commandFields) > 1 {
				commandArgs = commandFields[1:]
			}
			execCmd := osexec.Command(command, commandArgs...)
			execCmd.Stdout, execCmd.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
			err := execCmd.Run()

			if err != nil {
				log.WithError(err).Fatalf("Failed executing command: %q", command)
			}
		},
	}
	return cmd
}
