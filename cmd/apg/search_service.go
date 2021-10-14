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

var SearchConfig *viper.Viper
var SearchClient *gapic.SearchClient
var SearchSubCommands []string = []string{
	"index",
	"query",
}

func init() {
	rootCmd.AddCommand(SearchServiceCmd)

	SearchConfig = viper.New()
	SearchConfig.SetEnvPrefix("APG_SEARCH")
	SearchConfig.AutomaticEnv()

	SearchServiceCmd.PersistentFlags().Bool("insecure", false, "Make insecure client connection. Or use APG_SEARCH_INSECURE. Must be used with \"address\" option")
	SearchConfig.BindPFlag("insecure", SearchServiceCmd.PersistentFlags().Lookup("insecure"))
	SearchConfig.BindEnv("insecure")

	SearchServiceCmd.PersistentFlags().String("address", "", "Set API address used by client. Or use APG_SEARCH_ADDRESS.")
	SearchConfig.BindPFlag("address", SearchServiceCmd.PersistentFlags().Lookup("address"))
	SearchConfig.BindEnv("address")

	SearchServiceCmd.PersistentFlags().String("token", "", "Set Bearer token used by the client. Or use APG_SEARCH_TOKEN.")
	SearchConfig.BindPFlag("token", SearchServiceCmd.PersistentFlags().Lookup("token"))
	SearchConfig.BindEnv("token")

	SearchServiceCmd.PersistentFlags().String("api_key", "", "Set API Key used by the client. Or use APG_SEARCH_API_KEY.")
	SearchConfig.BindPFlag("api_key", SearchServiceCmd.PersistentFlags().Lookup("api_key"))
	SearchConfig.BindEnv("api_key")
}

var SearchServiceCmd = &cobra.Command{
	Use:       "search",
	Short:     "Build and search an index of APIs.",
	Long:      "Build and search an index of APIs.",
	ValidArgs: SearchSubCommands,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		var opts []option.ClientOption

		address := SearchConfig.GetString("address")
		if address != "" {
			opts = append(opts, option.WithEndpoint(address))
		}

		if SearchConfig.GetBool("insecure") {
			if address == "" {
				return fmt.Errorf("Missing address to use with insecure connection")
			}

			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				return err
			}
			opts = append(opts, option.WithGRPCConn(conn))
		}

		if token := SearchConfig.GetString("token"); token != "" {
			opts = append(opts, option.WithTokenSource(oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: token,
					TokenType:   "Bearer",
				})))
		}

		if key := SearchConfig.GetString("api_key"); key != "" {
			opts = append(opts, option.WithAPIKey(key))
		}

		SearchClient, err = gapic.NewSearchClient(ctx, opts...)
		return
	},
}
