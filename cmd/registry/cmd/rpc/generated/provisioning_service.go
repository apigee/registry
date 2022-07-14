// Code generated. DO NOT EDIT.

package generated

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	gapic "github.com/apigee/registry/gapic"
	"github.com/apigee/registry/pkg/connection"
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

		ProvisioningClient, err = connection.NewProvisioningClient(ctx)
		return
	},
}
