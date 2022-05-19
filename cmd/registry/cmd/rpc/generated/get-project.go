// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var GetProjectInput rpcpb.GetProjectRequest

var GetProjectFromFile string

func init() {
	AdminServiceCmd.AddCommand(GetProjectCmd)

	GetProjectCmd.Flags().StringVar(&GetProjectInput.Name, "name", "", "Required. The name of the project to retrieve.  Format:...")

	GetProjectCmd.Flags().StringVar(&GetProjectFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var GetProjectCmd = &cobra.Command{
	Use:   "get-project",
	Short: "GetProject returns a specified project.",
	Long:  "GetProject returns a specified project.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if GetProjectFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if GetProjectFromFile != "" {
			in, err = os.Open(GetProjectFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &GetProjectInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Admin", "GetProject", &GetProjectInput)
		}
		resp, err := AdminClient.GetProject(ctx, &GetProjectInput)
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
