// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	gapic "github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
)

var RegistryClient *gapic.RegistryClient
var RegistrySubCommands []string = []string{
	"list-apis",
	"get-api",
	"create-api",
	"update-api",
	"delete-api",
	"list-api-versions",
	"get-api-version",
	"create-api-version",
	"update-api-version",
	"delete-api-version",
	"list-api-specs",
	"get-api-spec",
	"get-api-spec-contents",
	"create-api-spec",
	"update-api-spec",
	"delete-api-spec",
	"tag-api-spec-revision",
	"list-api-spec-revisions",
	"rollback-api-spec",
	"delete-api-spec-revision",
	"list-api-deployments",
	"get-api-deployment",
	"create-api-deployment",
	"update-api-deployment",
	"delete-api-deployment",
	"tag-api-deployment-revision",
	"list-api-deployment-revisions",
	"rollback-api-deployment",
	"delete-api-deployment-revision",
	"list-artifacts",
	"get-artifact",
	"get-artifact-contents",
	"create-artifact",
	"replace-artifact",
	"delete-artifact",
}

func init() {
	rootCmd.AddCommand(RegistryServiceCmd)

}

var RegistryServiceCmd = &cobra.Command{
	Use:       "registry",
	Short:     "The Registry service allows teams to manage...",
	Long:      "The Registry service allows teams to manage descriptions of APIs.",
	ValidArgs: RegistrySubCommands,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		RegistryClient, err = connection.NewRegistryClient(ctx)
		return
	},
}
