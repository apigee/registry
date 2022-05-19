// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var DeleteArtifactInput rpcpb.DeleteArtifactRequest

var DeleteArtifactFromFile string

func init() {
	RegistryServiceCmd.AddCommand(DeleteArtifactCmd)

	DeleteArtifactCmd.Flags().StringVar(&DeleteArtifactInput.Name, "name", "", "Required. The name of the artifact to delete. ...")

	DeleteArtifactCmd.Flags().StringVar(&DeleteArtifactFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var DeleteArtifactCmd = &cobra.Command{
	Use:   "delete-artifact",
	Short: "DeleteArtifact removes a specified artifact.",
	Long:  "DeleteArtifact removes a specified artifact.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if DeleteArtifactFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if DeleteArtifactFromFile != "" {
			in, err = os.Open(DeleteArtifactFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &DeleteArtifactInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "DeleteArtifact", &DeleteArtifactInput)
		}
		err = RegistryClient.DeleteArtifact(ctx, &DeleteArtifactInput)
		if err != nil {
			return err
		}

		return err
	},
}
