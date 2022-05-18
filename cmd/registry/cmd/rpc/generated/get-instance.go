// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var GetInstanceInput rpcpb.GetInstanceRequest

var GetInstanceFromFile string

func init() {
	ProvisioningServiceCmd.AddCommand(GetInstanceCmd)

	GetInstanceCmd.Flags().StringVar(&GetInstanceInput.Name, "name", "", "Required. The name of the Instance to retrieve. ...")

	GetInstanceCmd.Flags().StringVar(&GetInstanceFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var GetInstanceCmd = &cobra.Command{
	Use:   "get-instance",
	Short: "Gets details of a single Instance.",
	Long:  "Gets details of a single Instance.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if GetInstanceFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if GetInstanceFromFile != "" {
			in, err = os.Open(GetInstanceFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &GetInstanceInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Provisioning", "GetInstance", &GetInstanceInput)
		}
		resp, err := ProvisioningClient.GetInstance(ctx, &GetInstanceInput)
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
