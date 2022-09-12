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

var ListArtifactsInput rpcpb.ListArtifactsRequest

var ListArtifactsFromFile string

func init() {
	RegistryServiceCmd.AddCommand(ListArtifactsCmd)

	ListArtifactsCmd.Flags().StringVar(&ListArtifactsInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	ListArtifactsCmd.Flags().Int32Var(&ListArtifactsInput.PageSize, "page_size", 10, "Default is 10. The maximum number of artifacts to return.  The...")

	ListArtifactsCmd.Flags().StringVar(&ListArtifactsInput.PageToken, "page_token", "", "A page token, received from a previous...")

	ListArtifactsCmd.Flags().StringVar(&ListArtifactsInput.Filter, "filter", "", "An expression that can be used to filter the...")

	ListArtifactsCmd.Flags().StringVar(&ListArtifactsInput.OrderBy, "order_by", "", "A comma-separated list of fields, e.g. 'foo,bar' ...")

	ListArtifactsCmd.Flags().StringVar(&ListArtifactsFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var ListArtifactsCmd = &cobra.Command{
	Use:   "list-artifacts",
	Short: "ListArtifacts returns matching artifacts.",
	Long:  "ListArtifacts returns matching artifacts.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if ListArtifactsFromFile == "" {

			cmd.MarkFlagRequired("parent")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if ListArtifactsFromFile != "" {
			in, err = os.Open(ListArtifactsFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &ListArtifactsInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "ListArtifacts", &ListArtifactsInput)
		}
		iter := RegistryClient.ListArtifacts(ctx, &ListArtifactsInput)

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
