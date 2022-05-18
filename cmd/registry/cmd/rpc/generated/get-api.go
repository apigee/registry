// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var GetApiInput rpcpb.GetApiRequest

var GetApiFromFile string

func init() {
	RegistryServiceCmd.AddCommand(GetApiCmd)

	GetApiCmd.Flags().StringVar(&GetApiInput.Name, "name", "", "Required. The name of the API to retrieve. ...")

	GetApiCmd.Flags().StringVar(&GetApiFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var GetApiCmd = &cobra.Command{
	Use:   "get-api",
	Short: "GetApi returns a specified API.",
	Long:  "GetApi returns a specified API.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if GetApiFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if GetApiFromFile != "" {
			in, err = os.Open(GetApiFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &GetApiInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "GetApi", &GetApiInput)
		}
		resp, err := RegistryClient.GetApi(ctx, &GetApiInput)
		if err != nil {
			return err
		}

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(resp)

		return err
	},
}
