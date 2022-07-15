// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	gapic "github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
)

var ProvisioningClient *gapic.ProvisioningClient
var ProvisioningSubCommands []string = []string{
	"create-instance",
	"poll-create-instance", "delete-instance",
	"poll-delete-instance", "get-instance",
}

func init() {
	rootCmd.AddCommand(ProvisioningServiceCmd)

}

var ProvisioningServiceCmd = &cobra.Command{
	Use:       "provisioning",
	Short:     "The service that is used for managing the data...",
	Long:      "The service that is used for managing the data plane provisioning of the  Registry.",
	ValidArgs: ProvisioningSubCommands,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		ProvisioningClient, err = connection.NewProvisioningClient(ctx)
		return
	},
}
