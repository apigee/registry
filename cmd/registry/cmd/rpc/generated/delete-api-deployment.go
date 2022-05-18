// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var DeleteApiDeploymentInput rpcpb.DeleteApiDeploymentRequest

var DeleteApiDeploymentFromFile string

func init() {
	RegistryServiceCmd.AddCommand(DeleteApiDeploymentCmd)

	DeleteApiDeploymentCmd.Flags().StringVar(&DeleteApiDeploymentInput.Name, "name", "", "Required. The name of the deployment to delete. ...")

	DeleteApiDeploymentCmd.Flags().BoolVar(&DeleteApiDeploymentInput.Force, "force", false, "If set to true, any child resources will also be...")

	DeleteApiDeploymentCmd.Flags().StringVar(&DeleteApiDeploymentFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var DeleteApiDeploymentCmd = &cobra.Command{
	Use:   "delete-api-deployment",
	Short: "DeleteApiDeployment removes a specified...",
	Long:  "DeleteApiDeployment removes a specified deployment, all revisions, and all  child resources (e.g. artifacts).",
	PreRun: func(cmd *cobra.Command, args []string) {

		if DeleteApiDeploymentFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if DeleteApiDeploymentFromFile != "" {
			in, err = os.Open(DeleteApiDeploymentFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &DeleteApiDeploymentInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "DeleteApiDeployment", &DeleteApiDeploymentInput)
		}
		err = RegistryClient.DeleteApiDeployment(ctx, &DeleteApiDeploymentInput)
		if err != nil {
			return err
		}

		return err
	},
}
