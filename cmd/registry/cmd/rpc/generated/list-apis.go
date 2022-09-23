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

var ListApisInput rpcpb.ListApisRequest

var ListApisFromFile string

func init() {
	RegistryServiceCmd.AddCommand(ListApisCmd)

	ListApisCmd.Flags().StringVar(&ListApisInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	ListApisCmd.Flags().Int32Var(&ListApisInput.PageSize, "page_size", 10, "Default is 10. The maximum number of APIs to return.  The...")

	ListApisCmd.Flags().StringVar(&ListApisInput.PageToken, "page_token", "", "A page token, received from a previous `ListApis`...")

	ListApisCmd.Flags().StringVar(&ListApisInput.Filter, "filter", "", "An expression that can be used to filter the...")

	ListApisCmd.Flags().StringVar(&ListApisInput.OrderBy, "order_by", "", "A comma-separated list of fields, e.g. 'foo,bar' ...")

	ListApisCmd.Flags().StringVar(&ListApisFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var ListApisCmd = &cobra.Command{
	Use:   "list-apis",
	Short: "ListApis returns matching APIs.",
	Long:  "ListApis returns matching APIs.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if ListApisFromFile == "" {

			cmd.MarkFlagRequired("parent")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if ListApisFromFile != "" {
			in, err = os.Open(ListApisFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &ListApisInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "ListApis", &ListApisInput)
		}
		iter := RegistryClient.ListApis(ctx, &ListApisInput)

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
