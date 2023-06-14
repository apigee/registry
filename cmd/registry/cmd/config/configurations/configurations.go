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
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configurations",
		Short: "Maintain named configurations of properties",
		Long: `Maintain named configurations of properties.
		
Configurations are sets of properties used by a client when connecting to
the API Registry. These commands manipulate these named sets of properties
stored in the '$HOME/.config/registry folder.

When a configuration is 'active', its properties are automatically loaded 
and used, although they can be overridden by flags and environment variables.

See 'registry config' for commands to manipulate the properties within a
configuration.`,
	}

	cmd.AddCommand(activateCommand())
	cmd.AddCommand(createCommand())
	cmd.AddCommand(deleteCommand())
	cmd.AddCommand(describeCommand())
	cmd.AddCommand(listCommand())
	return cmd
}
