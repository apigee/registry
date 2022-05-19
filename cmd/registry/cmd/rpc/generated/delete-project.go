// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var DeleteProjectInput rpcpb.DeleteProjectRequest

var DeleteProjectFromFile string

func init() {
	AdminServiceCmd.AddCommand(DeleteProjectCmd)

	DeleteProjectCmd.Flags().StringVar(&DeleteProjectInput.Name, "name", "", "Required. The name of the project to delete.  Format:...")

	DeleteProjectCmd.Flags().BoolVar(&DeleteProjectInput.Force, "force", false, "If set to true, any child resources will also be...")

	DeleteProjectCmd.Flags().StringVar(&DeleteProjectFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var DeleteProjectCmd = &cobra.Command{
	Use:   "delete-project",
	Short: "DeleteProject removes a specified project and all...",
	Long:  "DeleteProject removes a specified project and all of the resources that it  owns.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if DeleteProjectFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if DeleteProjectFromFile != "" {
			in, err = os.Open(DeleteProjectFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &DeleteProjectInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Admin", "DeleteProject", &DeleteProjectInput)
		}
		err = AdminClient.DeleteProject(ctx, &DeleteProjectInput)
		if err != nil {
			return err
		}

		return err
	},
}
