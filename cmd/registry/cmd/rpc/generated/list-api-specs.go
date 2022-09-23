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

var ListApiSpecsInput rpcpb.ListApiSpecsRequest

var ListApiSpecsFromFile string

func init() {
	RegistryServiceCmd.AddCommand(ListApiSpecsCmd)

	ListApiSpecsCmd.Flags().StringVar(&ListApiSpecsInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	ListApiSpecsCmd.Flags().Int32Var(&ListApiSpecsInput.PageSize, "page_size", 10, "Default is 10. The maximum number of specs to return.  The...")

	ListApiSpecsCmd.Flags().StringVar(&ListApiSpecsInput.PageToken, "page_token", "", "A page token, received from a previous...")

	ListApiSpecsCmd.Flags().StringVar(&ListApiSpecsInput.Filter, "filter", "", "An expression that can be used to filter the...")

	ListApiSpecsCmd.Flags().StringVar(&ListApiSpecsInput.OrderBy, "order_by", "", "A comma-separated list of fields, e.g. 'foo,bar' ...")

	ListApiSpecsCmd.Flags().StringVar(&ListApiSpecsFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var ListApiSpecsCmd = &cobra.Command{
	Use:   "list-api-specs",
	Short: "ListApiSpecs returns matching specs.",
	Long:  "ListApiSpecs returns matching specs.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if ListApiSpecsFromFile == "" {

			cmd.MarkFlagRequired("parent")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if ListApiSpecsFromFile != "" {
			in, err = os.Open(ListApiSpecsFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &ListApiSpecsInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "ListApiSpecs", &ListApiSpecsInput)
		}
		iter := RegistryClient.ListApiSpecs(ctx, &ListApiSpecsInput)

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
