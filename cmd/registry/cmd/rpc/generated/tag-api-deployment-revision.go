// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var TagApiDeploymentRevisionInput rpcpb.TagApiDeploymentRevisionRequest

var TagApiDeploymentRevisionFromFile string

func init() {
	RegistryServiceCmd.AddCommand(TagApiDeploymentRevisionCmd)

	TagApiDeploymentRevisionCmd.Flags().StringVar(&TagApiDeploymentRevisionInput.Name, "name", "", "Required. The name of the deployment to be...")

	TagApiDeploymentRevisionCmd.Flags().StringVar(&TagApiDeploymentRevisionInput.Tag, "tag", "", "Required. The tag to apply.  The tag should be at...")

	TagApiDeploymentRevisionCmd.Flags().StringVar(&TagApiDeploymentRevisionFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var TagApiDeploymentRevisionCmd = &cobra.Command{
	Use:   "tag-api-deployment-revision",
	Short: "TagApiDeploymentRevision adds a tag to a...",
	Long:  "TagApiDeploymentRevision adds a tag to a specified revision of a  deployment.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if TagApiDeploymentRevisionFromFile == "" {

			cmd.MarkFlagRequired("name")

			cmd.MarkFlagRequired("tag")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if TagApiDeploymentRevisionFromFile != "" {
			in, err = os.Open(TagApiDeploymentRevisionFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &TagApiDeploymentRevisionInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "TagApiDeploymentRevision", &TagApiDeploymentRevisionInput)
		}
		resp, err := RegistryClient.TagApiDeploymentRevision(ctx, &TagApiDeploymentRevisionInput)
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
