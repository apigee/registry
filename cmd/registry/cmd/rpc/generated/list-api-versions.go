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

var ListApiVersionsInput rpcpb.ListApiVersionsRequest

var ListApiVersionsFromFile string

func init() {
	RegistryServiceCmd.AddCommand(ListApiVersionsCmd)

	ListApiVersionsCmd.Flags().StringVar(&ListApiVersionsInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	ListApiVersionsCmd.Flags().Int32Var(&ListApiVersionsInput.PageSize, "page_size", 10, "Default is 10. The maximum number of versions to return.  The...")

	ListApiVersionsCmd.Flags().StringVar(&ListApiVersionsInput.PageToken, "page_token", "", "A page token, received from a previous...")

	ListApiVersionsCmd.Flags().StringVar(&ListApiVersionsInput.Filter, "filter", "", "An expression that can be used to filter the...")

	ListApiVersionsCmd.Flags().StringVar(&ListApiVersionsInput.OrderBy, "order_by", "", "A comma-separated list of fields, e.g. 'foo,bar' ...")

	ListApiVersionsCmd.Flags().StringVar(&ListApiVersionsFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var ListApiVersionsCmd = &cobra.Command{
	Use:   "list-api-versions",
	Short: "ListApiVersions returns matching versions.",
	Long:  "ListApiVersions returns matching versions.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if ListApiVersionsFromFile == "" {

			cmd.MarkFlagRequired("parent")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if ListApiVersionsFromFile != "" {
			in, err = os.Open(ListApiVersionsFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &ListApiVersionsInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "ListApiVersions", &ListApiVersionsInput)
		}
		iter := RegistryClient.ListApiVersions(ctx, &ListApiVersionsInput)

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
