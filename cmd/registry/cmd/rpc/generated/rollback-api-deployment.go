// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var RollbackApiDeploymentInput rpcpb.RollbackApiDeploymentRequest

var RollbackApiDeploymentFromFile string

func init() {
	RegistryServiceCmd.AddCommand(RollbackApiDeploymentCmd)

	RollbackApiDeploymentCmd.Flags().StringVar(&RollbackApiDeploymentInput.Name, "name", "", "Required. The deployment being rolled back.")

	RollbackApiDeploymentCmd.Flags().StringVar(&RollbackApiDeploymentInput.RevisionId, "revision_id", "", "Required. The revision ID to roll back to.  It...")

	RollbackApiDeploymentCmd.Flags().StringVar(&RollbackApiDeploymentFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var RollbackApiDeploymentCmd = &cobra.Command{
	Use:   "rollback-api-deployment",
	Short: "RollbackApiDeployment sets the current revision...",
	Long:  "RollbackApiDeployment sets the current revision to a specified prior  revision. Note that this creates a new revision with a new revision ID.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if RollbackApiDeploymentFromFile == "" {

			cmd.MarkFlagRequired("name")

			cmd.MarkFlagRequired("revision_id")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if RollbackApiDeploymentFromFile != "" {
			in, err = os.Open(RollbackApiDeploymentFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &RollbackApiDeploymentInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "RollbackApiDeployment", &RollbackApiDeploymentInput)
		}
		resp, err := RegistryClient.RollbackApiDeployment(ctx, &RollbackApiDeploymentInput)
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
