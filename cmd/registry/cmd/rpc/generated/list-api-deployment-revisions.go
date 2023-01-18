// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"google.golang.org/api/iterator"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var ListApiDeploymentRevisionsInput rpcpb.ListApiDeploymentRevisionsRequest

var ListApiDeploymentRevisionsFromFile string

func init() {
	RegistryServiceCmd.AddCommand(ListApiDeploymentRevisionsCmd)

	ListApiDeploymentRevisionsCmd.Flags().StringVar(&ListApiDeploymentRevisionsInput.Name, "name", "", "Required. The name of the deployment to list...")

	ListApiDeploymentRevisionsCmd.Flags().Int32Var(&ListApiDeploymentRevisionsInput.PageSize, "page_size", 10, "Default is 10. The maximum number of revisions to return per...")

	ListApiDeploymentRevisionsCmd.Flags().StringVar(&ListApiDeploymentRevisionsInput.PageToken, "page_token", "", "The page token, received from a previous...")

	ListApiDeploymentRevisionsCmd.Flags().StringVar(&ListApiDeploymentRevisionsInput.Filter, "filter", "", "An expression that can be used to filter the...")

	ListApiDeploymentRevisionsCmd.Flags().StringVar(&ListApiDeploymentRevisionsFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var ListApiDeploymentRevisionsCmd = &cobra.Command{
	Use:   "list-api-deployment-revisions",
	Short: "ListApiDeploymentRevisions lists all revisions of...",
	Long:  "ListApiDeploymentRevisions lists all revisions of a deployment.  Revisions are returned in descending order of revision creation time.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if ListApiDeploymentRevisionsFromFile == "" {

			cmd.MarkFlagRequired("name")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if ListApiDeploymentRevisionsFromFile != "" {
			in, err = os.Open(ListApiDeploymentRevisionsFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &ListApiDeploymentRevisionsInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "ListApiDeploymentRevisions", &ListApiDeploymentRevisionsInput)
		}
		iter := RegistryClient.ListApiDeploymentRevisions(ctx, &ListApiDeploymentRevisionsInput)

		// populate iterator with a page
		_, err = iter.Next()
		if err != nil && err != iterator.Done {
			return err
		}

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(iter.Response)

		return err
	},
}
