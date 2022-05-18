// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var DeleteApiSpecInput rpcpb.DeleteApiSpecRequest

var DeleteApiSpecFromFile string

func init() {
	RegistryServiceCmd.AddCommand(DeleteApiSpecCmd)

	DeleteApiSpecCmd.Flags().StringVar(&DeleteApiSpecInput.Name, "name", "", "Required. The name of the spec to delete. ...")

	DeleteApiSpecCmd.Flags().BoolVar(&DeleteApiSpecInput.Force, "force", false, "If set to true, any child resources will also be...")

	DeleteApiSpecCmd.Flags().StringVar(&DeleteApiSpecFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var DeleteApiSpecCmd = &cobra.Command{
	Use:   "delete-api-spec",
	Short: "DeleteApiSpec removes a specified spec, all...",
	Long:  "DeleteApiSpec removes a specified spec, all revisions, and all child  resources (e.g. artifacts).",
	PreRun: func(cmd *cobra.Command, args []string) {

		if DeleteApiSpecFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if DeleteApiSpecFromFile != "" {
			in, err = os.Open(DeleteApiSpecFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &DeleteApiSpecInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "DeleteApiSpec", &DeleteApiSpecInput)
		}
		err = RegistryClient.DeleteApiSpec(ctx, &DeleteApiSpecInput)
		if err != nil {
			return err
		}

		return err
	},
}
