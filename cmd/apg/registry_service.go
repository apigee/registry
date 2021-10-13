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

var RegistryConfig *viper.Viper
var RegistryClient *gapic.RegistryClient
var RegistrySubCommands []string = []string{
	"get-status",
	"list-projects",
	"get-project",
	"create-project",
	"update-project",
	"delete-project",
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
	"list-artifacts",
	"get-artifact",
	"get-artifact-contents",
	"create-artifact",
	"replace-artifact",
	"delete-artifact",
}

func init() {
	rootCmd.AddCommand(RegistryServiceCmd)

	RegistryConfig = viper.New()
	RegistryConfig.SetEnvPrefix("APG_REGISTRY")
	RegistryConfig.AutomaticEnv()

	RegistryServiceCmd.PersistentFlags().Bool("insecure", false, "Make insecure client connection. Or use APG_REGISTRY_INSECURE. Must be used with \"address\" option")
	RegistryConfig.BindPFlag("insecure", RegistryServiceCmd.PersistentFlags().Lookup("insecure"))
	RegistryConfig.BindEnv("insecure")

	RegistryServiceCmd.PersistentFlags().String("address", "", "Set API address used by client. Or use APG_REGISTRY_ADDRESS.")
	RegistryConfig.BindPFlag("address", RegistryServiceCmd.PersistentFlags().Lookup("address"))
	RegistryConfig.BindEnv("address")

	RegistryServiceCmd.PersistentFlags().String("token", "", "Set Bearer token used by the client. Or use APG_REGISTRY_TOKEN.")
	RegistryConfig.BindPFlag("token", RegistryServiceCmd.PersistentFlags().Lookup("token"))
	RegistryConfig.BindEnv("token")

	RegistryServiceCmd.PersistentFlags().String("api_key", "", "Set API Key used by the client. Or use APG_REGISTRY_API_KEY.")
	RegistryConfig.BindPFlag("api_key", RegistryServiceCmd.PersistentFlags().Lookup("api_key"))
	RegistryConfig.BindEnv("api_key")
}

var RegistryServiceCmd = &cobra.Command{
	Use:       "registry",
	Short:     "The Registry service allows teams to manage...",
	Long:      "The Registry service allows teams to manage descriptions of APIs.",
	ValidArgs: RegistrySubCommands,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		var opts []option.ClientOption

		address := RegistryConfig.GetString("address")
		if address != "" {
			opts = append(opts, option.WithEndpoint(address))
		}

		if RegistryConfig.GetBool("insecure") {
			if address == "" {
				return fmt.Errorf("Missing address to use with insecure connection")
			}

			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				return err
			}
			opts = append(opts, option.WithGRPCConn(conn))
		}

		if token := RegistryConfig.GetString("token"); token != "" {
			opts = append(opts, option.WithTokenSource(oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: token,
					TokenType:   "Bearer",
				})))
		}

		if key := RegistryConfig.GetString("api_key"); key != "" {
			opts = append(opts, option.WithAPIKey(key))
		}

		RegistryClient, err = gapic.NewRegistryClient(ctx, opts...)
		return
	},
}
