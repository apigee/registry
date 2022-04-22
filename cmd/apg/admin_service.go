// Code generated. DO NOT EDIT.

package apg

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	gapic "github.com/apigee/registry/gapic"
)

var AdminConfig *viper.Viper
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

	AdminConfig = viper.New()
	AdminConfig.SetEnvPrefix("APG_REGISTRY")
	AdminConfig.AutomaticEnv()
}

var AdminServiceCmd = &cobra.Command{
	Use:       "admin",
	Short:     "The Admin service supports setup and operation of...",
	Long:      "The Admin service supports setup and operation of an API registry.  It is typically not included in hosted versions of the API.",
	ValidArgs: AdminSubCommands,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		var opts []option.ClientOption

		address := AdminConfig.GetString("address")
		if address != "" {
			opts = append(opts, option.WithEndpoint(address))
		}

		if AdminConfig.GetBool("insecure") {
			if address == "" {
				return fmt.Errorf("Missing address to use with insecure connection")
			}

			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				return err
			}
			opts = append(opts, option.WithGRPCConn(conn))
		}

		if token := AdminConfig.GetString("token"); token != "" {
			opts = append(opts, option.WithTokenSource(oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: token,
					TokenType:   "Bearer",
				})))
		}

		if key := AdminConfig.GetString("api_key"); key != "" {
			opts = append(opts, option.WithAPIKey(key))
		}

		AdminClient, err = gapic.NewAdminClient(ctx, opts...)
		return
	},
}
