// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	gapic "github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
)

var AdminClient *gapic.AdminClient
var AdminSubCommands []string = []string{
	"get-status",
	"get-storage",
	"migrate-database",
	"poll-migrate-database", "list-projects",
	"get-project",
	"create-project",
	"update-project",
	"delete-project",
}

func init() {
	rootCmd.AddCommand(AdminServiceCmd)

}

var AdminServiceCmd = &cobra.Command{
	Use:       "admin",
	Short:     "The Admin service supports setup and operation of...",
	Long:      "The Admin service supports setup and operation of an API registry.  It is typically not included in hosted versions of the API.",
	ValidArgs: AdminSubCommands,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		AdminClient, err = connection.NewAdminClient(ctx)
		return
	},
}
