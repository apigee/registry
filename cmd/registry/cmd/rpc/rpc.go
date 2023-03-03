// Copyright 2022 Google LLC
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

package rpc

import (
	"github.com/apigee/registry/cmd/registry/cmd/rpc/generated"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := generated.RegistryServiceCmd
	cmd.Use = "rpc"
	cmd.Short = "Make direct calls to RPC methods"
	cmd.Long = cmd.Short
	cmd.PersistentFlags().BoolVarP(&generated.Verbose, "verbose", "v", false, "Print verbose output")
	cmd.PersistentFlags().BoolVarP(&generated.OutputJSON, "json", "j", false, "Print JSON output")

	generated.AdminServiceCmd.Use = "admin"
	generated.AdminServiceCmd.Short = "Make direct calls to Admin RPC methods (self-hosted installations only)"
	generated.AdminServiceCmd.Long = generated.AdminServiceCmd.Short
	cmd.AddCommand(generated.AdminServiceCmd)

	generated.ProvisioningServiceCmd.Use = "provisioning"
	generated.ProvisioningServiceCmd.Short = "Make direct calls to Provisioning RPC methods (Google-hosted installations only)"
	generated.ProvisioningServiceCmd.Long = generated.ProvisioningServiceCmd.Short
	cmd.AddCommand(generated.ProvisioningServiceCmd)

	return cmd
}
