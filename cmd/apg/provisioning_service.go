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

var ProvisioningConfig *viper.Viper
var ProvisioningClient *gapic.ProvisioningClient
var ProvisioningSubCommands []string = []string{
	"create-instance",
	"poll-create-instance", "delete-instance",
	"poll-delete-instance", "get-instance",
}

func init() {
	rootCmd.AddCommand(ProvisioningServiceCmd)

	ProvisioningConfig = viper.New()
	ProvisioningConfig.SetEnvPrefix("APG_REGISTRY")
	ProvisioningConfig.AutomaticEnv()

	ProvisioningServiceCmd.PersistentFlags().Bool("insecure", false, "Make insecure client connection. Or use APG_REGISTRY_INSECURE. Must be used with \"address\" option")
	ProvisioningConfig.BindPFlag("insecure", ProvisioningServiceCmd.PersistentFlags().Lookup("insecure"))
	ProvisioningConfig.BindEnv("insecure")

	ProvisioningServiceCmd.PersistentFlags().String("address", "", "Set API address used by client. Or use APG_REGISTRY_ADDRESS.")
	ProvisioningConfig.BindPFlag("address", ProvisioningServiceCmd.PersistentFlags().Lookup("address"))
	ProvisioningConfig.BindEnv("address")

	ProvisioningServiceCmd.PersistentFlags().String("token", "", "Set Bearer token used by the client. Or use APG_REGISTRY_TOKEN.")
	ProvisioningConfig.BindPFlag("token", ProvisioningServiceCmd.PersistentFlags().Lookup("token"))
	ProvisioningConfig.BindEnv("token")

	ProvisioningServiceCmd.PersistentFlags().String("api_key", "", "Set API Key used by the client. Or use APG_REGISTRY_API_KEY.")
	ProvisioningConfig.BindPFlag("api_key", ProvisioningServiceCmd.PersistentFlags().Lookup("api_key"))
	ProvisioningConfig.BindEnv("api_key")
}

var ProvisioningServiceCmd = &cobra.Command{
	Use:       "provisioning",
	Short:     "The service that is used for managing the data...",
	Long:      "The service that is used for managing the data plane provisioning of the  Registry.",
	ValidArgs: ProvisioningSubCommands,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		var opts []option.ClientOption

		address := ProvisioningConfig.GetString("address")
		if address != "" {
			opts = append(opts, option.WithEndpoint(address))
		}

		if ProvisioningConfig.GetBool("insecure") {
			if address == "" {
				return fmt.Errorf("Missing address to use with insecure connection")
			}

			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				return err
			}
			opts = append(opts, option.WithGRPCConn(conn))
		}

		if token := ProvisioningConfig.GetString("token"); token != "" {
			opts = append(opts, option.WithTokenSource(oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: token,
					TokenType:   "Bearer",
				})))
		}

		if key := ProvisioningConfig.GetString("api_key"); key != "" {
			opts = append(opts, option.WithAPIKey(key))
		}

		ProvisioningClient, err = gapic.NewProvisioningClient(ctx, opts...)
		return
	},
}
