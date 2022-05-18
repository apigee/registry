// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var DeleteApiVersionInput rpcpb.DeleteApiVersionRequest

var DeleteApiVersionFromFile string

func init() {
	RegistryServiceCmd.AddCommand(DeleteApiVersionCmd)

	DeleteApiVersionCmd.Flags().StringVar(&DeleteApiVersionInput.Name, "name", "", "Required. The name of the version to delete. ...")

	DeleteApiVersionCmd.Flags().BoolVar(&DeleteApiVersionInput.Force, "force", false, "If set to true, any child resources will also be...")

	DeleteApiVersionCmd.Flags().StringVar(&DeleteApiVersionFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var DeleteApiVersionCmd = &cobra.Command{
	Use:   "delete-api-version",
	Short: "DeleteApiVersion removes a specified version and...",
	Long:  "DeleteApiVersion removes a specified version and all of the resources that  it owns.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if DeleteApiVersionFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if DeleteApiVersionFromFile != "" {
			in, err = os.Open(DeleteApiVersionFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &DeleteApiVersionInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "DeleteApiVersion", &DeleteApiVersionInput)
		}
		err = RegistryClient.DeleteApiVersion(ctx, &DeleteApiVersionInput)
		if err != nil {
			return err
		}

		return err
	},
}
