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

var ListApiDeploymentsInput rpcpb.ListApiDeploymentsRequest

var ListApiDeploymentsFromFile string

func init() {
	RegistryServiceCmd.AddCommand(ListApiDeploymentsCmd)

	ListApiDeploymentsCmd.Flags().StringVar(&ListApiDeploymentsInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	ListApiDeploymentsCmd.Flags().Int32Var(&ListApiDeploymentsInput.PageSize, "page_size", 10, "Default is 10. The maximum number of deployments to return.  The...")

	ListApiDeploymentsCmd.Flags().StringVar(&ListApiDeploymentsInput.PageToken, "page_token", "", "A page token, received from a previous...")

	ListApiDeploymentsCmd.Flags().StringVar(&ListApiDeploymentsInput.Filter, "filter", "", "An expression that can be used to filter the...")

	ListApiDeploymentsCmd.Flags().StringVar(&ListApiDeploymentsInput.OrderBy, "order_by", "", "A comma-separated list of fields, e.g. 'foo,bar' ...")

	ListApiDeploymentsCmd.Flags().StringVar(&ListApiDeploymentsFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var ListApiDeploymentsCmd = &cobra.Command{
	Use:   "list-api-deployments",
	Short: "ListApiDeployments returns matching deployments.",
	Long:  "ListApiDeployments returns matching deployments.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if ListApiDeploymentsFromFile == "" {

			cmd.MarkFlagRequired("parent")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if ListApiDeploymentsFromFile != "" {
			in, err = os.Open(ListApiDeploymentsFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &ListApiDeploymentsInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "ListApiDeployments", &ListApiDeploymentsInput)
		}
		iter := RegistryClient.ListApiDeployments(ctx, &ListApiDeploymentsInput)

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
