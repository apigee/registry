// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var RollbackApiSpecInput rpcpb.RollbackApiSpecRequest

var RollbackApiSpecFromFile string

func init() {
	RegistryServiceCmd.AddCommand(RollbackApiSpecCmd)

	RollbackApiSpecCmd.Flags().StringVar(&RollbackApiSpecInput.Name, "name", "", "Required. The spec being rolled back.")

	RollbackApiSpecCmd.Flags().StringVar(&RollbackApiSpecInput.RevisionId, "revision_id", "", "Required. The revision ID to roll back to.  It...")

	RollbackApiSpecCmd.Flags().StringVar(&RollbackApiSpecFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var RollbackApiSpecCmd = &cobra.Command{
	Use:   "rollback-api-spec",
	Short: "RollbackApiSpec sets the current revision to a...",
	Long:  "RollbackApiSpec sets the current revision to a specified prior revision.  Note that this creates a new revision with a new revision ID.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if RollbackApiSpecFromFile == "" {

			cmd.MarkFlagRequired("name")

			cmd.MarkFlagRequired("revision_id")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if RollbackApiSpecFromFile != "" {
			in, err = os.Open(RollbackApiSpecFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &RollbackApiSpecInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "RollbackApiSpec", &RollbackApiSpecInput)
		}
		resp, err := RegistryClient.RollbackApiSpec(ctx, &RollbackApiSpecInput)
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
