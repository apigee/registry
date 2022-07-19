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
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configurations",
		Short: "Maintain configuration profiles",
	}

	// TODO
	// activate       Activates an existing named configuration.
	// create         Creates a new named configuration.
	// delete         Deletes a named configuration.
	// describe       Describes a named configuration by listing its properties.
	// list           Lists existing named configurations.

	// cmd.AddCommand(activateCommand())
	// cmd.AddCommand(createCommand())
	// cmd.AddCommand(deleteCommand())
	// cmd.AddCommand(describeCommand())
	cmd.AddCommand(listCommand())
	return cmd
}
