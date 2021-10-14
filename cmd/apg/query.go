// Code generated. DO NOT EDIT.

package main

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"google.golang.org/api/iterator"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var QueryInput rpcpb.QueryRequest

var QueryFromFile string

func init() {
	SearchServiceCmd.AddCommand(QueryCmd)

	QueryCmd.Flags().StringVar(&QueryInput.Q, "q", "", "Required. Search query string")

	QueryCmd.Flags().Int32Var(&QueryInput.PageSize, "page_size", 10, "Default is 10. Page size")

	QueryCmd.Flags().StringVar(&QueryInput.PageToken, "page_token", "", "Page token")

	QueryCmd.Flags().StringVar(&QueryFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var QueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query the index.",
	Long:  "Query the index.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if QueryFromFile == "" {

			cmd.MarkFlagRequired("q")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if QueryFromFile != "" {
			in, err = os.Open(QueryFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &QueryInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Search", "Query", &QueryInput)
		}
		iter := SearchClient.Query(ctx, &QueryInput)

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
