// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var CreateProjectInput rpcpb.CreateProjectRequest

var CreateProjectFromFile string

func init() {
	AdminServiceCmd.AddCommand(CreateProjectCmd)

	CreateProjectInput.Project = new(rpcpb.Project)

	CreateProjectCmd.Flags().StringVar(&CreateProjectInput.Project.Name, "project.name", "", "Resource name.")

	CreateProjectCmd.Flags().StringVar(&CreateProjectInput.Project.DisplayName, "project.display_name", "", "Human-meaningful name.")

	CreateProjectCmd.Flags().StringVar(&CreateProjectInput.Project.Description, "project.description", "", "A detailed description.")

	CreateProjectCmd.Flags().StringVar(&CreateProjectInput.ProjectId, "project_id", "", "Required. The ID to use for the project, which will become...")

	CreateProjectCmd.Flags().StringVar(&CreateProjectFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var CreateProjectCmd = &cobra.Command{
	Use:   "create-project",
	Short: "CreateProject creates a specified project.  (--...",
	Long:  "CreateProject creates a specified project.  (-- api-linter: standard-methods=disabled --)  (-- api-linter: core::0133::http-uri-parent=disabled     ...",
	PreRun: func(cmd *cobra.Command, args []string) {

		if CreateProjectFromFile == "" {

			cmd.MarkFlagRequired("project_id")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if CreateProjectFromFile != "" {
			in, err = os.Open(CreateProjectFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &CreateProjectInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Admin", "CreateProject", &CreateProjectInput)
		}
		resp, err := AdminClient.CreateProject(ctx, &CreateProjectInput)
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
