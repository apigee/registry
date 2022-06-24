// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var DeleteApiInput rpcpb.DeleteApiRequest

var DeleteApiFromFile string

func init() {
	RegistryServiceCmd.AddCommand(DeleteApiCmd)

	DeleteApiCmd.Flags().StringVar(&DeleteApiInput.Name, "name", "", "Required. The name of the API to delete.  Format:...")

	DeleteApiCmd.Flags().BoolVar(&DeleteApiInput.Force, "force", false, "If set to true, any child resources will also be...")

	DeleteApiCmd.Flags().StringVar(&DeleteApiFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var DeleteApiCmd = &cobra.Command{
	Use:   "delete-api",
	Short: "DeleteApi removes a specified API and all of the...",
	Long:  "DeleteApi removes a specified API and all of the resources that it  owns.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if DeleteApiFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if DeleteApiFromFile != "" {
			in, err = os.Open(DeleteApiFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &DeleteApiInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "DeleteApi", &DeleteApiInput)
		}
		err = RegistryClient.DeleteApi(ctx, &DeleteApiInput)
		if err != nil {
			return err
		}

		return err
	},
}
