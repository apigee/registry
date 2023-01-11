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
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication to the API Registry",
		Long: "Authentication to the registry is via tokens. These tokens can be provided " +
			"directly by using '--registry.token' in any command, or can be generated as needed " +
			"by setting the token-source property in a configuration. The token-source can be set " +
			"to any executable command that prints a registry token by using the 'config set token-source'. " +
			"Once set, the command will be executed before a registry command is run and the output used " +
			"as the token passed to the registry. The output can be verified by 'auth print-token'.",
		Example: "config set token-source 'gcloud auth print-access-token email@example.com'",
	}

	cmd.AddCommand(printTokenCommand())
	return cmd
}

func genToken(command string) (string, error) {
	cmdArgs := strings.Split(command, " ")
	execCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := execCmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
