// Code generated. DO NOT EDIT.

package main

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var IndexInput rpcpb.IndexRequest

var IndexFromFile string

func init() {
	SearchServiceCmd.AddCommand(IndexCmd)

	IndexCmd.Flags().StringVar(&IndexInput.ResourceName, "resource_name", "", "")

	IndexCmd.Flags().StringVar(&IndexFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var IndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Add a resource to the search index.",
	Long:  "Add a resource to the search index.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if IndexFromFile == "" {

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if IndexFromFile != "" {
			in, err = os.Open(IndexFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &IndexInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Search", "Index", &IndexInput)
		}
		resp, err := SearchClient.Index(ctx, &IndexInput)

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(resp)

		return err
	},
}
