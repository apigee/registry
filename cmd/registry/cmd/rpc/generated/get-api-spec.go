// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var GetApiSpecInput rpcpb.GetApiSpecRequest

var GetApiSpecFromFile string

func init() {
	RegistryServiceCmd.AddCommand(GetApiSpecCmd)

	GetApiSpecCmd.Flags().StringVar(&GetApiSpecInput.Name, "name", "", "Required. The name of the spec to retrieve. ...")

	GetApiSpecCmd.Flags().StringVar(&GetApiSpecFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var GetApiSpecCmd = &cobra.Command{
	Use:   "get-api-spec",
	Short: "GetApiSpec returns a specified spec.",
	Long:  "GetApiSpec returns a specified spec.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if GetApiSpecFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if GetApiSpecFromFile != "" {
			in, err = os.Open(GetApiSpecFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &GetApiSpecInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "GetApiSpec", &GetApiSpecInput)
		}
		resp, err := RegistryClient.GetApiSpec(ctx, &GetApiSpecInput)
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
