// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var UpdateProjectInput rpcpb.UpdateProjectRequest

var UpdateProjectFromFile string

func init() {
	AdminServiceCmd.AddCommand(UpdateProjectCmd)

	UpdateProjectInput.Project = new(rpcpb.Project)

	UpdateProjectInput.UpdateMask = new(fieldmaskpb.FieldMask)

	UpdateProjectCmd.Flags().StringVar(&UpdateProjectInput.Project.Name, "project.name", "", "Resource name.")

	UpdateProjectCmd.Flags().StringVar(&UpdateProjectInput.Project.DisplayName, "project.display_name", "", "Human-meaningful name.")

	UpdateProjectCmd.Flags().StringVar(&UpdateProjectInput.Project.Description, "project.description", "", "A detailed description.")

	UpdateProjectCmd.Flags().StringSliceVar(&UpdateProjectInput.UpdateMask.Paths, "update_mask.paths", []string{}, "The set of field mask paths.")

	UpdateProjectCmd.Flags().BoolVar(&UpdateProjectInput.AllowMissing, "allow_missing", false, "If set to true, and the project is not found, a...")

	UpdateProjectCmd.Flags().StringVar(&UpdateProjectFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var UpdateProjectCmd = &cobra.Command{
	Use:   "update-project",
	Short: "UpdateProject can be used to modify a specified...",
	Long:  "UpdateProject can be used to modify a specified project.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if UpdateProjectFromFile == "" {

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if UpdateProjectFromFile != "" {
			in, err = os.Open(UpdateProjectFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &UpdateProjectInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Admin", "UpdateProject", &UpdateProjectInput)
		}
		resp, err := AdminClient.UpdateProject(ctx, &UpdateProjectInput)
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
