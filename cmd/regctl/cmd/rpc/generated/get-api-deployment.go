// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var GetApiDeploymentInput rpcpb.GetApiDeploymentRequest

var GetApiDeploymentFromFile string

func init() {
	RegistryServiceCmd.AddCommand(GetApiDeploymentCmd)

	GetApiDeploymentCmd.Flags().StringVar(&GetApiDeploymentInput.Name, "name", "", "Required. The name of the deployment to retrieve....")

	GetApiDeploymentCmd.Flags().StringVar(&GetApiDeploymentFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var GetApiDeploymentCmd = &cobra.Command{
	Use:   "get-api-deployment",
	Short: "GetApiDeployment returns a specified deployment.",
	Long:  "GetApiDeployment returns a specified deployment.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if GetApiDeploymentFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if GetApiDeploymentFromFile != "" {
			in, err = os.Open(GetApiDeploymentFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &GetApiDeploymentInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "GetApiDeployment", &GetApiDeploymentInput)
		}
		resp, err := RegistryClient.GetApiDeployment(ctx, &GetApiDeploymentInput)
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
