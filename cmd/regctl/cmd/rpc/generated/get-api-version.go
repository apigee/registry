// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var GetApiVersionInput rpcpb.GetApiVersionRequest

var GetApiVersionFromFile string

func init() {
	RegistryServiceCmd.AddCommand(GetApiVersionCmd)

	GetApiVersionCmd.Flags().StringVar(&GetApiVersionInput.Name, "name", "", "Required. The name of the version to retrieve. ...")

	GetApiVersionCmd.Flags().StringVar(&GetApiVersionFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var GetApiVersionCmd = &cobra.Command{
	Use:   "get-api-version",
	Short: "GetApiVersion returns a specified version.",
	Long:  "GetApiVersion returns a specified version.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if GetApiVersionFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if GetApiVersionFromFile != "" {
			in, err = os.Open(GetApiVersionFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &GetApiVersionInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "GetApiVersion", &GetApiVersionInput)
		}
		resp, err := RegistryClient.GetApiVersion(ctx, &GetApiVersionInput)
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
