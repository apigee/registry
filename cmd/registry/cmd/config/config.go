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

package config

import (
	"github.com/apigee/registry/cmd/registry/cmd/config/configurations"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Maintain properties in the active configuration",
		Long: `Maintain properties in the active configuration.
		
Configurations manage sets of properties used when a client connects to the 
API Registry. These commands display manipulate the property values in the 
active configuration.

The following are valid configuration properties:
	- address
	- insecure
	- location
	- project
	- token-source

See 'config configurations --help' for information on how to manage configurations.
See 'config set --help' for the list of properties available.`,
	}

	cmd.AddCommand(configurations.Command())
	cmd.AddCommand(getCommand())
	cmd.AddCommand(listCommand())
	cmd.AddCommand(setCommand())
	cmd.AddCommand(unsetCommand())
	return cmd
}
