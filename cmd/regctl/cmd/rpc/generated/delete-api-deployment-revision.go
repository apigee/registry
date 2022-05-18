// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var DeleteApiDeploymentRevisionInput rpcpb.DeleteApiDeploymentRevisionRequest

var DeleteApiDeploymentRevisionFromFile string

func init() {
	RegistryServiceCmd.AddCommand(DeleteApiDeploymentRevisionCmd)

	DeleteApiDeploymentRevisionCmd.Flags().StringVar(&DeleteApiDeploymentRevisionInput.Name, "name", "", "Required. The name of the deployment revision to...")

	DeleteApiDeploymentRevisionCmd.Flags().StringVar(&DeleteApiDeploymentRevisionFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var DeleteApiDeploymentRevisionCmd = &cobra.Command{
	Use:   "delete-api-deployment-revision",
	Short: "DeleteApiDeploymentRevision deletes a revision of...",
	Long:  "DeleteApiDeploymentRevision deletes a revision of a deployment.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if DeleteApiDeploymentRevisionFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if DeleteApiDeploymentRevisionFromFile != "" {
			in, err = os.Open(DeleteApiDeploymentRevisionFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &DeleteApiDeploymentRevisionInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "DeleteApiDeploymentRevision", &DeleteApiDeploymentRevisionInput)
		}
		resp, err := RegistryClient.DeleteApiDeploymentRevision(ctx, &DeleteApiDeploymentRevisionInput)
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
