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
	"strings"

	"github.com/apigee/registry/pkg/connection"
	"github.com/spf13/cobra"
)

func deleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete CONFIGURATION_NAME [CONFIGURATION_NAME ...]",
		Short: "Deletes a named configuration.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			for _, name := range args {
				if err := connection.ValidateConfigName(name); err != nil {
					return err
				}
			}

			cmd.Println("The following configs will be deleted:")
			for _, name := range args {
				cmd.Printf(" - %s\n", name)
			}
			cmd.Print("Do you want to continue (Y/n)? ")
			for {
				var yn string
				fmt.Fscanln(cmd.InOrStdin(), &yn)
				yn = strings.ToLower(yn)
				if yn == "" || yn == "y" || yn == "yes" {
					break
				} else if yn == "n" || yn == "no" {
					return fmt.Errorf("Aborted by user.")
				}
				cmd.Print("Please enter 'y' or 'n': ")
			}

			for _, name := range args {
				err := connection.DeleteConfig(name)
				if err != nil {
					return fmt.Errorf("Cannot delete config %q: %v.", name, err)
				}
				cmd.Printf("Deleted %q.\n", name)
			}
			return nil
		},
	}
	return cmd
}
