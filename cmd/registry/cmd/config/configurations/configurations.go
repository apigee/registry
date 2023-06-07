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
		Long: "Configurations manage sets of properties used when connecting to the registry. " +
			"These commands manipulate named sets of these properties, called 'configurations'. " +
			"When a configuration is 'active', it's properties are automatically loaded and used " +
			"by registry commands. See `registry config` for manipulating the properties themselves. " +
			"Configuration files are stored in the $HOME/.config/registry folder.",
	}

	cmd.AddCommand(activateCommand())
	cmd.AddCommand(createCommand())
	cmd.AddCommand(deleteCommand())
	cmd.AddCommand(describeCommand())
	cmd.AddCommand(listCommand())
	return cmd
}
