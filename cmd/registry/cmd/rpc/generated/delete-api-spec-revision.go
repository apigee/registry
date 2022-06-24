// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var DeleteApiSpecRevisionInput rpcpb.DeleteApiSpecRevisionRequest

var DeleteApiSpecRevisionFromFile string

func init() {
	RegistryServiceCmd.AddCommand(DeleteApiSpecRevisionCmd)

	DeleteApiSpecRevisionCmd.Flags().StringVar(&DeleteApiSpecRevisionInput.Name, "name", "", "Required. The name of the spec revision to be...")

	DeleteApiSpecRevisionCmd.Flags().StringVar(&DeleteApiSpecRevisionFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var DeleteApiSpecRevisionCmd = &cobra.Command{
	Use:   "delete-api-spec-revision",
	Short: "DeleteApiSpecRevision deletes a revision of a...",
	Long:  "DeleteApiSpecRevision deletes a revision of a spec.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if DeleteApiSpecRevisionFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if DeleteApiSpecRevisionFromFile != "" {
			in, err = os.Open(DeleteApiSpecRevisionFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &DeleteApiSpecRevisionInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "DeleteApiSpecRevision", &DeleteApiSpecRevisionInput)
		}
		resp, err := RegistryClient.DeleteApiSpecRevision(ctx, &DeleteApiSpecRevisionInput)
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
