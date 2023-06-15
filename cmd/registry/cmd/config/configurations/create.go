// Copyright 2022 Google LLC.
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
	"os/exec"
	"strings"

	"github.com/apigee/registry/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func createCommand() *cobra.Command {
	var gcloud bool

	cmd := &cobra.Command{
		Use:   "create CONFIGURATION",
		Short: "Creates and activates a new named configuration",
		Long: `Creates and activates a new named configuration. 
		
Values in the new configuration will default to the currently active
configuration (if any) but can be overridden by setting property flags
(unless using --gcloud).`,
		Example: `registry config configurations create localhost --registry.address='locahost:8080'
registry config configurations create hosted --gcloud`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := config.ValidateName(name); err != nil {
				return err
			}

			if _, err := config.Read(name); err == nil {
				return fmt.Errorf("cannot create config %q, it already exists", name)
			}

			s := config.Configuration{}
			var err error
			if gcloud {
				cmd.Flags().Visit(func(f *pflag.Flag) {
					if strings.HasPrefix(f.Name, "registry.") || f.Name == "address" {
						err = fmt.Errorf("--gcloud is mutually exclusive from --address and --registry.* flags")
					}
				})
				if err != nil {
					return err
				}

				project, err := activeGCloudProject()
				if err != nil {
					return fmt.Errorf("cannot execute `gcloud`: %v", err)
				}
				s.Registry.Project = project
				s.Registry.Address = "apigeeregistry.googleapis.com:443"
				s.Registry.Insecure = false
				s.TokenSource = "gcloud auth print-access-token"
			} else {
				// attempt to clone the active config or at least the current flags
				s, err = config.Active()
				if err != nil {
					if s, err = config.ReadValid(""); err != nil {
						s = config.Configuration{}
					}
				}
			}

			err = s.Write(name)
			if err != nil {
				return fmt.Errorf("cannot create config %q: %v", name, err)
			}
			cmd.Printf("Created %q.\n", name)

			err = config.Activate(name)
			if err != nil {
				return fmt.Errorf("cannot activate config %q: %v", name, err)
			}
			cmd.Printf("Activated %q.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&gcloud, "gcloud", false, "if specified, uses values from active gcloud config")
	return cmd
}

func activeGCloudProject() (string, error) {
	cmdArgs := strings.Split("gcloud config get project", " ")
	execCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := execCmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
