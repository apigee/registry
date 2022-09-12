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

var ListProjectsInput rpcpb.ListProjectsRequest

var ListProjectsFromFile string

func init() {
	AdminServiceCmd.AddCommand(ListProjectsCmd)

	ListProjectsCmd.Flags().Int32Var(&ListProjectsInput.PageSize, "page_size", 10, "Default is 10. The maximum number of projects to return.  The...")

	ListProjectsCmd.Flags().StringVar(&ListProjectsInput.PageToken, "page_token", "", "A page token, received from a previous...")

	ListProjectsCmd.Flags().StringVar(&ListProjectsInput.Filter, "filter", "", "An expression that can be used to filter the...")

	ListProjectsCmd.Flags().StringVar(&ListProjectsInput.OrderBy, "order_by", "", "A comma-separated list of fields, e.g. 'foo,bar' ...")

	ListProjectsCmd.Flags().StringVar(&ListProjectsFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var ListProjectsCmd = &cobra.Command{
	Use:   "list-projects",
	Short: "ListProjects returns matching projects.  (--...",
	Long:  "ListProjects returns matching projects.  (-- api-linter: standard-methods=disabled --)  (-- api-linter: core::0132::method-signature=disabled     ...",
	PreRun: func(cmd *cobra.Command, args []string) {

		if ListProjectsFromFile == "" {

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if ListProjectsFromFile != "" {
			in, err = os.Open(ListProjectsFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &ListProjectsInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Admin", "ListProjects", &ListProjectsInput)
		}
		iter := AdminClient.ListProjects(ctx, &ListProjectsInput)

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
