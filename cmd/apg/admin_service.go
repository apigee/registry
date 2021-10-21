// Code generated. DO NOT EDIT.

package main

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
	"list-projects",
	"get-project",
	"create-project",
	"update-project",
	"delete-project",
}

func init() {
	rootCmd.AddCommand(AdminServiceCmd)

	AdminConfig = viper.New()
	AdminConfig.SetEnvPrefix("APG_ADMIN")
	AdminConfig.AutomaticEnv()

	AdminServiceCmd.PersistentFlags().Bool("insecure", false, "Make insecure client connection. Or use APG_ADMIN_INSECURE. Must be used with \"address\" option")
	AdminConfig.BindPFlag("insecure", AdminServiceCmd.PersistentFlags().Lookup("insecure"))
	AdminConfig.BindEnv("insecure")

	AdminServiceCmd.PersistentFlags().String("address", "", "Set API address used by client. Or use APG_ADMIN_ADDRESS.")
	AdminConfig.BindPFlag("address", AdminServiceCmd.PersistentFlags().Lookup("address"))
	AdminConfig.BindEnv("address")

	AdminServiceCmd.PersistentFlags().String("token", "", "Set Bearer token used by the client. Or use APG_ADMIN_TOKEN.")
	AdminConfig.BindPFlag("token", AdminServiceCmd.PersistentFlags().Lookup("token"))
	AdminConfig.BindEnv("token")

	AdminServiceCmd.PersistentFlags().String("api_key", "", "Set API Key used by the client. Or use APG_ADMIN_API_KEY.")
	AdminConfig.BindPFlag("api_key", AdminServiceCmd.PersistentFlags().Lookup("api_key"))
	AdminConfig.BindEnv("api_key")
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
